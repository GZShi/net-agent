package tunnel

import (
	"errors"
	"io"
)

// NewWriteCloser 请避免直接使用_conn对象进行写入，会产生时序错乱问题
// 当需要调用原始conn连接写入数据时，需要创建临时的WriteCloser
// 创建时会请求对原始连接进行上锁
// 写入数据完毕后需要手动调用Close，释放锁
func (t *tunnel) NewWriteCloser() io.WriteCloser {
	t.writerLock.Lock()
	return &writeCloser{t}
}

//
// implement of io.WriteCloser
//

type writeCloser struct {
	t *tunnel
}

func (w *writeCloser) Write(buf []byte) (int, error) {
	if w.t == nil {
		return 0, errors.New("writeCloser closed")
	}
	return w.t._conn.Write(buf)
}

func (w *writeCloser) Close() error {
	if w.t == nil {
		return errors.New("writeCloser closed")
	}
	w.t.writerLock.Unlock()
	w.t = nil
	return nil
}
