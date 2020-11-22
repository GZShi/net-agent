package tunnel

import (
	"net"
	"sync"
	"sync/atomic"
)

// Tunnel 通道协议
type Tunnel interface {
	Run() error
	NewStream() (Stream, uint32)
	SendJSON(string, interface{}, interface{}) error
	SendText(string, string) (string, error)

	Listen(string, OnRequestFunc)
}

// New 创建
func New(conn net.Conn) Tunnel {
	return &tunnel{
		idSequece: 0,
		_conn:     conn,
	}
}

// Server 双向数据传输服务
type tunnel struct {
	idSequece    uint32
	_conn        net.Conn
	respGuards   sync.Map
	streamGuards sync.Map
	cmdFuncMap   map[string]OnRequestFunc

	writerLock sync.Mutex
}

type frameGuard struct {
	ch chan *Frame
}

// NewID 生成不断自增的唯一ID
func (t *tunnel) NewID() uint32 {
	return atomic.AddUint32(&t.idSequece, 1)
}

// Run 执行读取程序，不断解析收到的数据包
func (t *tunnel) Run() error {

	var err error
	for {
		frame := &Frame{}
		_, err = frame.ReadFrom(t._conn)
		if err != nil {
			return err
		}
		switch frame.Type {
		case FrameRequest:
			go t.onRequest(frame)
		case FrameResponseOK, FrameResponseErr:
			go t.onResponse(frame)
		case FrameStreamData:
			t.onStreamData(frame)
		}
	}
}
