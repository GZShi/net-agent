package cluster

import (
	"net"

	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

type stReqDial struct {
	WriteSID uint32 `json:"writeSID"`
	Vhost    string `json:"vhost"`
	Vport    uint32 `json:"vport"`
}

type stRespDial struct {
	ReadSID uint32 `json:"readSID"`
}

func (c *client) Dial(vhost string, vport uint32) (conn net.Conn, err error) {
	stream, sid := c.t.NewStream()
	defer func() {
		if err != nil {
			stream.Close()
		}
	}()

	var resp stRespDial
	err = c.t.SendJSON(
		c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "Dial"),
		&stReqDial{
			WriteSID: sid,
			Vhost:    vhost,
			Vport:    vport,
		},
		&resp,
	)
	if err != nil {
		return nil, err
	}
	stream.Bind(resp.ReadSID)

	return stream, nil
}

func (s *svc) Dial(ctx tunnel.Context) {
	var req stReqDial
	var resp stRespDial
	var err error

	err = ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	// 找到对应的tunnel，并且访问其vport上的服务，建立连接
	conn, err := s.impl.Dial(req.Vhost, req.Vport)
	if err != nil {
		ctx.Error(err)
		return
	}

	// 与调用方创建对端传输通道，并把写id回传
	stream, sid := s.t.NewStream()
	stream.Bind(req.WriteSID)
	resp.ReadSID = sid

	// 两侧的连接创建成功后，启动协程，不停传输数据
	go func(a, b net.Conn) {
		utils.LinkReadWriteCloser(a, b)
	}(stream, conn)

	ctx.JSON(&resp)
}
