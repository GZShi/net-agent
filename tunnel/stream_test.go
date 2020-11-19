package tunnel

import (
	"bytes"
	"io"
	"math/rand"
	"sync"
	"testing"
)

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
