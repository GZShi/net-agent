package main

import (
	"io"
	"net"
	"strings"

	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/transport"
	"github.com/sirupsen/logrus"
)

func startPortproxyServer(tunnelCluster *transport.TunnelCluster, rawRules string) {
	rawRules = strings.Trim(rawRules, " ")
	if rawRules == "" {
		return
	}
	rules := strings.Split(rawRules, ",")

	for index, rule := range rules {
		go func(index int, rule string) {
			logger := log.Get().WithFields(logrus.Fields{
				"rule":  rule,
				"index": index,
			})
			params := strings.Split(rule, "=")
			localAddr := strings.Trim(params[0], " ")
			agentInfo := params[1]

			params = strings.Split(agentInfo, "@")
			agentAddr := strings.Trim(params[0], " ")
			agentName := strings.Trim(params[1], " ")

			if localAddr == "" || agentAddr == "" || agentName == "" {
				logger.Warn("portproxy rule has empty fields")
				return
			}

			l, err := net.Listen("tcp", localAddr)
			if err != nil {
				logger.WithError(err).Warn("portproxy server listen failed")
				return
			}

			logger.Info("portproxy server is running")

			for {
				conn, err := l.Accept()
				if err != nil {
					break
				}

				go func() {
					defer conn.Close()
					// dial by tunnel
					remote, err := tunnelCluster.Dial("", "tcp", agentAddr, agentName, "*")
					if err != nil || remote == nil {
						logger.WithError(err).Warn("portproxy dial by tunnel failed")
						return
					}

					// relay data
					defer remote.Close()

					logger.WithField("raddr", conn.RemoteAddr().String()).Info("portproxy relay data")
					go io.Copy(remote, conn)
					io.Copy(conn, remote)
				}()
			}
			logger.Info("portproxy server stopped")
		}(index, rule)
	}
}
