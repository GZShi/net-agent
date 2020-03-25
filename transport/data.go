package transport

const (
	cmdDialReq = iota
	cmdDialSuccess
	cmdDialFailed
	cmdData
	cmdClose
	cmdHeartbeat
	cmdTextMessages
)

// Data 在隧道中传输的数据包
type Data struct {
	ConnID cid
	Cmd    uint8
	Bytes  []byte
}

func newDialReq(connID cid, addr string) Data {
	return Data{connID, cmdDialReq, []byte(addr)}
}

func newDialAns(connID cid, err error) Data {
	if err != nil {
		return Data{connID, cmdDialFailed, []byte(err.Error())}
	}
	return Data{connID, cmdDialSuccess, nil}
}

func newCloseData(connID cid) Data {
	return Data{connID, cmdClose, nil}
}

func newHeartbeatData(connID cid) Data {
	return Data{connID, cmdHeartbeat, nil}
}

func newTextMessageData(connID cid, text string) Data {
	return Data{connID, cmdTextMessages, []byte(text)}
}
