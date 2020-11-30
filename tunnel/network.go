package tunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
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
	return l
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
func (t *tunnel) Dial(virtualPort uint32) (conn net.Conn, err error) {
	stream, readSID := t.NewStream()
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], virtualPort)
	binary.BigEndian.PutUint32(buf[4:8], readSID)

	sid := t.NewID()
	guard := make(chan *Frame)
	t.dialGuards.Store(sid, guard)
	defer func() {
		// 先执行Delete，再close
		t.dialGuards.Delete(sid)
		close(guard)
		guard = nil
	}()

	dialRequest := t.NewFrame(FrameDialRequest)
	dialRequest.SessionID = sid
	dialRequest.DataType = BinaryData
	dialRequest.Data = buf

	w := t.NewWriteCloser()
	_, err = dialRequest.WriteTo(w)
	w.Close()
	if err != nil {
		return nil, err
	}

	var respFrame *Frame
	var ok bool
	err = nil
	select {
	case respFrame, ok = <-guard:
		if !ok {
			err = errors.New("wait dial response failed")
		}
	case <-time.After(time.Second * 5):
		err = errors.New("wait dial response timeout")
	}
	if err != nil {
		return nil, err
	}

	if respFrame.DataType != BinaryData {
		if respFrame.DataType == TextData {
			return nil, fmt.Errorf(string(respFrame.Data))
		}
		return nil, errors.New("response data decode failed")
	}

	if respFrame.Data == nil || len(respFrame.Data) != 4 {
		return nil, errors.New("length of response data invalid")
	}

	writeSID := binary.BigEndian.Uint32(respFrame.Data)
	if err := stream.Bind(writeSID); err != nil {
		return nil, err
	}

	return stream, nil
}

func (t *tunnel) onDial(frame *Frame) {
	resp := t.NewFrame(FrameDialResponse)
	resp.SessionID = frame.SessionID

	if frame.DataType != BinaryData || frame.Data == nil || len(frame.Data) != 8 {
		// logo error
		// dial failed
		resp.DataType = TextData
		resp.Data = []byte("decode dial request failed")
		w := t.NewWriteCloser()
		resp.WriteTo(w)
		w.Close()
		return
	}

	virtualPort := binary.BigEndian.Uint32(frame.Data[0:4])
	writeSID := binary.BigEndian.Uint32(frame.Data[4:8])

	val, loaded := t.acceptGuards.Load(virtualPort)
	if !loaded {
		resp.DataType = TextData
		resp.Data = []byte("connect to virtualPort failed")
		w := t.NewWriteCloser()
		resp.WriteTo(w)
		w.Close()
		return
	}

	// get writeSID
	stream, readSID := t.NewStream()
	stream.Bind(writeSID)

	resp.Data = make([]byte, 4)
	binary.BigEndian.PutUint32(resp.Data, readSID)

	w := t.NewWriteCloser()
	_, err := resp.WriteTo(w)
	w.Close()
	if err != nil {
		// todo: logo net error
		stream.Close()
		return
	}

	// 放入监听队列中
	val.(*listener).streamCh <- stream
}

func (t *tunnel) onDialResponse(frame *Frame) {
	defer func() {
		if r := recover(); r != nil {
			log.Get().Error("recoverd:", r)
		}
	}()

	val, loaded := t.dialGuards.Load(frame.SessionID)
	if !loaded {
		// 没有找到
		log.Get().Warn("dial guard not found")
		return
	}

	// may closed and panic
	val.(chan *Frame) <- frame
}
