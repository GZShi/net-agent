package service

import (
	"errors"

	"github.com/GZShi/net-agent/rpc/cluster/def"
)

var errNotImplement = errors.New("method not implement")

// New 获取新的实例
func New() def.Cluster {
	return &impl{}
}

type impl struct{}

func (p *impl) Login() error {
	return errNotImplement
}

func (p *impl) Logout() error {
	return errNotImplement
}

func (p *impl) SetLabel(label string) error {
	return errNotImplement
}

func (p *impl) CreateGroup(name, password, desc string, canBeSearch bool) error {
	return errNotImplement
}

func (p *impl) JoinGroup(groupID uint32, password string) error {
	return errNotImplement
}

func (p *impl) LeaveGroup(groupID uint32) error {
	return errNotImplement
}

func (p *impl) SendGroupMessage(groupID uint32, message string, msgType int) error {
	return errNotImplement
}
