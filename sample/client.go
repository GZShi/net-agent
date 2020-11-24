package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
)

func connectAsClient(addr, socks5Addr, password string) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Get().WithError(err).Error("connect ", addr, " failed")
		return
	}
	defer conn.Close()
	log.Get().Info("connect ", addr, " success")

	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.Get().WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)

	if socks5Addr != "" {
		s := socks5.NewServer()
		s.SetRequster(makeRequester(t))
		go func() {
			s.ListenAndRun(socks5Addr)
			log.Get().Info("socks5 server stopped")
		}()
		log.Get().Info("socks5 listen on ", socks5Addr)
	}

	log.Get().Info("tunnel[client] created")
	t.Run()
	log.Get().Info("tunnel[client] closed")
}

func makeRequester(t tunnel.Tunnel) socks5.Requester {
	return func(req socks5.Request) (net.Conn, error) {
		if req.GetCommand() != socks5.ConnectCommand {
			return nil, socks5.ErrCommandNotSupport
		}
		addr := req.GetAddrPortStr()

		// dial with tunnel
		// conn, err := net.Dial("tcp4", addr)
		conn, err := dialWithTunnel(t, addr)
		if err != nil {
			log.Get().WithError(err).Error("dial failed: ", addr)
			return nil, err
		}
		log.Get().Info("dial success: ", addr)
		return conn, nil
	}
}
