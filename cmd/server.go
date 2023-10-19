package main

import (
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/control_link"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/control_task"
	"collab-net-v2/task/service_task"
	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func OnTaskChange(idTask string, evt int, x []byte) (e error) {
	log.Println("USER [OnTaskChange] idTask = ", idTask, ", evt = ", evt)

	return nil
}

func main() {
	log.Println("main [init] : ")
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		log.Println("Failed to get POD_NAME environment variable")
	} else {
		log.Printf("Pod Name: %s\n", podName)
	}

	logServer := os.Getenv("LOG_SERVER")
	if logServer == "" {
		log.Println("Failed to get LOG_SERVER environment variable")
	} else {
		log.Printf("logServer: %s\n", logServer)
	}

	//
	logger = logrus.New()
	logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	hook, err := logrustash.NewHook("tcp", logServer, ""+podName)
	if err != nil {
		log.Fatal(err)
	}
	logger.Hooks.Add(hook)

	log := logger.WithFields(logrus.Fields{
		"method": "main",
	})

	service_task.SetTaskCallback(OnTaskChange)

	go func() {
		e := control_link.InitGinService(config_link.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_link.InitGinService e: ", e)
		}
	}()

	go func() {
		e := control_task.InitGinService(config_task.LISTEN_PORT)
		if e != nil {
			log.Fatal("control_task.InitGinService4 e: ", e)
		}
	}()

	log.Println("[main] waiting select{}")
	select {}
}
