package tunnel

import (
	"net"
)

func makePipe() (Tunnel, Tunnel) {
	send, recv := net.Pipe()

	s1 := New(send)
	s2 := New(recv)

	go s1.Run()
	go s2.Run()

	return s1, s2
}
