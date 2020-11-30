package socks5

import (
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
	case IPv6:
		// 读取IPv6的地址
		buf = make([]byte, net.IPv6len+2)
		rn, err = io.ReadFull(r, buf)
		readed += int64(rn)
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
	default:
		return readed, ErrAddressTypeNotSupport
	}

	// 把port读取出来，并把addressBuf中的port裁减掉
	portPos := len(buf) - 2
	req.addressBuf = buf[1:portPos]
	req.port = binary.BigEndian.Uint16(buf[portPos:])

	return readed, err
}

func (req *request) WriteTo(w io.Writer) (written int64, err error) {
	if req.version != VersionSocks5 {
		return 0, ErrVersionNotSupport
	}
	if len(req.addressBuf) > 255 {
		return 0, ErrAddressBufTooLong
	}

	var buf []byte
	var portPos uint8
	switch req.addressType {
	case IPv4:
		portPos = 4 + net.IPv4len
	case IPv6:
		portPos = 4 + net.IPv6len
	case Domain:
		portPos = 4 + 1 + uint8(len(req.addressBuf))
	default:
		return 0, ErrAddressTypeNotSupport
	}

	buf = make([]byte, portPos+2)

	buf[0] = req.version
	buf[1] = req.command
	buf[3] = req.addressType
	buf[4] = byte(len(req.addressBuf))
	copy(buf[5:portPos], req.addressBuf)
	binary.BigEndian.PutUint16(buf[portPos:portPos+2], req.port)

	wn, err := w.Write(buf)
	written = int64(wn)
	return
}
