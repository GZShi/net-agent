package main

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/GZShi/net-agent/bin/common"
	"github.com/GZShi/net-agent/bin/ws"
	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/mixlistener"
	"github.com/GZShi/net-agent/rpc/cluster"
	clssvc "github.com/GZShi/net-agent/rpc/cluster/service"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "./config.json", "filepath of config json file")
	flag.Parse()

	var cfg common.Config
	err := utils.LoadJSONFile(configPath, &cfg)
	if err != nil {
		log.Get().WithError(err).WithField("path", configPath).Error("load config file failed")
		return
	}

	// run tunnel server
	mixl := mixlistener.Listen("tcp4", cfg.Tunnel.Address)
	mixl.RegisterBuiltIn(
		mixlistener.HTTPName,
		mixlistener.TunnelName,
		mixlistener.Socks5Name,
	)

	var wg sync.WaitGroup
	go runTunnelServer(mixl, &cfg, &wg)
	go runWsUpgraderServer(mixl, &cfg, &wg)
	go runSocks5Server(mixl, &cfg, &wg)

	log.Get().Info("server running on ", cfg.Tunnel.Address)

	err = mixl.Run()
	if err != nil {
		log.Get().WithError(err).WithField("addr", cfg.Tunnel.Address).Error("listener stopped")
	} else {
		log.Get().WithField("addr", cfg.Tunnel.Address).Warn("listener stopped")
	}

	log.Get().Info("wait server all close")
	wg.Wait()
}

func serve(conn net.Conn, cfg *common.Config) {
	defer conn.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Get().Warn("serve recover: ", r)
		}
	}()

	cc, err := cipherconn.New(conn, cfg.Tunnel.Password)
	if err != nil {
		log.Get().WithError(err).Error("new cipherconn failed")
		return
	}

	t := tunnel.New(cc)

	t.BindServices(cluster.NewService())

	log.Get().Info("tunnel connected")
	t.Run()
	log.Get().Info("tunnel stopped")
}

func runTunnelServer(mixl mixlistener.MixListener, cfg *common.Config, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	l, err := mixl.GetListener(mixlistener.TunnelName)
	if err != nil {
		log.Get().WithError(err).Error("get listener failed")
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept conn failed")
			return
		}

		go serve(conn, cfg)
	}
}

func runWsUpgraderServer(mixl mixlistener.MixListener, cfg *common.Config, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	l, err := mixl.GetListener(mixlistener.HTTPName)
	if err != nil {
		log.Get().WithError(err).Error("get listener failed")
		return
	}

	if !cfg.Websocket.Enable {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}

	http.HandleFunc(cfg.Websocket.Path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.Upgrade(w, r)
		if err != nil {
			w.Write([]byte("upgrade failed"))
			return
		}

		go serve(conn, cfg)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello,world"))
	})

	log.Get().Info("websocket server enabled")
	http.Serve(l, nil)
}

func runSocks5Server(mixl mixlistener.MixListener, cfg *common.Config, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	l, err := mixl.GetListener(mixlistener.Socks5Name)
	if err != nil {
		log.Get().WithError(err).Error("get listener failed")
		return
	}

	cls := clssvc.New(nil)
	s := socks5.NewServer()

	s.SetAuthChecker(socks5.PswdAuthChecker(func(username, password string, ctx map[string]string) (err error) {
		defer func() {
			if err != nil {
				log.Get().WithError(err).Error("auth failed")
			}
		}()

		if ctx == nil {
			return errors.New("auth ctx is nil")
		}

		parts := strings.Split(username, "@")
		if len(parts) != 2 {
			return errors.New("parse username failed")
		}

		ctx["username"] = parts[0]
		ctx["password"] = password
		ctx["proxy"] = parts[1]

		if parts[0] == "" || password == "" {
			return errors.New("invalid username and password")
		}

		return nil
	}))

	s.SetRequster(func(req socks5.Request, ctx map[string]string) (conn net.Conn, err error) {
		defer func() {
			if err != nil {
				log.Get().WithError(err).Error("auth failed")
			}
		}()

		if ctx == nil {
			return nil, errors.New("request ctx is nil")
		}

		if cls == nil {
			return nil, errors.New("cluster is nil")
		}

		vhost, port, err := net.SplitHostPort(ctx["proxy"])
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(vhost, ".tunnel") {
			return nil, errors.New("host not support")
		}
		vport, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}

		conn, err = cls.Dial(vhost, uint32(vport))
		if err != nil {
			return nil, err
		}

		auth := socks5.AuthPswd(ctx["username"], ctx["password"])
		return socks5.Upgrade(conn, req.GetAddrPortStr(), auth)
	})

	s.Run(l)
}
