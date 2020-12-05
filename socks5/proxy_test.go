package socks5

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestSocks5(t *testing.T) {
	echoAddr := "localhost:9998"
	proxy := &ProxyInfo{
		Network:  "tcp4",
		Address:  "localhost:9999",
		NeedAuth: false,
	}
	// run echo server
	go func() {
		l, err := net.Listen("tcp4", echoAddr)
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
			io.Copy(conn, conn)
		}
	}()

	// run socks5 server
	go func() {
		server := NewServer()
		server.ListenAndRun(proxy.Address)
	}()

	<-time.After(time.Millisecond * 50)

	conn, err := proxy.Dial(echoAddr)
	if err != nil {
		t.Error(err)
		return
	}

	payload := []byte("aklfjlajfklasjfjafkiwerioqwoirklsjf")
	_, err = conn.Write(payload)
	if err != nil {
		t.Error(err)
		return
	}

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

func TestSocks5PswdAuth(t *testing.T) {
	echoAddr := "localhost:9990"
	proxy := &ProxyInfo{
		Network:  "tcp4",
		Address:  "localhost:9991",
		NeedAuth: true,
		Username: "test-username",
		Password: "test-pswdsdf",
	}

	// run echo server
	go func() {
		l, err := net.Listen("tcp4", echoAddr)
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
			io.Copy(conn, conn)
		}
	}()

	// run socks5 server
	go func() {
		server := NewServer()
		server.SetAuthChecker(PswdAuthChecker(func(u, p string, ctx map[string]string) error {
			if u == proxy.Username && p == proxy.Password {
				return nil
			}
			return errors.New("pswd checker error")
		}))
		server.ListenAndRun(proxy.Address)
	}()

	<-time.After(time.Millisecond * 50)

	conn, err := proxy.Dial(echoAddr)
	if err != nil {
		t.Error(err)
		return
	}

	payload := []byte("aklfjlajfklasjfjafkiwerioqwoirklsjf")
	_, err = conn.Write(payload)
	if err != nil {
		t.Error(err)
		return
	}

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
