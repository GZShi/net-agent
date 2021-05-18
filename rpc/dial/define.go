package dial

import (
	"fmt"

	"github.com/GZShi/net-agent/rpc/dial/service"

	"github.com/GZShi/net-agent/rpc/dial/def"
	"github.com/GZShi/net-agent/tunnel"
)

//
// NewClient 获取新的实例
//
func NewClient(t tunnel.Tunnel, ctx tunnel.Context) def.Dial {
	return &client{t, ctx, "dial"}
}

type client struct {
	t      tunnel.Tunnel
	ctx    tunnel.Context
	prefix string
}

func (c *client) SetPrefix(prefix string) {
	c.prefix = prefix
}

//
// NewService 创建rpc服务
//
func NewService() tunnel.Service {
	return &svc{"dial", nil, nil}
}

type svc struct {
	prefix string
	t      tunnel.Tunnel
	impl   def.Dial
}

func (s *svc) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *svc) Prefix() string {
	return s.prefix
}

func (s *svc) Hello(t tunnel.Tunnel) error {
	s.t = t
	s.impl = service.New(t)
	return nil
}

func (s *svc) Exec(ctx tunnel.Context) error {
	switch ctx.GetMethod() {
	case "Dial":
		s.Dial(ctx)
		return nil
	}

	return fmt.Errorf("route failed: '%v' not found in '%v'", ctx.GetMethod(), ctx.GetService())
}
