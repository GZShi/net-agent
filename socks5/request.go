package socks5

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	IPv4   = uint8(1)
	IPv6   = uint8(4)
	Domain = uint8(3)
)

var (
	AddressTypeNotSupport = errors.New("address type not supported")
	AddressBufTooLong     = errors.New("address buf is too long")
)

type request struct {
	version     uint8
	command     uint8
	addressType uint8
	addressBuf  []byte
	port        uint16
}

func (req *request) ReadFrom(r io.Reader) (readed int64, err error) {
	header := make([]byte, 1+1+1+1)
	rn, err := io.ReadFull(r, header)
	readed += int64(rn)
	if err != nil {
		return readed, err
	}

	if header[0] != VersionSocks5 {
		return readed, VersionNotSupport
	}
	req.version = header[0]
	req.command = header[1]
	req.addressType = header[3]

	switch req.addressType {
	case IPv4:
		// 读取IPv4的地址
		req.addressBuf = make([]byte, net.IPv4len+2)
		rn, err = io.ReadFull(r, req.addressBuf)
		readed += int64(rn)
	case IPv6:
		// 读取IPv6的地址
		req.addressBuf = make([]byte, net.IPv6len+2)
		rn, err = io.ReadFull(r, req.addressBuf)
		readed += int64(rn)
	case Domain:
		// 读取域名地址
		buf := make([]byte, 1+255+2)
		rn, err = io.ReadAtLeast(r, buf, 1+2)
		readed += int64(rn)
		if err == nil {
			pos := rn
			end := buf[0] + 1 + 2
			rn, err = io.ReadFull(r, buf[pos:end])
			readed += int64(rn)
		}
	default:
		return readed, AddressTypeNotSupport
	}

	return readed, err
}

func (req *request) WriteTo(w io.Writer) (written int64, err error) {
	if req.version != VersionSocks5 {
		return 0, VersionNotSupport
	}
	if len(req.addressBuf) > 255 {
		return 0, AddressBufTooLong
	}

	var buf []byte
	var portPos uint8
	switch req.addressType {
	case IPv4:
		portPos = 4 + net.IPv4len
	case IPv6:
		portPos = 4 + net.IPv6len
	case Domain:
		portPos = 4 + uint8(len(req.addressBuf))
	default:
		return 0, AddressTypeNotSupport
	}

	buf = make([]byte, portPos+2)

	buf[0] = req.version
	buf[1] = req.command
	buf[3] = req.addressType
	copy(buf[4:portPos], req.addressBuf)
	binary.BigEndian.PutUint16(buf[portPos:portPos+2], req.port)

	wn, err := w.Write(buf)
	written = int64(wn)
	return
}
