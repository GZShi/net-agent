package cipherconn

import (
	"bytes"
	"crypto/rand"
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
