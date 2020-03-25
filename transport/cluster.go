package transport

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

// TunnelCluster 集群
type TunnelCluster struct {
	groups      sync.Map
	secret      string
	activeCount int32

	// agent通过tunnel连过来后，服务端会在auth之后下发uid，保证agent的唯一性
	// 需要注意的是，服务端重启后依然要保证唯一性，所以需要加上时间因子
	// to be done
	agentUID int32
}

// NewTunnelCluster 创建集群
func NewTunnelCluster(secret string) *TunnelCluster {
	return &TunnelCluster{
		secret:      secret,
		activeCount: 0,
	}
}

// Run 开启服务
func (p *TunnelCluster) Run(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept failed")
			break
		}

		// 等待客户端channel连进来
		go func(client net.Conn) {
			name, randKey, err := CheckAgentConn(client, p.secret)
			if err != nil {
				log.Get().WithError(err).Error("auth failed")
				return
			}

			tunnel, err := NewTunnel(client, name, p.secret, randKey, false)
			if err != nil {
				log.Get().WithError(err).Error("create tunnel failed")
				return
			}
			group, _ := p.groups.LoadOrStore(name, NewTunnelList(name))
			tList := group.(*TunnelList)
			tList.ZombTunnelCheck(time.Second*30, time.Second*120)
			tList.Add(tunnel)

			tunnel.conn.Write([]byte{0x00})
			atomic.AddInt32(&p.activeCount, 1)
			log.Get().WithField("name", name).WithField("addr", client.RemoteAddr()).Info("new tunnel created")
			tunnel.StartHeartbeat()

			// 当tunnel还在工作的时候，会一直在这里block
			tunnel.Serve()

			tList.Del(tunnel)
			atomic.AddInt32(&p.activeCount, -1)
		}(conn)
	}
}

// Dial 通过channelName拨号
func (p *TunnelCluster) Dial(sourceAddr, network, addr, channelName, userName string) (net.Conn, error) {
	group, ok := p.groups.Load(channelName)

	if !ok {
		return nil, fmt.Errorf("channel(%v) not found", channelName)
	}

	tList := group.(*TunnelList)
	t := tList.PollTunnel()
	if t == nil {
		return nil, fmt.Errorf("channel(%v) is empty", channelName)
	}

	return t.Dial(sourceAddr, network, addr, userName)
}
