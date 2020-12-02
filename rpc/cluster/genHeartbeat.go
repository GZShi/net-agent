package cluster

import (
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

func (c *client) Heartbeat() error {
	var err error
	for {
		<-time.After(time.Second * 4)
		err = c.t.SendJSON(c.ctx,
			tunnel.JoinServiceMethod(c.prefix, "Heartbeat"),
			nil, nil)
		if err != nil {
			log.Get().WithError(err).Error("heartbeat stopped")
			return err
		}
	}
}

func (s *svc) Heartbeat(ctx tunnel.Context) {
	if err := s.impl.Heartbeat(); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(nil)
}
