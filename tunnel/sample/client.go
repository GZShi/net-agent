package main

import (
	"io"
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
)

func connectTunnel(addr string) (tunnel.Tunnel, error) {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	log.Get().Info("tunnel connect to ", addr, " success")
	return tunnel.New(conn), nil
}

func runClient(socks5Addr, serverAddr string) error {
	t, err := connectTunnel(serverAddr)
	if err != nil {
		return err
	}
	go t.Run()

	s := socks5.NewSocks5Server("", func(sourceAddr, network, targetAddr, clientName string) (io.ReadWriteCloser, error) {
		stream, sid := t.NewStream()
		req := &dialReqeust{
			Network:   "tcp4",
			Address:   targetAddr,
			SessionID: sid,
		}
		resp := &dialResponse{}
		log.Get().Info("try to dial remote")
		err := t.SendJSON("dial", req, resp)
		if err != nil {
			return nil, err
		}
		stream.Bind(resp.SessionID)
		return stream, nil
	})

	l, err := net.Listen("tcp4", socks5Addr)
	if err != nil {
		return err
	}

	log.Get().Info("socks5 listen on ", socks5Addr)

	return s.Run(l)
}
