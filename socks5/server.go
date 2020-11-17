package socks5

import (
	"errors"
	"net"
	"strings"

	"github.com/GZShi/net-agent/transport"
)

// Dialer 拨号函数
type Dialer func(string, string, string, string) (net.Conn, error)
type BlockChecker func(string, string, string) error

// Server Socks5服务
type Server struct {
	secret string
	dialer Dialer
}

// NewSocks5Server 创建新的socks5协议服务端
func NewSocks5Server(secret string, dialer Dialer) *Server {
	return &Server{secret, dialer}
}

// Run 将服务跑起来
func (p *Server) Run(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go HandleSocks5Request(conn, conn, conn, p.secret, p.dialer)
	}
}

// ListenAndRun 监听并服务
func (p *Server) ListenAndRun(addr string) error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return p.Run(l)
}

func NewDefaultDialer() Dialer {
	return func(sourceAddr, network, targetAddr, clientName string) (net.Conn, error) {
		return nil, nil
	}
}

func ParseClientName(clientName string) (userName, channelName string, err error) {
	authInfos := strings.Split(clientName, "@")
	if len(authInfos) != 2 {
		return "", "", errors.New("bad client name")
	}
	userName = authInfos[0]
	channelName = authInfos[1]
	if len(channelName) < 3 {
		return "", "", errors.New("channel name is too short")
	}

	err = nil
	return
}

// NewTunnelClusterDialer 基于Tunnel Cluster创建网络连接
func NewTunnelClusterDialer(cluster *transport.TunnelCluster, checker BlockChecker) Dialer {
	return func(sourceAddr, network, targetAddr, clientName string) (net.Conn, error) {
		userName, channelName, err := ParseClientName(clientName)
		if err != nil {
			return nil, err
		}
		// 黑白名单访问限制
		if err := checker(network, targetAddr, channelName); err != nil {
			return nil, err
		}
		return cluster.Dial(sourceAddr, network, targetAddr, channelName, userName)
	}
}

// NewTunnelDialer 基于Tunnel创建网络连接
func NewTunnelDialer(t *transport.Tunnel, channelName, userName string) Dialer {
	return func(sourceAddr, network, targetAddr, clientName string) (net.Conn, error) {
		return t.Dial(sourceAddr, network, targetAddr, channelName, userName)
	}
}
