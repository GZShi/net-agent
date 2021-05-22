package tunnel

import (
	"errors"
	"net"
	"time"
)

// Network todo:5
func (s *streamRWC) Network() string {
	return "tcp4"
}

// String todo:制定标准
func (s *streamRWC) String() string {
	return "127.0.0.1:65535"
}

func (s *streamRWC) LocalAddr() net.Addr {
	return s
}

func (s *streamRWC) RemoteAddr() net.Addr {
	return s
}

func (s *streamRWC) SetDeadline(t time.Time) error {
	// todo:5
	return errors.New("streamRWC.SetDeadline not implemented")
}

func (s *streamRWC) SetReadDeadline(t time.Time) error {
	// todo:5
	return errors.New("streamRWC.SetReadDeadline not implemented")
}
func (s *streamRWC) SetWriteDeadline(t time.Time) error {
	// todo:5
	return errors.New("streamRWC.SetWriteDeadline not implemented")
}
