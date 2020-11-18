package api

import (
	"net"

	"github.com/GZShi/net-agent/protocol"
	"github.com/GZShi/net-agent/transport"
	"github.com/kataras/iris"
)

// HTTPServer http api 服务，支持管理端功能等
type HTTPServer struct {
	app     *iris.Application
	cluster *transport.TunnelCluster
	proto   *protocol.ProtoManager
	totpMap map[string]*TotpInfo
}

// NewHTTPServer 创建新的http api服务
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{
		app:     iris.New(),
		cluster: nil,
		proto:   nil,
		totpMap: nil,
	}
}

// Run 运行服务
func (s *HTTPServer) Run(l net.Listener) error {
	ns := s.app.Party("/naapi")
	s.enableClusterRoute(ns, "tunnel")
	s.enableTotpRoute(ns, "totp")
	s.enableBasicRoute(ns, "basic")
	s.enableWebSocketRoute(ns, "ws")

	return s.app.Run(iris.Listener(l))
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
