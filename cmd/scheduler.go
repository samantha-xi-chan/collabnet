package main

import (
	"collab-net-v2/link"
	"log"
	"time"
)

func init() {
	log.Println("init: ")
}

func main() {
	go link.NewServer()

	for true {
		time.Sleep(time.Second * 10)
		link.SendDataToEndpoint(
			"M1",
			link.GetPackageBytes(time.Now().UnixMilli(),
				"v1.0",
				link.PACKAGE_TYPE_BIZ,
				link.BizData{
					Code: 0,
					Msg:  "start a docker",
				}))
	}

	log.Println("waiting in select")
	select {}
}
