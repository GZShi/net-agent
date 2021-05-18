package cluster

import (
	"fmt"

	"github.com/GZShi/net-agent/tunnel"
)

type stReqSendGroupMessage struct {
	GroupID uint32 `json:"groupID"`
	Message string `json:"message"`
	MsgType int    `json:"msgType"`
}

// SendGroupMessage 发送群组消息
func (s *svc) SendGroupMessage(ctx tunnel.Context) {
	var req stReqSendGroupMessage
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.impl.SendGroupMessage(req.GroupID, req.Message, req.MsgType)
	if err != nil {
		ctx.Error(err)
		return
	}
}

// SendGroupMessage 客户端发送消息接口
func (c *client) SendGroupMessage(groupID uint32, message string, msgType int) error {
	err := c.t.SendJSON(
		c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "SendGroupMessage"),
		&stReqSendGroupMessage{
			groupID, message, msgType,
		},
		nil)
	if err != nil {
		fmt.Printf("SendGroupMessage error: %v\n", err)
	}
	return err
}
