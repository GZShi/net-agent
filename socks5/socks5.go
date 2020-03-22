package socks5

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"strconv"

	"github.com/GZShi/net-agent/logger"
)

const (
	dataVersion = uint8(0x05)
	dataVerPswd = uint8(0x01)
	dataRsv     = uint8(0x00)

	MethodNoAuth       = uint8(0x00)
	MethodGssapi       = uint8(0x01)
	MethodAuthPswd     = uint8(0x02)
	MethodNoAcceptable = uint8(0xff)

	cmdConnect = uint8(0x01)
	cmdBind    = uint8(0x02)
	cmdUDP     = uint8(0x03)

	atypeIPV4   = uint8(0x01)
	atypeIPV6   = uint8(0x04)
	atypeDomain = uint8(0x03)

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
	errVersion             = errors.New("socks version not supported")
	errDataSize            = errors.New("socks data is too large")
	errPswdAuthFailed      = errors.New("socks password auth failure")
	errMethodNotSupported  = errors.New("socks methods not supported")
	errCommandNotSupported = errors.New("socks command not supported")
	errAtypeNotSupported   = errors.New("socks atype not supported")
)

func methodHandshake(reader io.Reader, writer io.Writer) (authMethod uint8, err error) {
	defer func() {
		if err != nil {
			writer.Write([]byte{dataVersion, MethodNoAcceptable})
			return
		}
		_, err = writer.Write([]byte{dataVersion, authMethod})
	}()
	// VER + NMETHOD + METHODS(255)
	maxBufSize := 1 + 1 + 255
	// VER + NMETHOD + METHOD(1)
	minBufSize := 1 + 1 + 1
	buf := make([]byte, maxBufSize)

	rn, err := io.ReadAtLeast(reader, buf, minBufSize)
	if err != nil {
		return
	}

	if buf[0] != dataVersion {
		return 0, errVersion
	}

	nmethods := int(buf[1])
	bufSize := nmethods + 2
	if rn > bufSize {
		return 0, errDataSize
	}
	if rn < bufSize {
		_, err = io.ReadFull(reader, buf[rn:bufSize])
		if err != nil {
			return
		}
	}

	// check if Auth methods supported
	supportedMethods := []byte{MethodNoAuth, MethodAuthPswd}

	for _, supported := range supportedMethods {
		for _, method := range buf[2 : 2+nmethods] {
			if method == supported {
				authMethod = method
				return method, nil
			}
		}
	}

	return MethodNoAcceptable, errMethodNotSupported
}

func parseRequest(reader io.Reader, writer io.Writer) (cmd uint8, host string, port uint16, err error) {
	defer func() {
		if err != nil {
			var repErr uint8
			switch err {
			case errCommandNotSupported:
				repErr = repCmdNotSupported
			case errAtypeNotSupported:
				repErr = repAtypeNotSupported
			default:
				repErr = repFailure
			}
			writer.Write([]byte{dataVersion, repErr,
				0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			})
			return
		}
		_, err = writer.Write([]byte{dataVersion, repSuccess,
			0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		})
	}()

	// VER + CMD + RSV + ATYPE
	headerSize := 4
	// HEADER + ADDR_LEN + 255(MAX_ADDR_LEN) + PORT
	maxBufSize := headerSize + 1 + 255 + 2
	// HEADER + ADDR_LEN + PORT
	minBufSize := headerSize + 1 + 2
	buf := make([]byte, maxBufSize)

	rn, err := io.ReadAtLeast(reader, buf, minBufSize)
	if err != nil {
		return
	}
	if buf[0] != dataVersion {
		err = errVersion
		return
	}
	cmd = buf[1]

	bufSize := -1

	// parse address type
	switch buf[3] {
	case atypeIPV4:
		bufSize = headerSize + net.IPv4len + 2
	case atypeIPV6:
		bufSize = headerSize + net.IPv6len + 2
	case atypeDomain:
		bufSize = headerSize + 1 + int(buf[4]) + 2
	default:
		err = errAtypeNotSupported
		return
	}

	if rn > bufSize {
		err = errDataSize
		return
	}
	if rn < bufSize {
		_, err = io.ReadFull(reader, buf[rn:bufSize])
		if err != nil {
			return
		}
	}

	switch buf[3] {
	case atypeIPV4:
		host = net.IP(buf[4 : 4+net.IPv4len]).String()
	case atypeIPV6:
		host = net.IP(buf[4 : 4+net.IPv6len]).String()
	case atypeDomain:
		host = string(buf[5 : bufSize-2])
	default:
		err = errAtypeNotSupported
		return
	}

	port = binary.BigEndian.Uint16(buf[bufSize-2 : bufSize])

	err = nil
	return
}

func CalcSum(uname, secret string) string {
	h := sha256.New()
	io.WriteString(h, "$$"+uname+"##"+secret+"**")
	sum := hex.EncodeToString(h.Sum(nil))
	return sum[0:12]
}

func authWithPswd(reader io.Reader, writer io.Writer, secret string) (uname string, err error) {
	defer func() {
		if err != nil {
			writer.Write([]byte{0x01, 0x01})
			return
		}
		_, err = writer.Write([]byte{0x01, 0x00})
	}()

	maxBufSize := 1 + 1 + 255 + 1 + 255
	minBufSize := 1 + 1 + 1 + 1 + 1
	buf := make([]byte, maxBufSize)

	rn, err := io.ReadAtLeast(reader, buf, minBufSize)
	if err != nil {
		return
	}
	if buf[0] != dataVerPswd {
		return "", errVersion
	}
	unameLen := uint8(buf[1])
	bufSize := int(1 + 1 + unameLen + 1)
	if rn < bufSize {
		_, err := io.ReadFull(reader, buf[rn:bufSize])
		if err != nil {
			return "", err
		}
		rn = bufSize
	}
	uname = string(buf[2 : 2+unameLen])

	pswdLen := uint8(buf[2+unameLen])
	bufSize = int(1 + 1 + unameLen + 1 + pswdLen)
	if rn > bufSize {
		return "", errDataSize
	}
	if rn < bufSize {
		_, err := io.ReadFull(reader, buf[rn:bufSize])
		if err != nil {
			return "", err
		}
		rn = bufSize
	}
	pswd := string(buf[2+unameLen+1 : rn])

	if CalcSum(uname, secret) != pswd {
		return "", errPswdAuthFailed
	}

	return uname, nil
}

// HandleSocks5Request 处理socks5代理请求
func HandleSocks5Request(
	conn net.Conn,
	reader io.Reader,
	writer io.Writer,
	secret string,
	dialer Dialer,
) {
	defer conn.Close()
	// step1: method handshake
	method, err := methodHandshake(reader, writer)
	if err != nil {
		log.Get().WithError(err).Error("socks handshake failed")
		return
	}

	// step2: auth
	uname := ""
	if method == MethodAuthPswd {
		if uname, err = authWithPswd(reader, writer, secret); err != nil {
			log.Get().WithError(err).Error("socks auth password failed")
			return
		}
	}

	// step3: parse request
	_, host, port, err := parseRequest(reader, writer)
	if err != nil {
		log.Get().WithError(err).Error("socks parse request failed")
		return
	}

	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))

	// normal socks5
	target, err := dialer(conn.RemoteAddr().String(), "tcp", addr, uname)
	if err != nil {
		log.Get().WithError(err).WithField("addr", addr).Error("dial target failed")
		return
	}
	if target == nil {
		log.Get().WithField("addr", addr).Error("target is nil")
		return
	}
	defer target.Close()

	go func() {
		defer conn.Close()
		defer target.Close()
		io.Copy(target, reader)
	}()
	func() {
		defer conn.Close()
		defer target.Close()
		io.Copy(writer, target)
	}()
}
