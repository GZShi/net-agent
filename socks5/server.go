package socks5

import "net"

// Dialer 拨号函数
type Dialer func(string, string, string, string) (net.Conn, error)

// Server Socks5服务
type Server struct {
	secret string
	dialer Dialer
}

// NewSocks5Server 创建新的socks5协议服务端
func NewSocks5Server(secret string, dialer Dialer) *Server {
	return &Server{secret, dialer}
}

// Run 将服务跑起来
func (p *Server) Run(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		go ServeSocks5(conn, conn, conn, p.secret, p.dialer)
	}
}
