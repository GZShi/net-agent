package transport

import "encoding/json"

const (
	// cmdDialReq 让对端直接拨号
	cmdDialReq = iota
	cmdDialSuccess
	cmdDialFailed
	cmdData
	cmdClose
	cmdHeartbeat
	cmdTextMessages
	// cmdChannelDialReq 让对端在集群中进行选择拨号
	cmdChannelDialReq
)

// Data 在隧道中传输的数据包
type Data struct {
	ConnID cid
	Cmd    uint8
	Bytes  []byte
}

type channelDialData struct {
	SourceAddr  string `json:"srcAddr"`
	TargetAddr  string `json:"addr"`
	ChannelName string `json:"channel"`
	UserName    string `json:"user"`
}

func newClusterDialReq(connID cid, srcAddr, addr, channelName, userName string) Data {
	data, err := json.Marshal(&channelDialData{srcAddr, addr, channelName, userName})
	if err != nil {
		data = nil
	}
	return Data{connID, cmdChannelDialReq, data}
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
