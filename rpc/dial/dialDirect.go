package dial

import (
	"net"
	"time"

	"github.com/GZShi/net-agent/exchanger"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
	"github.com/sirupsen/logrus"
)

type dialRequest struct {
	Network     string        `json:"network"`
	Address     string        `json:"address"`
	SessionID   uint32        `json:"sid"`
	TunnelID    exchanger.TID `json:"tid"`
	TunnelLabel string        `json:"label"`
}

type dialResponse struct {
	SessionID uint32 `json:"sid"`
}

func (c *client) DialDirect(network, address string) (net.Conn, error) {
	stream, sid := c.t.NewStream() // ready to read

	var req dialRequest
	req.Network = network
	req.Address = address
	req.SessionID = sid

	var resp dialResponse
	err := c.t.SendJSON(nil, nameOfDialDirect, &req, &resp)
	if err != nil {
		return nil, err
	}

	stream.Bind(resp.SessionID) // ready to write
	stream.SetInfo(req.Address)
	return stream, nil
}

func (s *service) DialDirect(ctx tunnel.Context) {
	var req dialRequest
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
	stream, sid := ctx.GetTunnel().NewStream() // ready to read
	resp.SessionID = sid
	stream.Bind(req.SessionID) // ready to write
	stream.SetInfo(req.Address)

	go func(start time.Time, address string) {
		sent, received, err := utils.LinkReadWriteCloser(stream, conn)
		duration := time.Now().Sub(start)
		log.Get().WithError(err).WithFields(logrus.Fields{
			"c_sent": sent,
			"c_recv": received,
			"c_t":    duration,
		}).Debug("closed: ", address)
	}(start, req.Address)

	log.Get().WithField("delay", time.Now().Sub(start)).Info("opened: ", req.Address)

	ctx.JSON(&resp)
}
