package cluster

import (
	"errors"
	"fmt"

	"github.com/GZShi/net-agent/exchanger"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

// NewService 创建rpc服务模块
func NewService(cluster exchanger.Cluster) tunnel.Service {
	return &service{
		route:   make(map[string]tunnel.OnRequestFunc),
		cluster: cluster,
		t:       nil,
	}
}

type service struct {
	route   map[string]tunnel.OnRequestFunc
	cluster exchanger.Cluster
	t       tunnel.Tunnel
}

func (s *service) Hello(t tunnel.Tunnel) error {
	if t == nil {
		return errors.New("tunnel is nil")
	}
	s.t = t
	log.Get().Info("service.", s.Prefix(), " enabled")
	return nil
}

func (s *service) Prefix() string {
	return namePrefix
}

func (s *service) Exec(ctx tunnel.Context) error {
	if s.cluster == nil {
		return errors.New("cluster is nil")
	}

	cmd := ctx.GetCmd()
	switch cmd {
	case nameOfJoin:
		s.Join(ctx)
		return nil
	case nameOfDetach:
	case nameOfSetLabels:
		s.SetLabels(ctx)
		return nil
	case nameOfRemoveLabels:
	}

	return fmt.Errorf("handler of '%v' not found", cmd)
}
