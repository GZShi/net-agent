package cluster

import (
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

type stReqLogin struct {
	Vhost string `json:"vhost"`
}
type stRespLogin struct {
	TID def.TID `json:"tid"`
}

func (c *client) Login(vhost string) (def.TID, error) {
	var resp stRespLogin
	err := c.t.SendJSON(c.ctx, tunnel.JoinServiceMethod(c.prefix, "Login"),
		&stReqLogin{vhost}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.TID, nil
}

func (s *svc) Login(ctx tunnel.Context) {
	var req stReqLogin
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}
	tid, err := s.impl.Login(req.Vhost)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(&stRespLogin{tid})
}

func (c *client) Logout() error {
	return c.t.SendJSON(c.ctx, tunnel.JoinServiceMethod(c.prefix, "Logout"), nil, nil)
}
