package exchanger

import (
	"errors"

	"github.com/GZShi/net-agent/tunnel"
)

type Tunnels struct {
	list   []tunnel.Tunnel
	ids    map[uint32]tunnel.Tunnel
	labels map[string]tunnel.Tunnel
}

func (ts *Tunnels) Join(t tunnel.Tunnel) error {
	return nil
}

func (ts *Tunnels) FindTunnelByID(id uint32) (tunnel.Tunnel, error) {
	return nil, errors.New("not found")
}

func (ts *Tunnels) FindTunnelByLabel(label string) (tunnel.Tunnel, error) {
	return nil, errors.New("not found")
}
