package cluster

func (c *client) Login() error {
	return c.t.SendJSON(c.ctx, "Login", nil, nil)
}

func (c *client) Logout() error {
	return c.t.SendJSON(c.ctx, "Logout", nil, nil)
}

func (c *client) SetLabel(label string) error {
	return c.t.SendJSON(c.ctx, "SetLabel", &struct {
		Label string `json:"label"`
	}{label}, nil)
}

func (c *client) CreateGroup(name, password, desc string, canBeSearch bool) error {
	return c.t.SendJSON(c.ctx, "CreateGroup", &struct {
		Name        string `json:"name"`
		Password    string `json:"password"`
		Description string `json:"desc"`
		CanBeSearch bool   `json:"canBeSearch"`
	}{name, password, desc, canBeSearch}, nil)
}

func (c *client) JoinGroup(groupID uint32, password string) error {
	return c.t.SendJSON(c.ctx, "JoinGroup", &struct {
		GroupID  uint32 `json:"groupID"`
		Password string `json:"password"`
	}{groupID, password}, nil)
}

func (c *client) LeaveGroup(groupID uint32) error {
	return c.t.SendJSON(c.ctx, "LeaveGroup", &struct {
		GroupID uint32 `json:"groupID"`
	}{groupID}, nil)
}

func (c *client) SendGroupMessage(groupID uint32, message string, msgType int) error {
	return c.t.SendJSON(c.ctx, "SendGroupMessage", &struct {
		GroupID uint32 `json:"groupID"`
		Message string `json:"message"`
		MsgType int    `json:"msgType"`
	}{groupID, message, msgType}, nil)
}
