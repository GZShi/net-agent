package cipherconn

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"io"
)

const (
	checksumLen = 16
)

type ivdata struct {
	code     byte
	iv       []byte
	ivLen    int
	checksum []byte
}

func newIvData() *ivdata {
	return &ivdata{
		code:  0x09,
		ivLen: 16,
	}
}

func (p *ivdata) Gen(password string) error {
	p.code = 0x09
	p.iv = make([]byte, p.ivLen)
	_, err := rand.Read(p.iv)

	h := md5.New()
	h.Write(p.iv)
	h.Write([]byte(password))
	checksum := h.Sum(nil)[0:checksumLen]
	if len(checksum) < checksumLen {
		return errors.New("length of checksum buf is too long")
	}
	p.checksum = checksum[0:checksumLen]
	return err
}

func (p *ivdata) Verify(password string) bool {
	h := md5.New()
	h.Write(p.iv)
	h.Write([]byte(password))
	checksum := h.Sum(nil)
	if len(checksum) < checksumLen {
		return false
	}
	checksum = checksum[0:checksumLen]
	if !bytes.Equal(checksum, p.checksum) {
		return false
	}
	return p.code == 0x09
}

func (p *ivdata) ReadFrom(r io.Reader) (readed int64, err error) {
	buf := make([]byte, 1+p.ivLen+checksumLen)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	p.code = buf[0]
	p.iv = buf[1 : 1+p.ivLen]
	p.checksum = buf[1+p.ivLen:]

	return readed, nil
}

func (p *ivdata) WriteTo(w io.Writer) (written int64, err error) {
	buf := make([]byte, 1+p.ivLen+checksumLen)
	buf[0] = p.code
	copy(buf[1:1+p.ivLen], p.iv)
	copy(buf[1+p.ivLen:], p.checksum)

	wn, err := w.Write(buf)
	written = int64(wn)
	return written, err
}
