package mixlistener

type tunnelListener struct {
	ProtoListener
}

// Tunnel 监听HTTP协议
func Tunnel() ProtoListener {
	return &tunnelListener{
		ProtoListener: NewProtobase(TunnelName),
	}
}

func (proto *tunnelListener) Taste(buf []byte) bool {
	return buf != nil && buf[0] == 0x09
}
