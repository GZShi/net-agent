package tunnel

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Tunnel 通道协议
type Tunnel interface {
	Run() error
	Stop() error
	BindServices(s ...Service) error
	Ready(func(t Tunnel))

	NewStream() (Stream, uint32)
	GetStreamStates() []StreamState
	FindStreamBySID(uint32) (Stream, error)
	SendJSON(Context, string, interface{}, interface{}) error
	SendText(Context, string, string) (string, error)
	Ping() error

	// net interface
	Listen(virtualPort uint32) (net.Listener, error)
	Dial(virtualPort uint32) (net.Conn, error)
}

// New 创建
func New(conn net.Conn, logStreamState bool) Tunnel {
	return &tunnel{
		idSequece:      1,
		_conn:          conn,
		pongChan:       make(chan int, 5),
		logStreamState: logStreamState,
	}
}

// Server 双向数据传输服务
type tunnel struct {
	idSequece    uint32
	_conn        net.Conn
	respGuards   sync.Map
	dialGuards   sync.Map
	streamGuards sync.Map
	serviceMap   map[string]Service

	writerLock sync.Mutex

	readyFunc    func(t Tunnel)
	acceptGuards sync.Map
	pongChan     chan int

	logStreamState bool
	historyStreams []*streamRWC
	historyLock    sync.RWMutex
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
		if frame.Type == FrameStreamData {
			t.onStreamData(frame)
		} else {
			go func(f *Frame) {
				switch f.Type {
				case FrameRequest:
					t.onRequest(f)
				case FrameResponseOK, FrameResponseErr:
					t.onResponse(f)
				case FrameDialRequest:
					t.onDial(f)
				case FrameDialResponse:
					t.onDialResponse(f)
				case FramePing:
					resp := t.NewFrame(FramePong)
					w := t.NewWriteCloser()
					resp.WriteTo(w)
					w.Close()
				case FramePong:
					t.pongChan <- 0
				}
			}(frame)
		}
	}
}

func (t *tunnel) Ping() error {
	req := t.NewFrame(FramePing)
	w := t.NewWriteCloser()
	_, err := req.WriteTo(w)
	w.Close()

	if err != nil {
		return err
	}

	select {
	case <-t.pongChan:
	case <-time.After(time.Second * 3):
		return errors.New("wait pong timeout")
	}

	return nil
}

func (t *tunnel) Stop() error {
	return t._conn.Close()
}
