package main

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/control_link"
	"collab-net-v2/pkg/external/message"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/control_task"
	"collab-net-v2/task/service_task"
	"collab-net-v2/util/util_minio"
	"collab-net-v2/util/util_net"
	"collab-net-v2/workflow"
	"collab-net-v2/workflow/config_workflow"
	"context"
	"fmt"
	"log"
	"time"
)

func OnTaskChange(idTask string, evt int, x []byte) (e error) {
	log.Println("USER [OnTaskChange] idTask = ", idTask, ", evt = ", evt)

	return nil
}

func init() {
	log.Println("main [init] start ")
	defer func() {
		log.Println("main [init] end ")
	}()

	for true {
		bAllOK, _ := util_net.CheckTcpService(
			[]string{
				"go-message-waiter:10051", // todo: coding style
				"go-message-notify:9102",
				"mysql-collabnet:3306",
				config_workflow.MINIO_API_URL,
				"redis-service:6379",
				"rmq-cluster:5672",
			},
		)

		log.Println("CheckTcpService bAllOK: ", bAllOK)
		if bAllOK {
			break
		}

		time.Sleep(time.Second)
	}

	config.Init()

	// init storage
	util_minio.CreateBucketIfNotExist(context.Background(), config_workflow.MINIO_API_URL, config_workflow.MINIO_AK, config_workflow.MINIO_SK, config_workflow.MINIO_BUCKET_NAME_WF, config_workflow.SSSDefaultObjectName)
	e := util_minio.InitDistFileMs(context.Background(), config_workflow.MINIO_API_URL, config_workflow.MINIO_AK, config_workflow.MINIO_SK, config_workflow.MINIO_BUCKET_NAME_INTERTASK, false)
	if e != nil {
		log.Fatal("util_minio.InitDistFileMs: ", e)
	}
}

func main() {
	ctx := context.Background()
	//var logger *logrus.Logger
	log.Println("main [init] : ")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	fmt.Printf("GitCommit: %s\n", GitCommit)

	instance := config.GetRunningInstance()
	log.Println("GetRunningInstance: ", instance)
	//logServer := config.GetLogServer()
	//
	//logger = logrus.New()
	//logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	//hook, err := logrustash.NewHook("tcp", logServer, instance)
	//if err != nil {
	//	log.Fatal("logrustash.NewHook: ", err)
	//}
	//logger.Hooks.Add(hook)

	//log := logger.WithFields(logrus.Fields{
	//	"method": "main",
	//})

	v := config.GetDependMsgRpc()
	if v == "" {
		log.Fatal("config.GetDependMsgRpc()  v=empty ")
	}
	log.Println("GetDependMsgRpc: ", v)
	ok := message.GetMsgCtl().Init(v)
	if !ok {
		log.Fatal("message.GetMsgCtl().Init(v) error, addr = ", v)
	}
	message.GetMsgCtl().UpdateTaskWrapper("init", api.TASK_STATUS_RUNNING, "nil")

	service_task.Init()
	log.Println("[main] SetTaskCallback")
	service_task.SetTaskCallback(OnTaskChange)

	time.Sleep(time.Millisecond * 100)
	go func() {
		e := control_link.InitGinService(ctx, config_link.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_link.InitGinService e: ", e)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	service_sched.Init()

	time.Sleep(time.Millisecond * 100)
	go func() {
		e := control_task.InitGinService(context.Background(), config_task.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_task.InitGinService4 e: ", e)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	go func() {
		//control_workflow.StartHttpServer(":30001")
		workflow.StartService()
	}()

	log.Println("[main] waiting select{}")
	select {}
}
