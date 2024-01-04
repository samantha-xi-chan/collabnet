package service_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/util/util_mq"
	"collab-net-v2/workflow/repo_workflow"
	"log"
	"sync"
)

var instance *MySingleton
var once sync.Once

func GetMqInstance() *MySingleton {
	once.Do(func() {
		instance = &MySingleton{}
	})

	return instance
}

type MySingleton struct {
	data int
	mq   util_mq.RabbitMQManager
}

func (s *MySingleton) SetData(data int) {
	s.data = data
}
func (s *MySingleton) GetData() int {
	return s.data
}

func (s *MySingleton) PostMsgToQueue(queueName string, msg string, priority int64) int {

	repo_workflow.GetTaskCtl().UpdateItemByID(msg, map[string]interface{}{
		"status": api.TASK_STATUS_QUEUEING,
	})
	s.mq.Publish(queueName, []byte(msg), uint8(4))

	return 0
}

func (s *MySingleton) Init(rabbitUrl string, queueName string, priorityMax int64) int {

	SIZE_PRODUCER := 1

	// 定义操作对象
	_mq := util_mq.RabbitMQManager{}
	//defer mq.Release()

	if err := _mq.InitQ(rabbitUrl, SIZE_PRODUCER, false); err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	log.Println("PlayAsProducerBlock init ok")

	_mq.DeclarePublishQueue(queueName, priorityMax, true)

	s.mq = _mq

	return 0
}
