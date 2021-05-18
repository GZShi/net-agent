package msgclient

import (
	"fmt"

	"github.com/GZShi/net-agent/rpc/msgclient/def"
	"github.com/GZShi/net-agent/rpc/msgclient/service"
	"github.com/GZShi/net-agent/tunnel"
)

const defaultPrefix = "msgclient"

type client struct {
	t      tunnel.Tunnel
	ctx    tunnel.Context
	prefix string
}

func NewClient(t tunnel.Tunnel, ctx tunnel.Context) def.MsgClient {
	return &client{t, ctx, defaultPrefix}
}

func (c *client) SetPrefix(prefix string) {
	c.prefix = prefix
}

//
//
//

type svc struct {
	prefix string
	t      tunnel.Tunnel
	impl   def.MsgClient
}

func NewService() tunnel.Service {
	return &svc{defaultPrefix, nil, nil}
}

func (s *svc) Prefix() string {
	return s.prefix
}

func (s *svc) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *svc) Hello(t tunnel.Tunnel) error {
	s.t = t
	s.impl = service.New(t)
	return nil
}

func (s *svc) Exec(ctx tunnel.Context) error {
	switch ctx.GetMethod() {
	case "PushGM":
		s.PushGroupMessage(ctx)
		return nil
	case "PushSN":
		s.PushSysNotify(ctx)
		return nil
	}
	return fmt.Errorf("route failed: '%v' not found in '%v'", ctx.GetMethod(), ctx.GetService())
}
