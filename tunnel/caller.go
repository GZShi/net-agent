package tunnel

// Caller rpc调用者信息
// 用于跟踪调用链路，避免形成环路
type Caller interface {
	GetTunnelID() uint32
	GetCMD() string
	GetFrameID() uint32
}

type caller struct {
	tunnelID uint32
	cmd      string
	frameID  uint32
}

func (t *tunnel) NewCaller(cmd string, frameID uint32) Caller {
	return &caller{
		tunnelID: 0, // todo: get tunnel id
		cmd:      cmd,
		frameID:  frameID,
	}
}

func (c *caller) GetTunnelID() uint32 {
	return c.tunnelID
}

func (c *caller) GetCMD() string {
	return c.cmd
}

func (c *caller) GetFrameID() uint32 {
	return c.frameID
}
