package protocol

import (
	"bufio"
	"io"
	"net"
	"time"
)

const (
	protoUnknown = iota
	protoHTTP
	protoSocks5
	protoAgentClient
	protoShadowsocks
)

// Conn 支持协议侦测的连接
type Conn struct {
	rawConn net.Conn
	Reader  *bufio.Reader
	Writer  io.Writer

	protocol int
}

// NewConn 封装新的连接
func NewConn(raw net.Conn) *Conn {
	conn := &Conn{
		rawConn: raw,
		Reader:  bufio.NewReader(raw),
		Writer:  raw,
	}

	// b, err := conn.Reader.ReadByte()
	// if err != nil {
	// 	conn.protocol = protoUnknown
	// 	return conn
	// }
	// err = conn.Reader.UnreadByte()
	// if err != nil {
	// 	conn.protocol = protoUnknown
	// }
	headBytes, err := conn.Reader.Peek(3)
	if err != nil && err != bufio.ErrBufferFull {
		conn.protocol = protoUnknown
		return conn
	}

	b := headBytes[0]

	switch b {
	case 0x05:
		conn.protocol = protoSocks5
	case 0x01, 0x03, 0x04:
		conn.protocol = protoShadowsocks
	default:
		if b >= 'a' && b <= 'z' {
			conn.protocol = protoAgentClient
		} else if b >= 'A' && b <= 'Z' {
			conn.protocol = protoHTTP
		} else {
			conn.protocol = protoUnknown
		}
	}
	return conn
}

// IsHTTP 是不是http协议
func (p *Conn) IsHTTP() bool {
	return p.protocol == protoHTTP
}

// IsSocks5 是不是socks5协议
func (p *Conn) IsSocks5() bool {
	return p.protocol == protoSocks5
}

// IsAgent 是不是agent协议
func (p *Conn) IsAgent() bool {
	return p.protocol == protoAgentClient
}

// IsSS 是不是Shadowsocks协议
func (p *Conn) IsSS() bool {
	return p.protocol == protoShadowsocks
}

func (p *Conn) Read(b []byte) (int, error) {
	return p.Reader.Read(b)
}

func (p *Conn) Write(b []byte) (int, error) {
	return p.Writer.Write(b)
}

// Close 关闭连接
func (p *Conn) Close() error {
	return p.rawConn.Close()
}

// LocalAddr 获取连接的本地地址
func (p *Conn) LocalAddr() net.Addr {
	return p.rawConn.LocalAddr()
}

// RemoteAddr 获取连接的远端地址
func (p *Conn) RemoteAddr() net.Addr {
	return p.rawConn.RemoteAddr()
}

// SetDeadline ...
func (p *Conn) SetDeadline(t time.Time) error {
	return p.rawConn.SetDeadline(t)
}

// SetReadDeadline ...
func (p *Conn) SetReadDeadline(t time.Time) error {
	return p.rawConn.SetReadDeadline(t)
}

// SetWriteDeadline ...
func (p *Conn) SetWriteDeadline(t time.Time) error {
	return p.rawConn.SetWriteDeadline(t)
}
