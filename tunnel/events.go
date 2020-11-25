package tunnel

import (
	"fmt"
	"strings"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

// onRequest 基础事件
// 原始连接接收到对端的一个完整RequestFrame
// 这个事件产生后，需要根据Header的解析情况使用对应的回调函数进行处理
// 处理完成后，应该向对端发送ResponseFrame
func (t *tunnel) onRequest(req *Frame) {
	// process frame
	ctx := t.newContext(req)
	cmd := ctx.GetCmd()
	parts := strings.Split(cmd, "/")
	prefix := parts[0]

	if t.serviceMap == nil {
		ctx.Error(fmt.Errorf("service '%v' not found", prefix))
		return
	}

	s, found := t.serviceMap[prefix]
	if !found {
		ctx.Error(fmt.Errorf("service '%v' not found", prefix))
		return
	}
	if s == nil {
		ctx.Error(fmt.Errorf("service '%v' is nil", prefix))
		return
	}

	err := s.Exec(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
}

// onResponse 基础事件
// 原始连接接收到对端的一个ResponseFrame
// 根据Frame.SessionID来确定该应答属于哪个请求
// 要严格遵循“一问一答”的原则，收到应答后应该马上把Guard从Map中删除
func (t *tunnel) onResponse(f *Frame) {
	val, has := t.respGuards.Load(f.SessionID)
	if !has {
		// 丢弃不用
		log.Get().WithField("sessionID", f.SessionID).Error("can't find responseGuard")
		return
	}
	// Todo:5 go1.15中提供了LoadAndDelete方法
	t.respGuards.Delete(f.SessionID)

	guard := val.(*frameGuard)
	guard.ch <- f
	close(guard.ch)
}

// onSteramData 基础事件
// 接收到对端的一个数据传输包
// 数据传输包专门用于进行原始二进制通信
func (t *tunnel) onStreamData(f *Frame) {
	var guard *frameGuard
	// f.Data为nil，代表收到一个EOF信号
	// 如果存在guard则通知guard，否则应该丢弃这个信号
	if f.Data == nil {
		val, loaded := t.streamGuards.Load(f.SessionID)
		if loaded {
			guard = val.(*frameGuard)
		}
	} else {
		temp := &frameGuard{
			ch: make(chan *Frame, 256),
		}

		val, loaded := t.streamGuards.LoadOrStore(f.SessionID, temp)
		if loaded {
			guard = val.(*frameGuard)
		} else {
			guard = temp
		}
	}

	if guard != nil {
		// 如果当前guard深度过长，则消费端可能出现了阻塞的情况
		// 需要关闭连接，否则会阻塞其它Stream正常传输
		if len(guard.ch) > 200 {
			go func(sid uint32) {
				log.Get().Warn("len(chan) > 200")
				<-time.After(time.Second * 3)
				if guard != nil && guard.ch != nil && len(guard.ch) > 200 {
					t.streamGuards.Delete(sid)
					log.Get().Error("stream guard deleted")
				}
			}(f.SessionID)
		}
		guard.ch <- f
	}
}
