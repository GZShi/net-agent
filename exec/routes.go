package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/GZShi/net-agent/totp"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/transport"
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
)

var cluster *transport.TunnelCluster
var totpMap map[string]totpInfo
var registerLocker sync.Mutex // 避免接口被暴力破解，减少接口的并发能力

func initTotp(infos []totpInfo) {
	totpMap = make(map[string]totpInfo)

	for _, info := range infos {
		totpMap[info.Account] = info
		log.Get().WithField("account", info.Account).Info("add totp info")
	}
}

func response(ctx iris.Context, err error, data interface{}) {
	if err != nil {
		ctx.JSON(struct {
			ErrCode int         `json:"ErrCode"`
			ErrInfo string      `json:"ErrMsg"`
			Data    interface{} `json:"Data"`
		}{-1, err.Error(), nil})
	} else {
		ctx.JSON(struct {
			ErrCode int         `json:"ErrCode"`
			ErrInfo string      `json:"ErrMsg"`
			Data    interface{} `json:"Data"`
		}{0, "", data})
	}
}

func setTunnelRoute(app *iris.Application, tunnelCluster *transport.TunnelCluster) {

	cluster = tunnelCluster
	if cluster == nil {
		return
	}

	gss := app.Party("/reworkapi")
	{
		gss.Get("/register-socks5", handleRegisterSocks5)
		gss.Get("/register-portproxy", handleRegisterPortproxy)
		gss.Post("/register-portproxy", handleRegisterPortproxy)
		gss.Get("/register-totp", handleRegisterTotp)
		gss.Post("/login", handleLogin)

		admin := gss.Party("/admin", handleCheckSSO)
		{
			admin.Get("/base-info", handleGetBaseInfo)
			admin.Post("/logout", handleLogout)
		}
	}
}

func handleRegisterSocks5(ctx iris.Context) {
	registerLocker.Lock()
	<-time.After(time.Second * 3)
	defer registerLocker.Unlock()

	name := ctx.URLParam("name")
	secret := ctx.URLParam("secret")

	if len(name) <= 3 || len(secret) <= 2 {
		response(ctx, errors.New("参数长度过短"), nil)
		return
	}

	pswd := socks5.CalcSum(name, secret)

	response(ctx, nil, struct {
		Name string `json:"username"`
		Pswd string `json:"password"`
	}{name, pswd})
}

func handleLogin(ctx iris.Context) {
}

func handleLogout(ctx iris.Context) {

}

func handleCheckSSO(ctx iris.Context) {
	// todo
	ctx.Next()
}

func handleGetBaseInfo(ctx iris.Context) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := ctx.ReadJSON(&payload); err != nil {
		response(ctx, err, nil)
		return
	}
	response(ctx, nil, cluster.GetStatus(payload.Name))
}

func handleRegisterPortproxy(ctx iris.Context) {
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

	info, has := totpMap[payload.Account]
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
	if err := waitConn(payload.Port, ip, payload.Account, payload.Target); err != nil {
		response(ctx, err, nil)
		return
	}

	response(ctx, nil, ip)
}

func waitConn(port uint16, ip, name, target string) error {
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

func handleRegisterTotp(ctx iris.Context) {
	registerLocker.Lock()
	<-time.After(time.Second * 3)
	defer registerLocker.Unlock()

	account := ctx.URLParam("account")
	secret := totp.GenSecret(0)

	response(ctx, nil, totpInfo{
		Account: account,
		Secret:  secret,
		URL:     totp.GetSecretURL(account, secret),
	})
}
