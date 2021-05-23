package tunnel

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/GZShi/net-agent/v2/logger"
)

// Stream 数据通道流
type Stream interface {
	io.ReadWriteCloser
	net.Conn
	Bind(sessionID uint32) error
	SetInfo(info string)
	Info() string
	Cache(f *Frame)
}

const (
	streamWarningLevel = 500  // 警告位：readCh队列长度超过这个值将会进行提示
	streamFusingLevel  = 1000 // 熔断位：readCh队列长度超过这个值，将会触发熔断
	streamChanSize     = 1200
)

type streamRWC struct {
	t              *tunnel
	readSessionID  uint32
	writeSessionID uint32
	readCh         chan *Frame
	readingFrame   *Frame
	readingPos     int
	closed         bool
	info           string
	isStayAlert    bool
	closeOnce      sync.Once
}

// NewStream 根据指定Session创建流式数据通道
func (t *tunnel) NewStream() (Stream, uint32) {
	sid := t.NewID()
	stream := &streamRWC{
		readSessionID:  sid,
		writeSessionID: 0,
		t:              t,
		readCh:         make(chan *Frame, streamChanSize),
		readingFrame:   nil,
		readingPos:     0,
		closed:         false,
		isStayAlert:    false,
	}
	_, loaded := t.streamGuards.LoadOrStore(sid, stream)
	if loaded {
		panic("unexpceted stream stored")
	}
	return stream, sid
}

func (t *tunnel) FindStreamBySID(sid uint32) (Stream, error) {
	val, loaded := t.streamGuards.Load(sid)
	if !loaded {
		return nil, fmt.Errorf("stream not found, sid=%v", sid)
	}
	return val.(Stream), nil
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
	// 处于积压状态时，优先把readChan中的数据消费掉
	if stream.isStayAlert || stream.readingFrame == nil {
		var head *Frame
		select {
		case head = <-stream.readCh:
		case <-time.After(time.Minute * 10):
			// 如果超过10分钟都无法读取到数据，则这个连接很可能要关掉了
			// 这个超时时间不能太短，少于多数应用的心跳包将导致长连接应用频繁断线重连
			stream.Close()
			return 0, errors.New("read stream timeout")
		}

		if head == nil {
			stream.Close()
			return 0, errors.New("read a closed stream")
		}

		if head.Data == nil {
			stream.Close()
			return 0, io.EOF
		}

		lenOfReadCh := len(stream.readCh)
		var f *Frame
		for i := 0; i < lenOfReadCh; i++ {
			f = nil
			select {
			case f = <-stream.readCh:
			case <-time.After(time.Microsecond * 3):
			}
			if f != nil {
				if f.Data == nil {
					defer stream.Close()
				} else {
					head.Data = append(head.Data, f.Data...)
				}
			}
		}

		if stream.isStayAlert {
			log.Get().Info("merged size: ", len(head.Data))
			stream.isStayAlert = false
		}

		if stream.readingFrame == nil {
			stream.readingFrame = head
			stream.readingPos = 0
		} else {
			stream.readingFrame.Data = append(stream.readingFrame.Data, head.Data...)
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
	frame := stream.t.NewFrame(FrameStreamData)
	frame.SessionID = stream.writeSessionID
	frame.Data = buf

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
	stream.closeOnce.Do(func() {
		stream.close()
	})
	return nil
}

func (stream *streamRWC) close() error {
	if stream.closed {
		return errors.New("stream closed")
	}
	stream.closed = true

	// 向对端发送一个EOF，可能成功也可能失败
	frame := stream.t.NewFrame(FrameStreamData)
	frame.SessionID = stream.writeSessionID

	wc := stream.t.NewWriteCloser()
	frame.WriteTo(wc)
	wc.Close()

	// 清除guard，停止接收数据
	stream.t.streamGuards.Delete(stream.readSessionID)
	stream.t = nil

	// 清空channels
	for len(stream.readCh) > 0 {
		<-stream.readCh
	}
	close(stream.readCh)
	stream.readCh = nil

	stream.readingFrame = nil

	return nil
}

// SetInfo 设置连接信息，用于错误输出
func (stream *streamRWC) SetInfo(info string) {
	stream.info = info
}

// Info 获取信息
func (stream *streamRWC) Info() string {
	return stream.info
}

func (stream *streamRWC) Cache(f *Frame) {
	// 如果当前guard深度过长，则消费端可能出现了阻塞的情况
	// 需要关闭连接，否则会阻塞其它Stream正常传输
	if len(stream.readCh) > streamWarningLevel && !stream.isStayAlert {
		stream.isStayAlert = true // 超过了警告线，进入戒备状态，随时准备熔断
		go func(stream *streamRWC) {
			log.Get().Warn("stream will fusing: ", stream.Info())
			<-time.After(time.Second * 3)
			if len(stream.readCh) > streamFusingLevel {
				stream.t.streamGuards.Delete(stream.readSessionID)
				log.Get().Error("stream already fusing: ", stream.Info())
				stream.Close()
			}
		}(stream)
	}
	stream.readCh <- f
}

// todo:4 stream的生命周期管理，超时、关闭。流速控制
