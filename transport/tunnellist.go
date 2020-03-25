package transport

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/sirupsen/logrus"
)

// TunnelList 同名通道集合
type TunnelList struct {
	name            string
	list            []*Tunnel
	listMut         sync.RWMutex
	pollIndex       uint32
	onceDoZombCheck sync.Once
}

// NewTunnelList 创建同名列表
func NewTunnelList(name string) *TunnelList {
	return &TunnelList{
		name:      name,
		pollIndex: 0,
	}
}

// ZombTunnelCheck 检查僵尸通道，只允许调用一次
func (p *TunnelList) ZombTunnelCheck(checkDuration time.Duration, heartbeatTimeout time.Duration) {
	p.onceDoZombCheck.Do(func() {
		log.Get().WithField("name", p.name).Info("zomb checking starting")
		go func() {
			for {
				<-time.After(checkDuration)
				p.listMut.Lock()
				now := time.Now()
				size := len(p.list)
				deleted := 0
				for i := 0; i < size; i++ {
					hbeatTime := p.list[i].heartbeatTime
					if now.Sub(hbeatTime) > heartbeatTimeout {
						t := p.list[i]
						log.Get().WithFields(logrus.Fields{
							"name":      t.name,
							"created":   t.created,
							"hbeatTime": t.heartbeatTime,
						}).Warn("zomb tunnel detected")

						// delete tunnel
						p.list[i] = p.list[size-1]
						size--
						i--
						deleted++
					}
				}
				if deleted > 0 {
					p.list = p.list[:size]
				}
				p.listMut.Unlock()

				if deleted > 0 {
					p.BroadcastTextMessage(fmt.Sprintf("new tunnel join, count=%v", size))
				}
			}
		}()
	})
}

// Add 增加同名通道
func (p *TunnelList) Add(t *Tunnel) {
	if t != nil {
		p.listMut.Lock()
		p.list = append(p.list, t)
		size := len(p.list)
		p.listMut.Unlock()
		p.BroadcastTextMessage(fmt.Sprintf("new tunnel join, count=%v", size))
	}
}

// Del 删除同名通道
func (p *TunnelList) Del(t *Tunnel) {
	if t != nil {
		p.listMut.Lock()
		targetIndex := -1
		size := len(p.list)
		for i := 0; i < size; i++ {
			if p.list[i] == t {
				targetIndex = i
				break
			}
		}
		if targetIndex >= 0 {
			// 快速删除，不保证剩下元素的顺序
			p.list[targetIndex] = p.list[size-1]
			p.list[size-1] = nil
			p.list = p.list[:size-1]
			size--
		}
		p.listMut.Unlock()

		p.BroadcastTextMessage(fmt.Sprintf("one tunnel leaved, count=%v", size))
	}
}

// BroadcastTextMessage 向列表里的所有通道进行广播
func (p *TunnelList) BroadcastTextMessage(text string) {
	p.listMut.Lock()
	list := make([]*Tunnel, len(p.list))
	copy(list, p.list)
	p.listMut.Unlock()

	size := len(list)
	for i := 0; i < size; i++ {
		list[i].SendTextMessage(text)
	}
}

// PollTunnel 通过轮询获取tunnel
func (p *TunnelList) PollTunnel() *Tunnel {
	var t *Tunnel
	index := atomic.AddUint32(&p.pollIndex, 1)

	p.listMut.Lock()
	size := uint32(len(p.list))
	if size > 0 {
		t = p.list[index%size]
	}
	p.listMut.Unlock()

	return t
}
