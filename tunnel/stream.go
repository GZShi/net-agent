package tunnel

import (
	"errors"
	"io"

	log "github.com/GZShi/net-agent/logger"
)

type streamRW struct {
	writeSessionID uint32
	readSessionID  uint32
	server         *Server
	guard          *streamGuard
	readingFrame   *Frame
	readingPos     int
}

// NewStreamRW 根据指定Session创建流式数据通道
func (s *Server) NewStreamRW(readSessionID, writeSessionID uint32) io.ReadWriter {
	guard := &streamGuard{
		ch: make(chan *Frame, 256),
	}
	val, loaded := s.streamGuards.LoadOrStore(readSessionID, guard)
	if loaded {
		guard = val.(*streamGuard)
	}
	stream := &streamRW{
		readSessionID:  readSessionID,
		writeSessionID: writeSessionID,
		server:         s,
		guard:          guard,
		readingFrame:   nil,
		readingPos:     0,
	}
	return stream
}

func (stream *streamRW) Read(buf []byte) (int, error) {
	if stream.readingFrame == nil {
		stream.readingPos = 0
		stream.readingFrame = <-stream.guard.ch
		if stream.readingFrame == nil {
			return 0, errors.New("stream closed")
		}
		log.Get().WithField("d", stream.readingFrame).Debug("recv data frame")
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
		ID:        stream.server.NewID(),
		Type:      FrameStreamData,
		SessionID: stream.writeSessionID,
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
