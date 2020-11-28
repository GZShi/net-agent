package dial

import "github.com/GZShi/net-agent/tunnel"

//
// gen code for: Dial
//

type stReqDial struct {
	WriteSID uint32 `json:"writeSID"`
	Network  string `json:"network"`
	Address  string `json:"address"`
}

type stRespDial struct {
	ReadSID uint32 `json:"readSID"`
}

func (c *client) Dial(writeSID uint32, network, address string) (readSID uint32, err error) {
	var resp stRespDial
	err = c.t.SendJSON(c.ctx, "Dial", &stReqDial{writeSID, network, address}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.ReadSID, nil
}

func (s *svc) Dial(ctx tunnel.Context) {
	var req stReqDial
	err := ctx.GetJSON(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ReadSID, err := s.impl.Dial(req.WriteSID, req.Network, req.Address)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(&stRespDial{ReadSID})
}
