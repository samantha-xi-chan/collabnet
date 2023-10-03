package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	err    error
	done   = make(chan struct{})
	sigint = make(chan os.Signal, 1)

	notify     = make(chan int, 1024)
	readChanEx = make(chan []byte, 1024)
	writeChan  = make(chan []byte, 1024)
)

func OnNewBizData(bytes []byte) {
	log.Println("OnNewBizData: ", string(bytes))
}

func SendBizData(bytes []byte) {
	writeChan <- bytes
}

// 状态: 连接异常、连接正常-业务空、连接正常-认证通过、连接正常-认证失败
func init() {

}

// 需要实现： 登进登出、心跳、容器饿死杀灭、指令执行（启动容器、停止容器、清除容器）
func main() {
	signal.Notify(sigint, os.Interrupt)

	go func() {
		for true {
			msg, ok := <-readChanEx
			if !ok {
				return
			}
			log.Println("string(msg): ", string(msg))
		}
	}()

	hostname, _ := os.Hostname()
	link.NewClientConnection(
		link.Config{
			Ver:      "v1.0",
			Auth:     config.AuthTokenForDev,
			HostName: hostname,
			HostAddr: fmt.Sprintf("%s%s", config.SCHEDULER_LISTEN_DOMAIN, config.SCHEDULER_LISTEN_PORT),
		},
		//notify,
		readChanEx,
		//writeChan,
	)

	log.Println("[main] waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
		//writeChan <- service.GetPackageBytes(time.Now().UnixMilli(), "v1.1", service.PACKAGE_TYPE_GOODBYE, nil)
		time.Sleep(time.Millisecond * 200)
		log.Println("Interrupted by user sleep")
	case <-done:
		log.Println(" chan done")
	}

}
