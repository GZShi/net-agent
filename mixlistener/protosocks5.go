package mixlistener

import (
	"net"
)

type socks5Listener struct {
	ProtoListener
}

// Socks5 监听HTTP协议
func Socks5() ProtoListener {
	return &protobase{
		name:    Socks5Name,
		ch:      make(chan net.Conn, 255),
		network: "",
		addr:    "",
	}
}

func (proto *socks5Listener) Taste(buf []byte) bool {
	return buf != nil && buf[0] == 0x05
}
