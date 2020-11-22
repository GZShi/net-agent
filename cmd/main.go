package main

import "flag"

func main() {
	var mode string
	var addr string
	var socks5Addr string

	flag.StringVar(&mode, "m", "c", "mode select")
	flag.StringVar(&addr, "addr", "127.0.0.1:20035", "listen or connect addr")
	flag.StringVar(&socks5Addr, "socks5", "127.0.0.1:20034", "work port")
	flag.Parse()

	if mode == "c" {
		runClient(socks5Addr, addr)
	} else {
		runServer(addr)
	}
}
