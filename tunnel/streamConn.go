package tunnel

import (
	"errors"
	"net"
	"time"
)

func (s *streamRWC) LocalAddr() net.Addr {
	// todo:5
	return nil
}

func (s *streamRWC) RemoteAddr() net.Addr {
	// todo:5
	return nil
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
