package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"encoding/json"
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

	notify      = make(chan int, 1024)
	readChanEx  = make(chan []byte, 1024)
	writeChanEx = make(chan []byte, 1024)
)

func OnNewBizData(bytes []byte) {
	log.Println("OnNewBizData: ", string(bytes))

	var body link.BizData
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		log.Println("[OnNewBizData] json.Unmarshal")
		return
	}

	idSched := body.Id

	log.Println("[OnNewBizData]  idSched = ", idSched)

	SendBizData(link.GetPackageBytes(
		time.Now().UnixMilli(),
		"1.0",
		link.PACKAGE_TYPE_BIZ,
		link.BizData{
			Id:   body.Id,
			Code: 0,
			Msg:  "STATUS_SCHED_CMD_ACKED",
		},
	))
	log.Println(" STATUS_SCHED_CMD_ACKED [OnNewBizData]  idSched = ", idSched)

	time.Sleep(time.Second * config.TEST_TIME_PREPARE)
	SendBizData(link.GetPackageBytes(
		time.Now().UnixMilli(),
		"1.0",
		link.PACKAGE_TYPE_BIZ,
		link.BizData{
			Id:   body.Id,
			Code: 0,
			Msg:  "STATUS_SCHED_PRE_ACKED",
		},
	))
	log.Println(" STATUS_SCHED_PRE_ACKED [OnNewBizData]  idSched = ", idSched)

	for i := 0; i < config.TEST_TIME_RUN/config.SCHED_HEARTBEAT_INTERVAL; i++ {
		time.Sleep(time.Second * time.Duration(body.HbInterval))
		SendBizData(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				Id:   body.Id,
				Code: 0,
				Msg:  "STATUS_SCHED_HEARTBEAT",
			},
		))
		log.Println(" HeartBeat [OnNewBizData]  idSched = ", idSched)
	}

	time.Sleep(time.Millisecond * 100)
	SendBizData(link.GetPackageBytes(
		time.Now().UnixMilli(),
		"1.0",
		link.PACKAGE_TYPE_BIZ,
		link.BizData{
			Id:   body.Id,
			Code: 0,
			Msg:  "STATUS_SCHED_END",
		},
	))
	log.Println(" Finished [OnNewBizData]  idSched = ", idSched)
}

func SendBizData(bytes []byte) {
	log.Println("SendBizData: ", string(bytes))
	writeChanEx <- bytes
}

// 状态: 连接异常、连接正常-业务空、连接正常-认证通过、连接正常-认证失败
func init() {

}

// 需要实现： 登进登出、心跳、容器饿死杀灭、指令执行（启动容器、停止容器、清除容器）
func main() {
	signal.Notify(sigint, os.Interrupt)

	go func() {
		for true {
			bytes, ok := <-readChanEx
			if !ok {
				return
			}

			log.Println("string(msg): ", string(bytes))

			go func() {
				OnNewBizData(bytes)
			}()
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
		writeChanEx,
	)

	log.Println("[main] waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
		//writeChan <- service.GetPackageBytes(time.Now().UnixMilli(), "v1.1", service.PACKAGE_TYPE_GOODBYE, nil)
		time.Sleep(time.Millisecond * 20)
	case <-done:
		log.Println(" chan done")
	}

}
