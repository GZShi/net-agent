package socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// 测试认证方法握手过程
func TestMethodHandshake(t *testing.T) {
	client, server := net.Pipe()
	methods := []byte{MethodNoAuth, MethodNoAuth, MethodNoAuth}
	var wg sync.WaitGroup
	wg.Add(2)

	// client simulation
	go func() {
		defer wg.Done()
		defer client.Close()

		for _, method := range methods {
			if _, err := client.Write([]byte{0x05, 0x01, method}); err != nil {
				t.Error(err)
				return
			}
			buf := make([]byte, 2)
			if _, err := io.ReadFull(client, buf); err != nil {
				t.Error(err)
				return
			}
		}
	}()

	// server simulation
	go func() {
		defer wg.Done()
		defer server.Close()

		for _, method := range methods {
			resMethod, err := methodHandshake(server, server)
			if err != nil {
				t.Error(err)
				return
			}
			if resMethod != method {
				t.Error("method is not same. res:", resMethod, "want:", method)
				return
			}
		}
	}()

	wg.Wait()
}

func TestMethodHandshake_badVersion(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		payloads := []byte{0x03, 0x01, 0x01}
		wn, err := client.Write(payloads)
		if err != nil {
			t.Error(err)
			return
		}
		if wn != len(payloads) {
			t.Error("size not match")
			return
		}
		resp := make([]byte, 2)
		_, err = io.ReadFull(client, resp)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	_, err := methodHandshake(server, server)
	if err != errVersion {
		t.Error("error info not match")
		return
	}
}

func TestMethodHandshake_badDataSize(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		payloads := []byte{0x05, 0x02, 0x00, 0x01, 0x02}
		rn, err := client.Write(payloads)
		if err != nil {
			t.Error(err)
			return
		}
		if rn != len(payloads) {
			t.Error("size not match")
			return
		}
		resp := make([]byte, 2)
		_, err = io.ReadFull(client, resp)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	_, err := methodHandshake(server, server)
	if err != errDataSize {
		t.Error("error info not match")
		return
	}
}

func TestMethodHandshake_coverSlowSending(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		payloads := []byte{0x05, 0x05, 0x00, 0x01, 0x02, 0x03, 0x04}
		wn1, err := client.Write(payloads[0:4])
		if err != nil {
			t.Error(err)
			return
		}
		// 等待2秒钟再写剩下的数据
		<-time.After(time.Second * 2)
		wn2, err := client.Write(payloads[4:len(payloads)])
		if err != nil {
			t.Error(err)
			return
		}
		wn := wn1 + wn2
		if wn != len(payloads) {
			t.Error("size not match")
			return
		}
		resp := make([]byte, 2)
		_, err = io.ReadFull(client, resp)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	_, err := methodHandshake(server, server)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestParseRequest(t *testing.T) {
	client, server := net.Pipe()

	portNum := uint16(12345)
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, portNum)
	addrs := []string{
		"FE80::202:B3FF:FE1E:8329",
		"192.168.1.1",
		"weibo.com",
		"google.com.cn",
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// client simulation
	go func() {
		defer wg.Done()
		defer client.Close()

		for _, addr := range addrs {
			atype := atypeDomain
			addrBuf := append([]byte{uint8(len(addr))}, []byte(addr)...)

			ip := net.ParseIP(addr)
			if ip != nil {
				if ipv4 := ip.To4(); ipv4 != nil {
					atype = atypeIPV4
					addrBuf = []byte(ipv4)
				} else if ipv6 := ip.To16(); ipv6 != nil {
					atype = atypeIPV6
					addrBuf = []byte(ip)
				} else {
					t.Error(addr, "Unknown address type")
					return
				}
			}
			header := []byte{0x05, cmdConnect, 0x00}

			buf := append(header, atype)
			buf = append(buf, addrBuf...)
			buf = append(buf, port...)
			if _, err := client.Write(buf); err != nil {
				t.Error(addr, err)
				return
			}

			resp := make([]byte, 10)
			if _, err := io.ReadFull(client, resp); err != nil {
				t.Error(addr, err)
				return
			}
		}
	}()

	// server simulation
	go func() {
		defer wg.Done()
		defer client.Close()

		for _, addr := range addrs {
			cmd, host, parsedPort, err := parseRequest(server, server)
			if err != nil {
				t.Error(err)
				return
			}

			if cmd != cmdConnect {
				t.Error("cmd is not connect", cmd, cmdConnect)
				return
			}

			if host != addr && false == bytes.Equal(net.ParseIP(host), net.ParseIP(addr)) {
				t.Error("addr parse failed", host, addr)
				return
			}

			if parsedPort != portNum {
				t.Error("port parse failed", parsedPort, portNum)
				return
			}
		}
	}()

	wg.Wait()
}

func TestAuthWithPswd(t *testing.T) {
	secret := "testhelloworld"
	users := []string{"abcd", "12345", "xisith"}
	pswds := []string{}

	for _, user := range users {
		pswds = append(pswds, CalcSum(user, secret))
	}

	client, server := net.Pipe()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer client.Close()

		for index, user := range users {
			pswd := pswds[index]

			payload := []byte{0x01}
			payload = append(payload, uint8(len(user)))
			payload = append(payload, []byte(user)...)
			payload = append(payload, uint8(len(pswd)))
			payload = append(payload, []byte(pswd)...)

			_, err := client.Write(payload)
			if err != nil {
				t.Error(err)
				return
			}

			resp := make([]byte, 100)
			rn, err := io.ReadAtLeast(client, resp, 2)
			if err != nil {
				t.Error(err)
				return
			}
			if rn > 2 {
				t.Error("data size error", rn)
				return
			}
		}
	}()

	defer server.Close()

	for _, user := range users {
		var uname string
		var err error
		if uname, err = authWithPswd(server, server, secret); err != nil {
			t.Error(err, user)
			return
		}
		if uname != user {
			t.Error("username not match")
			return
		}
	}
}

// 测试总流程
func TestServeSocks5(t *testing.T) {
	client, server1 := net.Pipe()
	server2, target := net.Pipe()
	host := "hello.world.com"
	port := uint16(18081)
	requestStr := "socks5 is coming..."
	responseStr := "reply from fake socket"

	dialer := func(sourceAddr, net, addr, uname string) (net.Conn, error) {
		if addr != fmt.Sprintf("%s:%d", host, port) {
			err := errors.New("address not match")
			t.Error(err)
			return nil, err
		}

		// 此处模拟目标服务器，等待客户端发送 requestStr
		// 然后应答 responseStr
		// 最后关闭连接
		go func() {
			defer target.Close()
			buf := make([]byte, 512)
			rn, _ := io.ReadAtLeast(target, buf, len(requestStr))
			if string(buf[0:rn]) != requestStr {
				t.Error("request not match")
				return
			}
			target.Write([]byte(responseStr))
		}()

		return server2, nil
	}

	// client simulation
	go func() {
		defer client.Close()
		client.Write([]byte{dataVersion, 0x01, MethodNoAuth})
		resp1 := make([]byte, 2)
		io.ReadFull(client, resp1)

		client.Write([]byte{dataVersion, cmdConnect, 0x00,
			atypeDomain, uint8(len(host)),
		})
		client.Write([]byte(host))
		portBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(portBuf, port)
		client.Write(portBuf)
		resp2 := make([]byte, 1+1+1+1+net.IPv4len+2)
		io.ReadFull(client, resp2)

		client.Write([]byte(requestStr))
		resp3 := make([]byte, 4096)
		rn, _ := io.ReadAtLeast(client, resp3, len(responseStr))
		if string(resp3[0:rn]) != responseStr {
			t.Error("test failed")
			return
		}
	}()

	go func() {
		defer server1.Close()
		ServeSocks5(server1, server1, server1, "", dialer)
	}()
}
