package ws

import (
	"errors"
	"net"
	"time"

	log "github.com/GZShi/net-agent/v2/logger"

	"github.com/gorilla/websocket"
)

var ErrClosedEvent = errors.New("closed")

type frameBuf struct {
	err  error
	data []byte
	pos  int
}

// Conn 满足net.Conn协议的封装
type Conn struct {
	wsconn        *websocket.Conn
	readBufs      chan *frameBuf
	currBuf       *frameBuf
	writeWait     time.Duration
	heartbeatWait time.Duration
}

// NewConn 创建新的连接
func NewConn(wsconn *websocket.Conn) net.Conn {
	c := &Conn{
		wsconn:        wsconn,
		readBufs:      make(chan *frameBuf, 256),
		writeWait:     time.Second * 10,
		heartbeatWait: time.Second * 15,
	}
	c.setHandlers()
	go c.runDataReader()
	go c.runPingPong()

	return c
}

func (p *Conn) setHandlers() {
	p.wsconn.SetPingHandler(func(d string) error {
		return nil
	})
	p.wsconn.SetPongHandler(func(d string) error {
		return nil
	})
	p.wsconn.SetCloseHandler(func(code int, d string) error {
		p.Close()
		return nil
	})
}

// RunPingPong
func (p *Conn) runPingPong() {
	pingTicker := time.NewTicker(p.heartbeatWait)
	defer func() {
		pingTicker.Stop()
		p.Close()
	}()

	for {
		select {
		case <-pingTicker.C:
			p.wsconn.SetWriteDeadline(time.Now().Add(p.writeWait))
			err := p.wsconn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				return
			}
		}
	}
}

// RunDataReader 不断读取数据包
func (p *Conn) runDataReader() error {
	for {
		p.wsconn.SetReadDeadline(time.Now().Add(time.Minute * 3))
		msgType, msg, err := p.wsconn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Get().WithError(err).Error("ws read message error")
			}
			p.readBufs <- &frameBuf{err, nil, 0}
			close(p.readBufs)
			return err
		}

		switch msgType {
		case websocket.BinaryMessage:
			p.readBufs <- &frameBuf{nil, msg, 0}
		case websocket.CloseMessage:
			p.readBufs <- nil
			close(p.readBufs)
		}
	}
}

// implement net.Conn interface

func (p *Conn) Read(b []byte) (int, error) {
	var frame *frameBuf
	if p.currBuf == nil {
		p.currBuf = <-p.readBufs
		if p.currBuf == nil {
			p.Close()
			return 0, ErrClosedEvent
		}

		if p.currBuf.err != nil {
			p.Close()
			return 0, p.currBuf.err
		}
	}
	frame = p.currBuf

	remainSize := len(frame.data) - frame.pos
	if len(b) < remainSize {
		copy(b, frame.data[frame.pos:frame.pos+len(b)])
		frame.pos += len(b)
		return len(b), nil
	}

	copy(b, frame.data[frame.pos:])
	p.currBuf = nil
	return remainSize, nil
}

func (p *Conn) Write(b []byte) (int, error) {
	pos := 0
	end := pos + writeBufferSize
	for pos < len(b) {
		if end > len(b) {
			end = len(b)
		}
		p.wsconn.SetWriteDeadline(time.Now().Add(p.writeWait))
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
