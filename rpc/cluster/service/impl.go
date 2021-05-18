package service

import (
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/dial"
	msgclientdef "github.com/GZShi/net-agent/rpc/msgclient/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

var (
	errNotImplement       = errors.New("method not implement")
	errTunnelNotFoundByID = errors.New("tunnel not found by id")
)

// New 获取新的实例
func New(t tunnel.Tunnel) def.Cluster {
	return &impl{
		t:   t,
		cls: getCluster(),
		mc:  getMsgCenter(),
	}
}

// connectContext 当客户端调用cluster.Login成功后，生成上下文信息
type connectContext struct {
	t             tunnel.Tunnel
	tid           def.TID
	vhost         string
	msgClient     msgclientdef.MsgClient
	connectTime   time.Time
	lastHeartbeat time.Time
}

func (ctx *connectContext) String() string {
	beatGap := time.Since(ctx.lastHeartbeat).Round(time.Millisecond)
	alive := time.Since(ctx.connectTime).Round(time.Second)
	return fmt.Sprintf("%v(%v) beat_gap=%v alive=%v", ctx.vhost, ctx.tid, beatGap, alive)
}

type impl struct {
	t   tunnel.Tunnel // 这个Tunnel是发起请求的通道，相当于客户端应用
	cls *cluster

	// connCtx info
	connCtx *connectContext

	// msg center for chat
	mc *msgCenter
}

func (p *impl) Heartbeat() error {
	if p.connCtx != nil {
		p.connCtx.lastHeartbeat = time.Now()
		return nil
	}
	return errors.New("cache ctx not found")
}

func (p *impl) Login(vhost string) (def.TID, string, error) {
	d, err := p.cls.Join(p.t, vhost)
	if err == nil {
		p.connCtx = d
	}

	// start checking heartbeat
	go func() {
		ctx := p.connCtx
		log.Get().WithField("ctx", ctx).Info("heartbeat run")
		defer func() {
			log.Get().WithField("ctx", ctx).Info("heartbeat stopped")
		}()

		for {
			d := p.connCtx
			if d == nil {
				return
			}
			<-time.After(time.Second * 10)
			now := time.Now()
			if now.Sub(d.lastHeartbeat) > time.Second*10 {
				p.Logout()
				return
			}
		}
	}()

	return d.tid, d.vhost, err
}

func (p *impl) Logout() error {
	ctx := p.connCtx
	if ctx == nil {
		return errors.New("you need login first")
	}

	log.Get().WithField("ctx", ctx).Info("logout")

	p.connCtx = nil
	p.cls.Detach(p.t)
	return nil
}

func (p *impl) DialByTID(tid def.TID, writeSID uint32, network, address string) (readSID uint32, err error) {
	target, err := p.cls.FindTunnelByID(tid)
	if err != nil {
		return 0, err
	}
	dialer := dial.NewClient(target, nil)

	// 第一个虚拟连接，用于访问目标站点
	conn, wSID := target.NewStream()
	rSID, err := dialer.Dial(wSID, network, address)
	conn.Bind(rSID)
	conn.SetInfo(address)

	// 第二个虚拟连接，用于连接代理服务器
	stream, sid := p.t.NewStream()
	stream.Bind(writeSID)
	stream.SetInfo(address)

	go utils.LinkReadWriteCloser(stream, conn)

	return sid, nil
}

func (p *impl) Dial(vhost string, vport uint32) (net.Conn, error) {
	tid, err := p.cls.Lookup(vhost)
	if err != nil {
		return nil, err
	}
	target, err := p.cls.FindTunnelByID(tid)
	if err != nil {
		return nil, err
	}
	return target.Dial(vport)
}

func (p *impl) SetLabel(label string) error {
	return errNotImplement
}

func (p *impl) CreateGroup(name, password, desc string, canBeSearch bool) error {
	return errNotImplement
}

func (p *impl) JoinGroup(groupID uint32, password string) error {
	return errNotImplement
}

func (p *impl) LeaveGroup(groupID uint32) error {
	return errNotImplement
}

// SendGroupMessage 处理客户端向某个群组发送消息的请求
func (p *impl) SendGroupMessage(groupID uint32, message string, msgType int) error {
	return p.mc.PushMessage(&msg{
		// 从可信区域获取数据
		SenderVhost: p.connCtx.vhost, // 直接从连接上下文中获取vhost信息
		Date:        time.Now(),

		// 登记客户端送上来的信息
		GroupID: groupID,
		Message: message,
		MsgType: msgType,
	})
}
