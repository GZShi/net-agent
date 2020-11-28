package cluster

const (
	namePrefix         = "cluster"
	nameOfJoin         = "cluster/join"
	nameOfDetach       = "cluster/detach"
	nameOfSetLabels    = "cluster/setLabels"
	nameOfRemoveLabels = "cluster/removeLabels"
)

type Cluster interface {
	Login() error
	Logout() error

	SetLabel(label string) error

	CreateGroup(name, password, desc string, canBeSearch bool) error
	JoinGroup(groupID uint32, password string) error
	LeaveGroup(groupID uint32) error
	SendGroupMessage(groupID uint32, message string, msgType int) error
}
