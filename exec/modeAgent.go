package main

import (
	"fmt"
	"io"
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
	"github.com/sirupsen/logrus"
)

func runAsAgent(cfg *config) {
	defer func() {
		err := recover()
		if err != nil {
			log.Get().Error("panic recover success")
		}
	}()

	log.Get().WithField("addr", cfg.Addr).Info("try to connect server")

	client, err := net.Dial("tcp", cfg.Addr)
	if err != nil {
		log.Get().WithError(err).WithField("addr", cfg.Addr).Error("dial failed")
		return
	}

	randKey, err := transport.RequireAuth(client, cfg.Secret, cfg.ChannelName)
	if err != nil {
		log.Get().WithError(err).Error("auth failed")
		return
	}

	// auth response
	// byte1: 成功失败标识。
	buf := []byte{0x01}
	if _, err = io.ReadFull(client, buf); err != nil {
		log.Get().WithError(err).Error("receive bind info failed")
		return
	}
	if buf[0] != 0x00 {
		log.Get().WithField("name", cfg.ChannelName).Error("channel register failed")
		return
	}

	// tunnel
	t, err := transport.NewTunnel(client, cfg.ChannelName, cfg.Secret, randKey, true, true)
	if err != nil {
		log.Get().WithError(err).Error("create client failed")
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
