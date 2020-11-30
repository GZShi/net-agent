package cluster

import (
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

type stRespLogin struct {
	TID def.TID `json:"tid"`
}

func (c *client) Login() (def.TID, error) {
	var resp stRespLogin
	err := c.t.SendJSON(c.ctx, tunnel.JoinServiceMethod(c.prefix, "Login"), nil, &resp)
	if err != nil {
		return 0, err
	}
	return resp.TID, nil
}

func (s *svc) Login(ctx tunnel.Context) {
	tid, err := s.impl.Login()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(&stRespLogin{tid})
}

func (c *client) Logout() error {
	return c.t.SendJSON(c.ctx, tunnel.JoinServiceMethod(c.prefix, "Logout"), nil, nil)
}
