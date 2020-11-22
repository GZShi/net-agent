package socks5

import (
	"errors"
	"io"
	"net"
	"sync"
)

const (
	ConnectCommand = uint8(0x01)
	BindCommand    = uint8(0x02)
	UDPCommand     = uint8(0x03)
)

var (
	// ErrCommandNotSupport ...
	ErrCommandNotSupport = errors.New("socks5 command not supported")
)

// DefaultRequester 执行net.Dial创建连接，并将两个net.Conn进行连接
func DefaultRequester(req Request) (net.Conn, error) {
	if req.GetCommand() != ConnectCommand {
		return nil, ErrCommandNotSupport
	}
	return net.Dial("tcp4", req.GetAddrPortStr())
}

// Link 连接两个io.ReadWriteCloser
// 互相从对方读取数据和写入数据
// 直到发生一个错误为止
// 发生错误后，执行Close操作
func Link(a io.ReadWriteCloser, b io.ReadWriteCloser) (a2bN, b2aN int64, err error) {
	var wg sync.WaitGroup
	var once sync.Once

	clean := func(copyErr error) {
		if copyErr != nil {
			once.Do(func() {
				err = copyErr
			})
		}
		a.Close()
		b.Close()
		wg.Done()
	}

	wg.Add(1)
	go func() {
		var cpErr error
		b2aN, cpErr = io.Copy(a, b)
		clean(cpErr)
	}()

	wg.Add(1)
	go func() {
		var cpErr error
		a2bN, err = io.Copy(b, a)
		clean(cpErr)
	}()

	wg.Wait()
	return a2bN, b2aN, err
}
