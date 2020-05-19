package transport

import "time"

// GetStatus 获取集群中所有隧道的信息
func (p *TunnelCluster) GetStatus(tlistName string) interface{} {
	var info interface{}

	p.groups.Range(func(_, value interface{}) bool {
		tList := value.(*TunnelList)

		if tList.name == tlistName {
			info = tList.GetStatus()
		}
		return true
	})

	return struct {
		Now  time.Time   `json:"now"`
		Info interface{} `json:"info"`
	}{
		time.Now(),
		info,
	}
}
