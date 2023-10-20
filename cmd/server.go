package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/control_link"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/control_task"
	"collab-net-v2/task/service_task"
	"collab-net-v2/util/logrus_wrap"
	"context"
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
	instance := config.GetRunningInstance()
	logServer := config.GetLogServer()
	//
	logger = logrus.New()
	logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	hook, err := logrustash.NewHook("tcp", logServer, instance)
	if err != nil {
		log.Fatal(err)
	}
	logger.Hooks.Add(hook)

	log := logger.WithFields(logrus.Fields{
		"method": "main",
	})

	log.Println("[main] SetTaskCallback")
	service_task.SetTaskCallback(OnTaskChange)

	go func() {
		time.Sleep(time.Second * 1)
		e := control_link.InitGinService(logrus_wrap.SetContextLogger(ctx, logger), config_link.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_link.InitGinService e: ", e)
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)
		e := control_task.InitGinService(logrus_wrap.SetContextLogger(ctx, logger), config_task.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_task.InitGinService4 e: ", e)
		}
	}()

	log.Println("[main] waiting select{}")
	select {}
}
