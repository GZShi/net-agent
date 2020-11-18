package tunnel

import (
	"net"
	"sync"
	"sync/atomic"

	log "github.com/GZShi/net-agent/logger"
)

type responseGuard struct {
	response *Frame
	c        sync.Cond
}

type streamGuard struct {
	ch chan *Frame
}

type Server struct {
	idSequece    uint32
	conn         net.Conn
	respGuards   sync.Map
	streamGuards sync.Map
}

func NewServer(conn net.Conn) *Server {
	return &Server{
		idSequece: 0,
		conn:      conn,
	}
}

func (s *Server) NewID() uint32 {
	return atomic.AddUint32(&s.idSequece, 1)
}

func (s *Server) Run() error {

	var err error
	for {
		frame := &Frame{}
		_, err = frame.ReadFrom(s.conn)
		if err != nil {
			return err
		}
		switch frame.Type {
		case FrameRequest:
			go s.onRequest(frame)
		case FrameResponse:
			go s.onResponse(frame)
		case FrameStreamData:
			s.onStreamData(frame)
		}
	}
}

// request 发送一个RequestFrame，并等待对端返回一个ResponseFrame
func (s *Server) request(req *Frame) (*Frame, error) {
	guard := &responseGuard{}
	s.respGuards.Store(req.ID, guard)

	_, err := req.WriteTo(s.conn)
	if err != nil {
		s.respGuards.Delete(req.ID)
		return nil, err
	}

	guard.c.Wait()
	return guard.response, nil
}

// onRequest 接收到对端的一个RequestFrame
func (s *Server) onRequest(req *Frame) {
	// process frame

	// response
	resp := &Frame{
		ID:        0,
		Type:      FrameResponse,
		SessionID: req.ID,
		Header:    nil,
		DataType:  BinaryData,
		Data:      nil,
	}

	_, err := resp.WriteTo(s.conn)
	if err != nil {
		log.Get().WithError(err).Error("write response failed")
	}
}

// onResponse 接收到对端的一个Response
func (s *Server) onResponse(f *Frame) {
	val, has := s.respGuards.Load(f.SessionID)
	if !has {
		// 丢弃不用
		log.Get().WithField("sessionID", f.SessionID).Error("can't find responseGuard")
		return
	}
	// Todo: go1.15中提供了LoadAndDelete方法
	s.respGuards.Delete(f.SessionID)

	guard := val.(*responseGuard)
	guard.response = f
	guard.c.Signal()
}

// onSteramData 接收到对端的一个数据传输包
func (s *Server) onStreamData(f *Frame) {
	guard := &streamGuard{
		ch: make(chan *Frame, 256),
	}
	val, loaded := s.streamGuards.LoadOrStore(f.SessionID, guard)
	if loaded {
		guard = val.(*streamGuard)
	}

	guard.ch <- f
}
