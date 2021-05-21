package common

import (
	"io"
	"net"
	"time"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/msgclient"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
	"github.com/sirupsen/logrus"
)

func response(ctx iris.Context, data interface{}, err error) {
	if err != nil {
		ctx.JSON(&struct {
			ErrCode int
			ErrMsg  string
		}{-1, err.Error()})
		return
	}
	ctx.JSON(&struct {
		ErrCode int
		ErrMsg  string
		Data    interface{}
	}{0, "success", data})
}

// RunChatServer 文件服务
func RunChatServer(t tunnel.Tunnel, cls def.Cluster, param map[string]string, log *logrus.Entry) (io.Closer, error) {

	log.Debug("chat server init")

	l, err := net.Listen("tcp4", "127.0.0.1:2021")
	if err != nil {
		return nil, err
	}

	// 在客户端运行推送客户端服务，用于处理远端推送消息
	svc := msgclient.NewService()
	t.BindServices(svc)

	app := iris.New()
	app.StaticWeb("/example", "./ui-html/example")
	app.StaticWeb("/site", "./ui-html/site")
	app.Get("/say-hello", func(ctx iris.Context) {
		ctx.Write([]byte("who you are~ ?"))
	})
	api := app.Party("/agent-api")
	{
		api.Get("/ctx-info", func(ctx iris.Context) {
			info, err := cls.GetCtxInfo()
			response(ctx, &info, err)
		})
		api.Post("/new-message", func(ctx iris.Context) {
			var msg struct {
				GroupID uint32 `json:"groupID"`
				MsgType int    `json:"msgType"`
				Message string `json:"message"`
			}
			ctx.ReadJSON(&msg)

			err := cls.SendGroupMessage(msg.GroupID, msg.Message, msg.MsgType)
			response(ctx, nil, err)
		})
		api.Post("/recent-message", func(ctx iris.Context) {
			msgs, err := cls.GetGroupMessages([]uint32{0}, time.Now().Add(-7*24*time.Hour), 100)
			response(ctx, msgs, err)
		})
	}

	// 设置websocket接口
	ws := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})
	ws.OnConnection(func(c websocket.Connection) {
		log.Debug("websocket connected")
		svc.RegisterWSClient(c)
		c.On("new-message", func(msg string) {
			log.WithField("msg", msg).Debug("new-message")
		})
		c.OnDisconnect(func() {
			svc.UnregisterWSClient(c)
		})
	})
	app.Get("ws-conn", ws.Handler())

	go app.Run(iris.Listener(l))

	// todo:
	return nil, nil
}
