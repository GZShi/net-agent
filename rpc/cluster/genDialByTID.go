package cluster

import (
	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

type stReqDialByTID struct {
	TID      def.TID `json:"tid"`
	WriteSID uint32  `json:"writeSID"`
	Network  string  `json:"network"`
	Address  string  `json:"address"`
}

type stRespDialByTID struct {
	ReadSID uint32 `json:"readSID"`
}

func (c *client) DialByTID(
	tid def.TID,
	writeSID uint32,
	network, address string,
) (readSID uint32, err error) {
	var resp stRespDialByTID
	err = c.t.SendJSON(
		c.ctx,
		tunnel.JoinServiceMethod(c.prefix, "DialByTID"),
		&stReqDialByTID{tid, writeSID, network, address},
		&resp,
	)
	if err != nil {
		return 0, err
	}

	return resp.ReadSID, nil
}

func (s *svc) DialByTID(ctx tunnel.Context) {
	var req stReqDialByTID
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ReadSID, err := s.impl.DialByTID(req.TID, req.WriteSID, req.Network, req.Address)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(&stRespDialByTID{ReadSID})
}
