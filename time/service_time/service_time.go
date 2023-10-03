package service_time

import (
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/time/api"
	"collab-net-v2/time/repo_time"
	"collab-net-v2/time/util/rmq_util"
	"log"

	"time"
	//"github.com/apex/log"
	//"github.com/apex/log/handlers/cli"
	//"github.com/jinzhu/configor"
)

func shouldAck(x []byte) bool {
	strVal := string(x)
	log.Printf("in shouldAck(), strVal=%s\n", strVal)

	item, _ := repo_time.GetTimeCtl().GetItemByKeyValue("id_once", strVal)
	if item.Status == api.STATUS_TIMER_DISABLED {
		log.Println("STATUS_TIMER_DISABLED ")
		return true
	}

	log.Println("item: ", item)
	return true
}

var rmq *rmq_util.RabbitMQ
var err error

func Init() {
	repo_time.Init("root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local", 5, 200)

	rmq, err = rmq_util.InitRabbitMQ(rmq_util.AMQP{
		URL:      "amqp://guest:guest@192.168.31.7:5672",
		Exchange: "amq.direct",
	}, shouldAck)
	if err != nil {
		log.Fatalf("run: failed to init rabbitmq: %v", err)
	}

}

func NewTimer(timeoutSecond int, desc string) (id string, e error) {

	idTimer := idgen.GetIDWithPref("time")
	idOnce := idgen.GetIDWithPref("once")

	repo_time.GetTimeCtl().CreateItem(repo_time.Time{
		ID:       idTimer,
		Desc:     desc,
		Status:   api.STATUS_TIMER_INITED,
		CreateAt: time.Now().UnixMilli(),
		IdOnce:   idOnce,
		CreateBy: 0,
		Timeout:  timeoutSecond,
	})

	err := rmq.PublishWithDelay("user.event.publish", []byte(idOnce), int64(1000*timeoutSecond))
	if err != nil {
		log.Fatalf("run: failed to publish into rabbitmq: %v", err)
	}

	repo_time.GetTimeCtl().UpdateItemByID(id, map[string]interface{}{
		"status": api.STATUS_TIMER_RUNNING,
	})

	return idTimer, nil
}

func DisableTimer(id string) (e error) {
	repo_time.GetTimeCtl().UpdateItemByID(id, map[string]interface{}{
		"status": api.STATUS_TIMER_DISABLED,
	})

	return nil
}
