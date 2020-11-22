package main

import (
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
)

func runClient(socks5Addr, serverAddr string) error {
	t, err := connectTunnel(serverAddr)
	if err != nil {
		return err
	}
	go t.Run()

	s := socks5.NewServer()
	s.SetRequster(makeRequester(t))
	log.Get().Info("socks5 listen on ", socks5Addr)
	return s.ListenAndRun(socks5Addr)
}

func connectTunnel(addr string) (tunnel.Tunnel, error) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	log.Get().Info("tunnel connect to ", addr, " success")
	return tunnel.New(conn), nil
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

func dialWithTunnel(t tunnel.Tunnel, addr string) (net.Conn, error) {
	stream, sid := t.NewStream()
	resp := &dialResponse{}
	err := t.SendJSON("dial", &dialReqeust{"tcp4", addr, sid}, resp)
	if err != nil {
		return nil, err
	}
	stream.Bind(resp.SessionID)
	return stream, nil
}
