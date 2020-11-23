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

type Protobase struct {
	name    string
	ch      chan net.Conn
	network string
	addr    string
}

func NewProtobase(name string) *Protobase {
	return &Protobase{
		name:    name,
		ch:      make(chan net.Conn, 255),
		network: "",
		addr:    "",
	}
}

func (base *Protobase) Name() string {
	return base.name
}

func (base *Protobase) Accept() (net.Conn, error) {
	conn := <-base.ch
	if conn == nil {
		return nil, errors.New("conn channel closed")
	}
	return conn, nil
}

func (base *Protobase) Close() error {
	close(base.ch)
	return nil
}

func (base *Protobase) SetAddr(network, addr string) {
	base.network = network
	base.addr = addr
}

func (base *Protobase) Addr() net.Addr {
	// todo
	return base
}

func (base *Protobase) Network() string {
	return base.network
}

func (base *Protobase) String() string {
	return base.addr
}

func (base *Protobase) Taste(buf []byte) bool {
	return false
}

func (base *Protobase) Recieve(conn net.Conn) error {
	base.ch <- conn
	return nil
}
