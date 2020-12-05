package socks5

import "errors"

const (
	dataVersion = uint8(0x05)
	dataVerPswd = uint8(0x01)
	dataRsv     = uint8(0x00)

	MethodNoAuth       = uint8(0x00)
	MethodGssapi       = uint8(0x01)
	MethodAuthPswd     = uint8(0x02)
	MethodNoAcceptable = uint8(0xff)

	repSuccess              = uint8(0x00)
	repFailure              = uint8(0x01)
	repConnectionNotAllow   = uint8(0x02)
	repNetworkUnRereachable = uint8(0x03)
	repHostUnreachable      = uint8(0x04)
	repConnectionRefused    = uint8(0x05)
	repTTLExpired           = uint8(0x06)
	repCmdNotSupported      = uint8(0x07)
	repAtypeNotSupported    = uint8(0x08)
)

var (
	ReplyErrFailure              = errors.New("general SOCKS server failure")
	ReplyErrConnectionNotAllow   = errors.New("connection not allowed by ruleset")
	ReplyErrNetworkUnRereachable = errors.New("Network unreachable")
	ReplyErrHostUnreachable      = errors.New("Host unreachable")
	ReplyErrConnectionRefused    = errors.New("Connection refused")
	ReplyErrTTLExpired           = errors.New("TTL expired")
	ReplyErrCmdNotSupported      = errors.New("Command not supported")
	ReplyErrAtypeNotSupported    = errors.New("Address type not supported")
)
