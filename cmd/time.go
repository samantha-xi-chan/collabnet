package main

import (
	//"collab-net-v2/internal/config"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/time/control_time"
	"collab-net-v2/time/service_time"
	"collab-net-v2/util/logrus_wrap"
	"context"
	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logServer := "192.168.36.101:5000"
	logger.SetLevel(logrus.TraceLevel) // 后续改为 配置中心处理
	//logger.SetFormatter(&logrus_wrap.TextFormatter{
	//	FullTimestamp: true,
	//})
	hook, err := logrustash.NewHook("tcp", logServer, "serviceName002")
	if err != nil {
		log.Fatal(err)
	}
	logger.Hooks.Add(hook)

	log := logger.WithFields(logrus.Fields{
		"method": "init",
	})

	log.Println("main [init] : ")
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		log.Println("Failed to get POD_NAME environment variable")
	} else {
		log.Printf("Pod Name: %s\n", podName)
	}

	if true { // if in k8s
		service_time.Init(
			"amqp://guest:guest@rmq-cluster:5672",
			config_sched.AMQP_EXCH,
			"root:password@tcp(mysql:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local")

		log.Println("service_time.Init ok  222")
	} else { // if native
		//mqDsn, e := config.GetMqDsn()
		//if e != nil {
		//	log.Fatal("config.GetMqDsn() e=", e)
		//}
		//log.Println("mqDsn: ", mqDsn)
		//
		//mySqlDsn, e := config.GetMySqlDsn()
		//if e != nil {
		//	log.Fatal("config.GetMySqlDsn: ", e)
		//}
		//log.Println("mySqlDsn", mySqlDsn)
		//
		//service_time.Init(
		//	mqDsn,
		//	config_sched.AMQP_EXCH,
		//	mySqlDsn)
	}

	log.Println("main [init]  end ")
}

func main() {
	// 业务逻辑
	log := logger.WithFields(logrus.Fields{
		"method": "main",
	})

	log.Println("[main] start ... ")

	go func() {
		addr := ":8088"
		log.Printf("goting to InitTimeHttpService on  %s\n", addr)
		e := control_time.InitTimeHttpService(logrus_wrap.SetContextLogger(context.Background(), logger), addr)
		if e != nil {
			log.Fatal("control_time.InitTimeHttpService e: ", e)
		}
	}()

	log.Println("[main] waiting select{}")
	select {}
}
