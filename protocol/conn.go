package protocol

import (
	"bufio"
	"io"
	"net"
	"time"
)

const (
	ProtoUnknown = iota
	ProtoHTTP
	ProtoSocks5
	ProtoAgentClient
	ProtoShadowsocks
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
	// 	conn.protocol = ProtoUnknown
	// 	return conn
	// }
	// err = conn.Reader.UnreadByte()
	// if err != nil {
	// 	conn.protocol = ProtoUnknown
	// }
	headBytes, err := conn.Reader.Peek(3)
	if err != nil && err != bufio.ErrBufferFull {
		conn.protocol = ProtoUnknown
		return conn
	}

	b := headBytes[0]

	switch b {
	case 0x05:
		conn.protocol = ProtoSocks5
	case 0x01, 0x03, 0x04:
		conn.protocol = ProtoShadowsocks
	default:
		if b >= 'a' && b <= 'z' {
			conn.protocol = ProtoAgentClient
		} else if b >= 'A' && b <= 'Z' {
			conn.protocol = ProtoHTTP
		} else {
			conn.protocol = ProtoUnknown
		}
	}
	return conn
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
