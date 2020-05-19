package transport

import (
	"io"
	"net"

	"github.com/golang/snappy"
)

// CompressConn 能够压缩传输的连接封装
type CompressConn struct {
	net.Conn
	r io.Reader
	w io.Writer
}

func (c *CompressConn) Read(b []byte) (rn int, err error) {
	return c.r.Read(b)
}

func (c *CompressConn) Write(b []byte) (wn int, err error) {
	return c.w.Write(b)
}

// NewCompressConn 封装的压缩连接
func NewCompressConn(conn net.Conn) net.Conn {
	cc := &CompressConn{
		Conn: conn,
		w:    snappy.NewWriter(io.Writer(conn)),
		r:    snappy.NewReader(io.Reader(conn)),
	}

	return cc
}
