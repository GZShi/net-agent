package socks5

import (
	"net"
)

const (
	ConnectCommand = uint8(0x01)
	BindCommand    = uint8(0x02)
	UDPCommand     = uint8(0x03)
)

// DefaultRequester 执行net.Dial创建连接，并将两个net.Conn进行连接
func DefaultRequester(req Request, ctx map[string]string) (net.Conn, error) {
	if req.GetCommand() != ConnectCommand {
		return nil, ReplyErrCmdNotSupported
	}
	return net.Dial("tcp4", req.GetAddrPortStr())
}
