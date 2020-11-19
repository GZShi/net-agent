package tunnel

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"sync"
	"testing"
)

func makePipe() (*Server, *Server) {
	send, recv := net.Pipe()

	s1 := NewServer(send)
	s2 := NewServer(recv)

	go s1.Run()
	go s2.Run()

	return s1, s2
}
func TestServerStream(t *testing.T) {
	s1, s2 := makePipe()

	t.Run("test stream", func(t *testing.T) {
		largeData := []byte{}
		kbData := make([]byte, 1024)
		for i := 0; i < 100; i++ {
			rand.Read(kbData)
			largeData = append(largeData, kbData...)
		}
		payloads := [][]byte{
			[]byte("helloworld"),
			[]byte("1234567788sasdxfklajfasjfklasf"),
			[]byte("large file~~~"),
			largeData,
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw := s1.NewStreamRW(1, 2)
			for _, payload := range payloads {
				buf := make([]byte, len(payload))
				_, err := io.ReadFull(rw, buf)
				if err != nil {
					t.Error(err)
					return
				}
				if !bytes.Equal(payload, buf) {
					t.Error("not equal", len(payload), len(buf))
					t.Log("want", string(payload))
					t.Log("recv", string(buf))
					return
				}
			}
		}()

		rw := s2.NewStreamRW(2, 1)
		for _, payload := range payloads {
			pos := 0
			for pos < len(payload) {
				end := pos + 1024
				if end > len(payload) {
					end = len(payload)
				}
				_, err := rw.Write(payload[pos:end])
				if err != nil {
					t.Error(err)
					return
				}
				pos = end
			}
		}
		wg.Wait()
	})
}

func TestServerRequest(t *testing.T) {
	s1, s2 := makePipe()

	s2.On("echo", func(ctx Context) {
		text, err := ctx.GetText()
		if err != nil {
			t.Error(err)
			return
		}
		ctx.Text(text)
	})

	t.Run("test request", func(t *testing.T) {
		req := &Frame{
			ID:        0,
			Type:      FrameRequest,
			SessionID: 0,
			Header:    nil,
			DataType:  TextData,
			Data:      []byte("hello,world"),
		}
		resp, err := s1.request(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp == nil {
			t.Error("resp is nil")
			return
		}
	})
}
