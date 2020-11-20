package socks5

import (
	"errors"
	"io"
)

const (
	VersionSocks5 = uint8(5)
)

var (
	VersionNotSupport  = errors.New("socks version not supported")
	MethodsIsEmpty     = errors.New("socks methods is empty")
	MethodsNotSupport  = errors.New("socks methods not supported")
	MethodsSizeIllegal = errors.New("socks methos size illegal")
)

type handshakeData struct {
	version uint8
	methods []uint8
}

func (s *handshakeData) ReadFrom(r io.Reader) (readed int64, err error) {
	buf := make([]byte, 2)
	rn, err := io.ReadFull(r, buf)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	if buf[0] != VersionSocks5 {
		return readed, VersionNotSupport
	}
	s.version = buf[0]

	if buf[1] == 0 {
		return readed, MethodsIsEmpty
	}
	s.methods = make([]byte, buf[1])
	rn, err = io.ReadFull(r, s.methods)
	readed += int64(rn)

	return readed, err
}

func (s *handshakeData) WriteTo(w io.Writer) (written int64, err error) {
	if len(s.methods) > 255 {
		return 0, MethodsSizeIllegal
	}
	wn, err := w.Write([]byte{s.version, uint8(len(s.methods))})
	written += int64(wn)
	if err != nil {
		return written, err
	}

	wn, err = w.Write(s.methods)
	written += int64(wn)

	return written, err
}
