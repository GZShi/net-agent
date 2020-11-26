package dial

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
	}
}

type service struct {
	route   map[string]tunnel.OnRequestFunc
	t       tunnel.Tunnel
	cluster exchanger.Cluster
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
	cmd := ctx.GetCmd()

	switch cmd {
	case nameOfDialDirect:
		s.DialDirect(ctx)

	case nameOfDialWithTunnelID:
		s.DialWithTunnelID(ctx)

	case nameOfDialWithTunnelLabel:
		s.DialWithTunnelLabel(ctx)

	default:
		return fmt.Errorf("handler of '%v' not found", cmd)
	}

	return nil
}
