package cluster

import (
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

///////////////////
type stReqLogin struct {
	Vhost string `json:"vhost"`
}
type stRespLogin struct {
	TID   def.TID `json:"tid"`
	Vhost string  `json:"vhost"`
}

func (c *client) Login(vhost string) (def.TID, string, error) {
	var resp stRespLogin
	err := c.t.SendJSON(c.ctx, tunnel.JoinServiceMethod(c.prefix, "Login"),
		&stReqLogin{vhost}, &resp)
	if err != nil {
		return 0, "", err
	}
	return resp.TID, resp.Vhost, nil
}

func (s *svc) Login(ctx tunnel.Context) {
	var req stReqLogin
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}
	tid, vhost, err := s.impl.Login(req.Vhost)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(&stRespLogin{tid, vhost})
}

///////////////////////

func (c *client) Logout() error {
	return c.t.SendJSON(c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "Logout"), nil, nil)
}

func (s *svc) Logout(ctx tunnel.Context) {
	s.impl.Logout()
	ctx.JSON(nil)
}

///////////////////////

type stRespGetCtxInfo struct {
	CtxInfo def.CtxInfo
}

func (c *client) GetCtxInfo() (def.CtxInfo, error) {
	var resp stRespGetCtxInfo
	err := c.t.SendJSON(c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "GetCtxInfo"), nil, &resp,
	)
	return resp.CtxInfo, err
}

func (s *svc) GetCtxInfo(ctx tunnel.Context) {
	info, err := s.impl.GetCtxInfo()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(&stRespGetCtxInfo{info})
}
