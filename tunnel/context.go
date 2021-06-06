package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/GZShi/net-agent/v2/logger"
)

// Context 用于处理RPC请求的上下文对象
type Context interface {
	// context
	GetTunnel() Tunnel
	GetCallStackStr(Caller) (string, error)

	// for route
	GetCmd() string     // service.method
	GetService() string // service
	GetMethod() string  // method

	// for request data
	GetData() []byte
	GetJSON(v interface{}) error
	GetText() (string, error)

	// for response data
	JSON(v interface{})
	Text(string)
	Data([]byte)
	Error(error)
	Flush()
}

// OnRequestFunc 处理请求的回调函数
type OnRequestFunc func(Context)

//
// context implement
//

type context struct {
	tunnel      *tunnel           //传输通道
	req         *Frame            //原始请求帧
	header      map[string]string //从原始帧解析出的头信息
	command     string            // 请求的完整路由信息
	serviceName string            // 从完整路由中解析出的服务名
	methodName  string            // 从完整路由中解析出的方法名
	caller      []Caller          // 调用链信息
	resp        *Frame            // 应答帧信息
	respChan    chan *Frame
	respLock    sync.Mutex
	respClosed  bool
	onceParse   sync.Once
}

// header keys
const (
	cmdKey     = "cmd"
	cmdSepByte = '/'
)

// JoinServiceMethod 拼接service和method
func JoinServiceMethod(service, method string) string {
	return service + string(cmdSepByte) + method
}

func (t *tunnel) newContext(req *Frame) Context {
	resp := t.NewFrame(FrameResponseErr)
	resp.SessionID = req.ID
	resp.DataType = TextData

	ctx := &context{
		tunnel:     t,
		req:        req,
		header:     nil,
		caller:     nil,
		resp:       resp,
		respChan:   make(chan *Frame, 1),
		respClosed: false,
	}
	return ctx
}

func (ctx *context) parse() {
	ctx.onceParse.Do(func() {
		if ctx.req.Header != nil {
			ctx.header, _ = ctx.req.ReadHeader()
		}
		cmd, found := ctx.header[cmdKey]
		if found {
			ctx.command = cmd
			pos := strings.IndexByte(cmd, cmdSepByte)
			if pos < 0 {
				ctx.serviceName = cmd
				ctx.methodName = ""
			} else {
				ctx.serviceName = cmd[0:pos]
				ctx.methodName = cmd[pos+1:]
			}
		}
	})
}

func (ctx *context) GetTunnel() Tunnel {
	return ctx.tunnel
}

func (ctx *context) GetCallStackStr(newCall Caller) (string, error) {
	ctx.parse()

	// todo:2 将header["stack"]进行解析，并且与newCall比对
	// 然后将stack+newCall进行序列化输出
	return "", nil
}

func (ctx *context) GetCmd() string {
	ctx.parse()
	return ctx.command
}

func (ctx *context) GetService() string {
	ctx.parse()
	return ctx.serviceName
}

func (ctx *context) GetMethod() string {
	ctx.parse()
	return ctx.methodName
}

func (ctx *context) GetData() []byte {
	return ctx.req.Data
}

func (ctx *context) GetJSON(v interface{}) error {
	if ctx.req.DataType != JSONData {
		return errors.New("typeof data is not json")
	}
	if ctx.req.Data == nil || len(ctx.req.Data) == 0 {
		return errors.New("data is empty")
	}
	return json.Unmarshal(ctx.req.Data, v)
}

func (ctx *context) GetText() (string, error) {
	if ctx.req.DataType != TextData {
		return "", errors.New("typeof data is not text")
	}
	if ctx.req.Data == nil {
		return "", errors.New("data is nil")
	}
	return string(ctx.req.Data), nil
}

func (ctx *context) JSON(v interface{}) {
	if v == nil {
		ctx.response(JSONData, nil, nil)
		return
	}
	data, err := json.Marshal(v)
	ctx.response(JSONData, data, err)
}

func (ctx *context) Text(text string) {
	ctx.response(TextData, []byte(text), nil)
}

func (ctx *context) Data(buf []byte) {
	ctx.response(BinaryData, buf, nil)
}

func (ctx *context) Error(err error) {
	ctx.response(0, nil, err)
}

func (ctx *context) send(resp *Frame) (err error) {
	ctx.respLock.Lock()
	if ctx.respClosed {
		return errors.New("resp channel closed")
	}
	defer func() {
		ctx.respLock.Unlock()
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
		if err != nil {
			// todo:3 log the error
		}
	}()
	ctx.respChan <- resp
	close(ctx.respChan)
	ctx.respClosed = true
	return nil
}

func (ctx *context) response(dataType uint8, data []byte, err error) {
	if err != nil {
		ctx.resp.Type = FrameResponseErr
		ctx.resp.DataType = TextData
		ctx.resp.Data = []byte(err.Error())
	} else {
		ctx.resp.Type = FrameResponseOK
		ctx.resp.DataType = dataType
		ctx.resp.Data = data
	}
	ctx.send(ctx.resp)
	ctx.Flush()
}

func (ctx *context) Flush() {
	select {
	case f := <-ctx.respChan:
		if f != nil {
			wc := ctx.tunnel.NewWriteCloser()
			_, err := f.WriteTo(wc)
			wc.Close()
			if err != nil {
				log.Get().WithError(err).Error("write response failed")
			}
		}
	case <-time.After(time.Second * 20):
		log.Get().Error("wait response frame timeout")
	}
}
