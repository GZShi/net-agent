package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/rpc/msgclient"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/GZShi/net-agent/utils"
)

type cluster struct {
	tunnelMapCtx sync.Map // map[tunnel.Tunnel]*connContext
	tidMapCtx    sync.Map // map[def.TID]*connContext
	vhostMapCtx  sync.Map // map[string]*connContext

	labelMap map[string]*tlist

	mut        sync.RWMutex
	idSequence uint32
}

var clusterInstance *cluster
var onceClusterInit sync.Once

// getCluster 单例模式
func getCluster() *cluster {
	onceClusterInit.Do(func() {
		clusterInstance = newCluster()
	})
	return clusterInstance
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
	val, found := ts.vhostMapCtx.Load(vhost[:len(vhost)-7])
	if !found {
		return 0, errors.New("vhost record not found: " + vhost)
	}
	d := val.(*connectContext)
	return d.tid, nil
}

func (ts *cluster) Join(t tunnel.Tunnel, vhost string) (*connectContext, error) {

	if len(vhost) < 4 {
		return nil, errors.New("vhost too short")
	}

	// 第一步，判断是否已经存在
	ctx := &connectContext{
		t:             t,
		tid:           ts.NextTID(),
		msgClient:     msgclient.NewClient(t, nil),
		connectTime:   time.Now(),
		lastHeartbeat: time.Now(),
	}
	_, loaded := ts.tunnelMapCtx.LoadOrStore(t, ctx)
	if loaded {
		return nil, errors.New("tunnel repeat login")
	}

	ts.tidMapCtx.Store(ctx.tid, ctx)

	// bind vhost to tunData
	// todo: 优化性能
	val, loaded := ts.vhostMapCtx.LoadOrStore(vhost, ctx)
	for loaded {
		existCtx := val.(*connectContext)
		if existCtx.t == nil {
			existCtx.vhost = "" // 要置为空，否则close时会误删tid和tunData的绑定
			ts.vhostMapCtx.Store(vhost, ctx)
			break
		}

		err := existCtx.t.Ping()
		if err != nil {
			existCtx.vhost = "" // 要置为空，否则close时会误删tid和tunData的绑定
			ts.vhostMapCtx.Store(vhost, ctx)
			break
		}

		vhost = utils.NextNameStr(vhost)
		val, loaded = ts.vhostMapCtx.LoadOrStore(vhost, ctx)
	}
	ctx.vhost = vhost

	return ctx, nil
}

func (ts *cluster) Detach(t tunnel.Tunnel) {
	val, loaded := ts.tunnelMapCtx.LoadAndDelete(t)
	if !loaded {
		return
	}
	d := val.(*connectContext)
	ts.tidMapCtx.Delete(d.tid)
	ts.vhostMapCtx.Delete(d.vhost)
}

func (ts *cluster) FindTunnelByID(tid def.TID) (tunnel.Tunnel, error) {
	val, found := ts.tidMapCtx.Load(tid)
	if found {
		return val.(*connectContext).t, nil
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

// 向所有在线链接派发推送消息
func (ts *cluster) DispatchGMToAll(m *msg) {
	// 找到所有的MsgClients
	ts.vhostMapCtx.Range(func(k, v interface{}) bool {
		msgc := v.(*connectContext).msgClient
		msgc.PushGroupMessage(m.SenderVhost, m.GroupID, m.Message, m.MsgType)
		return true
	})
}

// 向指定vhosts派发推送消息
func (ts *cluster) DispatchGMToVhosts(m *msg, vhosts []string) {
	for _, vhost := range vhosts {
		ctx, found := ts.vhostMapCtx.Load(vhost)
		if !found {
			continue
		}
		ctx.(*connectContext).msgClient.PushGroupMessage(m.SenderVhost, m.GroupID, m.Message, m.MsgType)
	}
}
