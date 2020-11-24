package exchanger

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/GZShi/net-agent/tunnel"
)

type Cluster interface {
	Join(t tunnel.Tunnel) error
	Detach(t tunnel.Tunnel)

	SetLabels(tid uint32, label []string) error
	RemoveLabels(tid uint32, label []string) error

	FindTunnelByID(id uint32) (tunnel.Tunnel, error)
	FindTunnelByLabel(label string) (tunnel.Tunnel, error)
}

type tunnelList struct {
	listByID []tunnel.Tunnel
	ids      map[uint32]tunnel.Tunnel

	mut        sync.RWMutex
	idSequence uint32
}

// NewCluster 创建新的tunnel集群
func NewCluster() Cluster {
	return &tunnelList{}
}

func (ts *tunnelList) NextID() uint32 {
	return atomic.AddUint32(&ts.idSequence, 1)
}

func (ts *tunnelList) Join(t tunnel.Tunnel) error {
	ts.mut.Lock()
	defer ts.mut.Unlock()

	id := ts.NextID()
	ts.ids[id] = t
	ts.listByID = append(ts.listByID, t)

	return nil
}

func (ts *tunnelList) Detach(t tunnel.Tunnel) {
	ts.mut.Lock()
	defer ts.mut.Unlock()
}

func (ts *tunnelList) FindTunnelByID(id uint32) (tunnel.Tunnel, error) {
	ts.mut.RLock()
	defer ts.mut.RUnlock()

	return nil, errors.New("not found")
}

func (ts *tunnelList) FindTunnelByLabel(label string) (tunnel.Tunnel, error) {
	ts.mut.RLock()
	defer ts.mut.RUnlock()

	return nil, errors.New("not found")
}
