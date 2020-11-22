package socks5

import (
	"errors"
	"io"
)

var (
	// ErrAuthPswdFailed 错误：校验用户名密码失败
	ErrAuthPswdFailed = errors.New("socks auth password failed")
)

// Auth 认证接口
type Auth interface {
	Start() io.WriterTo
	Next(fromServer []byte) (io.WriterTo, error)
}

// NoAuth 直接连接的方式
func NoAuth() Auth {
	return &noAuth{}
}

type noAuth struct{}

func (auth *noAuth) Start() io.WriterTo {
	return &handshakeData{
		version: VersionSocks5,
		methods: []byte{MethodNoAuth},
	}
}

func (auth *noAuth) Next(fromServer []byte) (io.WriterTo, error) {
	if fromServer[1] != MethodNoAuth {
		return nil, ErrMethodsNotSupport
	}

	return nil, nil
}

// AuthPswd 基础的用户名密码校验
func AuthPswd(username, password string) Auth {
	return &authPswd{username, password, 0}
}

type authPswd struct {
	username string
	password string
	round    int
}

func (auth *authPswd) Start() io.WriterTo {
	return &handshakeData{
		version: VersionSocks5,
		methods: []byte{MethodAuthPswd},
	}
}
func (auth *authPswd) Next(fromServer []byte) (io.WriterTo, error) {
	r := auth.round
	auth.round++
	switch r {
	case 0:
		return auth, nil
	case 1:
		if fromServer[0] != 0x01 || fromServer[1] != 0x00 {
			return nil, ErrAuthPswdFailed
		}
		return nil, nil
	default:
		return nil, errors.New("unexpected round")
	}
}

func (auth *authPswd) WriteTo(w io.Writer) (int64, error) {
	posPswd := 1 + len(auth.username)
	buf := make([]byte, posPswd+1+len(auth.password))
	buf[0] = dataVerPswd
	buf[1] = byte(len(auth.username))
	end := 2 + len(auth.username)
	copy(buf[2:end], auth.username)
	buf[end] = byte(len(auth.password))
	copy(buf[end+1:], auth.password)

	written, err := w.Write(buf)
	return int64(written), err
}
