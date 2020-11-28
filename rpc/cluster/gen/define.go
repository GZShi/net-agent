package gen

import (
	"errors"

	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/rpc/cluster/impl"
	"github.com/GZShi/net-agent/tunnel"
)

//
// NewClient 创建rpc客户端
//
func NewClient(t tunnel.Tunnel, ctx tunnel.Context) cluster.Cluster {
	return &client{t, ctx}
}

type client struct {
	t   tunnel.Tunnel
	ctx tunnel.Context
}

//
// NewService 创建rpc服务
//
func NewService(prefix string) tunnel.Service {
	return &service{prefix, nil, impl.New()}
}

type service struct {
	prefix string
	t      tunnel.Tunnel
	impl   cluster.Cluster
}

func (s *service) Prefix() string {
	return s.prefix
}

func (s *service) Hello(t tunnel.Tunnel) error {
	s.t = t
	return nil
}

func (s *service) Exec(ctx tunnel.Context) error {
	return errors.New("exec not implement")
}
