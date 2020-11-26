package dial

import (
	"errors"
	"net"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

func (c *client) DialWithTunnelID(tid exchanger.TID, network, address string) (net.Conn, error) {
	stream, sid := c.t.NewStream()

	var req dialRequest
	req.Network = network
	req.Address = address
	req.SessionID = sid
	req.TunnelID = tid

	var resp dialResponse
	err := c.t.SendJSON(nil, nameOfDialWithTunnelID, &req, &resp)
	if err != nil {
		return nil, err
	}

	stream.Bind(resp.SessionID) // ready to write
	stream.SetInfo(req.Address)
	return stream, nil
}

func (s *service) DialWithTunnelID(ctx tunnel.Context) {
	if s.cluster == nil {
		ctx.Error(errors.New("cluster not found"))
		return
	}

	var req dialRequest
	var resp dialResponse
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	t, err := s.cluster.FindTunnelByID(0)
	if err != nil {
		ctx.Error(err)
		return
	}

	client := NewClient(t)
	conn, err := client.DialDirect("tcp4", "")
	if err != nil {
		ctx.Error(err)
		return
	}

	stream, sid := ctx.GetTunnel().NewStream()
	resp.SessionID = sid
	stream.Bind(req.SessionID)
	stream.SetInfo(req.Address)

	go utils.LinkReadWriteCloser(stream, conn)

	ctx.JSON(&resp)
}
