package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	clusterDef "github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/dial"
	dialDef "github.com/GZShi/net-agent/rpc/dial/def"
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
	client := dial.NewClient(t, nil)

	var s socks5.Server
	if socks5Addr != "" {
		s = socks5.NewServer()
		s.SetRequster(makeTunnelDialer(t, client))
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

func makeTunnelDialer(t tunnel.Tunnel, remote dialDef.Dial) socks5.Requester {
	return func(req socks5.Request) (net.Conn, error) {
		if req.GetCommand() != socks5.ConnectCommand {
			return nil, socks5.ErrCommandNotSupport
		}
		addr := req.GetAddrPortStr()

		stream, writeSID := t.NewStream()
		readSID, err := remote.Dial(writeSID, "tcp4", addr)
		if err != nil {
			log.Get().WithError(err).Info("tunnel dialer failed")
			return nil, err
		}
		stream.Bind(readSID)

		return stream, nil
	}
}

func makeTunnelIDDialer(t tunnel.Tunnel, client clusterDef.Cluster) socks5.Requester {
	return func(req socks5.Request) (net.Conn, error) {
		return nil, nil
	}
}
