package service

import (
	"errors"
	"sync"
	"time"
)

var (
	errGroupNotFound error = errors.New("group not found")
)

//
// msgCenter
//

var msgCenterInstance *msgCenter
var onceMsgCenterInit sync.Once

// getCluster 单例模式
func getMsgCenter() *msgCenter {
	onceMsgCenterInit.Do(func() {
		msgCenterInstance = newMsgCenter()
	})
	return msgCenterInstance
}

type msgCenter struct {
	groups []*msgGroup
}

func newMsgCenter() *msgCenter {
	zeroGroup := &msgGroup{groupID: 0}
	mc := &msgCenter{
		groups: make([]*msgGroup, 0),
	}

	mc.groups = append(mc.groups, zeroGroup)
	return mc
}

func (mc *msgCenter) GetGroupByID(id uint32) (group *msgGroup, err error) {
	if id > uint32(len(mc.groups)) {
		return nil, errGroupNotFound
	}
	g := mc.groups[id]
	if g == nil {
		return nil, errGroupNotFound
	}
	return g, nil
}

func (mc *msgCenter) PushMessage(m *msg) error {
	g, err := mc.GetGroupByID(m.GroupID)
	if err != nil {
		return err
	}
	return g.PushMessage(m)
}

func (mc *msgCenter) GetMessages(groupID uint32) ([]*msg, error) {
	g, err := mc.GetGroupByID(groupID)
	if err != nil {
		return nil, err
	}

	return g.GetMessages()
}

//
// messages
//

type msg struct {
	SenderVhost string
	GroupID     uint32
	Date        time.Time
	Message     string
	MsgType     int
}

type memberInfo struct {
	JoinDate time.Time
	Nickname string
	Role     byte
}

type msgGroup struct {
	groupID uint32
	listMut sync.RWMutex
	list    []*msg

	members sync.Map
}

// PushMessage 往群里发送消息
func (g *msgGroup) PushMessage(m *msg) error {

	// 检查是否为群组成员。任何人都能在0号聊天室发言
	if g.groupID > 0 {
		_, found := g.members.Load(m.SenderVhost)
		if !found {
			return errors.New("you are not group member")
		}
	}

	g.listMut.Lock()
	if g.list == nil {
		g.list = make([]*msg, 0)
	}
	g.list = append(g.list, m)
	g.listMut.Unlock()

	if g.groupID == 0 {
		go getCluster().DispatchGMToAll(m)
	} else {
		vhosts := []string{}
		g.members.Range(func(k, v interface{}) bool {
			vhosts = append(vhosts, v.(string))
			return true
		})
		go getCluster().DispatchGMToVhosts(m, vhosts)
	}

	return nil
}

// GetMessage 获取群里消息
func (g *msgGroup) GetMessages() ([]*msg, error) {
	g.listMut.RLock()
	defer g.listMut.RUnlock()
	if len(g.list) > 100 {
		return g.list[len(g.list)-100:], nil
	}
	return g.list, nil
}

// Join 加入群
func (g *msgGroup) Join(vhost string) error {
	g.members.LoadOrStore(vhost, &memberInfo{
		// todo
	})
	return nil
}

// Leave 离开群
func (g *msgGroup) Leave(vhost string) error {
	g.members.Delete(vhost)
	return nil
}
