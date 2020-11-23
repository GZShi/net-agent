package main

import (
	ml "github.com/GZShi/net-agent/mixlistener"
)

type protoExchanger struct {
	ml.ProtoListener
}

// newExchanger 监听Exchanger协议
func newExchanger() *protoExchanger {
	return &protoExchanger{
		ml.NewProtobase("exchanger"),
	}
}

func (proto *protoExchanger) Taste(buf []byte) bool {
	return buf != nil && buf[0] == 0x09
}
