package main

import (
	"net"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

type dialTunnelRequest struct {
	SessionID      uint32        `json:"sid"`
	TargetTunnelID exchanger.TID `json:"tid"`
	Address        string        `json:"addr"`
}

type dialTunnelResponse struct {
	SessionID uint32 `json:"sid"`
}

func newDialTunnelHandler(ts exchanger.Cluster) tunnel.OnRequestFunc {
	return func(ctx tunnel.Context) {
		var req dialTunnelRequest
		var resp dialTunnelResponse
		err := ctx.GetJSON(&req)
		if err != nil {
			ctx.Error(err)
			return
		}

		t, err := ts.FindTunnelByID(req.TargetTunnelID)
		if err != nil {
			ctx.Error(err)
			return
		}

		conn, err := dialWithTunnel(t, req.Address)
		if err != nil {
			ctx.Error(err)
			return
		}

		stream, sid := ctx.GetTunnel().NewStream()
		resp.SessionID = sid
		stream.Bind(req.SessionID)

		go utils.LinkReadWriteCloser(stream, conn)
		ctx.JSON(&resp)
	}
}

func dialWithTunnel(t tunnel.Tunnel, addr string) (net.Conn, error) {
	stream, sid := t.NewStream()
	resp := &dialResponse{}
	err := t.SendJSON(nil, "dial", &dialReqeust{"tcp4", addr, sid}, resp)
	if err != nil {
		return nil, err
	}
	stream.Bind(resp.SessionID)
	return stream, nil
}
