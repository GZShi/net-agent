package main

import (
	"io"
	"net"
	"sync"

	"github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/tunnel"
)

func runServer(addr string) error {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	log.Get().Info("listen on ", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go serve(conn)
	}
}

func serve(conn net.Conn) {
	if conn == nil {
		return
	}
	log.Get().Info("a tunnel created")
	t := tunnel.New(conn)

	t.Listen("dial", func(ctx tunnel.Context) {
		var req dialReqeust
		var resp dialResponse
		err := ctx.GetJSON(&req)
		if err != nil {
			ctx.Error(err)
			return
		}

		// direct dial
		log.Get().Info("try to dial direct")
		conn, err := net.Dial(req.Network, req.Address)
		if err != nil {
			ctx.Error(err)
			return
		}

		// create and bind stream
		stream, sid := t.NewStream()
		resp.SessionID = sid
		stream.Bind(req.SessionID)

		go link(stream, conn)

		log.Get().Info("dial sucess")

		ctx.JSON(&resp)
	})

	t.Run()
	log.Get().Info("a tunnel closed")
}

func link(a io.ReadWriteCloser, b io.ReadWriteCloser) (a2bN, b2aN int64, err error) {
	var wg sync.WaitGroup
	var once sync.Once

	clean := func(copyErr error) {
		if copyErr != nil {
			once.Do(func() {
				err = copyErr
			})
		}
		a.Close()
		b.Close()
		wg.Done()
	}

	wg.Add(1)
	go func() {
		var cpErr error
		b2aN, cpErr = io.Copy(a, b)
		clean(cpErr)
	}()

	wg.Add(1)
	go func() {
		var cpErr error
		a2bN, err = io.Copy(b, a)
		clean(cpErr)
	}()

	wg.Wait()
	return a2bN, b2aN, err
}
