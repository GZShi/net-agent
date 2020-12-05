package socks5

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func runTestEchoServer(t *testing.T, addr string) {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		t.Error(err)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}

		go io.Copy(conn, conn)
	}
}

func runTestSocks5Server(t *testing.T, proxy *ProxyInfo) {
	l, err := net.Listen(proxy.Network, proxy.Address)
	if err != nil {
		t.Error(err)
		return
	}

	s := NewServer()
	if proxy.NeedAuth {
		s.SetAuthChecker(PswdAuthChecker(func(u, p string, ctx map[string]string) error {
			if u == proxy.Username && p == proxy.Password {
				return nil
			}
			return errors.New("check username password failed")
		}))
	}

	s.Run(l)
}

func TestProxyChain(t *testing.T) {
	echoAddr := "localhost:9000"
	proxy1 := &ProxyInfo{
		Network:  "tcp4",
		Address:  "localhost:9001",
		NeedAuth: true,
		Username: "hello",
		Password: "world",
	}
	proxy2 := &ProxyInfo{
		Network:  "tcp4",
		Address:  "localhost:9002",
		NeedAuth: false,
	}

	go runTestEchoServer(t, echoAddr)
	go runTestSocks5Server(t, proxy1)
	go runTestSocks5Server(t, proxy2)

	chain := &ProxyChain{proxy1, proxy2, proxy1, proxy2}

	<-time.After(time.Second * 1)

	conn, err := chain.Dial(echoAddr)
	if err != nil {
		t.Error(err)
		return
	}

	payload := []byte("jalkdsjfasjfajsdlfkjaslkfj")

	go func() {
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	resp := make([]byte, len(payload))
	_, err = io.ReadFull(conn, resp)
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(payload, resp) {
		t.Error("not equal")
		return
	}
}
