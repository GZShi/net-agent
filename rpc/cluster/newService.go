package cluster

import (
	"errors"
	"fmt"

	"github.com/GZShi/net-agent/exchanger"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

// NewService 创建rpc服务模块
func NewService() tunnel.Service {
	return &service{
		route:   make(map[string]tunnel.OnRequestFunc),
		cluster: exchanger.NewCluster(),
		t:       nil,
	}
}

type service struct {
	route   map[string]tunnel.OnRequestFunc
	cluster exchanger.Cluster
	t       tunnel.Tunnel
}

func (s *service) Hello(t tunnel.Tunnel) error {
	if t != nil {
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
	cmd := ctx.GetCmd()
	switch cmd {
	case nameOfJoin:
	case nameOfDetach:
	case nameOfSetLabels:
	case nameOfRemoveLabels:
	}

	return fmt.Errorf("handler of '%v' not found", cmd)
}
