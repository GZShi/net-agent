package common

import (
	"errors"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/socks5"
	"github.com/GZShi/net-agent/tunnel"
)

// RunSocks5Server 运行socks5服务
func RunSocks5Server(t tunnel.Tunnel, cls def.Cluster, desc string, param map[string]string) {
	listenAddr := param["listen"]
	username := param["username"]
	password := param["password"]

	l, err := Listen(t, "tcp4", listenAddr)
	if err != nil {
		log.Get().WithError(err).Error("socks5 listen failed")
		return
	}

	s := socks5.NewServer()
	checker := socks5.NoAuthChecker()
	if username != "" || password != "" {
		checker = socks5.PswdAuthChecker(func(u, p string) error {
			if u == username && p == password {
				return nil
			}
			return errors.New("username or password invalid")
		})
	}
	s.SetAuthChecker(checker)

	log.Get().Info("socks5 server running: ", desc)
	s.Run(l)
}
