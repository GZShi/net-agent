package main

import (
	"net"
	"sync"

	"github.com/GZShi/net-agent/cipherconn"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster"
	"github.com/GZShi/net-agent/rpc/dial"
	"github.com/GZShi/net-agent/tunnel"
)

func listenAndServe(addr, password string) {
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Get().WithError(err).Error("listen ", addr, " failed")
	}
	log.Get().Info("listen on ", addr)

	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept failed")
			break
		}
		wg.Add(1)
		go func(conn net.Conn) {
			serve(conn, password)
			wg.Done()
		}(conn)
	}

	log.Get().Info("wait all conn close")
	wg.Wait()
}

func serve(conn net.Conn, password string) {
	defer conn.Close()

	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.Get().WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)
	err = t.BindServices(
		dial.NewService(),
		cluster.NewService(),
	)
	if err != nil {
		log.Get().WithError(err).Error("bind service failed")
		return
	}

	log.Get().Info("tunnel created")
	t.Run()
	log.Get().Info("tunnel closed")
}
