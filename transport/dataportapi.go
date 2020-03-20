package transport

import (
	"fmt"
	"time"
)

func (p *dataPort) GetStatus() interface{} {

	stateStr := ""
	switch p.state {
	case portstInited:
		stateStr = "ready"
	case portstConnecting:
		stateStr = "connect"
	case portstWorking:
		stateStr = "working"
	case portstClosing:
		stateStr = "closing"
	case portstClosed:
		stateStr = "closed"
	default:
		stateStr = fmt.Sprintf("unknown(%v)", stateStr)
	}

	return struct {
		UserName   string    `json:"uname"`
		ConnID     int       `json:"cid"`
		State      string    `json:"state"`
		TargetAddr string    `json:"targetAddr"`
		SourceAddr string    `json:"sourceAddr"`
		Created    time.Time `json:"created"`
		Closed     time.Time `json:"closed"`

		ClientToTunnel uint64 `json:"c2t"`
		TunnelToClient uint64 `json:"t2c"`
	}{
		p.userName,
		int(p.connID),
		stateStr,
		p.addr,
		p.sourceAddr,
		p.created,
		p.closed,
		p.upToTunnel,
		p.downToClient,
	}
}
