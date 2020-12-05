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
	Start() (wt io.WriterTo, hasNext bool, err error)
	Next(r io.Reader) (wt io.WriterTo, hasNext bool, err error)
}

// NoAuth 直接连接的方式
func NoAuth() Auth {
	return &noAuth{
		round: 0,
	}
}

type noAuth struct {
	round int
}

func (auth *noAuth) Start() (io.WriterTo, bool, error) {
	return &handshakeData{
		version: VersionSocks5,
		methods: []byte{MethodNoAuth},
	}, true, nil
}

func (auth *noAuth) Next(reader io.Reader) (io.WriterTo, bool, error) {
	if auth.round != 0 {
		return nil, false, errors.New("unexpected round")
	}
	auth.round++

	buf := make([]byte, 2)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, false, err
	}
	if buf[0] != dataVersion || buf[1] != MethodNoAuth {
		return nil, false, errors.New("bad handshake response")
	}
	return nil, false, nil
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

func (auth *authPswd) Start() (io.WriterTo, bool, error) {
	return &handshakeData{
		version: VersionSocks5,
		methods: []byte{MethodAuthPswd},
	}, true, nil
}

func (auth *authPswd) Next(reader io.Reader) (io.WriterTo, bool, error) {
	var buf []byte
	r := auth.round
	auth.round++
	switch r {
	case 0:
		buf = make([]byte, 2)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, false, err
		}
		return auth, true, nil
	case 1:
		buf = make([]byte, 2)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return nil, false, err
		}
		if buf[0] != 0x01 || buf[1] != 0x00 {
			return nil, false, ErrAuthPswdFailed
		}
		return nil, false, nil
	default:
		return nil, false, errors.New("unexpected round")
	}
}

func (auth *authPswd) WriteTo(w io.Writer) (int64, error) {

	// https://tools.ietf.org/html/rfc1929
	//
	// +----+------+----------+------+----------+
	// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
	// +----+------+----------+------+----------+
	// | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
	// +----+------+----------+------+----------+

	bufLen := 1 + 1 + len(auth.username) + 1 + len(auth.password)

	buf := make([]byte, bufLen)

	// VER
	buf[0] = dataVerPswd

	// ULEN
	buf[1] = byte(len(auth.username))

	// UNAME
	end := 2 + len(auth.username)
	copy(buf[2:end], auth.username)

	// PLEN
	buf[end] = byte(len(auth.password))

	// PASSWD
	copy(buf[end+1:], auth.password)

	written, err := w.Write(buf)
	return int64(written), err
}
