package tunnel

import (
	"net"
)

func makePipe() (*Server, *Server) {
	send, recv := net.Pipe()

	s1 := NewServer(send)
	s2 := NewServer(recv)

	go s1.Run()
	go s2.Run()

	return s1, s2
}
