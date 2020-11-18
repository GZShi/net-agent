package tunnel

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
)

func TestNewServer(t *testing.T) {
	send, recv := net.Pipe()
	defer send.Close()
	defer recv.Close()

	s1 := NewServer(send)
	s2 := NewServer(recv)

	go s1.Run()
	go s2.Run()

	t.Run("test stream", func(t *testing.T) {
		payloads := [][]byte{
			[]byte("helloworld"),
			[]byte("1234567788sasdfklajfasjfklasf"),
			[]byte("large file~~~"),
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw := s1.NewStreamRW(1)
			for _, payload := range payloads {
				buf := make([]byte, len(payload))
				_, err := io.ReadFull(rw, buf)
				if err != nil {
					t.Error(err)
					return
				}
				if !bytes.Equal(payload, buf) {
					t.Error("not equal")
					return
				}
			}
		}()

		rw := s2.NewStreamRW(1)
		for _, payload := range payloads {
			_, err := rw.Write(payload)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
}
