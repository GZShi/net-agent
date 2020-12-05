package socks5

import "errors"

const (
	dataVersion = uint8(0x05)
	dataVerPswd = uint8(0x01)
	dataRsv     = uint8(0x00)

	// MethodNoAuth ...
	MethodNoAuth = uint8(0x00)
	// MethodGssapi ...
	MethodGssapi = uint8(0x01)
	// MethodAuthPswd ...
	MethodAuthPswd = uint8(0x02)
	// MethodNoAcceptable ...
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
	// ErrReplyFailure ...
	ErrReplyFailure = errors.New("general SOCKS server failure")
	// ErrReplyConnectionNotAllow ...
	ErrReplyConnectionNotAllow = errors.New("connection not allowed by ruleset")
	// ErrReplyNetworkUnRereachable ...
	ErrReplyNetworkUnRereachable = errors.New("network unreachable")
	// ErrReplyHostUnreachable ...
	ErrReplyHostUnreachable = errors.New("host unreachable")
	// ErrReplyConnectionRefused ...
	ErrReplyConnectionRefused = errors.New("connection refused")
	// ErrReplyTTLExpired ...
	ErrReplyTTLExpired = errors.New("TTL expired")
	// ErrReplyCmdNotSupported ...
	ErrReplyCmdNotSupported = errors.New("command not supported")
	// ErrReplyAtypeNotSupported ...
	ErrReplyAtypeNotSupported = errors.New("address type not supported")
)
