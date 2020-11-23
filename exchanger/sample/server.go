package main

import (
	"net"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

func main() {
	addr := "localhost:2000"
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Get().WithError(err).Error("listen failed: ", addr)
	}

	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept failed")
			break
		}
		wg.Add(1)
		go func(conn net.Conn) {
			serve(conn)
			wg.Done()
		}(conn)
	}

	log.Get().Info("wait all conn close")
	wg.Wait()
}

func serve(conn net.Conn) {
	defer conn.Close()
	_, err := conferReqOnServer(conn, "hello")
	if err != nil {
		return
	}

	t := tunnel.New(conn)
	t.Listen("check/acl", nil)
	t.Listen("dial/direct", nil)
	t.Listen("dial/tunnel", nil)
	t.Listen("dial/access-code", nil)
	t.Listen("register/access-code", nil)
	t.Listen("remove/access-code", nil)

	t.Run()
}
