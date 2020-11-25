package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

// Context 用于处理RPC请求的上下文对象
type Context interface {
	// context
	GetTunnel() Tunnel
	GetCallStackStr(Caller) (string, error)

	// for request
	GetCmd() string
	GetData() []byte
	GetJSON(v interface{}) error
	GetText() (string, error)

	// for response
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
	tunnel     *tunnel
	req        *Frame
	header     map[string]string
	caller     []Caller
	resp       *Frame
	respChan   chan *Frame
	respLock   sync.Mutex
	respClosed bool
	onceParse  sync.Once
}

// header keys
const (
	cmdKey = "cmd"
)

func (t *tunnel) newContext(req *Frame) Context {
	ctx := &context{
		tunnel: t,
		req:    req,
		header: nil,
		caller: nil,
		resp: &Frame{
			ID:        t.NewID(),
			Type:      FrameResponseErr,
			SessionID: req.ID,
			Header:    nil,
			DataType:  TextData,
			Data:      nil,
		},
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
	if ctx.header == nil {
		return ""
	}
	cmd, found := ctx.header[cmdKey]
	if found {
		return cmd
	}
	return ""
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
