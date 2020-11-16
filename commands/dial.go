package commands

import (
	"encoding/json"

	"github.com/GZShi/net-agent/transport"
)

// Dial 访问请求
type Dial struct {
	Network    string `json:"network"`
	SourceAddr string `json:"source"`
	ClientName string `json:"client"`
	TargetAddr string `json:"target"`
}

type DialAns struct {
	PortID int `json:"portID"`
}

// NewDial 构造新的请求
func NewDial(network, source, client, target string) *Dial {
	return &Dial{
		network, source, client, target,
	}
}

// Bytes 生成二进制数据
func (d *Dial) Bytes() ([]byte, error) {
	return json.Marshal(d)
}

// Parse 解析二进制数据
func (d *Dial) Parse(buf []byte) error {
	return json.Unmarshal(buf, d)
}

// Exec 执行命令
func (d *Dial) Exec(t *transport.Tunnel) ([]byte, error) {
	return nil, nil
}
