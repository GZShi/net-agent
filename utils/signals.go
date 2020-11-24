package utils

import (
	"os"
	"os/signal"
	"sync"
)

// WaitCtrlC 等待终端发出ctrl+c中断信号
func WaitCtrlC() {
	var wg sync.WaitGroup
	wg.Add(1)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		wg.Done()
	}()

	wg.Wait()
}
