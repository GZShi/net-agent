package main

import (
	"net"

	"github.com/GZShi/net-agent/cipherconn"
	logger "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/rpc/dial"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/sirupsen/logrus"
)

func connectAsClient(addr, socks5Addr, password string) {
	log := logger.Get().WithField("mode", "client")

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		log.WithError(err).Error("connect ", addr, " failed")
		return
	}
	defer conn.Close()
	log.Info("connect ", addr, " success")

	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)

	var s socks5.Server
	if socks5Addr != "" {
		s = socks5.NewServer()
		s.SetRequster(makeDialer2(t, log))
		go func() {
			s.ListenAndRun(socks5Addr)
			log.Info("socks5 server stopped")
			if t != nil {
				t.Stop()
			}
		}()
		log.Info("socks5 listen on ", socks5Addr)
	}

	log.Info("client created")
	t.Run()
	log.Info("client closed")

	// close socks server
	if s != nil {
		s.Stop()
	}
}

func makeDialer(t tunnel.Tunnel, log *logrus.Entry) socks5.Requester {
	dialClient := dial.NewClient(t, nil)
	clsClient := cluster.NewClient(t, nil)

	return func(req socks5.Request) (net.Conn, error) {
		if req.GetCommand() != socks5.ConnectCommand {
			return nil, socks5.ErrCommandNotSupport
		}
		addr := req.GetAddrPortStr()

		useClsClient := (globalTID > 0)

		stream, writeSID := t.NewStream()
		var readSID uint32
		var err error
		if useClsClient {
			readSID, err = clsClient.DialByTID(globalTID, writeSID, "tcp4", addr)
		} else {
			readSID, err = dialClient.Dial(writeSID, "tcp4", addr)
		}
		if err != nil {
			log.WithField("usecls", useClsClient).WithError(err).Info("tunnel dialer failed")
			return nil, err
		}
		stream.Bind(readSID)

		return stream, nil
	}
}

func makeDialer2(t tunnel.Tunnel, log *logrus.Entry) socks5.Requester {
	return func(req socks5.Request) (net.Conn, error) {
		if req.GetCommand() != socks5.ConnectCommand {
			return nil, socks5.ErrCommandNotSupport
		}
		addr := req.GetAddrPortStr()

		conn, err := t.Dial(1080)
		if err != nil {
			log.WithError(err).Info("tunnel dialer failed")
			return nil, err
		}

		return socks5.Upgrade(conn, addr, nil)
	}
}
