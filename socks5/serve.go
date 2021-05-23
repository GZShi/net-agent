package socks5

import (
	"net"

	"github.com/GZShi/net-agent/utils"
)

// Requester 解析客户端命令的函数
type Requester func(req Request, ctx map[string]string) (net.Conn, error)

// AuthPswdFunc 认证账号密码
type AuthPswdFunc func(username, password string) error

// Server Socks5服务
type Server interface {
	SetRequster(Requester)
	SetAuthChecker(AuthChecker)
	ListenAndRun(string) error
	Run(net.Listener) error
	Stop() error
}

type server struct {
	requester Requester
	checker   AuthChecker
	listener  net.Listener
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

func (s *server) Stop() error {
	return s.listener.Close()
}

// Run 将服务跑起来
func (s *server) Run(listener net.Listener) error {
	s.listener = listener
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

	ctx := make(map[string]string)

	//
	// 使用checker协议进行握手和身份校验
	//
	resp, next, err := s.checker.Start(conn, ctx)
	if err != nil {
		return err
	}
	_, err = conn.Write(resp)
	if err != nil {
		return err
	}

	for next {
		resp, next, err = s.checker.Next(conn, ctx)
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

	target, respErr := s.requester(&req, ctx)
	respCode := repSuccess
	if respErr != nil {
		respCode = repFailure
		switch respErr {
		case ErrReplyConnectionNotAllow:
			respCode = repConnectionNotAllow
		case ErrReplyNetworkUnRereachable:
			respCode = repNetworkUnRereachable
		case ErrReplyHostUnreachable:
			respCode = repHostUnreachable
		case ErrReplyConnectionRefused:
			respCode = repConnectionRefused
		case ErrReplyTTLExpired:
			respCode = repTTLExpired
		case ErrReplyCmdNotSupported:
			respCode = repCmdNotSupported
		case ErrReplyAtypeNotSupported:
			respCode = repAtypeNotSupported
		}
	}

	// 根据RFC1928，request与reply有相似结构
	var reply request
	reply.version = dataVersion
	reply.command = respCode // success
	reply.addressType = IPv4
	reply.addressBuf = make([]byte, net.IPv4len)
	reply.port = 0

	_, err = reply.WriteTo(conn)

	if respErr != nil {
		return respErr
	}

	if err != nil {
		return err
	}

	_, _, err = utils.LinkReadWriteCloser(conn, target)
	return err
}
