package ws

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

const readBufferSize = 1024 * 32
const writeBufferSize = 1024 * 32

var upgrader *websocket.Upgrader

func init() {
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
	}

	// for debug, very danger
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

// Upgrade 将http协议升级为net.Conn
func Upgrade(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}
