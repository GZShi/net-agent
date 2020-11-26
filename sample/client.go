package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/dial"
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
	remote := dial.NewClient(t)

	var s socks5.Server
	if socks5Addr != "" {
		s = socks5.NewServer()
		s.SetRequster(makeRequester(remote))
		go func() {
			s.ListenAndRun(socks5Addr)
			log.Get().Info("socks5 server stopped")
			if t != nil {
				t.Stop()
			}
		}()
		log.Get().Info("socks5 listen on ", socks5Addr)
	}

	log.Get().Info("client created")
	t.Run()
	log.Get().Info("client closed")

	// close socks server
	if s != nil {
		s.Stop()
	}
}

func makeRequester(remote dial.Client) socks5.Requester {
	return func(req socks5.Request) (net.Conn, error) {
		if req.GetCommand() != socks5.ConnectCommand {
			return nil, socks5.ErrCommandNotSupport
		}
		addr := req.GetAddrPortStr()

		// return remote.DialDirect("tcp4", addr)
		conn, err := remote.DialWithTunnelID(globalTID, "tcp4", addr)
		if err != nil {
			log.Get().WithError(err).Error("dial failed")
			return nil, err
		}
		return conn, nil
	}
}
