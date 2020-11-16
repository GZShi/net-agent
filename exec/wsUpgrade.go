package main

import (
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/ws"
	"github.com/kataras/iris"
)

func MakeAgentUpgrader(manager *protocol.ProtoManager) func(iris.Context) {
	return func(ctx iris.Context) {
		r := ctx.Request()
		w := ctx.ResponseWriter()

		conn, err := ws.Upgrade(w, r)
		if err != nil {
			ctx.Text("upgrade failed")
			return
		}

		manager.DispatchConn(conn)

		log.Get().Info("websocket upgrade success")
	}
}
