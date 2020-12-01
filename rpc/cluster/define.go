package cluster

import (
	"fmt"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/cluster/service"
	"github.com/GZShi/net-agent/tunnel"
)

//
// NewClient 创建rpc客户端
//
func NewClient(t tunnel.Tunnel, ctx tunnel.Context) def.Cluster {
	return &client{t, ctx, "cluster"}
}

type client struct {
	t      tunnel.Tunnel
	ctx    tunnel.Context
	prefix string
}

//
// NewService 创建rpc服务
//
func NewService() tunnel.Service {
	return &svc{"cluster", nil, nil}
}

type svc struct {
	prefix string
	t      tunnel.Tunnel
	impl   def.Cluster
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
	case "Login":
		s.Login(ctx)
		return nil
	case "DialByTID":
		s.DialByTID(ctx)
		return nil
	case "Dial":
		s.Dial(ctx)
		return nil
	}
	return fmt.Errorf("route failed: '%v' not found in '%v'", ctx.GetMethod(), ctx.GetService())
}
