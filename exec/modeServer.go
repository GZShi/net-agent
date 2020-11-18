package main

import (
	"path"
	"sync"

	api "github.com/GZShi/net-agent/httpapi"
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
)

func runAsServer(cfg *config, configDir string) {
	// 黑名单处理
	blockListPath := path.Join(configDir, "blocklist.json")
	go watchBlockList(blockListPath)
	if err := initBlockList(blockListPath); err != nil {
		log.Get().WithField("path", blockListPath).WithError(err).Error("初始化blocklist.json失败")
	}

	// 第一步：创建TCP监听器
	listener, err := protocol.NewListener("tcp", cfg.Addr)
	if err != nil {
		log.Get().WithField("addr", cfg.Addr).WithError(err).Error("listen on port failed")
		return
	}

	var servers sync.WaitGroup

	// 第二步：启动TunnelCluster服务
	tunnelCluster := transport.NewTunnelCluster(cfg.Secret)
	servers.Add(1)
	go func() {
		defer servers.Done()
		tunnelCluster.Run(listener.GetListener(protocol.ProtoAgentClient))
	}()

	// 第三步：启动Socks5服务
	servers.Add(1)
	go func() {
		defer servers.Done()
		dialer := socks5.NewTunnelClusterDialer(tunnelCluster, checkBlockList)
		server := socks5.NewSocks5Server(cfg.Secret, dialer)
		server.Run(listener.GetListener(protocol.ProtoSocks5))
	}()

	// 第三步：启动HTTP服务（支持管理接口与websocket）
	servers.Add(1)
	go func() {
		defer servers.Done()
		server := api.NewHTTPServer()
		server.SetTotpInfo(nil)
		server.SetCluster(tunnelCluster, listener)
		server.Run(listener.GetListener(protocol.ProtoHTTP))
	}()

	servers.Wait()
}
