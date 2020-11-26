package main

import (
	"net"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/rpc/dial"

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

	client := cluster.NewClient(t)
	tid, err := client.Join()
	if err != nil {
		log.Get().WithError(err).Error("join cluster failed")
		return
	}
	log.Get().Info("join cluster ok: ", tid)
	globalTID = tid

	// agent 默认只支持直接创建连接
	if err = t.BindService(dial.NewService()); err != nil {
		log.Get().WithError(err).Error("bind service failed")
		return
	}

	log.Get().Info("agent created")
	t.Run()
	log.Get().Info("agent closed")
}
