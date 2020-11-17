package main

import (
	"path"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
	"github.com/kataras/iris"
)

func runAsServer(cfg *config, configDir string) {
	blockListPath := path.Join(configDir, "blocklist.json")
	go watchBlockList(blockListPath)
	if err := initBlockList(blockListPath); err != nil {
		log.Get().WithField("path", blockListPath).WithError(err).Error("初始化blocklist.json失败")
	}
	tunnelCluster := transport.NewTunnelCluster(cfg.Secret)
	socks5Server := socks5.NewSocks5Server(cfg.Secret,
		socks5.NewTunnelClusterDialer(tunnelCluster, checkBlockList))

	// listen for http/socks5/tunnel
	listener, err := protocol.NewListener("tcp", cfg.Addr)
	if err != nil {
		log.Get().WithField("addr", cfg.Addr).WithError(err).Error("listen on port failed")
		return
	}

	// startPortproxyServer(tunnelCluster, cfg.PortProxy)
	httpServer := iris.New()
	initTotp(cfg.TotpList)
	setTunnelRoute(httpServer, tunnelCluster, listener)

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		tunnelCluster.Run(listener.GetListener(protocol.ProtoAgentClient))
	}()
	go func() {
		defer wg.Done()
		socks5Server.Run(listener.GetListener(protocol.ProtoSocks5))
	}()
	go func() {
		defer wg.Done()
		httpServer.Run(iris.Listener(listener.GetListener(protocol.ProtoHTTP)))
	}()

	wg.Wait()
}
