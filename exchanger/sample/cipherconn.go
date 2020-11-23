package main

import "net"

type cipherconn struct {
	net.Conn
}
