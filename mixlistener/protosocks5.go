package mixlistener

type socks5Listener struct {
	ProtoListener
}

// Socks5 监听HTTP协议
func Socks5() ProtoListener {
	return &socks5Listener{
		ProtoListener: NewProtobase(Socks5Name),
	}
}

func (proto *socks5Listener) Taste(buf []byte) bool {
	return buf != nil && buf[0] == 0x05
}
