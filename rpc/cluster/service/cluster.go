package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

type cluster struct {
	ids map[def.TID]tunnel.Tunnel

	vhostMap map[string]def.TID

	labelMap map[string]*tlist

	mut        sync.RWMutex
	idSequence uint32
}

var instance *cluster
var onceInit sync.Once

// getCluster 单例模式
func getCluster() *cluster {
	onceInit.Do(func() {
		instance = newCluster()
	})
	return instance
}

// newCluster 创建新的tunnel集群管理服务
func newCluster() *cluster {
	return &cluster{
		idSequence: 1,
		ids:        make(map[def.TID]tunnel.Tunnel),
		vhostMap:   make(map[string]def.TID),
		labelMap:   make(map[string]*tlist),
	}
}

func (ts *cluster) NextTID() def.TID {
	return def.TID(atomic.AddUint32(&ts.idSequence, 1))
}

func (ts *cluster) lookup(vhost string) (def.TID, error) {
	tid, found := ts.vhostMap[vhost]
	if !found {
		return 0, errors.New("vhost record not found")
	}
	return tid, nil
}

func (ts *cluster) Lookup(vhost string) (def.TID, error) {
	ts.mut.Lock()
	defer ts.mut.Unlock()
	return ts.lookup(vhost)
}

func (ts *cluster) Join(t tunnel.Tunnel, vhost string) (def.TID, error) {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	_, err := ts.lookup(vhost)
	if err == nil {
		return 0, errors.New("vhost exists")
	}

	tid := ts.NextTID()
	ts.ids[tid] = t
	ts.vhostMap[vhost] = tid

	return tid, nil
}

func (ts *cluster) Detach(tid def.TID) {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	delete(ts.ids, tid)
}

func (ts *cluster) findByID(id def.TID) (tunnel.Tunnel, error) {
	t, found := ts.ids[id]
	if found {
		return t, nil
	}

	return nil, fmt.Errorf("tid=%v not found", id)
}

func (ts *cluster) FindTunnelByID(tid def.TID) (tunnel.Tunnel, error) {
	ts.mut.RLock()
	defer ts.mut.RUnlock()
	return ts.findByID(tid)
}

func (ts *cluster) findByLabel(label string) (*tlist, error) {
	list, found := ts.labelMap[label]
	if !found || list == nil {
		return nil, errors.New("label not found")
	}
	return list, nil
}

func (ts *cluster) SelectTunnelByLabel(label string) (tunnel.Tunnel, error) {
	ts.mut.RLock()
	list, err := ts.findByLabel(label)
	ts.mut.RUnlock()

	tid, err := list.Select()
	if err != nil {
		return nil, err
	}

	return ts.FindTunnelByID(tid)
}

func (ts *cluster) SetLabels(tid def.TID, labels []string) error {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	for _, label := range labels {
		label = strings.Trim(label, " ")
		list, err := ts.findByLabel(label)
		if err == nil {
			list.Append(tid)
		}
	}

	return nil
}

func (ts *cluster) RemoveLabels(tid def.TID, labels []string) error {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	for _, label := range labels {
		label = strings.Trim(label, " ")
		list, err := ts.findByLabel(label)
		if err != nil {
			list.Remove(tid)
		}
	}

	return nil
}
