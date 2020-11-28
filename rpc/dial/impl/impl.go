package impl

import (
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/dial"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
	"github.com/sirupsen/logrus"
)

// New 创建服务端实现
func New(t tunnel.Tunnel) dial.Dial {
	return &impl{t}
}

type impl struct {
	t tunnel.Tunnel
}

func (s *impl) Dial(dialSessionID uint32, network, address string) (connSessionID uint32, err error) {
	start := time.Now()

	// direct dial
	conn, err := net.Dial(network, address)
	if err != nil {
		log.Get().WithError(err).Error("dial failed: ", address)
		return 0, err
	}

	// create and bind stream
	stream, sid := s.t.NewStream() // ready to read
	stream.Bind(dialSessionID)     // ready to write
	stream.SetInfo(address)

	go func(start time.Time, address string) {
		sent, received, err := utils.LinkReadWriteCloser(stream, conn)
		duration := time.Now().Sub(start)
		log.Get().WithError(err).WithFields(logrus.Fields{
			"c_sent": sent,
			"c_recv": received,
			"c_t":    duration,
		}).Debug("closed: ", address)
	}(start, address)

	log.Get().WithField("delay", time.Now().Sub(start)).Info("opened: ", address)

	return sid, nil
}
