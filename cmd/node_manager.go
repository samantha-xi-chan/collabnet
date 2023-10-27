package main

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/package/message"
	"collab-net-v2/package/util/docker_vol"
	"collab-net-v2/package/util/util_minio"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/service_workflow"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

func HandlerDockerTask(task api.PluginTask) (willHandle bool) {
	go func() { // 模拟任务处理的情况
		log.Println("任务准备 ing", task.TaskId)
		time.Sleep(time.Second * 1)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK,
				SchedId: task.Id,
				Para01:  api.TASK_EVT_PREACK,
			},
		))

		// 普通的命令行执行
		/*
			cmd := exec.Command("bash", "-c", task.Cmd)

			// 设置命令的输出和错误输出
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// 执行命令
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Command execution failed: %v\n", err)
			}
		*/

		// 启动 docker 执行命令
		log.Println("task.Cmd: ", task.Cmd)

		var containerReq api.PostContainerReq
		err = json.Unmarshal([]byte(task.Cmd), &containerReq)
		if err != nil {
			fmt.Println("JSON deserialization error:", err)
			return
		}

		quit := make(chan bool)

		go func() { // 心跳
			ticker := time.NewTicker(time.Second * 10)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					log.Println("Heartbeat")

					SendBizData2Platform(link.GetPackageBytes(
						time.Now().UnixMilli(),
						"1.0",
						link.PACKAGE_TYPE_BIZ,
						link.BizData{
							TypeId:  link.BIZ_TYPE_NEWTASK,
							SchedId: task.Id,
							Para01:  api.TASK_EVT_HEARTBEAT,
						},
					))

				case <-quit:
					// 停止发送心跳信息
					return
				}
			}
		}()

		// 启动任务
		_, exitCode, e := service_workflow.PostContainerBlock(context.Background(), containerReq)
		if e != nil {
			exitCode = 101 // todo: 特殊错误 进入编码表
		}
		quit <- true

		log.Println("任务执行 ed", task.TaskId)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:   link.BIZ_TYPE_NEWTASK,
				SchedId:  task.Id,
				Para01:   api.TASK_EVT_END,
				Para0101: exitCode, // exitCode: 0表示成功
			},
		))
	}()

	return true
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
	} else if body.TypeId == link.BIZ_TYPE_NEW_DOCKER_TASK {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			"1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK, // todo: 变为 docker
				SchedId: schedId,
				TaskId:  taskId,
				Para01:  api.TASK_EVT_CMDACK,
			},
		))
		log.Println(" [OnNewBizDataFromPlatform] SendBizData2Platform STATUS_SCHED_CMD_ACKED schedId = ", schedId)

		newTask := api.PluginTask{
			Id:         schedId,
			TaskId:     taskId,
			Msg:        "test",
			Cmd:        body.Para11,
			Valid:      true,
			TimeoutPre: body.Para02,
			TimeoutRun: body.Para03,
		}
		HandlerDockerTask(newTask)
	} else if body.TypeId == link.BIZ_TYPE_STOP_DOCKER_TASK {
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
		// todo: 在此执行关闭容器操作
	} else {
		log.Println("WARNING: unknown cmd ...")
	}

}

func SendBizData2Platform(bytes []byte) {
	log.Println("SendBizData: ", string(bytes))
	writeChanEx <- bytes
}

func init() {
	firstParty := config.GetFirstParty()
	log.Println("firstParty: ", firstParty)
	if firstParty {
		v, _ := config.GetDependMsgRpc()
		log.Println("GetDependMsgRpc: ", v)
		succ := message.GetMsgCtl().Init(v)
		if !succ {
			log.Fatal("message.GetMsgCtl().Init(v) error, addr = ", v)
		}

		e := docker_vol.CreateVolumeFromFile(context.Background(), config_workflow.VOL_TOOL, config_workflow.SCRIPT_FILENAME, config_workflow.SCRIPT_CONTENT)
		if e != nil {
			log.Fatal("CreateVolumeFromFile e: ", e)
		}

		dsn, e := config.GetMinioDsn()
		if e != nil {
			log.Fatal("GetMinioDsn: ", dsn)
		}
		log.Println("minioDsn: ", dsn)
		parts := strings.SplitN(dsn, ":", 2)
		if len(parts) != 2 {
			log.Fatal("format error ")
		}
		username := parts[0]
		lastIndex := strings.LastIndex(parts[1], "@")
		if lastIndex == -1 {
			log.Fatal(" @ not found")
		}
		password := parts[1][:lastIndex]
		address := parts[1][lastIndex+1:]
		log.Printf("username: %s , password: %s, address: %s\n", username, password, address)

		e = util_minio.Init(context.Background(), address, username, password, config_workflow.BUCKET_NAME, false)
		if e != nil {
			log.Fatal("util_minio.Init: ", e)
		}
	}
}

var pluginChan = make(chan api.PluginTask)

func main() {
	log.Println("main() ")
	para01 := api.FALSE

	firstParty := config.GetFirstParty()
	log.Println("firstParty: ", firstParty)
	if firstParty {
		para01 = api.TRUE
	}

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

	instance := config.GetRunningInstance()
	link.NewClientConnection(
		link.Config{
			Ver:      "v1.0",
			Auth:     config_sched.AuthTokenForDev,
			HostName: instance,
			Para01:   para01,
			HostAddr: schedServer, //fmt.Sprintf("%s%s", schedServer, config_sched.SCHEDULER_LISTEN_PORT),
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

	return r.Run(config.PLUGIN_SERVICE_PORT) // todo: 改为 仅监听 local

	//rand.Seed(time.Now().UnixMilli())
	//addr := fmt.Sprintf(":%d", 8090+rand.Intn(100))
	//return r.Run(addr)
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
