package common

import (
	log "github.com/GZShi/net-agent/logger"
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
func RunService(t tunnel.Tunnel, cls def.Cluster, info ServiceInfo) {
	if !info.Enable {
		return
	}

	switch info.Type {
	case "socks5":
		go RunSocks5Server(t, cls, info.Desc, info.Param)
	case "portproxy":
		go RunPortproxy(t, cls, info.Desc, info.Param)
	default:
		log.Get().Error("unknown service type: " + info.Type)
	}

}
