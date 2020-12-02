package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

type cluster struct {
	tdata  sync.Map // map[tunnel.Tunnel]*tunData
	ids    sync.Map // map[def.TID]tunnel.Tunnel
	vhosts sync.Map // map[string]def.TID

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
		labelMap:   make(map[string]*tlist),
	}
}

func (ts *cluster) NextTID() def.TID {
	return def.TID(atomic.AddUint32(&ts.idSequence, 1))
}

func (ts *cluster) Lookup(vhost string) (def.TID, error) {
	if !strings.HasSuffix(vhost, ".tunnel") {
		return 0, errors.New("can't resolve vhost: " + vhost)
	}
	val, found := ts.vhosts.Load(vhost[:len(vhost)-7])
	if !found {
		return 0, errors.New("vhost record not found: " + vhost)
	}
	d := val.(*tunData)
	return d.tid, nil
}

func (ts *cluster) Join(t tunnel.Tunnel, vhost string) (*tunData, error) {

	// 第一步，判断是否已经存在
	d := &tunData{
		t:   t,
		tid: ts.NextTID(),
	}
	_, loaded := ts.tdata.LoadOrStore(t, d)
	if loaded {
		return nil, errors.New("tunnel found")
	}

	ts.ids.Store(d.tid, d)

	// bind vhost to tunData
	// todo: 优化性能
	_, loaded = ts.vhosts.LoadOrStore(vhost, d)
	for loaded {
		vhost = utils.NextNameStr(vhost)
		_, loaded = ts.vhosts.LoadOrStore(vhost, d)
	}
	d.vhost = vhost

	return d, nil
}

func (ts *cluster) Detach(t tunnel.Tunnel) {
	val, loaded := ts.tdata.LoadAndDelete(t)
	if !loaded {
		return
	}
	d := val.(*tunData)
	ts.ids.Delete(d.tid)
	ts.vhosts.Delete(d.vhost)
}

func (ts *cluster) FindTunnelByID(tid def.TID) (tunnel.Tunnel, error) {
	val, found := ts.ids.Load(tid)
	if found {
		return val.(*tunData).t, nil
	}

	return nil, fmt.Errorf("tid=%v not found", tid)
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
