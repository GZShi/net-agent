package api

import (
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/ws"
	"github.com/kataras/iris"
)

func (s *HTTPServer) enableWebSocketRoute(ns iris.Party, pathstr string) {

	party := ns.Party(pathstr)

	party.Get("/agent", func(ctx iris.Context) {
		if s.proto == nil {
			ctx.Text("websocket upgrade failed, proto is nil")
			return
		}

		conn, err := ws.Upgrade(ctx.ResponseWriter(), ctx.Request())
		if err != nil {
			ctx.Text("websocket upgrade failed")
			return
		}
		s.proto.DispatchConn(conn)
		log.Get().Info("websocket upgrade success")
	})
}
