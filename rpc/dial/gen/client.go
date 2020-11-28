package gen

func (c *client) Dial(writeSID uint32, network, address string) (readSID uint32, err error) {
	var resp struct {
		ReadSID uint32 `json:"readSID"`
	}
	err = c.t.SendJSON(c.ctx, "Dial", &struct {
		WriteSessionID uint32 `json:"writeSID"`
		Network        string `json:"network"`
		Address        string `json:"address"`
	}{writeSID, network, address}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.ReadSID, nil
}
