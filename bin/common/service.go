package common

import (
	"github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	Enable bool              `json:"enable"`
	Desc   string            `json:"description"`
	Type   string            `json:"type"`
	Param  map[string]string `json:"param"`
}

// RunService 运行服务
func RunService(t tunnel.Tunnel, cls def.Cluster, index int, info ServiceInfo) {
	log := logger.Get().WithField("svcindex", index)

	if !info.Enable {
		log.WithField("desc", info.Desc).Warn("service disabled")
		return
	}

	log.WithField("desc", info.Desc).Info("init service")

	switch info.Type {
	case "socks5":
		RunSocks5Server(t, cls, info.Param, log)
	case "portproxy":
		RunPortproxy(t, cls, info.Param, log)
	default:
		log.Error("unknown service type: " + info.Type)
	}
}
