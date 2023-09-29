package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	Ping = "ping"
)

// 状态: 连接异常、连接正常-空、连接正常-认证通过、连接正常-认证失败、连接正常-被禁用、

func init() {

}

func main() {
	done := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	readBuff := make(chan []byte)
	writeBuff := make(chan []byte)
	service.NewConnection(fmt.Sprintf("%s%s",
		config.SCHEDULER_LISTEN_DOMAIN, config.SCHEDULER_LISTEN_PORT),
		Ping,
		readBuff,
		writeBuff,
	)

	go func() {
		for true {
			time.Sleep(time.Second)
			log.Println("<-readBuff: ", string(<-readBuff))
		}
	}()

	go func() {
		for true {
			time.Sleep(time.Second)
			str := fmt.Sprintln(time.Now().UnixMilli())
			writeBuff <- []byte(str)
		}
	}()

	log.Println("waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
	case <-done:
		log.Println(" chan done")
	}

}
