package main

import (
	"flag"
	"log"
	"net"

	"github.com/blessedvictim/rongxy/proxy"
)

func init() {
	flag.Parse()
}

func main() {

	proxyServer := proxy.New(
		&net.TCPAddr{
			IP:   net.IPv4zero,
			Port: 1080,
		})
	log.Panic(proxyServer.ServeConnections())
}
