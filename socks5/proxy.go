package socks5

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ProxyInfo 代理的信息
type ProxyInfo struct {
	Network  string
	Address  string
	NeedAuth bool
	Username string
	Password string
}

// Dial 通过代理服务连接目标地址
func (p *ProxyInfo) Dial(targetAddr string) (retConn net.Conn, retErr error) {
	conn, err := net.Dial(p.Network, p.Address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			conn.Close()
		}
	}()

	return p.Upgrade(conn, targetAddr)
}

// Upgrade 在已有连接上发出request请求
func (p *ProxyInfo) Upgrade(conn net.Conn, targetAddr string) (retConn net.Conn, retErr error) {
	defer func() {
		if retErr != nil {
			conn.Close()
		}
	}()

	// 构建验证信息
	var auther Auth
	if p.NeedAuth {
		auther = AuthPswd(p.Username, p.Password)
	} else {
		auther = NoAuth()
	}

	// 分离host和port
	parts := strings.Split(targetAddr, ":")
	if len(parts) != 2 {
		return nil, errors.New("missing port in address")
	}
	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	// 开始认证信息握手
	wt, next, err := auther.Start()
	if err != nil {
		return nil, err
	}

	for next {
		_, err = wt.WriteTo(conn)
		if err != nil {
			return nil, err
		}

		wt, next, err = auther.Next(conn)
		if err != nil {
			return nil, err
		}
	}

	// send request
	req := &request{
		version:     dataVersion,
		command:     ConnectCommand,
		addressType: Domain,
		addressBuf:  []byte(host),
		port:        uint16(port),
	}

	_, err = req.WriteTo(conn)
	if err != nil {
		return nil, err
	}

	var reply request
	_, err = reply.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	switch reply.command {
	case repSuccess:
		return conn, nil
	case repFailure:
		return nil, ErrReplyFailure
	case repConnectionNotAllow:
		return nil, ErrReplyConnectionNotAllow
	case repNetworkUnRereachable:
		return nil, ErrReplyNetworkUnRereachable
	case repHostUnreachable:
		return nil, ErrReplyHostUnreachable
	case repConnectionRefused:
		return nil, ErrReplyConnectionRefused
	case repTTLExpired:
		return nil, ErrReplyTTLExpired
	case repCmdNotSupported:
		return nil, ErrReplyCmdNotSupported
	case repAtypeNotSupported:
		return nil, ErrReplyAtypeNotSupported
	default:
		return nil, fmt.Errorf("connect failed with code: %v", reply.command)
	}
}
