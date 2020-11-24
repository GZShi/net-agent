package main

import (
	"net"

	"github.com/GZShi/net-agent/tunnel"
)

func connectAsClient(addr, password string) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return
	}

	defer conn.Close()
	cc, err := CipherConn(conn, password, true)
	if err != nil {
		// todo: log error
		return
	}

	t := tunnel.New(cc)
	t.Run()
}
