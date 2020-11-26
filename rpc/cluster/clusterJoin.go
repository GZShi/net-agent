package cluster

import (
	"github.com/GZShi/net-agent/exchanger"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

type joinResponse struct {
	TunnelID exchanger.TID `json:"tid"`
}

func (c *client) Join() (tid exchanger.TID, err error) {
	var resp joinResponse
	err = c.t.SendJSON(nil, nameOfJoin, nil, &resp)
	if err != nil {
		return exchanger.InvalidTID, err
	}
	return resp.TunnelID, nil
}

func (s *service) Join(ctx tunnel.Context) {

	tid, err := s.cluster.Join(ctx.GetTunnel())
	if err != nil {
		ctx.Error(err)
		return
	}

	log.Get().Info("new tunnel join: ", tid)

	ctx.JSON(&joinResponse{tid})
}
