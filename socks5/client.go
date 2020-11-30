package socks5

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// Dial 通过代理创建连接
func Dial(proxyAddr, targetAddr string, auther Auth) (_conn net.Conn, _err error) {
	conn, err := net.Dial("tcp4", proxyAddr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if _err != nil {
			conn.Close()
		}
	}()

	if auther == nil {
		auther = NoAuth()
	}

	return upgrade(conn, targetAddr, auther)
}

// Upgrade 在已有连接上进行socks5协议协商
func Upgrade(conn net.Conn, targetAddr string, auther Auth) (net.Conn, error) {
	if auther == nil {
		auther = NoAuth()
	}
	return upgrade(conn, targetAddr, auther)
}

func upgrade(conn net.Conn, address string, auther Auth) (net.Conn, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid address")
	}
	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	wt, next, err := auther.Start()
	if err != nil {
		return nil, err
	}

	for next {
		_, err = wt.WriteTo(conn)
		if err != nil {
			return nil, err
		}

		wt, next, err = auther.Next(conn)
		if err != nil {
			return nil, err
		}
	}

	// send request
	req := &request{
		version:     dataVersion,
		command:     ConnectCommand,
		addressType: Domain,
		addressBuf:  []byte(host),
		port:        uint16(port),
	}

	_, err = req.WriteTo(conn)
	if err != nil {
		return nil, err
	}

	var reply request
	_, err = reply.ReadFrom(conn)
	if err != nil {
		return nil, err
	}
	if reply.command != 0 {
		return nil, errors.New("dial failed")
	}

	return conn, nil
}
