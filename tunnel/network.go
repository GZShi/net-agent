package tunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

//
// net.Listen implement
//
func (t *tunnel) Listen(virtualPort uint32) (net.Listener, error) {
	l := newListener(t, virtualPort)
	_, loaded := t.acceptGuards.LoadOrStore(virtualPort, l)
	if loaded {
		return nil, errors.New("listen failed, v-port used")
	}

	return l, nil
}

type listener struct {
	t        *tunnel
	port     uint32
	streamCh chan net.Conn
	network  string
	host     string
}

func newListener(t *tunnel, port uint32) *listener {
	return &listener{
		t:        t,
		port:     port,
		streamCh: make(chan net.Conn, 128),
		network:  "tcp4",
		host:     "virtualhost",
	}
}

func (l *listener) Accept() (net.Conn, error) {
	conn, ok := <-l.streamCh
	if !ok {
		return nil, errors.New("listener closed")
	}
	return conn, nil
}

func (l *listener) Close() error {
	l.t.acceptGuards.Delete(l.port)
	close(l.streamCh)
	return nil
}

func (l *listener) Addr() net.Addr {
	return nil
}

func (l *listener) Network() string {
	return l.network
}

func (l *listener) String() string {
	return fmt.Sprintf("%v:%v", l.host, l.port)
}

//
// net.Dial implement
//
func (t *tunnel) Dial(virtualPort uint32) (net.Conn, error) {
	stream, _ := t.NewStream()

	buf := make([]byte, 4)

	binary.BigEndian.PutUint32(buf, virtualPort)
	if _, err := stream.Write(buf); err != nil {
		stream.Close()
		return nil, err
	}

	if _, err := io.ReadFull(stream, buf); err != nil {
		stream.Close()
		return nil, err
	}

	if err := stream.Bind(binary.BigEndian.Uint32(buf)); err != nil {
		return nil, err
	}

	return stream, nil
}
