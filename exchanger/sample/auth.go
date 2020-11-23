package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

type authRequest struct {
	version    byte
	nonce      [16]byte
	hashedSize uint16
	hashed     []byte
}

func newAuthRequest(secret string) *authRequest {
	return &authRequest{
		version: 0x09,
	}
}

func (p *authRequest) Gen(secret string) error {
	_, err := rand.Read(p.nonce[:])
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write(p.nonce[:])
	h.Write([]byte(secret))
	p.hashed = h.Sum(nil)
	p.hashedSize = uint16(len(p.hashed))

	return nil
}

func (p *authRequest) Verify(secret string) bool {
	h := sha256.New()
	h.Write(p.nonce[:])
	h.Write([]byte(secret))
	hashed := h.Sum(nil)
	if bytes.Equal(hashed, p.hashed) {
		return true
	}
	return false
}

func (p *authRequest) ReadFrom(r io.Reader) (readed int64, err error) {
	buf := make([]byte, 1+16+16)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	p.version = buf[0]
	copy(p.nonce[:], buf[1:1+16])
	copy(p.hashed[:], buf[1+16:])

	return readed, nil
}

func (p *authRequest) WriteTo(w io.Writer) (written int64, err error) {
	buf := make([]byte, 1+16+16)
	buf[0] = p.version
	copy(buf[1:1+16], p.nonce[:])
	copy(buf[1+16:], p.hashed[:])
	wn, err := w.Write(buf)
	written = int64(wn)
	return written, err
}

//
// auth response
//
type authResponse struct {
	code      byte
	challenge [16]byte
}

func newAuthResponse() *authResponse {
	return &authResponse{}
}

func (p *authResponse) Gen(code byte) error {
	p.code = code
	if code == 0 {
		_, err := rand.Read(p.challenge[:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *authResponse) ReadFrom(r io.Reader) (readed int64, err error) {
	buf := make([]byte, 1+16)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	p.code = buf[0]
	copy(p.challenge[:], buf[1:1+16])

	return readed, nil
}

func (p *authResponse) WriteTo(w io.Writer) (written int64, err error) {
	buf := make([]byte, 1+16)
	buf[0] = p.code
	copy(buf[1:1+16], p.challenge[:])
	wn, err := w.Write(buf)
	written = int64(wn)
	return written, err
}
