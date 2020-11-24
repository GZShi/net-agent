package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

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

	// agent 默认只支持直接创建连接
	t.Listen("dial/dierct", handleDialDirect)

	go registerAgentToServer(t)

	log.Get().Info("tunnel[agent] created")
	t.Run()
	log.Get().Info("tunnel[agent] closed")
}

func registerAgentToServer(t tunnel.Tunnel) error {
	if err := joinCluster(t); err != nil {
		return err
	}

	// use control c to stop agent work
	log.Get().Info("press ctrl+c to stop agent work")
	utils.WaitCtrlC()
	log.Get().Info("close agent tunnel...")

	detachCluster(t)
	t.Stop()

	log.Get().Info("agent stopped")
	return nil
}

func joinCluster(t tunnel.Tunnel) error {
	t.SendJSON(nil, "join/cluster", nil, nil)
	return nil
}

func detachCluster(t tunnel.Tunnel) error {
	t.SendJSON(nil, "detach/cluster", nil, nil)
	return nil
}
