package workflow

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/control_workflow"
	"collab-net-v2/workflow/repo_workflow"
	"collab-net-v2/workflow/service_workflow"
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

	go func() {
		service_workflow.PlayAsConsumerBlock(mqDsn, 1)
	}()
	time.Sleep(time.Millisecond * 200)

	log.Println("going to listen on : ", config_workflow.LISTEN_PORT)
	control_workflow.StartHttpServer(config_workflow.LISTEN_PORT)
}
