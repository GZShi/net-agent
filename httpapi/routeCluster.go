package api

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/totp"
	"github.com/GZShi/net-agent/transport"

	log "github.com/GZShi/net-agent/logger"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
)

// SetCluster 绑定TunnelCluster
func (s *HTTPServer) SetCluster(cluster *transport.TunnelCluster, proto *protocol.ProtoManager) {
	s.cluster = cluster
	s.proto = proto
}

func (s *HTTPServer) enableClusterRoute(ns iris.Party, pathstr string) {
	var registerLocker sync.Mutex // 避免接口被暴力破解，减少接口的并发能力

	party := ns.Party(pathstr)

	//
	// 获取集群的基本运行情况
	//
	party.Get("/base-info", func(ctx iris.Context) {
		var payload struct {
			Name string `json:"name"`
		}
		if err := ctx.ReadJSON(&payload); err != nil {
			response(ctx, err, nil)
			return
		}
		response(ctx, nil, s.cluster.GetStatus(payload.Name))
	})

	//
	// 申请开启临时端口转发程序
	//
	party.Get("/request-port-proxy", func(ctx iris.Context) {
		registerLocker.Lock()
		<-time.After(time.Second * 3)
		defer registerLocker.Unlock()

		var payload struct {
			Account string `json:"account"`
			Totp    string `json:"totp"`
			Port    uint16 `json:"port"`
			Target  string `json:"target"`
		}

		switch ctx.Method() {
		case "GET":
			port, err := ctx.URLParamInt("port")
			if err != nil {
				response(ctx, err, nil)
				return
			}
			payload.Account = ctx.URLParam("account")
			payload.Totp = ctx.URLParam("totp")
			payload.Port = uint16(port)
			payload.Target = ctx.URLParam("target")
		case "POST":
			if err := ctx.ReadJSON(&payload); err != nil {
				response(ctx, err, nil)
				return
			}
		default:
			response(ctx, fmt.Errorf("%s not supported", ctx.Method()), nil)
			return
		}

		if s.totpMap == nil {
			response(ctx, errors.New("totpMap is nil"), nil)
			return
		}
		info, has := s.totpMap[payload.Account]
		if !has {
			response(ctx, errors.New("failed"), nil)
			return
		}
		codes, err := totp.GetPassCode(info.Secret)
		if err != nil {
			response(ctx, errors.New("failed"), nil)
			return
		}
		if payload.Totp != codes[0] && payload.Totp != codes[1] {
			response(ctx, errors.New("failed"), nil)
			return
		}

		if payload.Port < 20000 || payload.Port > 20999 {
			response(ctx, errors.New("illegal port"), nil)
			return
		}

		// raddr := ctx.RemoteAddr()
		// ip := "127.0.0.1"
		ip := ctx.GetHeader("X-Real-IP")
		if ip == "" {
			ip = ctx.RemoteAddr()
		}

		// 随机选择可用端口，开放30秒，等待指定IP连接
		if err := waitConn(s.cluster, payload.Port, ip, payload.Account, payload.Target); err != nil {
			response(ctx, err, nil)
			return
		}

		response(ctx, nil, ip)
	})
}

func waitConn(cluster *transport.TunnelCluster, port uint16, ip, name, target string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	l := log.Get().WithFields(logrus.Fields{
		"port":   port,
		"ip":     ip,
		"name":   name,
		"target": target,
	})

	var closeListener sync.Once
	close := func() {
		closeListener.Do(func() {
			l.Info("listener closed")
			listener.Close()
		})
	}

	l.Info("wait ip connect")
	go func() {
		// 45秒后，强制关闭端口监听
		<-time.After(time.Second * 45)
		close()
	}()

	// 只允许最多3个连接存在
	for times := 0; times < 3; times++ {
		go func() {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			raddr := conn.RemoteAddr().String()
			raddr = strings.Split(raddr, ":")[0]
			if raddr != ip {
				log.Get().WithFields(logrus.Fields{
					"want":  ip,
					"raddr": raddr,
				}).Warn("illegal ip connected")
				return
			}

			// 通过tunnel进行代理
			names := strings.Split(name, "@")
			if len(names) != 2 {
				return
			}
			remote, err := cluster.Dial(ip, "tcp", target, names[1], names[0])
			if err != nil {
				log.Get().WithError(err).WithFields(logrus.Fields{
					"target": target,
					"name":   name,
				}).Error("dial by cluster failed")
				return
			}
			defer remote.Close()

			go io.Copy(remote, conn)
			io.Copy(conn, remote)
		}()
	}

	return nil
}
