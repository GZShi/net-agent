package main

import (
	"github.com/GZShi/net-agent/bin/common"
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

	t, err := common.ConnectTunnel(&cfg.Tunnel)
	if err != nil {
		log.Get().WithError(err).Error("connect tunnel failed")
		return
	}

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
			for _, svc := range cfg.Services {
				log.Get().Debug("start running service")
				go common.RunService(t, cls, svc)
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
