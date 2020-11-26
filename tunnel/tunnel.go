package tunnel

import (
	"net"
	"sync"
	"sync/atomic"
)

// Tunnel 通道协议
type Tunnel interface {
	Run() error
	Stop() error
	BindService(s Service) error
	Ready(func(t Tunnel))

	NewStream() (Stream, uint32)
	SendJSON(Context, string, interface{}, interface{}) error
	SendText(Context, string, string) (string, error)
}

// New 创建
func New(conn net.Conn) Tunnel {
	return &tunnel{
		idSequece: 1,
		_conn:     conn,
	}
}

// Server 双向数据传输服务
type tunnel struct {
	idSequece    uint32
	_conn        net.Conn
	respGuards   sync.Map
	streamGuards sync.Map
	serviceMap   map[string]Service

	writerLock sync.Mutex

	readyFunc func(t Tunnel)
}

type frameGuard struct {
	ch        chan *Frame
	sessionID uint32
}

// NewID 生成不断自增的唯一ID
func (t *tunnel) NewID() uint32 {
	return atomic.AddUint32(&t.idSequece, 1)
}

func (t *tunnel) Ready(fn func(Tunnel)) {
	t.readyFunc = fn
}

// Run 执行读取程序，不断解析收到的数据包
func (t *tunnel) Run() error {
	if t.readyFunc != nil {
		// todo:0
		// 在tunnel.Run执行之前，如果进行rpc调用，有可能会无法接收到回包
		// 需要排查，目前还未找到原因
		go t.readyFunc(t)
	}
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

func (t *tunnel) Stop() error {
	return t._conn.Close()
}
