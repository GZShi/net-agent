package common

import (
	"net"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
	"github.com/sirupsen/logrus"
)

// RunPortproxy 运行端口代理服务
func RunPortproxy(t tunnel.Tunnel, cls def.Cluster, param map[string]string, log *logrus.Entry) {
	listenAddr := param["listen"]
	targetAddr := param["target"]

	l, err := Listen(t, "tcp4", listenAddr)
	if err != nil {
		log.WithError(err).Error("portproxy listen failed")
		return
	}

	// 服务启动后不应该阻塞主线程
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.WithError(err).Error("accept failed")
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				to, err := Dial(t, cls, "tcp4", targetAddr)
				if err != nil {
					log.WithError(err).Error("dial target failed")
					return
				}

				utils.LinkReadWriteCloser(from, to)
			}(conn)
		}
	}()
}