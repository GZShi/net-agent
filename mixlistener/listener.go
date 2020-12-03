package mixlistener

import (
	"errors"
	"net"
)

// MixListener 混合协议监听程序
// 可以实现在同一个端口，根据不同协议特征进行连接分发
// 可以实现在同一个端口，根据不同协议特征进行连接分发
type MixListener interface {
	Register(proto ProtoListener) error
	RegisterBuiltIn(protoName ...string)
	GetListener(protoName string) (net.Listener, error)
	Run() error
}

// Listen ...
func Listen(network, addr string) MixListener {
	return &mixlisn{
		protoMap: make(map[string]ProtoListener),
		protos:   []ProtoListener{},
		network:  network,
		addr:     addr,
	}
}

//
// implement of ProtoListener
//
type mixlisn struct {
	protoMap map[string]ProtoListener
	protos   []ProtoListener
	network  string
	addr     string
}

func (ml *mixlisn) Register(proto ProtoListener) error {
	_, found := ml.protoMap[proto.Name()]
	if found {
		return errors.New("proto exists")
	}
	ml.protoMap[proto.Name()] = proto
	ml.protos = append(ml.protos, proto)
	return nil
}

func (ml *mixlisn) RegisterBuiltIn(protoNames ...string) {
	var err error
	for _, name := range protoNames {
		err = nil
		switch name {
		case HTTPName:
			err = ml.Register(HTTP())
		case Socks5Name:
			err = ml.Register(Socks5())
		case TunnelName:
			err = ml.Register(Tunnel())
		default:
			err = errors.New("built-in proto not found")
		}

		if err != nil {
			// todo:3 log error
		}
	}
}

func (ml *mixlisn) GetListener(name string) (net.Listener, error) {
	proto, found := ml.protoMap[name]
	if found {
		return proto, nil
	}
	return nil, errors.New("proto not found")
}

func (ml *mixlisn) Run() error {
	listener, err := net.Listen(ml.network, ml.addr)
	if err != nil {
		return err
	}

	for {
		raw, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(conn *bufconn) {
			defer func() {
				if conn != nil {
					conn.Close()
				}
			}()
			peeked, err := conn.Reader.Peek(3)
			if err != nil {
				// todo:3 log the error
				return
			}

			// 逐一对所有协议进行尝试
			for _, p := range ml.protos {
				if p.Taste(peeked) {
					err = p.Recieve(conn)
					if err == nil {
						conn = nil // p.Recieve成功后，conn的生命周期将由p进行管理
					}
					// todo:3 log error
					return
				}
			}

			// todo:5 log the error: can't match any proto
		}(newBufconn(raw))
	}
}
