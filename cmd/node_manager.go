package main

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/pkg/external/message"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/util/docker_container"
	"collab-net-v2/util/docker_vol"
	"collab-net-v2/util/filems"
	"collab-net-v2/util/stl"
	"collab-net-v2/util/util_minio"
	"collab-net-v2/util/util_net"
	"collab-net-v2/util/util_os"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/service_workflow"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	mapContainerHeartbeat = stl.SafeMap{Data: make(map[string]int64)}
	//mapProcessHeartbeat   = stl.SafeMap{Data: make(map[string]int64)}
)

const (
	HeartbeatIntervalSecond    = 10
	HeartbeatIntervalSecondMul = 6
)

func OnUpdateFromPlugin(id string, status int, para01 int) {
	log.Printf("[OnUpdateFromPlugin] id=%s, status=%d, para01=%d \n", id, status, para01)

	if status == api.TASK_EVT_START {
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    id,
				Para01:     api.TASK_EVT_PREACK,
			},
		))

		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    id,
				Para01:     api.TASK_EVT_HEARTBEAT,
			},
		))
	} else if status == api.TASK_EVT_END {
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    id,
				Para01:     api.TASK_EVT_END,
				Para0101:   para01,
			},
		))
	} else if status == api.TASK_EVT_HEARTBEAT {
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    id,
				Para01:     api.TASK_EVT_HEARTBEAT,
			},
		))
	} else {
		log.Printf("WARNING: unknown status = %d \n", status)
	}
}

func HandlerDockerTask(task api.PluginTask) (willHandle bool) {
	go func() {
		log.Println("task prepare ing", task.TaskId)
		time.Sleep(time.Second * 1)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_DOCKER,
				SchedId:    task.Id,
				Para01:     api.TASK_EVT_PREACK,
			},
		))

		log.Println("task.Cmd: ", task.Cmd)

		var containerReq api.PostContainerReq
		err = json.Unmarshal([]byte(task.Cmd), &containerReq)
		if err != nil {
			fmt.Println("JSON deserialization error:", err)
			return
		}

		schedId := task.Id

		quit := make(chan bool)

		go func() {
			ticker := time.NewTicker(time.Second * HeartbeatIntervalSecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					log.Println("HandlerDockerTask Heartbeat schedId = ", schedId)

					SendBizData2Platform(link.GetPackageBytes(
						time.Now().UnixMilli(),
						config.VerSched,
						link.PACKAGE_TYPE_BIZ,
						link.PlatformBiiData{
							ActionType: link.ACTION_TYPE_NEWTASK,
							TaskType:   link.TASK_TYPE_DOCKER,
							SchedId:    task.Id,
							Para01:     api.TASK_EVT_HEARTBEAT,
						},
					))

				case <-quit:
					return
				}
			}
		}()

		exitCode := 0
		containerId, e := service_workflow.CreateContainerWrapper(context.Background(), containerReq, schedId)
		if e != nil {
			SendBizData2Platform(link.GetPackageBytes(
				time.Now().UnixMilli(),
				config.VerSched,
				link.PACKAGE_TYPE_BIZ,
				link.PlatformBiiData{
					ActionType: link.ACTION_TYPE_NEWTASK,
					TaskType:   link.TASK_TYPE_DOCKER,
					SchedId:    schedId,
					Para01:     api.TASK_EVT_END,
					Para0101:   api.ERR_CREAT_CONTAINER,
					Para0102:   e.Error(),
				},
			))

			return
		}

		mapContainerHeartbeat.Set(schedId, time.Now().UnixMilli())

		log.Println("task running : container created， containerId = ", containerId)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_DOCKER,
				SchedId:    task.Id,
				Para01:     api.TASK_EVT_REPORT,
				Para0102:   containerId,
			},
		))

		exitCode, e = service_workflow.StartContainerAndWait(context.Background(), containerId, containerReq, schedId)
		errString := ""
		quit <- true
		if e != nil {
			errString = e.Error()
		}

		log.Println("task run  ed", task.TaskId)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_DOCKER,
				SchedId:    task.Id,
				Para01:     api.TASK_EVT_END,
				Para0101:   exitCode, // exitCode: 0表示成功
				Para0102:   errString,
			},
		))
	}()

	return true
}

func OnNewBizDataFromPlatform(bytes []byte) {
	onNewBizData := string(bytes)
	log.Println("OnNewBizData: ", onNewBizData)

	var body link.PlatformBiiData
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		log.Println("[OnNewBizData] json.Unmarshal")
		return
	}

	if body.ActionType == link.ACTION_TYPE_NEWTASK && body.TaskType == link.TASK_TYPE_RAW {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    schedId,
				TaskId:     taskId,
				Para01:     api.TASK_EVT_CMDACK,
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
	} else if body.ActionType == link.ACTION_TYPE_STOPTASK && body.TaskType == link.TASK_TYPE_RAW {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_STOPTASK,
				TaskType:   link.TASK_TYPE_RAW,
				SchedId:    schedId,
				TaskId:     taskId,
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
	} else if body.ActionType == link.ACTION_TYPE_NEWTASK && body.TaskType == link.TASK_TYPE_DOCKER {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_NEWTASK,
				TaskType:   link.TASK_TYPE_DOCKER,
				SchedId:    schedId,
				TaskId:     taskId,
				Para01:     api.TASK_EVT_CMDACK,
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
	} else if body.ActionType == link.ACTION_TYPE_STOPTASK && body.TaskType == link.TASK_TYPE_DOCKER {
		schedId := body.SchedId
		taskId := body.TaskId
		log.Println("[OnNewBizData]  schedId = ", schedId)

		time.Sleep(time.Millisecond * 200)
		SendBizData2Platform(link.GetPackageBytes(
			time.Now().UnixMilli(),
			config.VerSched,
			link.PACKAGE_TYPE_BIZ,
			link.PlatformBiiData{
				ActionType: link.ACTION_TYPE_STOPTASK,
				TaskType:   link.TASK_TYPE_DOCKER,
				SchedId:    schedId,
				TaskId:     taskId,
			},
		))
		log.Println(" [OnNewBizDataFromPlatform] SendBizData2Platform STATUS_SCHED_CMD_ACKED schedId = ", schedId)

		docker_container.StopContainerByName(schedId)
	} else if body.ActionType == link.ACTION_TYPE_STATUS_TASK && body.TaskType == link.TASK_TYPE_DOCKER {
		schedId := body.SchedId
		mapContainerHeartbeat.Set(schedId, time.Now().UnixMilli())

		log.Println("mapContainerHeartbeat update : schedId = ", schedId)
	} else if body.ActionType == link.ACTION_TYPE_STATUS_TASK && body.TaskType == link.TASK_TYPE_RAW {
		schedId := body.SchedId
		//mapProcessHeartbeat.Set(schedId, time.Now().UnixMilli())
		log.Println("mapProcessHeartbeat update : schedId = ", schedId)
	} else {
		log.Println("WARNING: unknown cmd, ", body.ActionType, " ", body.TaskType)
	}

}

func SendBizData2Platform(bytes []byte) {
	log.Println("SendBizData: ", string(bytes))
	writeChanEx <- bytes
}

func checkS3fsCmd() {
	cmd := exec.Command("s3fs", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("s3fs command is not available or encountered an error:", err)
		os.Exit(1)
	}

	fmt.Println("s3fs command is available.")
	fmt.Println("s3fs version information:")
	fmt.Println(string(output))
}

func mountFS(ak string, sk string, ipAndPort string) (e error) {
	content := fmt.Sprintf("%s:%s\n", ak, sk) // "admin:password\n"
	err := ioutil.WriteFile("/etc/passwd-s3fs", []byte(content), 0600)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		os.Exit(1)
	}

	fmt.Println("File /etc/passwd-s3fs written successfully.")

	err = os.Chmod("/etc/passwd-s3fs", 0600)
	if err != nil {
		fmt.Println("Error changing file permissions:", err)
		os.Exit(1)
	}

	fmt.Println("File permissions changed successfully.")

	urlString := fmt.Sprintf("url=http://%s", ipAndPort)
	cmd := exec.Command("s3fs",
		"-o", "passwd_file=/etc/passwd-s3fs",
		"-o", urlString,
		"-o", "use_path_request_style",
		"-o", "nonempty",
		"-o", "kernel_cache",
		"-o", "max_background=1000",
		"-o", "max_stat_cache_size=100000",
		"-o", "multipart_size=64",
		"-o", "parallel_count=30",
		"-o", "multireq_max=30",
		"-o", "dbglevel=warn",
		config_workflow.MINIO_BUCKET_NAME_WF,
		config_workflow.DockerGroupPref)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running s3fs command:", err)
		os.Exit(1)
	}

	fmt.Println("s3fs command output:", string(output))

	// echo "admin:password" > /etc/passwd-s3fs
	// chmod 600 /etc/passwd-s3fs
	// s3fs -o passwd_file=/etc/passwd-s3fs -o url=http://192.168.31.45:32000 -o use_path_request_style -o nonempty workflowshare /mnt/sss

	return nil
}

func init() {
	checkS3fsCmd()

	if !util_os.IsCurrentUserRoot() {
		log.Fatal("currentRunning User not Root")
	}

	maxOpenFiles, err := util_os.GetMaxOpenFiles()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	log.Println("maxOpenFiles: ", maxOpenFiles)

	config.Init()

	firstParty := config.GetFirstParty()
	log.Println("firstParty: ", firstParty)
	if firstParty {
		log.Printf("Running In Mode FirstParty Node !!!!!! ") // Info level log
	} else {
		log.Printf("Running In Mode UserProvided Node !!!!!! ") // Info level log
	}

	if firstParty {
		v := config.GetDependMsgRpc()
		log.Println("GetDependMsgRpc: ", v)

		for true {
			bAllOK, _ := util_net.CheckTcpService(
				[]string{
					v,
				},
			)

			log.Println("CheckTcpService bAllOK: ", bAllOK)
			if bAllOK {
				break
			}

			time.Sleep(time.Second)
		}

		succ := message.GetMsgCtl().Init(v)
		if !succ {
			log.Fatal("message.GetMsgCtl().Init(v) error, addr = ", v)
		}

		log.Println("message.GetMsgCtl().Init end ")

		e := docker_vol.CreateVolumeFromFile(context.Background(), config_workflow.VOL_TOOL, config_workflow.SCRIPT_FILENAME, config_workflow.SCRIPT_CONTENT)
		if e != nil {
			log.Fatal("CreateVolumeFromFile e: ", e)
		}

		log.Println("CreateVolumeFromFile end ")

		// todo: 抽象为函数
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

		// check
		for true {
			mountFS(username, password, address)

			if filems.CheckFileReady(config_workflow.DefaultServerSignPath) == nil {
				break
			}

			time.Sleep(time.Second)
		}

		for true {
			bAllOK, _ := util_net.CheckTcpService(
				[]string{
					address,
				},
			)

			log.Println("CheckTcpService bAllOK: ", bAllOK)
			if bAllOK {
				break
			}

			time.Sleep(time.Second)
		}

		e = util_minio.InitDistFileMs(context.Background(), address, username, password, config_workflow.MINIO_BUCKET_NAME_INTERTASK, false)
		if e != nil {
			log.Fatal("util_minio.InitDistFileMs: ", e)
		}
	}
}

var pluginChan = make(chan api.PluginTask)

func main() {
	log.Println("main() ")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	fmt.Printf("GitCommit: %s\n", GitCommit)

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
			Ver:      config.VerLink,
			Auth:     config_sched.AuthTokenForDev,
			HostName: instance,
			Para01:   para01,
			HostAddr: schedServer, //fmt.Sprintf("%s%s", schedServer, config_sched.SCHEDULER_LISTEN_PORT),
		},
		//notify,
		readChanEx,
		writeChanEx,
	)

	go func() {
		for true {

			iteratedMap := mapContainerHeartbeat.Iterate()
			for key, tick := range iteratedMap {
				if time.Now().UnixMilli()-tick > HeartbeatIntervalSecond*HeartbeatIntervalSecondMul*1000 {
					log.Printf("WARNING: mapContainerHeartbeat Key: %s, Value: %d time.Now().UnixMilli()-tick >  const \n", key, tick)
					// kill it and remove record
					docker_container.StopContainerByName(key)
					//delete(mapContainerHeartbeat, key)
					mapContainerHeartbeat.Delete(key)
				}
				time.Sleep(time.Millisecond * 10)
			}

			//for key, tick := range mapProcessHeartbeat {
			//	if time.Now().UnixMilli()-tick > HeartbeatIntervalSecond*HeartbeatIntervalSecondMul*1000 {
			//		log.Printf("WARNING: mapProcessHeartbeat Key: %s, Value: %d time.Now().UnixMilli()-tick >  const \n", key, tick)
			//		// kill it and remove record
			//		// todo: stop process
			//		//delete(mapContainerHeartbeat, key)
			//		//mapProcessHeartbeat.Delete(key)
			//	}
			//	time.Sleep(time.Millisecond * 10)
			//}

			time.Sleep(time.Second * 20)
		}
	}()

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
