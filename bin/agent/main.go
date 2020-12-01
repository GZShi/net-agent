package main

import (
	"net"

	"github.com/GZShi/net-agent/bin/config"
	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

func main() {
	var configPath = "./configs.json"
	var cfg config.Config
	err := utils.LoadJSONFile(configPath, &cfg)
	if err != nil {
		log.Get().WithError(err).WithField("path", configPath).Error("load config file failed")
		return
	}

	conn, err := net.Dial("tcp4", cfg.Tunnel.Address)
	if err != nil {
		log.Get().WithError(err).Error("connect to tunnel failed: ", cfg.Tunnel.Address)
		return
	}

	cc, err := cipherconn.New(conn, cfg.Tunnel.Password)
	if err != nil {
		log.Get().WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)
	t.Ready(func(t tunnel.Tunnel) {
		log.Get().Info("tunnel created: ", cfg.Tunnel.Address)

		cls := cluster.NewClient(t, nil)
		tid, err := cls.Login(cfg.Tunnel.VHost)
		if err != nil {
			log.Get().WithError(err).Error("join cluster failed")
			return
		}
		log.Get().Info("join cluster success, tid=", tid)
	})

	t.Run()
}
