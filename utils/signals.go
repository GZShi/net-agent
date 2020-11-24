package utils

import (
	"os"
	"os/signal"
	"sync"
)

// WaitCtrlC 等待终端发出ctrl+c中断信号
func WaitCtrlC() {
	ch := make(chan os.Signal, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-ch
		signal.Reset(os.Interrupt)
		wg.Done()
	}()

	signal.Notify(ch, os.Interrupt)
	wg.Wait()
}
