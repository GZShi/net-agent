package socks5

import (
	"bytes"
	"errors"
	"io"
)

// AuthChecker 进行身份校验的状态机
type AuthChecker interface {
	Start(r io.Reader, ctx map[string]string) (respData []byte, hasNext bool, err error)
	Next(r io.Reader, ctx map[string]string) (respData []byte, hasNext bool, err error)
}

// DefaultAuthChecker 缺省身份校验（无校验）
func DefaultAuthChecker() AuthChecker {
	return NoAuthChecker()
}

func checkHandshakeMethods(r io.Reader, want byte) error {
	var hs handshakeData
	_, err := hs.ReadFrom(r)
	if err != nil {
		return err
	}
	if bytes.IndexByte(hs.methods, want) < 0 {
		return ErrMethodsNotSupport
	}
	return nil
}

// NoAuthChecker 无校验
func NoAuthChecker() AuthChecker {
	return &noAuthChecker{}
}

type noAuthChecker struct{}

func (checker *noAuthChecker) Start(r io.Reader, ctx map[string]string) ([]byte, bool, error) {
	if err := checkHandshakeMethods(r, MethodNoAuth); err != nil {
		return []byte{dataVersion, MethodNoAcceptable}, false, err
	}
	return []byte{dataVersion, MethodNoAuth}, false, nil
}

func (checker *noAuthChecker) Next(r io.Reader, ctx map[string]string) ([]byte, bool, error) {
	return nil, false, errors.New("unexpected data")
}

// PswdAuthChecker 用户名密码校验
func PswdAuthChecker(checkPswd func(string, string, map[string]string) error) AuthChecker {
	return &pswdAuthChecker{checkPswd}
}

type pswdAuthChecker struct {
	checkPswdFunc func(string, string, map[string]string) error
}

func (checker *pswdAuthChecker) Start(r io.Reader, ctx map[string]string) ([]byte, bool, error) {
	if err := checkHandshakeMethods(r, MethodAuthPswd); err != nil {
		return []byte{dataVersion, MethodNoAcceptable}, false, err
	}
	return []byte{dataVersion, MethodAuthPswd}, true, nil
}

func (checker *pswdAuthChecker) Next(r io.Reader, ctx map[string]string) ([]byte, bool, error) {
	resp := []byte{0x01, 0x01}

	buf := make([]byte, 2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return resp, false, err
	}
	if buf[0] != 0x01 {
		return resp, false, errors.New("unknown protocol version")
	}

	buf2 := make([]byte, buf[1]+1)
	_, err = io.ReadFull(r, buf2)
	if err != nil {
		return resp, false, err
	}
	username := string(buf2[:len(buf2)-1])

	buf3 := make([]byte, buf2[len(buf2)-1])
	_, err = io.ReadFull(r, buf3)
	if err != nil {
		return resp, false, err
	}
	password := string(buf3)

	err = checker.checkPswdFunc(username, password, ctx)
	if err != nil {
		return resp, false, err
	}

	resp[1] = 0x00
	return resp, false, nil
}
