package tunnel

import (
	"encoding/json"
	"errors"
	"sync"

	log "github.com/GZShi/net-agent/logger"
)

// Context 用于处理RPC请求的上下文对象
type Context interface {
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
	tunnel    *tunnel
	req       *Frame
	header    map[string]string
	resp      *Frame
	respChan  chan *Frame
	onceParse sync.Once
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
		resp: &Frame{
			ID:        t.NewID(),
			Type:      FrameResponseErr,
			SessionID: req.ID,
			Header:    nil,
			DataType:  TextData,
			Data:      nil,
		},
		respChan: make(chan *Frame, 1),
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
	ctx.respChan <- ctx.resp
	close(ctx.respChan)
	ctx.Flush()
}

func (ctx *context) Flush() {
	f := <-ctx.respChan
	if f != nil {
		wc := ctx.tunnel.NewWriteCloser()
		_, err := f.WriteTo(wc)
		wc.Close()
		if err != nil {
			log.Get().WithError(err).Error("write response failed")
		}
	}
}
