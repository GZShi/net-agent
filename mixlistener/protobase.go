package mixlistener

import (
	"errors"
	"net"
)

const (
	// HTTPName ...
	HTTPName = "http"
	// Socks5Name ...
	Socks5Name = "socks5"
)

// ProtoListener 单一协议的特征识别与监听程序
type ProtoListener interface {
	Name() string
	SetAddr(network, addr string)
	Taste(peekBuf []byte) bool
	Recieve(net.Conn) error
	net.Listener
}

type protobase struct {
	name    string
	ch      chan net.Conn
	network string
	addr    string
}

func (base *protobase) Name() string {
	return base.name
}

func (base *protobase) Accept() (net.Conn, error) {
	conn := <-base.ch
	if conn == nil {
		return nil, errors.New("conn channel closed")
	}
	return conn, nil
}

func (base *protobase) Close() error {
	close(base.ch)
	return nil
}

func (base *protobase) SetAddr(network, addr string) {
	base.network = network
	base.addr = addr
}

func (base *protobase) Addr() net.Addr {
	// todo
	return base
}

func (base *protobase) Network() string {
	return base.network
}

func (base *protobase) String() string {
	return base.addr
}

func (base *protobase) Taste(buf []byte) bool {
	return false
}

func (base *protobase) Recieve(conn net.Conn) error {
	base.ch <- conn
	return nil
}
