package tunnel

import (
	"fmt"
	"strings"

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
		log.Get().WithField("sessionID", f.SessionID).Warn("response guard not found")
		return
	}
	guard := val.(chan *Frame)
	guard <- f
}

// onSteramData 基础事件
// 接收到对端的一个数据传输包
// 数据传输包专门用于进行原始二进制通信
func (t *tunnel) onStreamData(f *Frame) {
	val, loaded := t.streamGuards.Load(f.SessionID)
	if !loaded {
		if f.Data == nil {
			// f.Data为nil，代表收到对端的EOF信号，此时如果找不到guard，忽略丢弃即可
			return
		}

		// 此时存在丢弃数据的风险
		datalen := 0
		if f.Data != nil {
			datalen = len(f.Data)
		}
		log.Get().Warn("stream guard not found: id=", f.SessionID, " datalen=", datalen)
		return
	}

	if val == nil {
		panic("stream is nil")
	}

	stream := val.(*streamRWC)
	stream.Cache(f)
}
