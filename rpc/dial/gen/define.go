package gen

import (
	"errors"

	"github.com/GZShi/net-agent/rpc/dial"
	"github.com/GZShi/net-agent/tunnel"
)

//
// NewClient 获取新的实例
//
func NewClient(t tunnel.Tunnel, ctx tunnel.Context) dial.Dial {
	return &client{t, ctx}
}

type client struct {
	t   tunnel.Tunnel
	ctx tunnel.Context
}

//
// New Service 创建rpc服务
//
func NewService(prefix string) tunnel.Service {
	return &service{prefix, nil, nil}
}

type service struct {
	prefix string
	t      tunnel.Tunnel
	impl   dial.Dial
}

func (s *service) Prefix() string {
	return s.prefix
}

func (s *service) Hello(t tunnel.Tunnel) error {
	s.t = t
	return nil
}

func (s *service) Exec(ctx tunnel.Context) error {
	return errors.New("exec not implemenet")
}
