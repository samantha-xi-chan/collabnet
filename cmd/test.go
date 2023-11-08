package main

import (
	"collab-net-v2/util/util_net"
	"log"
)

func main() {
	b, _ := util_net.CheckTcpService(
		[]string{
			"192.168.36.106:80",
			"192.168.36.106:443",
			//"192.168.36.106:4443",
		},
	)

	log.Println("b: ", b)
}
