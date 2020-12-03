package common

// WebsocketInfo 基于websocket协议进行连接
type WebsocketInfo struct {
	Enable bool   `json:"enable"`
	Path   string `json:"path"`
}
