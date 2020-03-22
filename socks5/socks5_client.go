package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// MakeSocks5Request 连接socks5服务器
func MakeSocks5Request(
	conn net.Conn,
	username, password,
	targetHost string, targetPort uint16,
) error {
	// step1: method handshake
	conn.Write([]byte{dataVersion, 0x01, MethodAuthPswd})
	resp1 := make([]byte, 2)
	_, err := io.ReadFull(conn, resp1)
	if err != nil {
		return err
	}
	if resp1[0] != dataVersion || resp1[1] != MethodAuthPswd {
		return fmt.Errorf("socks5 error: server not support pswd-auth")
	}

	// step2: auth
	conn.Write([]byte{dataVerPswd, byte(len(username))})
	conn.Write([]byte(username))
	conn.Write([]byte{byte(len(password))})
	conn.Write([]byte(password))
	resp2 := make([]byte, 2)
	_, err = io.ReadFull(conn, resp2)
	if err != nil {
		return err
	}
	if resp2[0] != 0x01 || resp2[1] != 0x00 {
		return fmt.Errorf("socks5 error: password error")
	}

	// step3: make request
	conn.Write([]byte{
		dataVersion,
		cmdConnect,
		0x00, /*RVS*/
		atypeDomain,
		byte(len(targetHost)),
	})
	conn.Write([]byte(targetHost))
	portBuf := []byte{0x00, 0x00}
	binary.BigEndian.PutUint16(portBuf, targetPort)
	conn.Write(portBuf)
	// response长度可变
	// VER + REP + RSV + ATYPE
	headerSize := 4
	headerBuf := make([]byte, headerSize)
	_, err = io.ReadFull(conn, headerBuf)
	if err != nil {
		return err
	}
	if headerBuf[0] != dataVersion || headerBuf[1] != repSuccess {
		return fmt.Errorf("socks5 error: connect failed")
	}
	bufSize := -1
	switch headerBuf[3] {
	case atypeIPV4:
		bufSize = net.IPv4len + 2
	case atypeIPV6:
		bufSize = net.IPv6len + 2
	case atypeDomain:
		bufSize = 1 + int(headerBuf[4]) + 2
	default:
		err = errAtypeNotSupported
		return err
	}
	buf := make([]byte, bufSize)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return err
	}

	// todo, check buf data

	return nil
}
