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
	cls := cluster.NewClient(t, nil)

	t.Ready(func(t tunnel.Tunnel) {
		log.Get().Info("tunnel created: ", cfg.Tunnel.Address)

		tid, vhost, err := cls.Login(cfg.Tunnel.VHost)
		if err != nil {
			log.Get().WithError(err).Error("join cluster failed")
			return
		}
		log.Get().Info("join cluster success, tid=", tid, " vhost=", vhost)
		go cls.Heartbeat()

		// run service
		if cfg.Services != nil {
			for index, svc := range cfg.Services {
				common.RunService(t, cls, index, svc)
			}
		}
	})

	go func() {
		log.Get().Info("press ctrl+c to stop tunnel")
		utils.WaitCtrlC()
		cls.Logout()
		t.Stop()
	}()

	t.Run()
}
