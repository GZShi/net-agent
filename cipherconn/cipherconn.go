package cipherconn

import (
	"crypto/cipher"
	"net"
)

type cipherconn struct {
	net.Conn
	encoder cipher.Stream
	decoder cipher.Stream
}

func (conn *cipherconn) Write(b []byte) (int, error) {
	buf := make([]byte, len(b))
	conn.encoder.XORKeyStream(buf, b)
	return conn.Conn.Write(buf)
}

func (conn *cipherconn) Read(b []byte) (rn int, err error) {
	defer func() {
		if rn > 0 {
			conn.decoder.XORKeyStream(b[0:rn], b[0:rn])
		}
	}()

	return conn.Conn.Read(b)
}
