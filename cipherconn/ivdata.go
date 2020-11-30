package cipherconn

import (
	"crypto/rand"
	"io"
)

type ivdata struct {
	code  byte
	iv    []byte
	ivLen int
}

func newIvData() *ivdata {
	return &ivdata{
		code:  0x09,
		ivLen: 16,
	}
}

func (p *ivdata) Gen() error {
	p.iv = make([]byte, p.ivLen)
	_, err := rand.Read(p.iv)
	for i := 0; i < len(p.iv); i++ {
		p.iv[i] = 0
	}
	return err
}

func (p *ivdata) Verify() bool {
	return p.code == 0x09
}

func (p *ivdata) ReadFrom(r io.Reader) (readed int64, err error) {
	buf := make([]byte, 1+p.ivLen)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	p.code = buf[0]
	p.iv = buf[1 : 1+p.ivLen]

	return readed, nil
}

func (p *ivdata) WriteTo(w io.Writer) (written int64, err error) {
	buf := make([]byte, 1+p.ivLen)
	buf[0] = p.code
	copy(buf[1:1+p.ivLen], p.iv)
	wn, err := w.Write(buf)
	written = int64(wn)
	return written, err
}
