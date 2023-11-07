package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/control_link"
	"collab-net-v2/package/message"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/control_task"
	"collab-net-v2/task/service_task"
	"collab-net-v2/util/logrus_wrap"
	"collab-net-v2/workflow"
	"context"
	"fmt"
	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

func OnTaskChange(idTask string, evt int, x []byte) (e error) {
	log.Println("USER [OnTaskChange] idTask = ", idTask, ", evt = ", evt)

	return nil
}

func main() {
	ctx := context.Background()
	var logger *logrus.Logger
	log.Println("main [init] : ")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	fmt.Printf("GitCommit: %s\n", GitCommit)

	instance := config.GetRunningInstance()
	logServer := config.GetLogServer()
	//
	logger = logrus.New()
	logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	hook, err := logrustash.NewHook("tcp", logServer, instance)
	if err != nil {
		log.Fatal("logrustash.NewHook: ", err)
	}
	logger.Hooks.Add(hook)

	log := logger.WithFields(logrus.Fields{
		"method": "main",
	})

	v := config.GetDependMsgRpc()
	if v == "" {
		log.Fatal("config.GetDependMsgRpc()  v=empty ")
	}
	log.Println("GetDependMsgRpc: ", v)
	ok := message.GetMsgCtl().Init(v)
	if !ok {
		log.Fatal("message.GetMsgCtl().Init(v) error, addr = ", v)
	}
	message.GetMsgCtl().UpdateTaskWrapper("taskid_demo", 1001, "demo msg")

	log.Println("[main] SetTaskCallback")
	service_task.SetTaskCallback(OnTaskChange)

	time.Sleep(time.Millisecond * 100)
	go func() {
		e := control_link.InitGinService(logrus_wrap.SetContextLogger(ctx, logger), config_link.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_link.InitGinService e: ", e)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	service_sched.Init()

	time.Sleep(time.Millisecond * 100)
	go func() {
		e := control_task.InitGinService(logrus_wrap.SetContextLogger(ctx, logger), config_task.LISTEN_PORT)
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
