package tunnel

import (
	"bytes"
	"math/rand"
	"net"
	"sync"
	"testing"
)

func TestFrameReadWrite(t *testing.T) {
	var err error
	var rn int64
	var wn int64
	var rf Frame

	sf := &Frame{
		ID:        uint32(rand.Int()),
		Type:      FrameStreamData,
		SessionID: uint32(rand.Int()),
		Header:    []byte("this.is.header.data"),
		DataType:  BinaryData,
		Data:      []byte("data.body.ok.now~~"),
	}

	var wg sync.WaitGroup
	send, recv := net.Pipe()

	wg.Add(1)
	go func() {
		defer wg.Done()

		rn, err = rf.ReadFrom(recv)
		if err != nil {
			t.Error(err)
			return
		}

	}()

	wn, err = sf.WriteTo(send)
	if err != nil {
		t.Error(err)
		return
	}

	wg.Wait()

	// compare
	if wn == 0 || rn == 0 {
		t.Error("wn and rn invalid")
		return
	}
	if wn != rn {
		t.Error("wn != rn")
		return
	}
	if sf.ID != rf.ID || sf.Type != rf.Type || sf.SessionID != rf.SessionID ||
		sf.DataType != rf.DataType {
		t.Error("sf not equal rf")
		return
	}
	if !bytes.Equal(sf.Header, rf.Header) || !bytes.Equal(sf.Data, rf.Data) {
		t.Error("sf.buf not equal rf.buf")
		return
	}
}
