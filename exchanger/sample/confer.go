package main

import (
	"crypto/sha256"
	"io"
)

var mixRandSalt = []byte("salt-string-for-hash")

func conferReqOnServer(wr io.ReadWriter, secret string) ([]byte, error) {
	var req authRequest
	var err error
	if _, err = req.ReadFrom(wr); err != nil {
		return nil, err
	}
	if !req.Verify(secret) {
		return nil, err
	}

	var resp authResponse
	if err = resp.Gen(0); err != nil {
		return nil, err
	}
	if _, err = resp.WriteTo(wr); err != nil {
		return nil, err
	}

	return mixKey(secret, &req, &resp)
}

func conferKeyOnClient(wr io.ReadWriter, secret string) ([]byte, error) {
	var req authRequest
	var err error
	if err = req.Gen(secret); err != nil {
		return nil, err
	}
	if _, err = req.WriteTo(wr); err != nil {
		return nil, err
	}

	var resp authResponse
	if _, err = resp.ReadFrom(wr); err != nil {
		return nil, err
	}

	return mixKey(secret, &req, &resp)
}

func mixKey(secret string, req *authRequest, resp *authResponse) ([]byte, error) {
	buf := []byte{}
	buf = append(buf, req.hashed...)
	buf = append(buf, resp.challenge[:]...)
	buf = append(buf, []byte(secret)...)
	buf = append(buf, []byte(mixRandSalt)...)

	h := sha256.New()
	if _, err := h.Write(buf); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
