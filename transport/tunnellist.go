package transport

import (
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
		for {
			<-time.After(checkDuration)
			p.listMut.Lock()
			now := time.Now()
			size := len(p.list)
			for i := 0; i < size; i++ {
				if now.Sub(p.list[i].heartbeatTime) > heartbeatTimeout {
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
				}
			}

			p.listMut.Unlock()
		}
	})
}

// Add 增加同名通道
func (p *TunnelList) Add(t *Tunnel) {
	if t != nil {
		p.listMut.Lock()
		p.list = append(p.list, t)
		p.listMut.Unlock()
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
		}
		p.listMut.Unlock()
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
