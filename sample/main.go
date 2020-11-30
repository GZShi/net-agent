package main

import (
	"flag"
	"sync"
)

func main() {
	var mode string
	var addr string
	var socks5Addr string
	var password string

	flag.StringVar(&mode, "m", "c", "run mode. c=client, s=server")
	flag.StringVar(&addr, "a", "127.0.0.1:2036", "address for listen or connect")
	flag.StringVar(&socks5Addr, "socks5", "127.0.0.1:2037", "address for socks5 server")
	flag.StringVar(&password, "p", "default-pAs5w0rd", "set password")
	flag.Parse()

	switch mode {
	case "s":
		listenAndServe(addr, password)
	case "c":
		connectAsClient(addr, socks5Addr, password)
	case "a":
		connectAsAgent(addr, password)
	case "ac", "ca":
		connectAsClientAgent(addr, socks5Addr, password)
	}
}

func connectAsClientAgent(addr, socks5Addr, password string) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		connectAsClient(addr, socks5Addr, password)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		connectAsAgent(addr, password)
	}()

	wg.Wait()
}
