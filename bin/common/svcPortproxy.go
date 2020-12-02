package common

import (
	"net"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

// RunPortproxy 运行端口代理服务
func RunPortproxy(t tunnel.Tunnel, cls def.Cluster, desc string, param map[string]string) {
	listenAddr := param["listen"]
	targetAddr := param["target"]

	l, err := Listen(t, "tcp4", listenAddr)
	if err != nil {
		log.Get().WithError(err).Error("portproxy listen failed")
		return
	}

	log.Get().Info("portproxy running: ", desc)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Get().WithError(err).Error("accept failed")
			return
		}

		go func(from net.Conn) {
			defer from.Close()

			to, err := Dial(t, cls, "tcp4", targetAddr)
			if err != nil {
				log.Get().WithError(err).Error("dial target failed")
				return
			}

			utils.LinkReadWriteCloser(from, to)
		}(conn)
	}
}
