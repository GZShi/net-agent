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
	// TunnelName ...
	TunnelName = "tunnel"
)

// ProtoListener 单一协议的特征识别与监听程序
type ProtoListener interface {
	Name() string
	SetAddr(network, addr string)
	Taste(peekBuf []byte) bool
	Recieve(net.Conn) error
	net.Listener
}

// Protobase 实现协议解析的基类
type Protobase struct {
	name    string
	ch      chan net.Conn
	network string
	addr    string
}

// NewProtobase 创建新的协议基础
func NewProtobase(name string) *Protobase {
	return &Protobase{
		name:    name,
		ch:      make(chan net.Conn, 255),
		network: "",
		addr:    "",
	}
}

// Name 协议名称
func (base *Protobase) Name() string {
	return base.name
}

// Accept net.Listener协议方法
func (base *Protobase) Accept() (net.Conn, error) {
	conn := <-base.ch
	if conn == nil {
		return nil, errors.New("conn channel closed")
	}
	return conn, nil
}

// Close net.Listener协议方法
func (base *Protobase) Close() error {
	close(base.ch)
	return nil
}

// SetAddr 设置Listener的地址，支持net.Listener.Addr返回正确值
func (base *Protobase) SetAddr(network, addr string) {
	base.network = network
	base.addr = addr
}

// Addr net.Listener协议方法
func (base *Protobase) Addr() net.Addr {
	// todo:5
	return base
}

// Network net.Addr 协议方法
func (base *Protobase) Network() string {
	return base.network
}

// Network net.Addr 协议方法
func (base *Protobase) String() string {
	return base.addr
}

// Taste 默认的ProtoListener协议方法，需要重载
func (base *Protobase) Taste(buf []byte) bool {
	return false
}

// Recieve 默认的ProtoListener协议方法
func (base *Protobase) Recieve(conn net.Conn) error {
	base.ch <- conn
	return nil
}
