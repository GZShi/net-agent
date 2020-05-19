package transport

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

type teststruct struct {
	Index int
}

func TestMakeConnID(t *testing.T) {
	maker := makeConnIDGenerator()

	last := uint32(0)
	for i := 0; i < 100; i++ {
		value := uint32(maker())
		if last != 0 {
			if value-last != 1 {
				t.Error("bad generator", last, value)
				return
			}
		}

		last = value
	}
}

// 测试channel的安全close机制
func TestChanClose(t *testing.T) {
	c := make(chan *teststruct)

	go func() {
		for i := 0; i < 10; i++ {
			c <- &teststruct{i}
		}
		close(c)
	}()

	datas := []*teststruct{}

loop:
	for {
		select {
		case data := <-c:
			if data == nil {
				break loop
			}
			datas = append(datas, data)
		case <-time.After(time.Second * 2):
			break loop
		}
	}

	if len(datas) != 10 {
		t.Error("datas lenght not match")
		return
	}
	for i := 0; i < 10; i++ {
		if datas[i].Index != i {
			t.Error("datas index not match", datas[i], i)
			return
		}
	}

	// try close chan
loop2:
	for {
		select {
		case data := <-c:
			if data == nil {
				// closed
				break loop2
			}
		default:
			close(c)
		}
	}

	c2 := make(chan int)
	close(c2)

	d := <-c2
	if d == 0 {
	} else {
		t.Error("")
	}

	// multiple read
	c3 := make(chan *teststruct)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := <-c3
			if data != nil {
				t.Error("data is not error")
			}
		}()
	}
	close(c3)
	wg.Wait()

}

// 测试tunnel的基本传输功能
func TestTunnel_Serve(t *testing.T) {
	agentConn, serverConn := net.Pipe()

	transData, err := ioutil.ReadFile("./tunnel_test.go")
	if err != nil {
		t.Error(err)
		return
	}
	dataLen := len(transData)

	var wg sync.WaitGroup
	wg.Add(1)
	addr := "127.0.0.1:8013"
	go func() {
		defer wg.Done()
		l, err := net.Listen("tcp", addr)
		if err != nil {
			t.Error(err)
			return
		}
		target, err := l.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer target.Close()
		l.Close()

		buf := make([]byte, dataLen)
		_, err = io.ReadFull(target, buf)
		if err != nil {
			t.Error(err)
			return
		}
		if !bytes.Equal(buf, transData) {
			t.Error("data not equal")
			return
		}
		_, err = target.Write(buf)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	<-time.After(time.Millisecond * 100)

	clientName := "abcdefg"
	secret := "1234"
	randKey := []byte("1234567812345678")
	agent, err := NewTunnel(agentConn, clientName, secret, randKey, true, true)
	if err != nil {
		t.Error(err)
		return
	}
	server, err := NewTunnel(serverConn, clientName, secret, randKey, true, false)
	if err != nil {
		t.Error(err)
		return
	}

	go agent.Serve()
	go server.Serve()

	client, err := server.Dial("", "tcp", addr, "")
	if err != nil {
		t.Error(err)
		return
	}

	// 检查两边是否存在port
	if agent.activePortCount != 1 {
		t.Error("not exist")
		return
	}
	if server.activePortCount != 1 {
		t.Error("not exist")
		return
	}
	pageSize := 100
	dataPos := 0
	for {
		start := dataPos
		end := start + pageSize
		if end > len(transData) {
			end = len(transData)
		}

		_, err = client.Write(transData[start:end])
		if err != nil {
			t.Error(err)
			return
		}

		if end == len(transData) {
			break
		}
		dataPos = end
		<-time.After(time.Millisecond * 50)
	}
	buf := make([]byte, len(transData))
	_, err = io.ReadFull(client, buf)
	if err != nil {
		t.Error(err)
		return
	}
	if false == bytes.Equal(buf, transData) {
		t.Error("data not equal")
		return
	}

	wg.Wait()
	client.Close()

	<-time.After(time.Millisecond * 100)

	// 检查tunnel里的ports是否已经清理干净
	if agent.activePortCount != 0 {
		t.Error("exist")
		return
	}
	if server.activePortCount != 0 {
		t.Error("exist")
		return
	}
}

// 测试XORKeyStream操作
func TestXORStream(t *testing.T) {
	blockCipher, _ := aes.NewCipher([]byte("1234567812345678"))
	iv := []byte("secret-iv")
	for len(iv) < blockCipher.BlockSize() {
		iv = append(iv, '*')
	}
	makeCTR := func() cipher.Stream {
		return cipher.NewCTR(blockCipher, iv[0:blockCipher.BlockSize()])
	}

	enc := makeCTR()
	dec := makeCTR()

	data := []byte("helloworld")
	encoded := make([]byte, len(data))
	decoded := make([]byte, len(data))
	enc.XORKeyStream(encoded, data)
	dec.XORKeyStream(decoded, encoded)

	if !bytes.Equal(data, decoded) {
		t.Error("XORKeyStream test failed")
		return
	}
}
