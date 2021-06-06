package tunnel

import (
	"sync"
	"time"
)

// 用于统计stream状态的结构体
type StreamState struct {
	Closed            bool
	TimeCreated       time.Time
	TimeClosed        time.Time
	TimeFirstByte     time.Time
	firstByteOnce     sync.Once
	TimeLastReadByte  time.Time
	TimeLastWriteByte time.Time

	ReadBytesLen              int64
	WriteBytesLen             int64
	RecentSecondReadBytesLen  int64
	RecentSecondWriteBytesLen int64
	secticker                 *time.Ticker // init at LogCreated
	sectickerStop             chan int     // init at LogCreated

	SID            uint32
	CallerVhost    string
	CallerSID      uint32
	ServiceName    string
	ServiceInfo    string
	FirstWrite256B []byte
	FirstRead256B  []byte
}

func (ss *StreamState) LogCreated() {
	ss.TimeCreated = time.Now()
	ss.secticker = time.NewTicker(time.Second)
	ss.sectickerStop = make(chan int)
	go func() {
		defer func() {
			ss.secticker.Stop()
			close(ss.sectickerStop)
			ss.sectickerStop = nil
		}()
		lastRL := ss.ReadBytesLen
		lastWL := ss.WriteBytesLen
		for {
			select {
			case <-ss.secticker.C:
				ss.RecentSecondReadBytesLen = ss.ReadBytesLen - lastRL
				lastRL = ss.ReadBytesLen
				ss.RecentSecondWriteBytesLen = ss.WriteBytesLen - lastWL
				lastWL = ss.WriteBytesLen
			case <-ss.sectickerStop:
				return
			}
		}
	}()
}
func (ss *StreamState) LogClosed() {
	ss.Closed = true
	ss.TimeClosed = time.Now()
	if ss.sectickerStop != nil {
		ss.sectickerStop <- 0
	}
}
func (ss *StreamState) LogFirstByte() {
	ss.firstByteOnce.Do(func() {
		if ss.TimeFirstByte.Unix() < 0 {
			ss.TimeFirstByte = time.Now()
		}
	})
}
func (ss *StreamState) AddReadLen(size int) {
	ss.ReadBytesLen += int64(size)
	if size > 0 {
		ss.TimeLastReadByte = time.Now()
		ss.LogFirstByte()
	}
}
func (ss *StreamState) AddWriteLen(size int) {
	ss.WriteBytesLen += int64(size)
	if size > 0 {
		ss.TimeLastWriteByte = time.Now()
		ss.LogFirstByte()
	}
}
func (ss *StreamState) SetFirstWriteBytes(buf []byte) {
	if len(ss.FirstWrite256B) <= 0 && buf != nil && len(buf) > 0 {
		if len(buf) > 256 {
			buf = buf[:256]
		}
		ss.FirstWrite256B = append(ss.FirstWrite256B, buf...)
	}
}
func (ss *StreamState) SetFirstReadBytes(buf []byte) {
	if len(ss.FirstRead256B) <= 0 && buf != nil && len(buf) > 0 {
		if len(buf) > 256 {
			buf = buf[:256]
		}
		ss.FirstRead256B = append(ss.FirstWrite256B, buf...)
	}
}
