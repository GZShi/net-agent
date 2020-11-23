package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"io"
	"net"

	"golang.org/x/crypto/hkdf"
)

type cipherconn struct {
	net.Conn
	encoder cipher.Stream
	decoder cipher.Stream
}

// CipherConn 根据已有的连接，构建加密传输通道
func CipherConn(conn net.Conn, password string, isClient bool) (net.Conn, error) {
	var err error
	cc := &cipherconn{
		Conn: conn,
	}
	if isClient {
		cc.encoder, cc.decoder, err = clientSideConfer(conn, password)
	} else {
		cc.encoder, cc.encoder, err = serverSideConfer(conn, password)
	}

	if err != nil {
		return nil, err
	}
	return cc, nil
}

//
// help function
//
const (
	stepRead = iota
	stepWrite
)

func clientSideConfer(wr io.ReadWriter, password string) (enc cipher.Stream, dec cipher.Stream, err error) {
	return confer(wr, password, []byte{stepWrite, stepRead})
}

func serverSideConfer(wr io.ReadWriter, password string) (enc cipher.Stream, dec cipher.Stream, err error) {
	return confer(wr, password, []byte{stepRead, stepWrite})
}

func confer(wr io.ReadWriter, password string, steps []byte) (enc cipher.Stream, dec cipher.Stream, err error) {
	var req, resp ivdata
	for _, step := range steps {
		switch step {
		case stepRead:
			if _, err = req.ReadFrom(wr); err != nil {
				return nil, nil, err
			}
			if !req.Verify() {
				return nil, nil, err
			}
		case stepWrite:
			if err = resp.Gen(); err != nil {
				return nil, nil, err
			}
			if _, err = resp.WriteTo(wr); err != nil {
				return nil, nil, err
			}
		}
	}
	if enc, err = makeCipherStream(password, &resp); err != nil {
		return nil, nil, err
	}
	if dec, err = makeCipherStream(password, &req); err != nil {
		return nil, nil, err
	}

	return enc, dec, nil
}

func makeCipherStream(password string, data *ivdata) (cipher.Stream, error) {
	key := make([]byte, 16)
	if err := hkdfSha1([]byte(password), key); err != nil {
		return nil, err
	}
	bc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(bc, data.iv), nil
}

func hkdfSha1(secret, outbuf []byte) error {
	r := hkdf.New(sha1.New, secret, []byte("cipherconn-of-exchanger"), nil)
	if _, err := io.ReadFull(r, outbuf); err != nil {
		return err
	}
	return nil
}
