package common

import (
	"errors"
	"io"
	"sync/atomic"

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
func RunService(t tunnel.Tunnel, cls def.Cluster, index int, info ServiceInfo) (io.Closer, error) {
	log := logger.Get().WithField("svcindex", index)

	if !info.Enable {
		return nil, errors.New("service disabled")
	}

	log.WithField("desc", info.Desc).Info("init service")

	var closer io.Closer
	var err error

	switch info.Type {
	case "socks5":
		closer, err = RunSocks5Server(t, cls, info.Param, log)
	case "portproxy":
		closer, err = RunPortproxy(t, cls, info.Param, log)
	case "chatserver":
		closer, err = RunChatServer(t, cls, info.Param, log)
	default:
		err = errors.New("unknown service type: " + info.Type)
	}

	return closer, err
}

type closer struct {
	sigCh      chan int
	closeTimes int32
}

func newCloser() *closer {
	return &closer{
		sigCh:      make(chan int),
		closeTimes: 0,
	}
}

func (p *closer) Close() error {
	times := atomic.AddInt32(&p.closeTimes, 1)
	if times > 1 {
		return errors.New("closer closed")
	}
	p.sigCh <- 1
	return nil
}

func (p *closer) WaitClose() {
	<-p.sigCh
}
