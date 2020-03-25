package transport

// GetStatus 获取同名管道状态
func (p *TunnelList) GetStatus() interface{} {
	p.listMut.Lock()
	list := make([]*Tunnel, len(p.list))
	copy(list, p.list)
	p.listMut.Unlock()

	details := []interface{}{}
	for _, t := range list {
		status := t.GetStatus()
		activeConns := t.GetActiveConns()
		historyConns := t.GetHistoryConns()

		details = append(details, struct {
			Status       interface{} `json:"status"`
			ActiveConns  interface{} `json:"activeConns"`
			HistoryConns interface{} `json:"historyConns"`
		}{status, activeConns, historyConns})
	}

	return struct {
		Name    string `json:"name"`
		Count   int    `json:"count"`
		Details []interface{}
	}{
		p.name,
		len(list),
		details,
	}
}
