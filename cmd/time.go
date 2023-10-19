package main

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/time/control_time"
	"collab-net-v2/time/service_time"
	"log"
)

func init() {
	log.Println("main [init] : ")

	if true { // if in k8s
		service_time.Init(
			"amqp://guest:guest@rmq-cluster:5672",
			config_sched.AMQP_EXCH,
			"root:password@tcp(mysql:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local")

		log.Println("service_time.Init ok  222")
	} else { // if native
		mqDsn, e := config.GetMqDsn()
		if e != nil {
			log.Fatal("config.GetMqDsn() e=", e)
		}
		log.Println("mqDsn: ", mqDsn)

		mySqlDsn, e := config.GetMySqlDsn()
		if e != nil {
			log.Fatal("config.GetMySqlDsn: ", e)
		}
		log.Println("mySqlDsn", mySqlDsn)

		service_time.Init(
			mqDsn,
			config_sched.AMQP_EXCH,
			mySqlDsn)
	}

	log.Println("main [init]  end ")
}

func main() {
	log.Println("[main] start ... ")
	go func() {
	}()

	go func() {
		addr := ":8088"
		log.Printf("goting to InitTimeHttpService on  %s\n", addr)
		e := control_time.InitTimeHttpService(addr)
		if e != nil {
			log.Fatal("control_time.InitTimeHttpService e: ", e)
		}
	}()

	log.Println("[main] waiting select{}")
	select {}
}
