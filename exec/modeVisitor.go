package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/sirupsen/logrus"
)

func runPortServer(info *portProxyInfo, cfg *config) {
	targetHost, targetPortStr, err := net.SplitHostPort(info.TargetAddr)
	if err != nil {
		log.Get().WithField("targetAddr", info.TargetAddr).WithError(err).Error("bad target address")
		return
	}
	targetPortInt, err := strconv.Atoi(targetPortStr)
	if err != nil {
		log.Get().WithField("targetAddr", info.TargetAddr).WithError(err).Error("parse port failed")
		return
	}
	if targetPortInt <= 0 || targetPortInt >= 65536 {
		log.Get().WithField("targetAddr", info.TargetAddr).Error("bad port")
		return
	}
	targetPort := uint16(targetPortInt)

	l, err := net.Listen("tcp4", info.Listen)
	if err != nil {
		log.Get().WithField("listen", info.Listen).WithError(err).Error("listen failed")
		return
	}

	log.Get().WithFields(logrus.Fields{
		"listen": info.Listen,
		"target": info.TargetAddr,
	}).Info("portproxy rule is working")

	var waitActiveConn sync.WaitGroup
	for {
		client, err := l.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept conn failed")
			break
		}
		waitActiveConn.Add(1)
		go func(client net.Conn) {
			defer waitActiveConn.Done()

			// 与server建立socks5协议连接
			// 然后pipe两个连接

			server, err := net.Dial("tcp4", cfg.Addr)
			if err != nil {
				log.Get().WithError(err).WithField("cfg.addr", cfg.Addr).Error("cant connect to socks5 server")
				return
			}
			username := fmt.Sprintf("%v@%v", cfg.ClientName, cfg.ChannelName)
			password := socks5.CalcSum(username, cfg.Secret)
			err = socks5.MakeSocks5Request(server, username, password, targetHost, targetPort)
			if err != nil {
				log.Get().WithError(err).Error("connect to socks server failed")
				return
			}

			// 透传数据
			go func() {
				defer client.Close()
				defer server.Close()
				io.Copy(client, server)
			}()
			func() {
				defer client.Close()
				defer server.Close()
				io.Copy(server, client)
			}()
		}(client)
	}

	waitActiveConn.Wait()
}

func runAsVisitor(cfg *config) {
	var wg sync.WaitGroup
	size := len(cfg.PortProxy)
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func(info *portProxyInfo) {
			defer wg.Done()
			runPortServer(info, cfg)
		}(&cfg.PortProxy[i])
	}
	wg.Wait()
}
