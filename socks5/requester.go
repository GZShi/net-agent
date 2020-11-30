package socks5

import (
	"errors"
	"net"
)

const (
	ConnectCommand = uint8(0x01)
	BindCommand    = uint8(0x02)
	UDPCommand     = uint8(0x03)
)

var (
	// ErrCommandNotSupport ...
	ErrCommandNotSupport = errors.New("socks5 command not supported")
)

// DefaultRequester 执行net.Dial创建连接，并将两个net.Conn进行连接
func DefaultRequester(req Request) (net.Conn, error) {
	if req.GetCommand() != ConnectCommand {
		return nil, ErrCommandNotSupport
	}
	return net.Dial("tcp4", req.GetAddrPortStr())
}
