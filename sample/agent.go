package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

func connectAsAgent(addr, password string) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return
	}

	defer conn.Close()
	
	cc, err := cipherconn.New(conn, password)
	if err != nil {
		return
	}

	t := tunnel.New(cc)

	// agent 默认只支持直接创建连接
	t.Listen("dial/dierct", handleDialDirect)

	go registerAgentToServer(t)

	t.Run()
}

func registerAgentToServer(t tunnel.Tunnel) error {
	if err := joinCluster(t); err != nil {
		return err
	}

	// use control c to stop agent work
	utils.WaitCtrlC()

	detachCluster(t)
	t.Stop()

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
