package socks5

import (
	"net"
)

// Requester 解析客户端命令的函数
type Requester func(Request) (net.Conn, error)

// AuthPswdFunc 认证账号密码
type AuthPswdFunc func(username, password string) error

// Server Socks5服务
type Server interface {
	SetRequster(Requester)
	SetAuthChecker(AuthChecker)
	ListenAndRun(string) error
	Run(net.Listener) error
}

type server struct {
	requester Requester
	checker   AuthChecker
}

// NewServer 创建新的socks5协议服务端
func NewServer() Server {
	return &server{
		requester: DefaultRequester,
		checker:   DefaultAuthChecker(),
	}
}

func (s *server) SetRequster(in Requester) {
	s.requester = in
}
func (s *server) SetAuthChecker(in AuthChecker) {
	s.checker = in
}

// Run 将服务跑起来
func (s *server) Run(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			s.serve(conn)
		}()
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

	//
	// 使用checker协议进行握手和身份校验
	//
	resp, next, err := s.checker.Start(conn)
	if err != nil {
		return err
	}
	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	for next {
		resp, next, err = s.checker.Next(conn)
		if err != nil {
			return err
		}
		_, err = conn.Write(resp)
	}

	//
	// 执行命令
	//
	var req request
	_, err = req.ReadFrom(conn)
	if err != nil {
		return err
	}

	target, err := s.requester(&req)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte{dataVersion, repSuccess,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
	if err != nil {
		return err
	}

	_, _, err = Link(conn, target)
	return err
}
