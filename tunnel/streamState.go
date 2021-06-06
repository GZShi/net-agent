package tunnel

import "time"

// 用于统计stream状态的结构体
type StreamState struct {
	Closed            bool
	TimeCreated       time.Time
	TimeClosed        time.Time
	TimeFirstByte     time.Time
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
}

func (ss *StreamState) LogCreated() {
	ss.TimeClosed = time.Now()
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
	if ss.TimeFirstByte.Unix() < 0 {
		ss.TimeFirstByte = time.Now()
	}
}
func (ss *StreamState) AddReadLen(size int) {
	ss.ReadBytesLen += int64(size)
	if size > 0 {
		ss.TimeLastReadByte = time.Now()
	}
}
func (ss *StreamState) AddWriteLen(size int) {
	ss.WriteBytesLen += int64(size)
	if size > 0 {
		ss.TimeLastWriteByte = time.Now()
	}
}
func (ss *StreamState) SetFirstWriteBytes(buf []byte) {}
