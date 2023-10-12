package service_time

import (
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/time/api"
	"collab-net-v2/time/repo_time"
	"collab-net-v2/time/util/rmq_util"
	"fmt"
	"log"

	"time"
	//"github.com/apex/log"
	//"github.com/apex/log/handlers/cli"
	//"github.com/jinzhu/configor"
)

const (
	DEBUG = false
)

func shouldAck(x []byte) bool {
	go func() {
		strVal := string(x)
		log.Printf("in shouldAck(), strVal=%s\n", strVal)

		item, e := repo_time.GetTimeCtl().GetItemByKeyValue("id_once", strVal)
		if e != nil {
			if DEBUG {
				log.Println("repo_time.GetTimeCtl().GetItemByKeyValue e: ", e)
			}
			return
		}
		if item.Status == api.STATUS_TIMER_DISABLED {
			if DEBUG {
				log.Println("STATUS_TIMER_DISABLED ")
			}
			return
		}
		if DEBUG {
			log.Println("item: ", item)
		}

		callbackFunc(item.Id, item.Type, item.Holder, nil)
	}()

	return true
}

var rmq *rmq_util.RabbitMQ
var err error

var callbackFunc func(idTimer string, _type int, holder string, bytes []byte) (x error)

func SetCallback(tmp func(idTimer string, _type int, holder string, bytes []byte) (x error)) {
	callbackFunc = tmp
}

func Init(url string, exchange string) {
	//repo_time.Init("root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local", 5, 200)
	repo_time.Init(config_sched.RepoMySQLDsn, config_sched.RepoLogLevel, config_sched.RepoSlowMs)

	rmq, err = rmq_util.InitRabbitMQ(rmq_util.AMQP{
		URL:      url,
		Exchange: exchange,
	}, shouldAck)
	if err != nil {
		log.Fatalf("run: failed to init rabbitmq: %v", err)
	}

}

func NewTimer(timeoutSecond int, _type int, holder string, desc string) (id string, e error) {

	idTimer := idgen.GetIdWithPref("time")
	idOnce := idgen.GetIdWithPref(fmt.Sprintf("once_%s_", desc))

	repo_time.GetTimeCtl().CreateItem(repo_time.Time{
		Id:       idTimer,
		Type:     _type,
		Holder:   holder,
		Desc:     desc,
		Status:   api.STATUS_TIMER_INITED,
		CreateAt: time.Now().UnixMilli(),
		IdOnce:   idOnce,
		CreateBy: 0,
		Timeout:  timeoutSecond,
	})

	err := rmq.PublishWithDelay(rmq_util.KEY, []byte(idOnce), int64(1000*timeoutSecond))
	if err != nil {
		log.Printf("run: failed to publish into rabbitmq: %v", err)
	}

	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"status": api.STATUS_TIMER_RUNNING,
	})

	return idTimer, nil
}

func DisableTimer(id string) (e error) {
	item, e := repo_time.GetTimeCtl().GetItemById(id)
	if e != nil {
		log.Printf("repo_time.GetTimeCtl().GetItemById, e= %s , id = %s \n ", e, id)
		return
	}

	log.Printf("[DisableTimer]  id = %s, type =  %d , desc = %s\n", id, item.Type, item.Desc)

	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"status": api.STATUS_TIMER_DISABLED,
	})

	return nil
}

func RenewTimer(id string, timeoutSecond int) (e error) {
	idOnce := idgen.GetIdWithPref("once_renew")

	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"id_once": idOnce,
	})

	err := rmq.PublishWithDelay("user.event.publish", []byte(idOnce), int64(1000*timeoutSecond))
	if err != nil {
		log.Printf("run: failed to publish into rabbitmq: %v", err)
	}

	return nil
}
