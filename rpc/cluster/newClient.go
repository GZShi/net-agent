package cluster

import (
	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
)

// Client rpc客户端
type Client interface {
	Join() (exchanger.TID, error)
	// Detach()
	SetLabels(labels []string) (finnalLabels []string, err error)
	// RemoveLabels(labels []string) error
}

// NewClient 构建客户端
func NewClient(t tunnel.Tunnel) Client {
	return &client{t}
}

type client struct {
	t tunnel.Tunnel
}
