// genXXX 文件主要负责RPC数据格式转换
// *client 是开放给客户端调用的方法
// *svc 是服务端收到客户端调用时的数据处理过程

package cluster

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
