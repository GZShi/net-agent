package transport

import "time"

// GetStatus 获取集群中所有隧道的信息
func (p *TunnelCluster) GetStatus() interface{} {
	list := []interface{}{}

	// p.groups.Range(func(_, value interface{}) bool {
	// 	tList := value.(*TunnelList)

	// 	// TODO: 重新设计此接口
	// 	// list = append(list, t.GetStatus())
	// 	return true
	// })

	return struct {
		Now     time.Time     `json:"now"`
		Tunnels []interface{} `json:"tunnels"`
	}{
		time.Now(),
		list,
	}
}

// GetActiveConns 获取活跃连接的信息
func (p *TunnelCluster) GetActiveConns() interface{} {
	list := []interface{}{}

	// p.tunnels.Range(func(_, value interface{}) bool {
	// 	t := value.(*Tunnel)
	// 	list = append(list, t.GetActiveConns())
	// 	return true
	// })

	return struct {
		Now     time.Time     `json:"now"`
		Tunnels []interface{} `json:"tunnels"`
	}{
		time.Now(),
		list,
	}
}

// GetHistoryConns 获取历史连接
func (p *TunnelCluster) GetHistoryConns() interface{} {
	list := []interface{}{}
	// p.tunnels.Range(func(_, value interface{}) bool {
	// 	t := value.(*Tunnel)
	// 	list = append(list, t.GetHistoryConns())
	// 	return true
	// })

	return struct {
		Now     time.Time     `json:"now"`
		Tunnels []interface{} `json:"tunnels"`
	}{
		time.Now(),
		list,
	}
}
