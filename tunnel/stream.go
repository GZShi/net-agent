package tunnel

import (
	"errors"
	"io"
	"net"
)

// Stream 数据通道流
type Stream interface {
	io.ReadWriteCloser
	net.Conn
	Bind(sessionID uint32) error
}

type streamRWC struct {
	t              *tunnel
	readSessionID  uint32
	writeSessionID uint32
	guard          *frameGuard
	readingFrame   *Frame
	readingPos     int
	closed         bool
}

// NewStream 根据指定Session创建流式数据通道
func (t *tunnel) NewStream() (Stream, uint32) {
	sid := t.NewID()
	guard := &frameGuard{
		ch: make(chan *Frame, 256),
	}
	val, loaded := t.streamGuards.LoadOrStore(sid, guard)
	if loaded {
		guard = val.(*frameGuard)
	}
	stream := &streamRWC{
		readSessionID:  sid,
		writeSessionID: 0,
		t:              t,
		guard:          guard,
		readingFrame:   nil,
		readingPos:     0,
		closed:         false,
	}
	return stream, sid
}

func (stream *streamRWC) Bind(sessionID uint32) error {
	if sessionID == 0 {
		return errors.New("invalid session id")
	}
	if stream.writeSessionID != 0 {
		return errors.New("repeat bind")
	}
	stream.writeSessionID = sessionID
	return nil
}

func (stream *streamRWC) Read(buf []byte) (int, error) {
	if stream.closed {
		return 0, errors.New("stream closed")
	}
	if stream.readingFrame == nil {
		stream.readingPos = 0
		stream.readingFrame = <-stream.guard.ch
		if stream.readingFrame == nil {
			return 0, errors.New("read from closed stream")
		}
		if stream.readingFrame.Data == nil {
			// 收到nil数据，代表收到了EOF，此时可以把guard安全delete了
			stream.t.streamGuards.Delete(stream.readSessionID)
			return 0, io.EOF
		}
	}

	f := stream.readingFrame
	remainSize := len(f.Data) - stream.readingPos
	if len(buf) < remainSize {
		start := stream.readingPos
		end := start + len(buf)
		copy(buf, f.Data[start:end])
		stream.readingPos = end
		return len(buf), nil
	}

	copy(buf, f.Data[stream.readingPos:])
	stream.readingFrame = nil
	stream.readingPos = 0
	return remainSize, nil
}

func (stream *streamRWC) Write(buf []byte) (int, error) {
	if stream.closed {
		return 0, errors.New("stream closed")
	}
	if stream.writeSessionID == 0 {
		return 0, errors.New("write session id is 0")
	}
	frame := &Frame{
		ID:        stream.t.NewID(),
		Type:      FrameStreamData,
		SessionID: stream.writeSessionID,
		Header:    nil,
		DataType:  BinaryData,
		Data:      buf,
	}

	wc := stream.t.NewWriteCloser()
	wn, err := frame.WriteTo(wc)
	wc.Close()

	written := int(wn - frameStableBufSize)
	if written < 0 {
		written = 0
	}

	return written, err
}

func (stream *streamRWC) Close() error {
	if stream.closed {
		return errors.New("stream closed")
	}
	stream.closed = true

	// 向对端发送一个EOF，可能成功也可能失败
	frame := &Frame{
		ID:        stream.t.NewID(),
		Type:      FrameStreamData,
		SessionID: stream.writeSessionID,
		Header:    nil,
		DataType:  BinaryData,
		Data:      nil,
	}
	wc := stream.t.NewWriteCloser()
	frame.WriteTo(wc)
	wc.Close()

	// 清除guard，停止接收数据
	stream.t.streamGuards.Delete(stream.readSessionID)

	return nil
}

// todo: stream的生命周期管理，超时、关闭
