package msgclient

import (
	"errors"

	"github.com/GZShi/net-agent/tunnel"
)

type stReqPushGM struct {
	Sender  string `json:"sender"`
	GroupID uint32 `json:"groupID"`
	Message string `json:"message"`
	MsgType int    `json:"msgType"`
}
type stRespPushGM struct {
}

// PushGroupMessage 推送群组消息（思考：推送是否需要错误返回）
func (c *client) PushGroupMessage(sender string, groupID uint32, message string, msgType int) {
	c.t.SendJSON(c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "PushGM"),
		&stReqPushGM{sender, groupID, message, msgType},
		&stRespPushGM{},
	)
}

func (s *svc) PushGroupMessage(ctx tunnel.Context) {
	var req stReqPushGM
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(errors.New("get json data failed"))
		return
	}

	s.impl.PushGroupMessage(req.Sender, req.GroupID, req.Message, req.MsgType)
	ctx.JSON(&stRespPushGM{})
}

type stReqPushSN struct {
	GroupID uint32 `json:"groupID"`
	Message string `json:"message"`
	MsgType int    `json:"msgType"`
}
type stRespPushSN struct {
}

func (c *client) PushSysNotify(groupID uint32, message string, msgType int) {
	c.t.SendJSON(c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "PushSN"),
		&stReqPushSN{groupID, message, msgType},
		&stRespPushSN{},
	)
}
func (s *svc) PushSysNotify(ctx tunnel.Context) {
	var req stReqPushSN
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(errors.New("get json data failed"))
		return
	}

	s.impl.PushSysNotify(req.GroupID, req.Message, req.MsgType)
	ctx.JSON(&stRespPushSN{})
}
