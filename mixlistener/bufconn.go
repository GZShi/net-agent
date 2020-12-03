package mixlistener

import (
	"bufio"
	"net"
)

type bufconn struct {
	net.Conn
	Reader *bufio.Reader
}

// newBufconn 创建带读缓存的连接
func newBufconn(raw net.Conn) *bufconn {
	return &bufconn{
		Conn:   raw,
		Reader: bufio.NewReader(raw),
	}
}

func (conn *bufconn) Read(b []byte) (int, error) {
	return conn.Reader.Read(b)
}
