package tunnel

import "io"

// NewWriteCloser 请避免直接使用_conn对象进行写入，会产生时序错乱问题
// 当需要调用原始conn连接写入数据时，需要创建临时的WriteCloser
// 创建时会请求对原始连接进行上锁
// 写入数据完毕后需要手动调用Close，释放锁
func (s *Server) NewWriteCloser() io.WriteCloser {
	s.writerLock.Lock()
	return &writeCloser{s}
}

//
// implement of io.WriteCloser
//

type writeCloser struct {
	server *Server
}

func (w *writeCloser) Write(buf []byte) (int, error) {
	return w.server._conn.Write(buf)
}

func (w *writeCloser) Close() error {
	w.server.writerLock.Unlock()
	return nil
}
