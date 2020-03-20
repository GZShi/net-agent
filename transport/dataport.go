package transport

import (
	"io"
	"net"
	"sync"
	"time"
)

const (
	portstInited = iota
	portstConnecting
	portstWorking
	portstClosing
	portstClosed
)

type dataPort struct {
	userName   string
	connID     cid
	addr       string
	sourceAddr string
	dialer     net.Conn
	porter     net.Conn
	t          *Tunnel

	dialErr error
	dialWg  sync.WaitGroup

	onceClose sync.Once

	state        int
	created      time.Time
	closed       time.Time
	upToTunnel   uint64
	downToClient uint64
}

func newDataPort(id cid, sourceAddr, addr, userName string) *dataPort {
	return &dataPort{
		userName:   userName,
		connID:     id,
		addr:       addr,
		sourceAddr: sourceAddr,

		state:        portstInited,
		created:      time.Now(),
		upToTunnel:   0,
		downToClient: 0,
	}
}

// dialWithTunnel 通过通道进行连接创建
// 实际上连接通道已经建立好，只等那头回应是否能过访问目标地址
func (p *dataPort) dialWithTunnel(t *Tunnel) (dialer net.Conn, err error) {
	p.state = portstConnecting
	defer func() {
		if err == nil {
			p.state = portstWorking
		}
	}()

	p.t = t
	dialer, porter := net.Pipe()
	p.dialer = dialer
	p.porter = porter

	p.dialWg.Add(1)
	if _, _, err = t.writeData(newDialReq(p.connID, p.addr)); err != nil {
		p.dialWg.Done()
		return nil, err
	}

	p.dialWg.Wait()

	if p.dialErr != nil {
		return nil, p.dialErr
	}
	return p.dialer, nil
}

// dialCallback 在通道那头回应的时候被调用
func (p *dataPort) dialCallback(err error) {
	p.dialErr = err
	p.dialWg.Done()
}

// dialToNet
func (p *dataPort) dialToNet(t *Tunnel) (err error) {
	p.state = portstConnecting
	defer func() {
		if err == nil {
			p.state = portstWorking
		}
	}()

	p.t = t
	target, err := net.Dial("tcp", p.addr)
	if err != nil {
		return err
	}

	p.porter = target
	return nil
}

func (p *dataPort) serve() {
	p.writeDataToTunnel()
}

func (p *dataPort) writeDataToTunnel() {
	buf := make([]byte, 4096)
	for {
		rn, err := p.porter.Read(buf)
		if rn > 0 {
			_, wn, _ := p.t.writeData(Data{
				ConnID: p.connID,
				Cmd:    cmdData,
				Bytes:  buf[0:rn],
			})
			p.upToTunnel += uint64(wn)
		}
		if err != nil {
			if err != io.EOF {
				// log the error
			}
			break
		}
	}
}

func (p *dataPort) writeDataToPorter(bytes []byte) (int, error) {
	wn, err := p.porter.Write(bytes)
	if wn > 0 {
		p.downToClient += uint64(wn)
	}

	return wn, err
}

func (p *dataPort) close() {
	p.onceClose.Do(func() {
		p.state = portstClosing
		defer func() {
			p.state = portstClosed
			p.closed = time.Now()
		}()

		if p.dialer != nil {
			p.dialer.Close()
		}
		if p.porter != nil {
			p.porter.Close()
		}
	})
}
