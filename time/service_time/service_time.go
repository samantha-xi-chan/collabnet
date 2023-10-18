package service_time

import (
	"bytes"
	"collab-net-v2/internal/config"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/time/api_time"
	"collab-net-v2/time/repo_time"
	"collab-net-v2/time/util/rmq_util"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"

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
		if DEBUG {
			log.Printf("in shouldAck(), strVal=%s\n", strVal)
		}

		item, e := repo_time.GetTimeCtl().GetItemByKeyValue("id_once", strVal)
		if e != nil {
			if DEBUG {
				log.Println("repo_time.GetTimeCtl().GetItemByKeyValue e: ", e)
			}
			return
		}
		if item.Status == api_time.STATUS_TIMER_DISABLED {
			if DEBUG {
				log.Println("STATUS_TIMER_DISABLED ")
			}
			return
		}
		if DEBUG {
			log.Println("item: ", item)
		}

		if callbackFunc != nil {
			callbackFunc(item.Id, item.Type, item.Holder, nil)
		}
		if item.CallbackAddr != "" {
			go func() {
				data := api_time.CallbackReq{
					Id:      item.Id,
					Type:    item.Type,
					Holder:  item.Holder,
					Desc:    item.Desc,
					Timeout: item.Timeout,
				}

				jsonData, err := json.Marshal(data)
				if err != nil {
					log.Println("json.Marshal err:", err)
					return
				}
				response, err := http.Post(item.CallbackAddr, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Println("http.Post err:", err)
					return
				}
				defer response.Body.Close()
			}()
		}
	}()

	return true
}

var rmq *rmq_util.RabbitMQ
var err error

type FUNC_TIMEOUT_CB func(idTimer string, _type int, holder string, bytes []byte) (x error)

var callbackFunc FUNC_TIMEOUT_CB

func SetCallback(tmp FUNC_TIMEOUT_CB) {
	callbackFunc = tmp
}

func Init(url string, exchange string) {
	log.Println("service_time [Init] : ")

	mySqlDsn, e := config.GetMySqlDsn()
	if e != nil {
		log.Fatal("config.GetMySqlDsn: ", e)
	}
	log.Println("mySqlDsn", mySqlDsn)

	repo_time.Init(mySqlDsn, config_sched.RepoLogLevel, config_sched.RepoSlowMs)

	rmq, err = rmq_util.InitRabbitMQ(rmq_util.AMQP{
		URL:      url,
		Exchange: exchange,
	}, shouldAck)
	if err != nil {
		log.Fatalf("run: failed to init rabbitmq: %v", err)
	}

}

func NewTimer(timeoutSecond int, _type int, holder string, desc string, callbackAddr string) (id string, e error) {
	log.Printf("[NewTimer] timeoutSecond=%d , _type = %d, holder = %s, desc = %s , callbackAddr = %s \n", timeoutSecond, _type, holder, desc, callbackAddr)

	idTimer := idgen.GetIdWithPref("time")
	idOnce := idgen.GetIdWithPref(fmt.Sprintf("once_%s_", desc))

	repo_time.GetTimeCtl().CreateItem(repo_time.Time{
		Id:           idTimer,
		Type:         _type,
		Holder:       holder,
		Desc:         desc,
		Status:       api_time.STATUS_TIMER_INITED,
		CreateAt:     time.Now().UnixMilli(),
		IdOnce:       idOnce,
		CreateBy:     0,
		Timeout:      timeoutSecond,
		CallbackAddr: callbackAddr,
	})

	err := rmq.PublishWithDelay(rmq_util.KEY, []byte(idOnce), int64(1000*timeoutSecond))
	if err != nil {
		log.Printf("run: failed to publish into rabbitmq: %v", err)
	}

	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"status": api_time.STATUS_TIMER_RUNNING,
	})

	return idTimer, nil
}

func DisableTimer(id string) (e error) {
	log.Printf("[DisableTimer]  id=%s \n", id)
	item, e := repo_time.GetTimeCtl().GetItemById(id)
	if e != nil {
		log.Printf("repo_time.GetTimeCtl().GetItemById, e= %s , id = %s \n ", e, id)
		return
	}

	log.Printf("    [DisableTimer]  id = %s, type =  %d , desc = %s\n", id, item.Type, item.Desc)
	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"status": api_time.STATUS_TIMER_DISABLED,
	})

	return nil
}

func RenewTimer(id string, timeoutSecond int) (e error) {
	log.Printf("[RenewTimer] id=%s  timeoutSecond=%d \n", id, timeoutSecond)
	idOnce := idgen.GetIdWithPref("once_renew")

	repo_time.GetTimeCtl().UpdateItemById(id, map[string]interface{}{
		"id_once": idOnce,
	})

	err := rmq.PublishWithDelay("user.event.publish", []byte(idOnce), int64(1000*timeoutSecond))
	if err != nil {
		log.Printf("run: failed to publish into rabbitmq: %v", err)
		return errors.Wrap(err, "rmq.PublishWithDelay: ")
	}

	return nil
}
