package main

import (
	"net"
<<<<<<< HEAD
=======
	"net/http"
	"sync"
>>>>>>> b382036b062b797046b3581c117c0853381f930a

	"github.com/GZShi/net-agent/bin/common"
	"github.com/GZShi/net-agent/bin/ws"
	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/mixlistener"
	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

func main() {
	var configPath = "./configs.json"
	var cfg common.Config
	err := utils.LoadJSONFile(configPath, &cfg)
	if err != nil {
		log.Get().WithError(err).WithField("path", configPath).Error("load config file failed")
		return
	}

	// run tunnel server
	mixl := mixlistener.Listen("tcp4", cfg.Tunnel.Address)
	mixl.RegisterBuiltIn(mixlistener.HTTPName, mixlistener.TunnelName)

	var wg sync.WaitGroup
	go runTunnelServer(mixl, &cfg, &wg)
	go runWsUpgraderServer(mixl, &cfg, &wg)

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
<<<<<<< HEAD
=======

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

	http.HandleFunc("/tunnel", func(w http.ResponseWriter, r *http.Request) {
		conn, err := ws.Upgrade(w, r)
		if err != nil {
			w.Write([]byte("upgrade failed"))
			return
		}

		go serve(conn, cfg)
	})

	log.Get().Info("websocket server enabled")
	http.Serve(l, nil)
}
>>>>>>> b382036b062b797046b3581c117c0853381f930a
