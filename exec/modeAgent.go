package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
	"github.com/GZShi/net-agent/ws"
	"github.com/sirupsen/logrus"
)

func createTunnel(cfg *config, hiddenTunnelName bool) (*transport.Tunnel, error) {
	// tunnel
	tunnelName := cfg.ChannelName
	if hiddenTunnelName {
		tunnelName = transport.AnonymName
	} else if tunnelName == transport.AnonymName {
		return nil, errors.New("current channelName is illegal")
	}

	log.Get().WithField("addr", cfg.Addr).Info("try to connect server")
	var client net.Conn
	var err error

	// 如果是websocket协议地址，则使用ws.Dial
	if strings.HasPrefix(cfg.Addr, "ws://") || strings.HasPrefix(cfg.Addr, "wss://") {
		client, err = ws.Dial(cfg.Addr)
	} else {
		client, err = net.Dial("tcp", cfg.Addr)
	}
	if err != nil {
		log.Get().WithError(err).WithField("addr", cfg.Addr).Error("dial failed")
		return nil, err
	}

	randKey, err := transport.RequireAuth(client, cfg.Secret, tunnelName)
	if err != nil {
		log.Get().WithError(err).Error("auth failed")
		return nil, err
	}

	// auth response
	// byte1: 成功失败标识。
	buf := []byte{0x01}
	if _, err = io.ReadFull(client, buf); err != nil {
		log.Get().WithError(err).Error("receive bind info failed")
		return nil, err
	}
	if buf[0] != 0x00 {
		log.Get().WithField("name", tunnelName).Error("channel register failed")
		return nil, errors.New("register failed")
	}

	t, err := transport.NewTunnel(client, tunnelName, cfg.Secret, randKey, true, true)
	if err != nil {
		log.Get().WithError(err).Error("create client failed")
		return nil, err
	}

	return t, nil
}

func runAsAgent(cfg *config) {
	defer func() {
		err := recover()
		if err != nil {
			log.Get().Error("panic recover success")
		}
	}()

	t, err := createTunnel(cfg, false)
	if err != nil {
		return
	}

	socks5User := fmt.Sprintf("%v@%v", cfg.ClientName, cfg.ChannelName)
	log.Get().WithFields(logrus.Fields{
		"addr":     cfg.Addr,
		"username": socks5User,
		"password": socks5.CalcSum(socks5User, cfg.Secret),
	}).Info("access with socks5")

	log.Get().Info("start transferring data")
	t.Serve()

	log.Get().WithField("status", fmt.Sprintf("%v", t.GetStatus())).Info("transport closed")
}
