package main

import (
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
)

func runServer(addr string) error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	log.Get().Info("listen on ", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go serve(conn)
	}
}

func serve(conn net.Conn) {
	if conn == nil {
		return
	}
	log.Get().Info("a tunnel created")
	t := tunnel.New(conn)

	t.Listen("dial", func(ctx tunnel.Context) {
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
		stream, sid := t.NewStream()
		resp.SessionID = sid
		stream.Bind(req.SessionID)

		go socks5.Link(stream, conn)
		log.Get().Info("dial sucess")

		ctx.JSON(&resp)
	})

	t.Run()
	log.Get().Info("a tunnel closed")
}
