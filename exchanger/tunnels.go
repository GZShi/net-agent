package exchanger

import (
	"errors"

	"github.com/GZShi/net-agent/tunnel"
)

type Cluster interface {
	Join(t tunnel.Tunnel) error
	Detach(t tunnel.Tunnel)

	FindTunnelByID(id uint32) (tunnel.Tunnel, error)
	FindTunnelByLabel(label string) (tunnel.Tunnel, error)
}

type tunnelList struct {
	list   []tunnel.Tunnel
	ids    map[uint32]tunnel.Tunnel
	labels map[string]tunnel.Tunnel
}

// NewCluster 创建新的tunnel集群
func NewCluster() Cluster {
	return &tunnelList{}
}

func (ts *tunnelList) Join(t tunnel.Tunnel) error {
	return nil
}

func (ts *tunnelList) Detach(t tunnel.Tunnel) {}

func (ts *tunnelList) FindTunnelByID(id uint32) (tunnel.Tunnel, error) {
	return nil, errors.New("not found")
}

func (ts *tunnelList) FindTunnelByLabel(label string) (tunnel.Tunnel, error) {
	return nil, errors.New("not found")
}
