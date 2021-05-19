package main

import (
	"flag"
	"io"
	"time"

	"github.com/GZShi/net-agent/bin/common"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster"
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

	waitSeconds := 3 * time.Second
	for {
		start := time.Now()
		connectAndWork(&cfg)

		// 运行时间小于2分钟，需要进入等待冷却
		if time.Now().Sub(start) < time.Minute*2 {
			waitSeconds = waitSeconds + time.Second*7
			if waitSeconds > time.Second*60 {
				waitSeconds = time.Second * 40
			}
		} else {
			waitSeconds = 3 * time.Second
		}

		log.Get().Warn("reconnect after ", waitSeconds)
		<-time.After(waitSeconds)
	}

}

func connectAndWork(cfg *common.Config) {

	t, err := common.ConnectTunnel(&cfg.Tunnel)
	if err != nil {
		log.Get().WithError(err).Error("connect tunnel failed")
		return
	}

	cls := cluster.NewClient(t, nil)
	svcClosers := []io.Closer{}

	t.Ready(func(t tunnel.Tunnel) {
		log.Get().Info("tunnel created: ", cfg.Tunnel.Address)

		go func() {
			for {
				tid, vhost, err := cls.Login(cfg.Tunnel.VHost)
				if err != nil {
					log.Get().WithError(err).Error("join cluster failed")
					return
				}
				log.Get().Info("join cluster success, tid=", tid, " vhost=", vhost)
				cls.Heartbeat()
			}
		}()

		// run service
		if cfg.Services != nil {
			for index, svc := range cfg.Services {
				closer, err := common.RunService(t, cls, index, svc)
				if err != nil {
					log.Get().WithError(err).Error("run service failed, index=", index)
					continue
				}
				if closer != nil {
					svcClosers = append(svcClosers, closer)
				}
			}
		}
	})

	t.Run()
	t.Stop()

	for _, closer := range svcClosers {
		if closer != nil {
			closer.Close()
		}
	}
}
