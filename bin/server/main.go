package main

import (
	"net"
	"strings"

	"github.com/GZShi/net-agent/bin/common"
	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
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
	l, err := net.Listen("tcp4", cfg.Tunnel.Address)
	if err != nil {
		log.Get().WithError(err).WithField("addr", cfg.Tunnel.Address).Error("listen address failed")
		return
	}

	log.Get().Info("server running on ", cfg.Tunnel.Address)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept conn failed")
			return
		}

		go serve(conn, &cfg)
	}
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

func runService(t tunnel.Tunnel, svc common.ServiceInfo) {
	if !svc.Enable {
		return
	}

	log.Get().Info(svc.Desc)

	switch svc.Type {
	case "socks5":
	case "portproxy":
	default:
		log.Get().Error("unknown service type: ", svc.Type)
		return
	}
}

func listen(t tunnel.Tunnel, address string) (net.Listener, error) {
	if strings.Contains(address, ".tunnel:") && t != nil {
		return t.Listen(1080)
	}
	return net.Listen("tcp4", address)
}

func openSocks5Service(t tunnel.Tunnel, param interface{}) {

}
