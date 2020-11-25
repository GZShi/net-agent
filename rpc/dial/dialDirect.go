package dial

import (
	"net"
	"time"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
	"github.com/sirupsen/logrus"
)

type dialReqeust struct {
	Network   string `json:"network"`
	Address   string `json:"address"`
	SessionID uint32 `json:"sid"`
}

type dialResponse struct {
	SessionID uint32 `json:"sid"`
}

func (c *client) DialDirect(network, address string) (net.Conn, error) {
	stream, sid := c.t.NewStream()

	var req dialReqeust
	req.Network = network
	req.Address = address
	req.SessionID = sid

	var resp dialResponse
	err := c.t.SendJSON(nil, nameOfDialDirect, &req, &resp)
	if err != nil {
		return nil, err
	}

	stream.Bind(resp.SessionID)
	return stream, nil
}

func (s *service) DialDirect(ctx tunnel.Context) {
	var req dialReqeust
	var resp dialResponse
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	start := time.Now()

	// direct dial
	conn, err := net.Dial(req.Network, req.Address)
	if err != nil {
		log.Get().WithError(err).Error("dial failed: ", req.Address)
		ctx.Error(err)
		return
	}

	// create and bind stream
	stream, sid := ctx.GetTunnel().NewStream()
	resp.SessionID = sid
	stream.Bind(req.SessionID)

	go func(start time.Time, address string) {
		sent, received, err := utils.LinkReadWriteCloser(stream, conn)
		duration := time.Now().Sub(start)
		log.Get().WithError(err).WithFields(logrus.Fields{
			"c_sent": sent,
			"c_recv": received,
			"c_t":    duration,
		}).Debug("closed: ", address)
	}(start, req.Address)

	log.Get().WithField("duration", time.Now().Sub(start)).Info("dial sucess: ", req.Address)

	ctx.JSON(&resp)
}
