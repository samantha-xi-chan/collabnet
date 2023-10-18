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

	service_time.Init(mqDsn, config_sched.AMQP_EXCH, mySqlDsn)
}

func main() {

	go func() {
		e := control_time.InitTimeHttpService(":8088")
		if e != nil {
			log.Fatal("control_time.InitTimeHttpService e: ", e)
		}
	}()

	log.Println("[main] waiting select{}")
	select {}
}
