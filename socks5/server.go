package socks5

import (
	"io"
	"net"
)

// DialFunc 拨号函数
type DialFunc func(network string, address string) (io.ReadWriteCloser, error)

// AuthPswdFunc 认证账号密码
type AuthPswdFunc func(username, password string) error

// Server Socks5服务
type Server interface {
	SetDialFunc(DialFunc)
	EnableNoAuth()
	EnableAuthPswd(AuthPswdFunc)
	ListenAndRun(string) error
	Run(net.Listener) error
}

type server struct {
	secret     string
	dialFn     DialFunc
	authPswdFn AuthPswdFunc
}

// NewServer 创建新的socks5协议服务端
func NewServer() Server {
	return &server{}
}

// Run 将服务跑起来
func (s *server) Run(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go s.serve(conn)
	}
}

// ListenAndRun 监听并服务
func (s *server) ListenAndRun(addr string) error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return s.Run(l)
}

func (s *server) serve(conn net.Conn) error {
	defer conn.Close()

	var hs handshakeData
	_, err := hs.ReadFrom(conn)
	if err != nil {
		return err
	}

	// todo: auth check

	var req request
	_, err := req.ReadFrom(conn)
	if err != nil {
		return err
	}
}
