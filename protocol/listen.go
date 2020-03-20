package protocol

import (
	"errors"
	"net"
	"strings"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/sirupsen/logrus"
)

// ProtoListener 指定协议监听器
type ProtoListener struct {
	manager *ProtoManager
	proto   int
}

// Accept ...
func (p *ProtoListener) Accept() (net.Conn, error) {
	return p.manager.accept(p.proto)
}

// Close ...
func (p *ProtoListener) Close() error {
	return p.manager.close(p.proto)
}

// Addr ...
func (p *ProtoListener) Addr() net.Addr {
	return p.manager.raw.Addr()
}

// ProtoManager 混合协议监听器
type ProtoManager struct {
	raw    net.Listener
	initWg sync.WaitGroup

	httpConns   chan *Conn
	socks5Conns chan *Conn
	agentConns  chan *Conn
}

// NewListener 新的监听器
func NewListener(network, addr string) (*ProtoManager, error) {
	raw, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	p := &ProtoManager{
		raw:         raw,
		httpConns:   make(chan *Conn, 256),
		socks5Conns: make(chan *Conn, 256),
		agentConns:  make(chan *Conn, 256),
	}
	p.initWg.Add(1)

	go func() {
		rAddrMap := sync.Map{}
		log.Get().WithField("addr", addr).Info("server is running")
		p.initWg.Done()
		for {
			conn, err := raw.Accept()
			if err != nil {
				break
			}
			go func(c net.Conn) {
				remoteAddr := c.RemoteAddr().String()
				ip := remoteAddr
				if strings.HasPrefix(remoteAddr, "[") {
					// ipv6
					ip = strings.Split(remoteAddr[1:], "]:")[0]
				} else if len(remoteAddr) > 0 {
					// ipv4
					ip = strings.Split(remoteAddr, ":")[0]
				}
				_, isLoad := rAddrMap.LoadOrStore(ip, 1)
				if !isLoad {
					log.Get().WithField("raddr", remoteAddr).Info("new remote addr")
				}

				protoConn := NewConn(c)
				switch {
				case protoConn.IsHTTP():
					p.httpConns <- protoConn
				case protoConn.IsSocks5():
					p.socks5Conns <- protoConn
				case protoConn.IsAgent():
					p.agentConns <- protoConn
				default:
					lineData, _, err := protoConn.Reader.ReadLine()
					log.Get().WithError(err).WithFields(logrus.Fields{
						"data":  lineData,
						"raddr": remoteAddr,
					}).Warn("bad protocol")
					protoConn.Close()
				}
			}(conn)
		}
		close(p.httpConns)
		close(p.socks5Conns)
		close(p.agentConns)
		raw.Close()
	}()

	return p, nil
}

// GetHTTPListener 获取http协议监听器
func (p *ProtoManager) GetHTTPListener() *ProtoListener {
	p.initWg.Wait()
	return &ProtoListener{
		manager: p,
		proto:   protoHTTP,
	}
}

// GetSocks5Listener 获取socks5协议监听器
func (p *ProtoManager) GetSocks5Listener() *ProtoListener {
	p.initWg.Wait()
	return &ProtoListener{
		manager: p,
		proto:   protoSocks5,
	}
}

// GetAgentListener 获取agent协议监听器
func (p *ProtoManager) GetAgentListener() *ProtoListener {
	p.initWg.Wait()
	return &ProtoListener{
		manager: p,
		proto:   protoAgentClient,
	}
}

func (p *ProtoManager) accept(proto int) (net.Conn, error) {
	var conn net.Conn
	switch proto {
	case protoHTTP:
		conn = <-p.httpConns
	case protoSocks5:
		conn = <-p.socks5Conns
	case protoAgentClient:
		conn = <-p.agentConns
	default:
		conn = nil
	}

	if conn == nil {
		return nil, errors.New("channel closed")
	}
	return conn, nil
}

func (p *ProtoManager) close(proto int) error {
	var ch chan *Conn
	switch proto {
	case protoHTTP:
		ch = p.httpConns
	case protoSocks5:
		ch = p.socks5Conns
	case protoAgentClient:
		ch = p.agentConns
	default:
		ch = nil
	}

	if ch == nil {
		return errors.New("close failed, proto not found")
	}

	for {
		select {
		case data := <-ch:
			if data == nil {
				return errors.New("channel closed")
			}
			close(ch)
			return nil
		default:
			close(ch)
			return nil
		}
	}
}
