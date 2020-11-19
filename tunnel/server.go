package tunnel

import (
	"io"
	"net"
	"sync"
	"sync/atomic"

	log "github.com/GZShi/net-agent/logger"
)

type responseGuard struct {
	ch chan *Frame
}

type streamGuard struct {
	ch chan *Frame
}

type writeCloser struct {
	server *Server
}

func (w *writeCloser) Write(buf []byte) (int, error) {
	return w.server._conn.Write(buf)
}

func (w *writeCloser) Close() error {
	w.server.writerLock.Unlock()
	return nil
}

// NewWriteCloser 请避免直接使用_conn对象进行写入，会产生时序错乱问题
// 当需要调用原始conn连接写入数据时，需要创建临时的WriteCloser
// 创建时会请求对原始连接进行上锁
// 写入数据完毕后需要手动调用Close，释放锁
func (s *Server) NewWriteCloser() io.WriteCloser {
	s.writerLock.Lock()
	return &writeCloser{s}
}

// Server 双向数据传输服务
type Server struct {
	idSequece    uint32
	_conn        net.Conn
	respGuards   sync.Map
	streamGuards sync.Map
	cmdFuncMap   map[string]OnRequestFunc

	writerLock sync.Mutex
}

// NewServer 创建
func NewServer(conn net.Conn) *Server {
	return &Server{
		idSequece: 0,
		_conn:     conn,
	}
}

// NewID 生成不断自增的唯一ID
func (s *Server) NewID() uint32 {
	return atomic.AddUint32(&s.idSequece, 1)
}

// Run 执行读取程序，不断解析收到的数据包
func (s *Server) Run() error {

	var err error
	for {
		frame := &Frame{}
		_, err = frame.ReadFrom(s._conn)
		if err != nil {
			return err
		}
		switch frame.Type {
		case FrameRequest:
			go s.onRequest(frame)
		case FrameResponseOK, FrameResponseErr:
			go s.onResponse(frame)
		case FrameStreamData:
			s.onStreamData(frame)
		}
	}
}

// request 发送一个RequestFrame，并等待对端返回一个ResponseFrame
func (s *Server) request(req *Frame) (*Frame, error) {
	guard := &responseGuard{
		ch: make(chan *Frame),
	}
	s.respGuards.Store(req.ID, guard)

	w := s.NewWriteCloser()
	_, err := req.WriteTo(w)
	w.Close()
	if err != nil {
		s.respGuards.Delete(req.ID)
		return nil, err
	}

	resp := <-guard.ch
	return resp, nil
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
	guard.ch <- f
	close(guard.ch)
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
