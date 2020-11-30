package def

const (
	namePrefix         = "cluster"
	nameOfJoin         = "cluster/join"
	nameOfDetach       = "cluster/detach"
	nameOfSetLabels    = "cluster/setLabels"
	nameOfRemoveLabels = "cluster/removeLabels"
)

// TID tunnel id
type TID uint32

// Cluster 集群管理
type Cluster interface {
	Login() (TID, error)
	Logout() error
	DialByTID(tid TID, writeSID uint32, network, address string) (readSID uint32, err error)

	SetLabel(label string) error

	CreateGroup(name, password, desc string, canBeSearch bool) error
	JoinGroup(groupID uint32, password string) error
	LeaveGroup(groupID uint32) error
	SendGroupMessage(groupID uint32, message string, msgType int) error
}
