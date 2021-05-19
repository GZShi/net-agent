package msgclient

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/GZShi/net-agent/rpc/msgclient/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/kataras/iris/websocket"
)

const defaultPrefix = "msgclient"

type client struct {
	t      tunnel.Tunnel
	ctx    tunnel.Context
	prefix string
}

func NewClient(t tunnel.Tunnel, ctx tunnel.Context) def.MsgClient {
	return &client{t, ctx, defaultPrefix}
}

func (c *client) SetPrefix(prefix string) {
	c.prefix = prefix
}

//
// 推送终端服务
//

type MsgClientSvc struct {
	prefix string
	t      tunnel.Tunnel

	wsclients []websocket.Connection
	wsMut     sync.Mutex
}

func NewService() *MsgClientSvc {
	return &MsgClientSvc{
		prefix:    defaultPrefix,
		t:         nil,
		wsclients: make([]websocket.Connection, 0)}
}

func (s *MsgClientSvc) Prefix() string {
	return s.prefix
}

func (s *MsgClientSvc) SetPrefix(prefix string) {
	s.prefix = prefix
}

func (s *MsgClientSvc) Hello(t tunnel.Tunnel) error {
	s.t = t
	return nil
}

func (s *MsgClientSvc) Exec(ctx tunnel.Context) error {
	switch ctx.GetMethod() {
	case "PushGM":
		s.PushGroupMessage_(ctx)
		return nil
	case "PushSN":
		s.PushSysNotify_(ctx)
		return nil
	}
	return fmt.Errorf("route failed: '%v' not found in '%v'", ctx.GetMethod(), ctx.GetService())
}

//
// 直接调用websocket连接推送消息
//

type WSMessage struct {
	SenderType string `json:"senderType"`
	Sender     string `json:"sender"`
	GroupID    uint32 `json:"groupID"`
	Message    string `json:"message"`
	MsgType    int    `json:"msgType"`
}

func (p *MsgClientSvc) PushMsg(senderType string, sender string, groupID uint32, message string, msgType int) {
	msg := &WSMessage{senderType, sender, groupID, message, msgType}
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("json.Marshal failed: %v\n", err.Error())
		return
	}

	p.wsMut.Lock()
	defer p.wsMut.Unlock()
	for _, wc := range p.wsclients {
		wc.EmitMessage(data)
	}
	if len(p.wsclients) <= 0 {
		fmt.Printf("[%v:%v]->[%v] type=%v msg=%v\n", senderType, sender, groupID, msgType, message)
	}
}

//
// “私有”方法
//
func (p *MsgClientSvc) RegisterWSClient(conn websocket.Connection) {
	p.wsMut.Lock()
	defer p.wsMut.Unlock()

	for _, wc := range p.wsclients {
		if wc == conn {
			return
		}
	}
	p.wsclients = append(p.wsclients, conn)
}

func (p *MsgClientSvc) UnregisterWSClient(conn websocket.Connection) {
	p.wsMut.Lock()
	defer p.wsMut.Unlock()

	for index, wc := range p.wsclients {
		if wc == conn {
			p.wsclients[index] = p.wsclients[len(p.wsclients)-1]
			p.wsclients = p.wsclients[:len(p.wsclients)-1]
			return
		}
	}
}
