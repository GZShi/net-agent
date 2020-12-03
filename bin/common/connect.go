package common

import (
	"errors"
	"net"
	"strings"

	"github.com/GZShi/net-agent/bin/ws"
	"github.com/GZShi/net-agent/cipherconn"
	"github.com/GZShi/net-agent/tunnel"
)

// TunnelInfo 隧道连接信息
type TunnelInfo struct {
	Network  string `json:"network"`
	Address  string `json:"address"`
	Password string `json:"password"`
	VHost    string `json:"vhost"`
}

// ConnectTunnel 根据配置信息创建隧道连接
func ConnectTunnel(info *TunnelInfo) (tunnel.Tunnel, error) {

	var conn net.Conn
	var err error
	addr := info.Address

	if strings.HasPrefix(addr, "ws://") || strings.HasPrefix(addr, "wss://") {
		conn, err = ws.Dial(addr)
	} else {
		conn, err = net.Dial("tcp4", info.Address)
	}

	if err != nil {
		return nil, errors.New("connect to tunnel failed: " + info.Address)
	}

	cc, err := cipherconn.New(conn, info.Password)
	if err != nil {
		return nil, errors.New("create cipherconn failed")
	}

	return tunnel.New(cc), nil
}
