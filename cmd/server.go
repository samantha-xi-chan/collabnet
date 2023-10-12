package main

import (
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/control_link"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/control_task"
	"collab-net-v2/task/service_task"
	"log"
)

func OnTaskChange(idTask string, evt int, x []byte) (e error) {
	log.Println("[OnTaskChange] ", idTask, " ", evt)

	return nil
}

func main() {
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
