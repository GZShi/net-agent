package socks5

import (
	"bytes"
	"net"
	"testing"
)

func TestRequestReadFrom(t *testing.T) {
	payloads := []*request{
		{5, 2, IPv4, []byte(net.IPv4(127, 0, 0, 1)), 1024},
	}

	for i, p := range payloads {
		r, err := copyRequest(p)
		if err != nil {
			t.Error(err)
			return
		}
		if !equal(r, p) {
			t.Error("not equal", i, "payload:", p, "copied:", r)
			return
		}
	}

}

func copyRequest(req *request) (*request, error) {
	buf := bytes.NewBuffer(nil)
	_, err := req.WriteTo(buf)
	if err != nil {
		return nil, err
	}

	b := buf.Bytes()
	r := bytes.NewReader(b)

	var reply request
	reply.ReadFrom(r)

	return &reply, nil
}

func equal(r1 *request, r2 *request) bool {
	if r1 == r2 {
		return true
	}
	if r1 == nil || r2 == nil {
		return false
	}
	return r1.version == r2.version && r1.command == r2.command &&
		r1.addressType == r2.addressType &&
		r1.port == r2.port &&
		r1.GetAddrPortStr() == r2.GetAddrPortStr()
}
