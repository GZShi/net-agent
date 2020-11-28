package main

import (
	"net"

	"github.com/GZShi/net-agent/exchanger"
	clusterGen "github.com/GZShi/net-agent/rpc/cluster/gen"
	dialGen "github.com/GZShi/net-agent/rpc/dial/gen"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

var globalTID = exchanger.InvalidTID

func connectAsAgent(addr, password string) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Get().WithError(err).Error("connect ", addr, " failed")
		return
	}

	defer conn.Close()
	log.Get().Info("connect ", addr, " success")

	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.Get().WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)

	t.Ready(func(t tunnel.Tunnel) {
		client := clusterGen.NewClient(t, nil)
		err := client.Login()
		if err != nil {
			log.Get().WithError(err).Error("join cluster failed")
			return
		}
		log.Get().Info("login in success")
	})

	// agent 默认只支持直接创建连接
	err = t.BindServices(
		dialGen.NewService("dial"),
	)
	if err != nil {
		log.Get().WithError(err).Error("bind service failed")
		return
	}

	log.Get().Info("agent created")
	t.Run()
	log.Get().Info("agent closed")
}
