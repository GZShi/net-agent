package service

import (
	"fmt"

	"github.com/GZShi/net-agent/rpc/msgclient/def"
	"github.com/GZShi/net-agent/tunnel"
)

func New(t tunnel.Tunnel) def.MsgClient {
	return &impl{t}
}

type impl struct {
	t tunnel.Tunnel
}

func (p *impl) PushGroupMessage(sender string, groupID uint32, message string, msgType int) {
	fmt.Printf("[vhost:%v]->[%v] type=%v msg=%v\n", sender, groupID, msgType, message)
}

func (p *impl) PushSysNotify(groupID uint32, message string, msgType int) {
	fmt.Printf("[%v]->[%v] type=%v msg=%v\n", "system", groupID, msgType, message)
}
