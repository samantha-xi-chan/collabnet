package main

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/sched/config_sched"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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

func OnUpdateFromPlugin(id string, status int, para01 int) {

	//SendBizData(  )
}

func OnNewBizData(bytes []byte) {
	onNewBizData := string(bytes)
	log.Println("OnNewBizData: ", onNewBizData)

	var body link.BizData
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		log.Println("[OnNewBizData] json.Unmarshal")
		return
	}

	idTask := body.Id

	log.Println("[OnNewBizData]  idTask = ", idTask)

	time.Sleep(time.Millisecond * 200)
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
	log.Println(" STATUS_SCHED_CMD_ACKED [OnNewBizData]  idTask = ", idTask)

	time.Sleep(time.Second * config_sched.TEST_TIME_PREPARE)

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
	log.Println(" STATUS_SCHED_PRE_ACKED [OnNewBizData]  idTask = ", idTask)

	for i := 0; i < config_sched.TEST_TIME_RUN/config_sched.SCHED_HEARTBEAT_INTERVAL; i++ {
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
		log.Println(" HeartBeat [OnNewBizData]  idTask = ", idTask)
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
	log.Println(" Finished [OnNewBizData]  idTask = ", idTask)
}

func SendBizData(bytes []byte) {
	log.Println("SendBizData: ", string(bytes))
	writeChanEx <- bytes
}

func init() {

}

var pluginChan = make(chan string)

func main() {
	go func() {
		e := StartPluginService()
		if e != nil {
			log.Fatal("StartPluginService: e=", e)
		}
	}()
	log.Println("end StartPluginService()")

	signal.Notify(sigint, os.Interrupt)

	go func() {
		for true {
			bytes, ok := <-readChanEx
			if !ok {
				return
			}

			readChanExStr := string(bytes)
			log.Println("readChanExStr: ", readChanExStr)

			pluginChan <- readChanExStr

			go func() {
				OnNewBizData(bytes)
			}()
		}
	}()

	hostName, _ := os.Hostname()
	link.NewClientConnection(
		link.Config{
			Ver:      "v1.0",
			Auth:     config_sched.AuthTokenForDev,
			HostName: hostName,
			HostAddr: fmt.Sprintf("%s%s", config_sched.SCHEDULER_LISTEN_DOMAIN, config_sched.SCHEDULER_LISTEN_PORT),
		},
		//notify,
		readChanEx,
		writeChanEx,
	)

	log.Println("[main] waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
		//writeChan <- service_sched.GetPackageBytes(time.Now().UnixMilli(), "v1.1", service_sched.PACKAGE_TYPE_GOODBYE, nil)
		time.Sleep(time.Millisecond * 20)
	case <-done:
		log.Println(" chan done")
	}

}

func StartPluginService() (ee error) {
	r := gin.Default()
	r.GET(config.PLUGIN_SERVICE_ROUTER, getTaskCmd)
	r.POST(config.PLUGIN_SERVICE_ROUTER_ID, postTaskStatus)
	return r.Run(config.PLUGIN_SERVICE_PORT)
}

func getTaskCmd(c *gin.Context) {
	select {
	case msg := <-pluginChan:
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: 0,
			Msg:  msg,
			//Data: []byte(msg),
		})
	case <-time.After(30 * time.Second):
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: 9999,
			Msg:  "no data",
		})
	}
}

func postTaskStatus(c *gin.Context) { // 任务状态变更
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "postTaskStatus ok",
	})
}

//func longPollHandler(w http.ResponseWriter, r *http.Request) {
//	// Set response headers to allow cross-origin requests
//	w.Header().Set("Access-Control-Allow-Origin", "*")
//	w.Header().Set("Content-Type", "text/plain")
//
//	select {
//	case msg := <-pluginChan:
//		// When data is available, send it as a response
//		w.Write([]byte(msg))
//	case <-time.After(30 * time.Second):
//		// After a timeout, respond with a message indicating no new data
//		w.Write([]byte("No new data available."))
//	}
//}
