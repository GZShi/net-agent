package cluster

import (
	"time"

	"github.com/GZShi/net-agent/rpc/cluster/def"
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
	ctx.JSON(nil)
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
	return err
}

type stReqGetGMs struct {
	Vhost     string
	GroupIDs  []uint32
	StartTime time.Time
	Limit     int
}
type stRespGetGMs struct {
	Messages []def.Message
}

func (c *client) GetGroupMessages(groupIDs []uint32, startTime time.Time, limit int) ([]def.Message, error) {
	var resp stRespGetGMs
	err := c.t.SendJSON(
		c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "GetGroupMessages"),
		&stReqGetGMs{
			Vhost:     "",
			GroupIDs:  groupIDs,
			StartTime: startTime,
			Limit:     limit,
		},
		&resp)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

func (s *svc) GetGroupMessages(ctx tunnel.Context) {
	var req stReqGetGMs
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	messages, err := s.impl.GetGroupMessages(req.GroupIDs, req.StartTime, req.Limit)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(&stRespGetGMs{
		Messages: messages,
	})
}
