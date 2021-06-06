package tunnel

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestNetwork(t *testing.T) {
	connC, connS := net.Pipe()

	client := New(connC, true)
	server := New(connS, true)
	go client.Run()
	go server.Run()
	<-time.After(time.Millisecond * 50)

	go func() {
		l, err := server.Listen(1080)
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

	<-time.After(time.Millisecond * 50)
	conn, err := client.Dial(1080)
	if err != nil {
		t.Error(err)
		return
	}

	payload := []byte("hello,world,virtual,net,work")
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
