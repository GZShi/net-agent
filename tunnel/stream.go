package tunnel

import (
	"errors"
	"io"
)

type streamRW struct {
	sessionID    uint32
	server       *Server
	guard        *streamGuard
	readingFrame *Frame
	readingPos   int
}

// NewStreamRW 根据指定Session创建流式数据通道
func (s *Server) NewStreamRW(SessionID uint32) io.ReadWriter {
	guard := &streamGuard{
		ch: make(chan *Frame, 256),
	}
	stream := &streamRW{
		sessionID:    SessionID,
		server:       s,
		guard:        guard,
		readingFrame: nil,
		readingPos:   0,
	}

	s.streamGuards.Store(SessionID, stream.guard)
	return stream
}

func (stream *streamRW) Read(buf []byte) (int, error) {
	if stream.readingFrame == nil {
		stream.readingPos = 0
		stream.readingFrame = <-stream.guard.ch
		if stream.readingFrame == nil {
			return 0, errors.New("stream closed")
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

func (stream *streamRW) Write(buf []byte) (int, error) {
	frame := &Frame{
		ID:        0,
		Type:      FrameStreamData,
		SessionID: stream.sessionID,
		Header:    nil,
		DataType:  BinaryData,
		Data:      buf,
	}

	wn, err := frame.WriteTo(stream.server.conn)

	written := int(wn - frameStableBufSize)
	if written < 0 {
		written = 0
	}

	return written, err
}
