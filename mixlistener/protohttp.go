package mixlistener

import (
	"bytes"
)

type httpListener struct {
	ProtoListener
}

// HTTP 监听HTTP协议
func HTTP() ProtoListener {
	return &httpListener{
		ProtoListener: NewProtobase(HTTPName),
	}
}

func (proto *httpListener) Taste(buf []byte) bool {
	if buf == nil || len(buf) < 3 {
		return false
	}
	if len(buf) > 3 {
		buf = buf[0:3]
	}
	if bytes.Equal(buf, []byte("GET")) ||
		bytes.Equal(buf, []byte("HEA")) ||
		bytes.Equal(buf, []byte("POS")) ||
		bytes.Equal(buf, []byte("PUT")) ||
		bytes.Equal(buf, []byte("DEL")) ||
		bytes.Equal(buf, []byte("CON")) ||
		bytes.Equal(buf, []byte("OPT")) ||
		bytes.Equal(buf, []byte("TRA")) ||
		bytes.Equal(buf, []byte("PAT")) {
		return true
	}
	return false
}
