package transport

import (
	"time"
)

// GetStatus 获取通道的状态
func (t *Tunnel) GetStatus() interface{} {

	var currDataTransDuration time.Duration
	if t.currTransData != nil {
		currDataTransDuration = time.Since(t.currTransTime)
	}

	return struct {
		Name                  string        `json:"name"`
		RemoteAddr            string        `json:"addr"`
		Created               time.Time     `json:"created"`
		Finished              time.Time     `json:"finished"`
		UploadSize            uint64        `json:"upSize"`
		DownloadSize          uint64        `json:"downSize"`
		UploadPack            uint64        `json:"upPack"`
		DownloadPack          uint64        `json:"downPack"`
		ActivePortCount       int32         `json:"activePortCount"`
		FinishedPortCount     int32         `json:"finishedPortCount"`
		FailedPortCount       int32         `json:"failedPortCount"`
		CurrDataTransDuration time.Duration `json:"currDataTransDuration"`
	}{
		t.secretName,
		t.conn.RemoteAddr().String(),
		t.created,
		t.finished,
		t.uploadSize,
		t.downloadSize,
		t.uploadPack,
		t.downloadPack,
		t.activePortCount,
		t.finishedPortCount,
		t.failedPortCount,
		currDataTransDuration,
	}
}

// GetActiveConns ...
func (t *Tunnel) GetActiveConns() interface{} {
	conns := []interface{}{}
	t.ports.Range(func(_, value interface{}) bool {
		port := value.(*dataPort)
		conns = append(conns, port.GetStatus())
		return true
	})
	return struct {
		Name          string        `json:"name"`
		LastHeartbeat uint32        `json:"lastHeartbeat"`
		Conns         []interface{} `json:"conns"`
	}{
		t.secretName,
		t.lastHeartbeat,
		conns,
	}
}

// GetHistoryConns ...
func (t *Tunnel) GetHistoryConns() interface{} {
	t.donePortsMut.Lock()
	length := len(t.donePorts)
	ports := t.donePorts[0:]
	t.donePortsMut.Unlock()

	conns := []interface{}{}
	for _, port := range ports {
		conns = append(conns, port.GetStatus())
	}

	return struct {
		Name   string        `json:"name"`
		Length int           `json:"length"`
		Conns  []interface{} `json:"conns"`
	}{
		t.secretName,
		length,
		conns,
	}
}
