package socks5

import (
	"errors"
	"net"
)

// ProxyChain 代理链条
type ProxyChain []*ProxyInfo

// Dial 通过代理链创建连接
func (chain *ProxyChain) Dial(targetAddr string) (net.Conn, error) {
	if len(*chain) < 1 {
		return nil, errors.New("proxychain is empty")
	}

	proxy := (*chain)[0]
	conn, err := net.Dial(proxy.Network, proxy.Address)
	if err != nil {
		return nil, err
	}

	return chain.Upgrade(conn, targetAddr)
}

// Upgrade 向已有链接发送代理链Request
func (chain *ProxyChain) Upgrade(conn net.Conn, targetAddr string) (net.Conn, error) {
	if len(*chain) < 1 {
		return nil, errors.New("proxychain is empty")
	}

	var next *ProxyInfo
	var err error

	proxy := (*chain)[0]
	for i := 1; i < len(*chain); i++ {
		next = (*chain)[i]
		conn, err = proxy.Upgrade(conn, next.Address)
		if err != nil {
			conn.Close()
			return nil, err
		}
		proxy = next
	}

	return proxy.Upgrade(conn, targetAddr)
}
