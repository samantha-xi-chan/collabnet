package main

import (
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/task/service_task"
	"log"
	"time"
)

func OnTaskChange(idTask string, evt int, x []byte) (e error) {

	log.Println("[OnTaskChange] ", idTask, " ", evt)

	return nil
}

func main() {
	service_task.SetTaskCallback(OnTaskChange)

	go func() {
		time.Sleep(time.Second * 20)
		id, e := service_task.NewTask("ls -alh ", "M1", config_sched.CMD_ACK_TIMEOUT, config_sched.TEST_TIMEOUT_PREPARE, config_sched.TEST_TIMEOUT_RUN)
		if e != nil {
			log.Println("[main] service_task.NewTask, e=", e)
			return
		}

		log.Println("[main]  id :  ", id)
	}()

	log.Println("[main] waiting select{}")
	select {}
}
