package common

// Config 配置文件
type Config struct {
	Tunnel    TunnelInfo    `json:"tunnel"`
	Websocket WebsocketInfo `json:"websocket"`
	Services  []ServiceInfo `json:"services"`
}
