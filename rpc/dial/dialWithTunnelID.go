package dial

import (
	"errors"
	"net"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
)

func (c *client) DialWithTunnelID(tid exchanger.TID, network, address string) (net.Conn, error) {
	return nil, errors.New("not implement")
}

func (s *service) DialWithTunnelID(ctx tunnel.Context) {
	ctx.Error(errors.New("not implement"))
}
