package socks5

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	// IPv4 RFC1928/ATYP/IP_V4_address: X'01'
	IPv4 = uint8(1)
	// IPv6 RFC1928/ATYP/IP_V6_address: X'04'
	IPv6 = uint8(4)
	// Domain RFC1928/ATYP/DOMAINNAME: X'03'
	Domain = uint8(3)
)

var (
	// ErrAddressTypeNotSupport ...
	ErrAddressTypeNotSupport = errors.New("address type not supported")
	// ErrAddressBufTooLong ...
	ErrAddressBufTooLong = errors.New("address buf is too long")
)

// Request 客户端请求的数据上下文
type Request interface {
	GetCommand() uint8
	GetAddress() (uint8, []byte)
	GetPort() uint16
	GetAddrPortStr() string
}

type request struct {
	version     uint8
	command     uint8
	addressType uint8
	addressBuf  []byte
	port        uint16
}

func (req *request) GetCommand() uint8 {
	return req.command
}

func (req *request) GetAddress() (uint8, []byte) {
	return req.addressType, req.addressBuf
}

func (req *request) GetPort() uint16 {
	return req.port
}

func (req *request) GetAddrPortStr() string {
	var host string

	switch req.addressType {
	case IPv4, IPv6:
		host = net.IP(req.addressBuf).String()
	case Domain:
		host = string(req.addressBuf)
	}

	return fmt.Sprintf("%v:%v", host, req.GetPort())
}

func (req *request) ReadFrom(r io.Reader) (readed int64, err error) {
	header := make([]byte, 1+1+1+1)
	rn, err := io.ReadFull(r, header)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	if header[0] != VersionSocks5 {
		return readed, ErrVersionNotSupport
	}
	req.version = header[0]
	req.command = header[1]
	req.addressType = header[3]

	var buf []byte
	switch req.addressType {
	case IPv4:
		// 读取IPv4的地址
		buf = make([]byte, net.IPv4len+2)
		rn, err = io.ReadFull(r, buf)
		readed += int64(rn)
		req.addressBuf = buf[:len(buf)-2]
	case IPv6:
		// 读取IPv6的地址
		buf = make([]byte, net.IPv6len+2)
		rn, err = io.ReadFull(r, buf)
		readed += int64(rn)
		req.addressBuf = buf[:len(buf)-2]
	case Domain:
		// 读取域名地址
		buf = make([]byte, 1+255+2)
		rn, err = io.ReadAtLeast(r, buf, 1+2)
		readed += int64(rn)
		if err != nil {
			return readed, err
		}
		bufSize := int(buf[0]) + 1 + 2
		pos := rn
		if pos < bufSize {
			rn, err = io.ReadFull(r, buf[pos:bufSize])
			readed += int64(rn)
		}
		buf = buf[0:bufSize]
		req.addressBuf = buf[1 : len(buf)-2]
	default:
		return readed, ErrAddressTypeNotSupport
	}

	// 把port读取出来，并把addressBuf中的port裁减掉
	req.port = binary.BigEndian.Uint16(buf[len(buf)-2:])

	return readed, err
}

func (req *request) WriteTo(w io.Writer) (written int64, err error) {
	if req.version != VersionSocks5 {
		return 0, ErrVersionNotSupport
	}
	if len(req.addressBuf) > 255 {
		return 0, ErrAddressBufTooLong
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(req.version)
	buf.WriteByte(req.command)
	buf.WriteByte(0)
	buf.WriteByte(req.addressType)

	switch req.addressType {
	case IPv4:
		if len(req.addressBuf) < net.IPv4len {
			return 0, errors.New("invalid IPv4 buffer")
		}
		buf.Write(req.addressBuf[len(req.addressBuf)-net.IPv4len:])
	case IPv6:
		if len(req.addressBuf) != net.IPv6len {
			return 0, errors.New("invalid IPv6 buffer")
		}
		buf.Write(req.addressBuf)
	case Domain:
		buf.WriteByte(byte(len(req.addressBuf)))
		buf.Write(req.addressBuf)
	default:
		return 0, ErrAddressTypeNotSupport
	}

	portBuf := []byte{0, 0}
	binary.BigEndian.PutUint16(portBuf, req.port)
	buf.Write(portBuf)

	return buf.WriteTo(w)
}
