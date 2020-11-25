package dial

import (
	"fmt"

	"github.com/GZShi/net-agent/tunnel"
)

// NewService 创建rpc服务模块
func NewService() tunnel.Service {
	return &service{}
}

type service struct {
	route map[string]tunnel.OnRequestFunc
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
