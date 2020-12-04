package common

import (
	"errors"
	"fmt"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/sirupsen/logrus"
)

// RunSocks5Server 运行socks5服务
func RunSocks5Server(t tunnel.Tunnel, cls def.Cluster, param map[string]string, log *logrus.Entry) {
	listenAddr := param["listen"]
	username := param["username"]
	password := param["password"]

	l, err := Listen(t, "tcp4", listenAddr)
	if err != nil {
		log.WithError(err).Error("socks5 listen failed")
		return
	}

	s := socks5.NewServer()
	checker := socks5.NoAuthChecker()
	if username != "" || password != "" {
		checker = socks5.PswdAuthChecker(func(u, p string, ctx map[string]string) error {
			if u == username && p == password {
				return nil
			}
			return errors.New("username or password invalid")
		})
	}
	s.SetAuthChecker(checker)

	go s.Run(l)

	log.
		WithField("info", fmt.Sprintf("socks5://%v", param["listen"])).
		Info("service.socks5 is running")
}
