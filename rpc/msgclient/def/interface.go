package def

const (
	namePrefix = "msgclient"
)

type MsgClient interface {
	PushGroupMessage(sender string, groupID uint32, message string, msgType int)
	PushSysNotify(groupID uint32, message string, msgType int)
}
