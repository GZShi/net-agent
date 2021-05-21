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

	members sync.Map // map[vhost:string]*memberInfo
}

func (g *msgGroup) IsMember(vhost string) bool {
	if g.groupID == 0 {
		return true
	}
	_, found := g.members.Load(vhost)
	return found
}

func (g *msgGroup) GetMemberInfo(vhost string) (info *memberInfo, found bool) {
	val, found := g.members.Load(vhost)
	if !found {
		return nil, false
	}
	return val.(*memberInfo), true
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
func (g *msgGroup) GetMessages(startTime time.Time, limit int) ([]*msg, error) {
	g.listMut.RLock()
	msgs := g.list[:] // copy slice
	g.listMut.RUnlock()

	// 如果当前消息列表为空，则直接返回，不做多余判断
	if len(msgs) == 0 {
		return msgs, nil
	}

	// 如果当前消息列表长度大于limit，则优先截短
	if len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}

	// 如果第一条消息已经是startTime之后，则后面消息不用再判断时间，全部返回
	if msgs[0].Date.After(startTime) {
		return msgs, nil
	}
	if msgs[len(msgs)-1].Date.Before(startTime) {
		return nil, nil
	}

	// 根据利用二分法，找到大于startTime的记录
	mid := 0
	start := 0
	end := len(msgs)
	for start < end {
		mid = (start + end) >> 1
		if msgs[mid].Date.Before(startTime) {
			start = mid + 1
		} else if msgs[mid].Date.After(startTime) {
			end = mid
		} else {
			break
		}
	}

	return msgs[mid:], nil
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
