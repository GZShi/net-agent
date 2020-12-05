package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
)

// request 发送一个RequestFrame，并等待对端返回一个ResponseFrame
func (t *tunnel) request(req *Frame) (*Frame, error) {
	guard := make(chan *Frame)
	_, loaded := t.respGuards.LoadOrStore(req.ID, guard)
	if loaded {
		return nil, errors.New("dump req.id")
	}

	// 这个请求完毕后，就应该删除对应的guard
	defer func() {
		t.respGuards.Delete(req.ID)
		close(guard)
	}()

	w := t.NewWriteCloser()
	_, err := req.WriteTo(w)
	w.Close()
	if err != nil {
		return nil, err
	}

	resp := <-guard

	// 判断应答包是正确应答还是错误应答
	if resp == nil {
		return nil, errors.New("empty response from remote")
	}
	if resp.Type == FrameResponseErr {
		return nil, fmt.Errorf("rpc: %v", string(resp.Data))
	}

	return resp, nil
}

// ctxChunk 发起当前rpc.call时，所处的Context环境
func (t *tunnel) call(ctxChunk Context, cmd string, dataType uint8, data []byte) ([]byte, error) {
	req := t.NewFrame(FrameRequest)
	req.DataType = dataType
	req.Data = data

	stack := ""
	var err error

	// 检查是否会形成调用环路
	if ctxChunk != nil {
		stack, err = ctxChunk.GetCallStackStr(nil)
		if err != nil {
			return nil, err
		}
	}

	req.WriteHeader(Headers{
		"cmd":   cmd,
		"stack": stack,
	})

	resp, err := t.request(req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SendJSON 向对端以JSON方式请求数据
func (t *tunnel) SendJSON(ctxChunk Context, cmd string, in interface{}, out interface{}) error {
	var payload []byte
	var err error
	if in != nil {
		payload, err = json.Marshal(in)
		if err != nil {
			return err
		}
	}
	respData, err := t.call(ctxChunk, cmd, JSONData, payload)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(respData, out)
}

// SendText 向对端以Text方式请求数据
func (t *tunnel) SendText(ctxChunk Context, cmd string, text string) (string, error) {
	resp, err := t.call(ctxChunk, cmd, TextData, []byte(text))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("resp is nil")
	}
	return string(resp), nil
}
