package socks5

import (
	"errors"
	"fmt"
	"net"
	"testing"
)

func TestSocks5Client(t *testing.T) {
	client, server1 := net.Pipe()
	host := "hello.world.com"
	port := uint16(18081)
	secret := "hello,world"
	username := "tester"
	password := CalcSum(username, secret)
	// requestStr := "socks5 is coming..."
	// responseStr := "reply from fake socket"

	dialer := func(sourceAddr, net, addr, uname string) (net.Conn, error) {
		if addr != fmt.Sprintf("%s:%d", host, port) {
			err := errors.New("address not match")
			t.Error(err)
			return nil, err
		}
		return nil, errors.New("simulate")
	}

	// client simulation
	go func() {
		if err := MakeSocks5Request(client, username, password, host, port); err != nil {
			t.Error(err)
			return
		}

		// client.Write([]byte(requestStr))
		// resp3 := make([]byte, 4096)
		// rn, _ := io.ReadAtLeast(client, resp3, len(responseStr))
		// if string(resp3[0:rn]) != responseStr {
		// 	t.Error("test failed")
		// 	return
		// }
	}()

	HandleSocks5Request(server1, server1, server1, secret, dialer)
}
