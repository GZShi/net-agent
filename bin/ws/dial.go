package ws

import (
	"net"

	"github.com/gorilla/websocket"
)

// Dial 创建连接
func Dial(wsAddr string) (net.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr, nil)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}
