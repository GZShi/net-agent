package dial

import (
	"net"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
)

// Client dial client 协议
type Client interface {
	DialDirect(network, address string) (net.Conn, error)
	DialWithTunnelID(tid exchanger.TID, network, address string) (net.Conn, error)
	DialWithTunnelLabel(label, network, address string) (net.Conn, error)
}

// NewClient 创建rpc客户端
func NewClient(t tunnel.Tunnel) Client {
	return &client{t}
}

type client struct {
	t tunnel.Tunnel
}
