package service

import (
	"errors"
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/dial"
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
	}
}

type tunData struct {
	t             tunnel.Tunnel
	tid           def.TID
	vhost         string
	lastHeartbeat time.Time
}
type impl struct {
	t   tunnel.Tunnel
	cls *cluster

	// cache info
	cache *tunData
}

func (p *impl) Heartbeat() error {
	if p.cache != nil {
		p.cache.lastHeartbeat = time.Now()
		return nil
	}
	return errors.New("cache tdata not found")
}

func (p *impl) Login(vhost string) (def.TID, string, error) {
	d, err := p.cls.Join(p.t, vhost)
	if err == nil {
		p.cache = d
	}

	// start checking heartbeat
	go func() {
		tdata := p.cache
		log.Get().WithField("tdata", tdata).Info("heartbeat run")
		defer func() {
			log.Get().WithField("tdata", tdata).Info("heartbeat stopped")
		}()

		for {
			d := p.cache
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
	tdata := p.cache
	if tdata == nil {
		return errors.New("you need login first")
	}

	log.Get().WithField("tdata", tdata).Info("logout")

	p.cache = nil
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

func (p *impl) SendGroupMessage(groupID uint32, message string, msgType int) error {
	return errNotImplement
}
