package main

import (
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

type dialReqeust struct {
	Network   string `json:"network"`
	Address   string `json:"address"`
	SessionID uint32 `json:"sid"`
}

type dialResponse struct {
	SessionID uint32 `json:"sid"`
}

func handleDialDirect(ctx tunnel.Context) {
	var req dialReqeust
	var resp dialResponse
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	// direct dial
	log.Get().Info("try to dial direct")
	conn, err := net.Dial(req.Network, req.Address)
	if err != nil {
		ctx.Error(err)
		return
	}

	// create and bind stream
	stream, sid := ctx.GetTunnel().NewStream()
	resp.SessionID = sid
	stream.Bind(req.SessionID)

	go utils.LinkReadWriteCloser(stream, conn)
	log.Get().Info("dial sucess")

	ctx.JSON(&resp)
}
