package exchanger

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GZShi/net-agent/tunnel"
)

// Cluster 集群管理
type Cluster interface {
	Join(t tunnel.Tunnel) (TID, error)
	Detach(tid TID)

	SetLabels(tid TID, label []string) error
	RemoveLabels(tid TID, label []string) error

	FindTunnelByID(id TID) (tunnel.Tunnel, error)
	SelectTunnelByLabel(label string) (tunnel.Tunnel, error)
}

type cluster struct {
	ids map[TID]tunnel.Tunnel

	labelMap map[string]*tlist

	mut        sync.RWMutex
	idSequence uint32
}

// NewCluster 创建新的tunnel集群
func NewCluster() Cluster {
	return &cluster{
		idSequence: 1,
	}
}

func (ts *cluster) NextTID() TID {
	return TID(atomic.AddUint32(&ts.idSequence, 1))
}

func (ts *cluster) Join(t tunnel.Tunnel) (TID, error) {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	id := ts.NextTID()
	ts.ids[id] = t

	return id, nil
}

func (ts *cluster) Detach(tid TID) {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	delete(ts.ids, tid)
}

func (ts *cluster) findByID(id TID) (tunnel.Tunnel, error) {
	t, found := ts.ids[id]
	if found {
		return t, nil
	}

	return nil, errors.New("tid not found")
}

func (ts *cluster) FindTunnelByID(tid TID) (tunnel.Tunnel, error) {
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

func (ts *cluster) SetLabels(tid TID, labels []string) error {
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

func (ts *cluster) RemoveLabels(tid TID, labels []string) error {
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
