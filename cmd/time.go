package main

import (
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/time/control_time"
	"collab-net-v2/time/service_time"
	"log"
)

func init() {
	log.Println("main [init] : ")

	/*
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
	*/

}

func main() {
	log.Println("[main] start ... ")
	go func() {
		service_time.Init(
			"amqp://RABBITMQ_USER:RABBITMQ_PASS@rmq-cluster:5672/",
			config_sched.AMQP_EXCH,
			"root:password@tcp(mysql:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local")

		log.Println("service_time.Init ok ")
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
