package main

import (
	"collab-net-v2/internal/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/internal/repo"
	"collab-net-v2/link"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/time/service_time"
	"encoding/json"
	"log"
	"time"
)

var (
//map
)

func OnTimer(idTimer string, _type int, holder string, bytes []byte) (ee error) { // 都是不吉利的情况
	log.Printf("[OnTimer]holder %s,  idTimer: %s, bytes: %s", holder, idTimer, string(bytes))
	itemSched, e := repo.GetSchedCtl().GetItemById(holder)
	if e != nil {
		log.Println("repo.GetSchedCtl().GetItemById e :", e)
		return nil
	}
	if itemSched.Status == api.STATUS_SCHED_END {
		log.Println("itemSched.Status == api.STATUS_SCHED_END")
		return nil
	}
	if itemSched.Enabled == api.INT_DISABLED {
		log.Println("itemSched.Enabled == api.INT_DISABLED")
		return nil
	}

	if _type == api.STATUS_SCHED_CMD_ACKED {
		log.Printf("[OnTimer]  STATUS_SCHED_CMD_ACKED ")
		repo.GetSchedCtl().UpdateItemByID(itemSched.ID, map[string]interface{}{
			"status": api.STATUS_SCHED_CMD_ACKED,
			//"code":   api.RESULT_SCHED_KILLED_PRE_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_PRE_ACKED {
		repo.GetSchedCtl().UpdateItemByID(itemSched.ID, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			//"code":   api.RESULT_SCHED_KILLED_PRE_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_HEARTBEAT {
		repo.GetSchedCtl().UpdateItemByID(itemSched.ID, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			//"code":   api.RESULT_SCHED_KILLED_HB_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_END {
		repo.GetSchedCtl().UpdateItemByID(itemSched.ID, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			//"code":   api.RESULT_SCHED_KILLED_RUN_TIMEOUT,
		})
	} else {
		log.Println("OnTimer unknown")
	}

	return
}

func init() {
	log.Println("init: ")

	service_time.SetCallback(OnTimer)
	service_time.Init(config.AMQP_URL, config.AMQP_EXCH)
	repo.Init(config.RepoMySQLDsn, config.RepoLogLevel, config.RepoSlowMs)
}

func OnConnChange(endpoint string, _type int) (e error) {
	log.Println("[OnConnChange]: endpoint ", endpoint, ", _type ", _type)

	go func() {

		if _type == link.LINK_EVT_HANDSHAKE_OK {
			test(endpoint)
		} else if _type == link.LINK_EVT_BYE {
			log.Println("bye")
		} else {
			log.Println("unknown")
		}

	}()

	return nil
}

func OnBizDataFromRegisterEndpoint(endpoint string, bytes []byte) (e error) {
	log.Printf("[OnBizDataFromRegisterEndpoint] endpoint: %s, bytes: %s", endpoint, string(bytes))

	var body link.BizData
	err := json.Unmarshal(bytes, &body)
	if err != nil {
		log.Println("[OnBizDataFromRegisterEndpoint]    Error decoding JSON:", err, " string(bytes): ", string(bytes))
		return
	}
	log.Println("[OnBizDataFromRegisterEndpoint]    [<-readChan] body : ", body)
	idSched := body.Id

	itemSched, e := repo.GetSchedCtl().GetItemById(idSched)

	status := api.STATUS_SCHED_CMD_ACKED
	if body.Msg == "STATUS_SCHED_END" {
		status = api.STATUS_SCHED_END
		idTimer, _ := service_time.NewTimer(itemSched.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "prepare_timeout")

		repo.GetSchedCtl().UpdateItemByID(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
			"pre_timer": idTimer,
			"code":      api.RESULT_SCHED_OK,
		})
	} else if body.Msg == "STATUS_SCHED_PRE_ACKED" {
		status = api.STATUS_SCHED_PRE_ACKED

		service_time.DisableTimer(itemSched.PreTimer)
		idRunTimer, _ := service_time.NewTimer(itemSched.RunTimeout, api.STATUS_SCHED_END, idSched, "run_timeout")
		idHbTimer, _ := service_time.NewTimer(config.SCHED_HEARTBEAT_INTERVAL*3, api.STATUS_SCHED_HEARTBEAT, idSched, "hb_timeout")

		repo.GetSchedCtl().UpdateItemByID(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
			"run_timer": idRunTimer,
			"hb_timer":  idHbTimer,
		})
	} else if body.Msg == "STATUS_SCHED_HEARTBEAT" {
		status = api.STATUS_SCHED_HEARTBEAT

		repo.GetSchedCtl().UpdateItemByID(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
		})

		// reset timer
		service_time.RenewTimer(itemSched.HbTimer, config.SCHED_HEARTBEAT_INTERVAL*3)
	}

	return
}

func StopSched(id string) (ee error) {

	item, e := repo.GetSchedCtl().GetItemById(id)
	if e != nil {
		log.Println("repo.GetSchedCtl().GetItemById, e=", e)
		return
	}

	if item.Enabled == 0 {
		log.Println("item.Enabled == 0")
		return
	}

	if item.Status == api.STATUS_SCHED_END {
		log.Println("item.Status == api.STATUS_SCHED_END")
		return
	}

	repo.GetSchedCtl().UpdateItemByID(
		id,
		map[string]interface{}{
			"status":  api.STATUS_SCHED_END,
			"enabled": 0,
		},
	)

	return
}

func NewSched(cmd string, endpoint string, preTimeoutSecond int, runTimeoutSecond int) (_id string, e error) {
	id := idgen.GetIDWithPref("sched")
	repo.GetSchedCtl().CreateItem(repo.Sched{
		ID:         id,
		Desc:       "",
		Status:     api.STATUS_SCHED_INIT,
		Endpoint:   endpoint,
		CreateAt:   time.Now().UnixMilli(),
		ActiveAt:   time.Now().UnixMilli(),
		Enabled:    1,
		PreTimeout: preTimeoutSecond,
		RunTimeout: runTimeoutSecond,
		Code:       api.INT_INVALID,
	})

	code, e := link.SendDataToEndpoint(
		endpoint,
		link.GetPackageBytes(
			time.Now().UnixMilli(),
			"v1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				Id:         id,
				Code:       0,
				HbInterval: config.SCHED_HEARTBEAT_INTERVAL,
				PreTimeout: preTimeoutSecond,
				RunTimeout: runTimeoutSecond,
				Msg:        cmd,
			}))

	if e != nil || code != 0 {
		log.Println("link.SendDataToEndpoint failed ")

		repo.GetSchedCtl().UpdateItemByID(id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_LOCAL_FAIL,
		})
		return id, nil
	} else {
		repo.GetSchedCtl().UpdateItemByID(id, map[string]interface{}{
			"status": api.STATUS_SCHED_SENT,
		})
	}

	return id, nil
}

func test(endpoint string) {
	time.Sleep(time.Second * 1)
	for i := 0; i < 9999; i++ {
		log.Println("test starting ...")
		id, e := NewSched("ls -alh ", endpoint, 5, 40)
		if e != nil {
			log.Println("NewSched e: ", e)
		}
		log.Println("NewSched succ: id ", id)

		//time.Sleep(time.Second * 100)
		//StopSched(id)
	}
}

func main() {
	link.SetConnChangeCallback(OnConnChange)
	link.SetBizDataCallback(OnBizDataFromRegisterEndpoint)
	go link.NewServer()

	log.Println("[main] waiting select{}")
	select {}
}
