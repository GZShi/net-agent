package config

// TunnelInfo 隧道连接信息
type TunnelInfo struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	VHost    string `json:"vhost"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Enable bool              `json:"enable"`
	Desc   string            `json:"description"`
	Type   string            `json:"type"`
	Param  map[string]string `json:"param"`
}

// Config 配置文件
type Config struct {
	Tunnel   TunnelInfo    `json:"tunnel"`
	Services []ServiceInfo `json:"services"`
}
