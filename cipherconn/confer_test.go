package cipherconn

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestMakeCipherStream(t *testing.T) {
	password := "pas50rd"
	iv := make([]byte, 1024)
	_, err := rand.Read(iv)
	if err != nil {
		t.Error(err)
		return
	}

	enc, err := makeCipherStream(password, iv[0:16])
	dec, err := makeCipherStream(password, iv[0:16])

	d1 := make([]byte, 1024)
	_, err = rand.Read(d1)
	if err != nil {
		t.Error(err)
		return
	}
	d2 := make([]byte, 1024)
	d3 := make([]byte, 1024)

	enc.XORKeyStream(d2, d1)
	dec.XORKeyStream(d3, d2)

	if !bytes.Equal(d3, d1) {
		t.Error("not equal")
		return
	}
}

func TestHkdfsha1(t *testing.T) {
	buf := make([]byte, 10)
	if err := hkdfSha1([]byte("hello"), buf); err != nil {
		t.Error(err)
		return
	}
	want := []byte{0x12, 0x70, 0x98, 0x8B, 0xE9, 0x6F, 0x2E, 0xB1, 0xAD, 0x44}
	if !bytes.Equal(buf, want) {
		t.Error("not equal", buf, want)
		return
	}
}

func TestCipherStream(t *testing.T) {
	enc, err := makeCipherStream("1234", []byte("1234567812345678"))
	if err != nil {
		t.Error(err)
		return
	}
	content := []byte("helloworld")
	buf := make([]byte, len(content))

	enc.XORKeyStream(buf, content)
	out := hex.EncodeToString(buf)
	want := "8367265500129b524ce8"
	if out != want {
		t.Error("not equal", out, want)
		return
	}
}
