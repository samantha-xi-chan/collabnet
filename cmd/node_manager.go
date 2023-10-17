package main

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/sched/config_sched"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	log.Printf("[OnUpdateFromPlugin] id=%s, status=%d, para01=%d \n", id, status, para01)

	if status == api.TASK_EVT_START {
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK,
				SchedId: id,
				Para01:  api.TASK_EVT_PREACK,
			},
		))

		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK,
				SchedId: id,
				Para01:  api.TASK_EVT_HEARTBEAT,
			},
		))
	} else if status == api.TASK_EVT_END {
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:   link.BIZ_TYPE_NEWTASK,
				SchedId:  id,
				Para01:   api.TASK_EVT_END,
				Para0101: para01,
			},
		))
	}
}

func OnNewBizDataFromPlatform(bytes []byte) {
	onNewBizData := string(bytes)
	log.Println("OnNewBizData: ", onNewBizData)

	var body link.BizData
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		log.Println("[OnNewBizData] json.Unmarshal")
		return
	}

	if body.TypeId == link.BIZ_TYPE_NEWTASK {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK,
				SchedId: schedId,
				TaskId:  taskId,
				Para01:  api.TASK_EVT_CMDACK,
			},
		))
		log.Println(" [OnNewBizDataFromPlatform] SendBizData2Platform STATUS_SCHED_CMD_ACKED schedId = ", schedId)

		// 将任务的结构体转换进入 chan
		newTask := api.PluginTask{
			Id:         schedId,
			TaskId:     taskId,
			Msg:        "test",
			Cmd:        body.Para11,
			Valid:      true,
			TimeoutPre: body.Para02,
			TimeoutRun: body.Para03,
		}
		pluginChan <- newTask
	} else if body.TypeId == link.BIZ_TYPE_STOPTASK {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_STOPTASK,
				SchedId: schedId,
				TaskId:  taskId,
			},
		))
		log.Println(" [OnNewBizDataFromPlatform] SendBizData2Platform STATUS_SCHED_CMD_ACKED schedId = ", schedId)

		// 将任务的结构体转换进入 chan
		stopTask := api.PluginTask{
			Id:     schedId,
			TaskId: taskId,
			Msg:    "stopTtask",
			Valid:  false,
		}
		pluginChan <- stopTask
	}
}

func SendBizData2Platform(bytes []byte) {
	log.Println("SendBizData: ", string(bytes))
	writeChanEx <- bytes
}

func init() {

}

var pluginChan = make(chan api.PluginTask)

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

			//pluginChan <- readChanExStr

			go func() {
				OnNewBizDataFromPlatform(bytes)
			}()
		}
	}()

	schedServer, e := config.GetBizSchedServer()
	if e != nil {
		log.Fatal("config.GetBizSchedServer: ", e)
	}
	log.Println("schedServer: ", schedServer)

	hostName, _ := os.Hostname()
	link.NewClientConnection(
		link.Config{
			Ver:      "v1.0",
			Auth:     config_sched.AuthTokenForDev,
			HostName: hostName,
			HostAddr: fmt.Sprintf("%s%s", schedServer, config_sched.SCHEDULER_LISTEN_PORT),
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
	r.GET(config.PLUGIN_SERVICE_ROUTER, getPluginTaskCmd)
	r.POST(config.PLUGIN_SERVICE_ROUTER_ID, postPluginTaskStatus)
	return r.Run(config.PLUGIN_SERVICE_PORT)
}

func getPluginTaskCmd(c *gin.Context) {
	select {
	case pluginTaskInfo := <-pluginChan:
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: 0,
			Msg:  "",
			Data: pluginTaskInfo,
		})
	case <-time.After(30 * time.Second):
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: 9999,
			Msg:  "no data",
		})
	}
}

func postPluginTaskStatus(c *gin.Context) { // 任务状态变更
	id := c.Param("id")
	log.Println("postPluginTaskStatus id:  ", id)

	var dto api.PostPluginTaskStatusReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT: " + err.Error(),
		})
		return
	}

	log.Println("postPluginTaskStatus dto:  ", dto)

	OnUpdateFromPlugin(id, dto.Status, dto.Para01)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "postTaskStatus ok",
	})
}
