package tunnel

import (
	"encoding/json"
	"errors"
)

// request 发送一个RequestFrame，并等待对端返回一个ResponseFrame
func (t *tunnel) request(req *Frame) (*Frame, error) {
	guard := &frameGuard{
		ch: make(chan *Frame),
	}
	t.respGuards.Store(req.ID, guard)

	w := t.NewWriteCloser()
	_, err := req.WriteTo(w)
	w.Close()
	if err != nil {
		t.respGuards.Delete(req.ID)
		return nil, err
	}

	resp := <-guard.ch
	return resp, nil
}

func (t *tunnel) call(cmd string, dataType uint8, data []byte) ([]byte, error) {
	req := &Frame{
		ID:        t.NewID(),
		Type:      FrameRequest,
		SessionID: 0,
		Header:    nil,
		DataType:  dataType,
		Data:      data,
	}
	req.WriteHeader(Headers{"cmd": cmd})

	resp, err := t.request(req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SendJSON 向对端以JSON方式请求数据
func (t *tunnel) SendJSON(cmd string, in interface{}, out interface{}) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	respData, err := t.call(cmd, JSONData, payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(respData, out)
}

// SendText 向对端以Text方式请求数据
func (t *tunnel) SendText(cmd string, text string) (string, error) {
	resp, err := t.call(cmd, TextData, []byte(text))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("resp is nil")
	}
	return string(resp), nil
}
