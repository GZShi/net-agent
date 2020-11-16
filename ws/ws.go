package ws

import (
	"errors"
	"net"
	"net/http"
	"time"

	log "github.com/GZShi/net-agent/logger"
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

var ErrClosedEvent = errors.New("closed")

type frameBuf struct {
	err  error
	data []byte
}

// Conn 满足net.Conn协议的封装
type Conn struct {
	wsconn   *websocket.Conn
	readBufs chan *frameBuf
}

// NewConn 创建新的连接
func NewConn(wsconn *websocket.Conn) net.Conn {
	c := &Conn{
		wsconn:   wsconn,
		readBufs: make(chan *frameBuf, 256),
	}
	go c.runDataReader()

	return c
}

// Dial 创建连接
func Dial(wsAddr string) (net.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsAddr, nil)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}

// Upgrade 将http协议升级为net.Conn
func Upgrade(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}

// RunDataReader 不断读取数据包
func (p *Conn) runDataReader() error {
	for {
		msgType, msg, err := p.wsconn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Get().WithError(err).Error("ws read message error")
			}
			p.readBufs <- &frameBuf{err, nil}
			close(p.readBufs)
			return err
		}

		switch msgType {
		case websocket.BinaryMessage:
			p.readBufs <- &frameBuf{nil, msg}
		case websocket.CloseMessage:
			p.readBufs <- nil
			close(p.readBufs)
		}
	}
}

// implement net.Conn interface

func (p *Conn) Read(b []byte) (int, error) {
	// need lock here ?
	frame := <-p.readBufs
	if frame == nil {
		p.Close()
		return 0, ErrClosedEvent
	}
	if frame.err != nil {
		return 0, frame.err
	}
	copy(b, frame.data)
	return len(frame.data), nil
}

func (p *Conn) Write(b []byte) (int, error) {
	pos := 0
	end := pos + writeBufferSize
	for pos < len(b) {
		if end > len(b) {
			end = len(b)
		}
		err := p.wsconn.WriteMessage(websocket.BinaryMessage, b[pos:end])
		if err != nil {
			p.Close()
			return pos, err
		}

		// next round
		pos = end
		end += writeBufferSize
	}
	return len(b), nil
}

// Close 关闭连接
func (p *Conn) Close() error {
	return p.wsconn.Close()
}

// LocalAddr 获取连接的本地地址
func (p *Conn) LocalAddr() net.Addr {
	return p.wsconn.LocalAddr()
}

// RemoteAddr 获取连接的远端地址
func (p *Conn) RemoteAddr() net.Addr {
	return p.wsconn.RemoteAddr()
}

// SetDeadline ...
func (p *Conn) SetDeadline(t time.Time) error {
	err := p.wsconn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = p.wsconn.SetWriteDeadline(t)
	if err != nil {
		return err
	}
	return nil
}

// SetReadDeadline ...
func (p *Conn) SetReadDeadline(t time.Time) error {
	return p.wsconn.SetReadDeadline(t)
}

// SetWriteDeadline ...
func (p *Conn) SetWriteDeadline(t time.Time) error {
	return p.wsconn.SetWriteDeadline(t)
}
