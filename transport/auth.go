package transport

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

func calcSum(secret, name string, b byte, timestamp uint64, randKey []byte) []byte {
	char := string(b)
	str := fmt.Sprintf("%s#%s#%s$%d", secret, char, name, timestamp)
	hash := sha256.New()
	hash.Write([]byte(str))
	hash.Write(randKey)
	checksum := hash.Sum(nil)
	return checksum
}

// makeAuthData 创建认证串
func makeAuthData(secret, name string, b byte, timestamp uint64, randKey []byte) ([]byte, error) {
	if b < 'a' || b > 'z' {
		return nil, errors.New("b is not lower case alpha")
	}
	if len(name) <= 0 || len(name) > 255 {
		return nil, errors.New("length of name invalid")
	}

	checksum := calcSum(secret, name, b, timestamp, randKey)

	timebuf := make([]byte, 8)
	binary.BigEndian.PutUint64(timebuf, timestamp)

	buf := []byte{b}
	buf = append(buf, checksum...)
	buf = append(buf, timebuf...)
	buf = append(buf, randKey...)
	buf = append(buf, byte(len(name)))
	buf = append(buf, []byte(name)...)

	return buf, nil
}

// CheckAgentConn ...
// [a-z 1byte][check-sum 32bytes][timestamp 8bytes][rand-key 16bytes][var name 1+bytes]
func CheckAgentConn(conn net.Conn, secret string) (name string, randKey []byte, err error) {
	var b byte
	defer func() {
		if err != nil {
			conn.Write([]byte{b, 0x01})
		} else {
			_, err = conn.Write([]byte{b, 0x00})
		}
	}()
	checksumSize := 32
	maxSize := 1 + checksumSize + 8 + 16 + 1 + 255
	minSize := 1 + checksumSize + 8 + 16 + 1
	buf := make([]byte, maxSize)

	rn, err := io.ReadAtLeast(conn, buf, minSize)
	if err != nil {
		return "", nil, err
	}

	b = buf[0]
	checksum := buf[1 : 1+checksumSize]
	timestamp := binary.BigEndian.Uint64(buf[1+checksumSize : 1+checksumSize+8])
	randKey = buf[1+checksumSize+8 : 1+checksumSize+8+16]
	nameLen := int(buf[minSize-1])

	if nameLen <= 0 {
		return "", nil, errors.New("empty name error")
	}

	bufSize := minSize + nameLen
	if rn > bufSize {
		return "", nil, errors.New("data length error")
	}
	if rn < bufSize {
		_, err := io.ReadFull(conn, buf[rn:bufSize])
		if err != nil {
			return "", nil, errors.New("read name failed")
		}
	}
	name = string(buf[minSize:bufSize])

	sum := calcSum(secret, name, b, timestamp, randKey)
	if false == bytes.Equal(sum, checksum) {
		return "", nil, errors.New("checksum invalid")
	}

	return name, randKey, nil
}

// RequireAuth 请求校验连接
//   client：socket连接
//   secret：与服务端约定的密钥串，用于创建签名
//   name：客户端的名称，服务端不允许多个相同的name的客户端
//   randKey：随机key，用于后续AES加解密，长度固定位16位
func RequireAuth(client net.Conn, secret, name string) (randKey []byte, err error) {
	b := byte('a' + rand.Intn(int('z'-'a')))
	timestamp := uint64(time.Now().Unix())

	randKey = make([]byte, 16)
	_, err = rand.Read(randKey)
	if err != nil {
		log.Get().WithError(err).Error("make auth randKey failed")
		return
	}

	payload, err := makeAuthData(secret, name, b, timestamp, randKey)
	if err != nil {
		log.Get().WithError(err).Error("make auth payload failed")
		return
	}

	_, err = client.Write(payload)
	if err != nil {
		log.Get().WithError(err).Error("send auth payload failed")
		return
	}

	resp := make([]byte, 2)
	_, err = io.ReadFull(client, resp)
	if err != nil {
		log.Get().WithError(err).Error("read auth response failed")
		return
	}

	if resp[0] != b || resp[1] != 0 {
		return nil, errors.New("auth failed, please check your secret")
	}

	return randKey, nil
}
