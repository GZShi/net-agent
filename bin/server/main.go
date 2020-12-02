package main

import (
	"net"

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
