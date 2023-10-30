package workflow

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/control_workflow"
	"collab-net-v2/workflow/repo_workflow"
	"collab-net-v2/workflow/service_workflow"
	"context"
	"log"
	"time"
)

func StartService() {
	mysqlDsn, _ := config.GetMySqlDsn()
	log.Println("mysqlDsn: ", mysqlDsn)
	repo_workflow.Init(mysqlDsn, config_workflow.RepoLogLevel, config_workflow.RepoSlowMs)
	mqDsn, _ := config.GetMqDsn()
	log.Println("mqDsn: ", mqDsn)
	service_workflow.GetMqInstance().Init(mqDsn, config.QUEUE_NAME, config.PRIORITY_MAX)
	redisDsn, _ := config.GetRedisDsn()
	log.Println("redisDsn: ", redisDsn)
	e := repo_workflow.InitRedis(context.Background(), redisDsn, 9999, 0)
	if e != nil {
		log.Fatal("InitRedis: ", e)
	}

	taskConcurrent := config.GetTaskConcurrent()
	if taskConcurrent == 0 {
		log.Fatal("taskConcurrent == 0")
	}
	log.Println("taskConcurrent: ", taskConcurrent)

	go func() {
		service_workflow.PlayAsConsumerBlock(mqDsn, taskConcurrent)
	}()
	time.Sleep(time.Millisecond * 200)

	log.Println("going to listen on : ", config_workflow.LISTEN_PORT)
	control_workflow.StartHttpServer(config_workflow.LISTEN_PORT)
}
