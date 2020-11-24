package main

import (
	"net"
	"sync"

	"github.com/GZShi/net-agent/cipherconn"
	"github.com/GZShi/net-agent/exchanger"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

func listenAndServe(addr, password string) {
	ts := exchanger.NewCluster()

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
			serve(ts, conn, password)
			wg.Done()
		}(conn)
	}

	log.Get().Info("wait all conn close")
	wg.Wait()
}

func serve(ts exchanger.Cluster, conn net.Conn, password string) {
	defer conn.Close()
	cc, err := cipherconn.New(conn, password)
	if err != nil {
		log.Get().WithError(err).Error("create cipherconn failed")
		return
	}

	t := tunnel.New(cc)
	t.Listen("join/cluster", newJoinClusterHandler(ts))     // 加入集群
	t.Listen("detach/cluster", newDetachClusterHandler(ts)) // 离开集群
	t.Listen("dial/direct", handleDialDirect)               // 直接通过对端网络创建网络连接
	t.Listen("dial/tunnel", newDialTunnelHandler(ts))       // 通过对端网络进行路由，创建网络连接

	log.Get().Info("tunnel created")
	t.Run()
	log.Get().Info("tunnel closed")
}
