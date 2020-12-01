package main

import (
	"net"

	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/dial"

	"github.com/GZShi/net-agent/cipherconn"
	logger "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

var globalTID = def.TID(0)

func connectAsAgent(addr, password string) {
	log := logger.Get().WithField("mode", "agent")
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.WithError(err).Error("connect ", addr, " failed")
		return
	}

	defer conn.Close()
	log.Info("connect ", addr, " success")

	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)

	t.Ready(func(t tunnel.Tunnel) {
		go enableSocks5Server(t)

		client := cluster.NewClient(t, nil)
		tid, err := client.Login("test.tunnel")
		if err != nil {
			log.WithError(err).Error("join cluster failed")
			return
		}
		log.WithField("tid", tid).Info("login in success")
		globalTID = tid
	})

	// agent 默认只支持直接创建连接
	err = t.BindServices(
		dial.NewService(),
	)
	if err != nil {
		log.WithError(err).Error("bind service failed")
		return
	}

	log.Info("agent created")
	t.Run()
	log.Info("agent closed")
}
