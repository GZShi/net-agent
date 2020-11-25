package dial

import (
	"errors"
	"net"

	"github.com/GZShi/net-agent/tunnel"
)

func (c *client) DialWithTunnelLabel(label, network, address string) (net.Conn, error) {
	return nil, errors.New("not implement")
}

func (s *service) DialWithTunnelLabel(ctx tunnel.Context) {
	ctx.Error(errors.New("not implement"))
}
