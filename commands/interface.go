package commands

import "github.com/GZShi/net-agent/transport"

// Command 接口
type Command interface {
	Bytes() ([]byte, error)
	Parse([]byte) error
	Exec(*transport.Tunnel) ([]byte, error)
}
