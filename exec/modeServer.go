package main

import (
	"errors"
	"net"
	"strings"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
	"github.com/kataras/iris"
)

func runAsServer(cfg *config) {
	httpServer := iris.New()
	tunnelCluster := transport.NewTunnelCluster(cfg.Secret)
	onSocks5Conn := func(sourceAddr, network, targetAddr, clientName string) (net.Conn, error) {
		authInfos := strings.Split(clientName, "@")
		if len(authInfos) != 2 {
			return nil, errors.New("bad client name")
		}
		userName := authInfos[0]
		channelName := authInfos[1]
		if len(channelName) < 3 {
			return nil, errors.New("channel name is too short")
		}
		if channelName == "direct" {
			return net.Dial(network, targetAddr)
		}
		return tunnelCluster.Dial(sourceAddr, network, targetAddr, channelName, userName)
	}
	socks5Server := socks5.NewSocks5Server(cfg.Secret, onSocks5Conn)

	initTotp(cfg.TotpList)
	// startPortproxyServer(tunnelCluster, cfg.PortProxy)
	setTunnelRoute(httpServer, tunnelCluster)

	// listen
	listener, err := protocol.NewListener("tcp", cfg.Addr)
	if err != nil {
		log.Get().WithField("addr", cfg.Addr).WithError(err).Error("listen on port failed")
		return
	}
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		tunnelCluster.Run(listener.GetAgentListener())
	}()
	go func() {
		defer wg.Done()
		socks5Server.Run(listener.GetSocks5Listener())
	}()
	go func() {
		defer wg.Done()
		httpServer.Run(iris.Listener(listener.GetHTTPListener()))
	}()
	wg.Wait()
}
