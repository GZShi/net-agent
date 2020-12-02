package common

// TunnelInfo 隧道连接信息
type TunnelInfo struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	VHost    string `json:"vhost"`
}

// Config 配置文件
type Config struct {
	Tunnel   TunnelInfo    `json:"tunnel"`
	Services []ServiceInfo `json:"services"`
}
