package socks5

import "testing"

func TestServe(t *testing.T) {
	s := NewServer()
	s.ListenAndRun("127.0.0.1:20034")
}
