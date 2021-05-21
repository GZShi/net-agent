package def

import (
	"net"
	"time"
)

// TID tunnel id
type TID uint32

type Message struct {
	ID      int       `json:"id"`
	GroupID uint32    `json:"groupID"`
	Sender  string    `json:"sender"`
	Date    time.Time `json:"date"`
	Type    int       `json:"type"`
	Content string    `json:"content"`
}

type CtxInfo struct {
	VHost string `json:"vhost"`
}

// Cluster 集群管理
type Cluster interface {
	Login(vhost string) (TID, string, error)
	Logout() error
	Heartbeat() error
	GetCtxInfo() (CtxInfo, error)

	DialByTID(tid TID, writeSID uint32, network, address string) (readSID uint32, err error)
	Dial(vhost string, vport uint32) (net.Conn, error)

	SetLabel(label string) error

	CreateGroup(name, password, desc string, canBeSearch bool) error
	JoinGroup(groupID uint32, password string) error
	LeaveGroup(groupID uint32) error
	SendGroupMessage(groupID uint32, message string, msgType int) error
	GetGroupMessages(groupIDs []uint32, startTime time.Time, limit int) ([]Message, error)
}
