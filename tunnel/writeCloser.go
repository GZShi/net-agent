package tunnel

import (
	"bytes"
	"errors"
	"io"
)

// NewWriteCloser 请避免直接使用_conn对象进行写入，会产生时序错乱问题
// * 当需要调用原始conn连接写入数据时，需要创建临时的WriteCloser
// * 创建时会请求对原始连接进行上锁
// * 写入数据完毕后需要手动调用Close，释放锁
func (t *tunnel) NewWriteCloser() io.WriteCloser {
	return &writeCloser{t, bytes.NewBuffer(nil)}
}

//
// implement of io.WriteCloser
//

type writeCloser struct {
	t  *tunnel
	bf *bytes.Buffer
}

func (w *writeCloser) Write(b []byte) (int, error) {
	return w.bf.Write(b)
}

func (w *writeCloser) Close() error {
	if w.t == nil {
		return errors.New("writeCloser closed")
	}
	w.t.writerLock.Lock()
	w.bf.WriteTo(w.t._conn)
	w.t.writerLock.Unlock()

	w.t = nil
	w.bf = nil
	return nil
}
