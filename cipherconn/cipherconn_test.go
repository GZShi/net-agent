package cipherconn

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

func TestCipherConn(t *testing.T) {
	password := "hello,world,pass50rd"
	c1, c2 := net.Pipe()

	payload := []byte("hello, world, cipher, conn")
	var wg sync.WaitGroup

	wg.Add(1)
	go func(conn net.Conn) {
		defer func() {
			conn.Close()
			wg.Done()
		}()

		cc1, err := New(c1, password)
		if err != nil {
			t.Error(err)
			return
		}

		sendData := payload
		recvData := make([]byte, len(sendData))

		cc1.Write(sendData)
		io.ReadFull(cc1, recvData)

		if !bytes.Equal(sendData, recvData) {
			t.Error("not equal")
			return
		}
	}(c1)

	wg.Add(1)
	go func(conn net.Conn) {
		defer func() {
			conn.Close()
			wg.Done()
		}()

		cc2, err := New(c2, password)
		if err != nil {
			t.Error(err)
			return
		}

		_, err = io.Copy(cc2, cc2)
		if err != nil {
			t.Error(err)
			return
		}
	}(c2)

	wg.Wait()
}
