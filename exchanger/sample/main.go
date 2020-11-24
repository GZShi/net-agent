package main

import (
	"flag"
	"sync"
)

func main() {
	var mode string
	var addr string
	var password string

	flag.StringVar(&mode, "m", "c", "run mode. c=client, s=server")
	flag.StringVar(&addr, "a", "127.0.0.1:2036", "address for listen or connect")
	flag.StringVar(&password, "p", "default-pAs5w0rd", "set password")
	flag.Parse()

	switch mode {
	case "s":
		listenAndServe(addr, password)
	case "c":
		connectAsClient(addr, password)
	case "a":
		connectAsAgent(addr, password)
	case "ac", "ca":
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

		}()
		wg.Wait()
	}
}

func connectAsClientAgent(addr, password string) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		connectAsClient(addr, password)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		connectAsAgent(addr, password)
	}()

	wg.Wait()
}
