package cipherconn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/hkdf"
)

// New 根据已有连接，构建加密连接
func New(conn net.Conn, password string) (net.Conn, error) {
	// return conn, nil

	cc := &cipherconn{
		Conn: conn,
	}

	var err error
	cc.encoder, cc.decoder, err = confer(conn, password)
	return cc, err
}

// confer 协商加密的iv
func confer(wr io.ReadWriter, password string) (enc cipher.Stream, dec cipher.Stream, err error) {
	req := newIvData()
	resp := newIvData()
	var recvErr, sendErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		recvErr = recvIV(wr, password, req)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		sendErr = sendIV(wr, password, resp)
		wg.Done()
	}()

	wg.Wait()

	if recvErr != nil {
		return nil, nil, recvErr
	}
	if sendErr != nil {
		return nil, nil, sendErr
	}

	if enc, err = makeCipherStream(password, resp.iv); err != nil {
		return nil, nil, err
	}
	if dec, err = makeCipherStream(password, req.iv); err != nil {
		return nil, nil, err
	}

	return enc, dec, nil
}

func recvIV(r io.Reader, password string, data *ivdata) error {

	if _, err := data.ReadFrom(r); err != nil {
		return err
	}
	if !data.Verify(password) {
		return errors.New("verify failed")
	}

	return nil
}

func sendIV(w io.Writer, password string, data *ivdata) error {

	if err := data.Gen(password); err != nil {
		return err
	}
	if _, err := data.WriteTo(w); err != nil {
		return err
	}
	return nil
}

func makeCipherStream(password string, iv []byte) (cipher.Stream, error) {
	// log.Get().WithField("pswd", password).WithField("iv", iv).Debug("create cipher stream")
	key := make([]byte, 16)
	if err := hkdfSha1([]byte(password), key); err != nil {
		return nil, err
	}
	bc, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewCTR(bc, iv), nil
}

func hkdfSha1(secret, outbuf []byte) error {
	r := hkdf.New(sha1.New, secret, []byte("cipherconn-of-exchanger"), nil)
	if _, err := io.ReadFull(r, outbuf); err != nil {
		return err
	}
	return nil
}
