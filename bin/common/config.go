package common

// Config 配置文件
type Config struct {
	Tunnel   TunnelInfo    `json:"tunnel"`
	Services []ServiceInfo `json:"services"`
}
