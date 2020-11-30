package cipherconn

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestIvdata(t *testing.T) {
	data := newIvData()
	data.code = 0x03
	data.iv = make([]byte, data.ivLen)

	_, err := rand.Read(data.iv)
	if err != nil {
		t.Error(err)
		return
	}

	buf := bytes.NewBuffer(nil)

	_, err = data.WriteTo(buf)
	if err != nil {
		t.Error(err)
		return
	}

	d2 := newIvData()
	_, err = d2.ReadFrom(buf)
	if err != nil {
		t.Error(err)
		return
	}

	if data.code != d2.code {
		t.Error("data.code:", data.code, " d2.code:", d2.code)
		return
	}

	if !bytes.Equal(data.iv, d2.iv) {
		t.Error("not equal")
		return
	}

	if data.ivLen != d2.ivLen {
		t.Error("data ivLen not equal")
		return
	}
}
