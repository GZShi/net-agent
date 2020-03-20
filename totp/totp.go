package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Time-based One-Time Password
// Google Authenticator: sha1, 6
// Microsoft Authenticator: sha1, 6
func getTOTP(hashFunc func() hash.Hash, codeLen uint64, secret string, counter uint64) (string, error) {
	// 处理密钥
	secret = strings.TrimSpace(secret)
	secret = strings.ToUpper(secret)
	if n := len(secret) % 8; n != 0 {
		secret = secret + strings.Repeat("=", 8-n)
	}
	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	// 处理时间
	counterBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuf, counter)

	// 计算
	mac := hmac.New(hashFunc, secretBytes)
	mac.Write(counterBuf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	value := int64(((int(sum[offset]) & 0x7f) << 24) |
		((int(sum[offset+1] & 0xff)) << 16) |
		((int(sum[offset+2] & 0xff)) << 8) |
		(int(sum[offset+3]) & 0xff))

	// 输出为字符串
	format := fmt.Sprintf("%%0%dd", codeLen)
	code := fmt.Sprintf(format, value)
	custCodeLen := uint64(len(code))
	if custCodeLen > codeLen {
		code = code[custCodeLen-codeLen : custCodeLen]
	}
	return code, nil
}

func getTimeCounter(period uint64) uint64 {
	return uint64(math.Floor(float64(time.Now().Unix()) / float64(period)))
}

// GetPassCode 获取一次性验证码，同时返回当前时刻与上一时刻的验证码
func GetPassCode(secret string) ([]string, error) {
	hashFunc := sha1.New
	counter := getTimeCounter(30)

	curr, err := getTOTP(hashFunc, 6, secret, counter)
	if err != nil {
		return nil, err
	}
	prev, err := getTOTP(hashFunc, 6, secret, counter-1)
	if err != nil {
		return nil, err
	}

	return []string{curr, prev}, nil
}

// GetSecretURL 获取用于保存在客户端的URL
func GetSecretURL(account string, secret string) string {
	protocol := "otpauth"
	issuer := "Tunnel-TXY"
	return fmt.Sprintf("%s://totp/%s:%s?secret=%s&issuer=%s",
		protocol, issuer, account, secret, issuer)
}

// GenSecret 生成密钥
func GenSecret(size uint64) string {
	bytes := make([]byte, 20)
	for i := range bytes {
		bytes[i] = byte(rand.Intn(256))
	}
	return base32.StdEncoding.EncodeToString(bytes)
}
