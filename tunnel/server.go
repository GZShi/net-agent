package tunnel

import (
	"net"
	"sync"
	"sync/atomic"
)

type frameGuard struct {
	ch chan *Frame
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
