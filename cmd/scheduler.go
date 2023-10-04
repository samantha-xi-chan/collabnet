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

func init() {
	log.Println("init: ")

	service_time.SetCallback(OnTimerWrapper)
	service_time.Init(config.AMQP_URL, config.AMQP_EXCH)
	repo.Init(config.RepoMySQLDsn, config.RepoLogLevel, config.RepoSlowMs)
}

func OnConnChange(endpoint string, _type int) (e error) {
	log.Println("[OnConnChange]: endpoint ", endpoint, ", _type ", _type)

	// test
	go func() {
		if _type == link.LINK_EVT_HANDSHAKE_OK {
			for i := 0; i < config.TESTCASE_CNT; i++ {

				//go func() {

				log.Println("test starting ...")
				id, e := NewSched("ls -alh ", endpoint, config.CMD_ACK_TIMEOUT, config.TEST_TIMEOUT_PREPARE, config.TEST_TIMEOUT_RUN)
				if e != nil {
					log.Println("NewSched e: ", e)
				}
				log.Println("NewSched succ: id ", id)
				//}()
			}
		} else if _type == link.LINK_EVT_BYE {
			log.Println("bye: ", endpoint)
		} else {
			log.Println("unknown")
		}

	}()

	return nil
}

func OnTimerWrapper(idTimer string, _type int, holder string, bytes []byte) (ee error) {
	log.Printf("[OnTimerWrapper] holder %s,  idTimer: %s, bytes: %s", holder, idTimer, string(bytes))
	go func() {
		OnTimer(idTimer, _type, holder, bytes)
	}()
	return nil
}

func OnTimer(idTimer string, _type int, holder string, bytes []byte) (ee error) { // 定时器 事件
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

	// 更新状态 并关闭 已经不再需要的定时器
	if _type == api.STATUS_SCHED_CMD_ACKED {
		log.Printf("[OnTimer]  STATUS_SCHED_CMD_ACKED ")
		repo.GetSchedCtl().UpdateItemById(itemSched.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_CMDACK_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_PRE_ACKED {
		repo.GetSchedCtl().UpdateItemById(itemSched.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_PRE_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_HEARTBEAT {
		repo.GetSchedCtl().UpdateItemById(itemSched.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_HB_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_END {
		repo.GetSchedCtl().UpdateItemById(itemSched.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   0,
		})

		service_time.DisableTimer(itemSched.HbTimer)

	} else {
		log.Println("OnTimer unknown")
	}

	return
}

func OnBizDataFromRegisterEndpointWrapper(endpoint string, bytes []byte) (e error) { // 网络 事件
	//go func() {
	OnBizDataFromRegisterEndpoint(endpoint, bytes)
	//}()
	return nil
}

func OnBizDataFromRegisterEndpoint(endpoint string, bytes []byte) (e error) { // 网络 事件
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
	if e != nil {
		log.Println("repo.GetSchedCtl().GetItemById: ", e)
		return
	}

	if itemSched.Status == api.STATUS_SCHED_END {
		log.Println("itemSched.Status == api.STATUS_SCHED_END")
		return nil
	}
	if itemSched.Enabled == api.INT_DISABLED {
		log.Println("itemSched.Enabled == api.INT_DISABLED")
		return nil
	}

	status := api.INT_INVALId
	if body.Msg == "STATUS_SCHED_CMD_ACKED" {
		status = api.STATUS_SCHED_CMD_ACKED

		service_time.DisableTimer(itemSched.CmdackTimer)
		idPreTimer, _ := service_time.NewTimer(itemSched.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "pre_ack_timeout")

		repo.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
			"pre_timer": idPreTimer,
		})
	} else if body.Msg == "STATUS_SCHED_PRE_ACKED" {
		status = api.STATUS_SCHED_PRE_ACKED

		service_time.DisableTimer(itemSched.PreTimer)
		idRunTimer, _ := service_time.NewTimer(itemSched.RunTimeout, api.STATUS_SCHED_END, idSched, "run_finish_timeout")
		idHbTimer, _ := service_time.NewTimer(config.SCHED_HEARTBEAT_TIMEOUT, api.STATUS_SCHED_HEARTBEAT, idSched, "hb_timeout")

		repo.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
			"run_timer": idRunTimer,
			"hb_timer":  idHbTimer,
		})
	} else if body.Msg == "STATUS_SCHED_HEARTBEAT" {
		status = api.STATUS_SCHED_HEARTBEAT

		repo.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
		})

		// reset timer
		service_time.RenewTimer(itemSched.HbTimer, config.SCHED_HEARTBEAT_TIMEOUT)
	} else if body.Msg == "STATUS_SCHED_END" {
		status = api.STATUS_SCHED_END
		//idTimer, _ := service_time.NewTimer(itemSched.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "prepare_timeout")

		repo.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"finish_at": time.Now().UnixMilli(),
			"status":    status,
			"code":      api.RESULT_SCHED_OK,
		})

		service_time.DisableTimer(itemSched.HbTimer)
		service_time.DisableTimer(itemSched.RunTimer)

	} else {

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

	repo.GetSchedCtl().UpdateItemById(
		id,
		map[string]interface{}{
			"status":  api.STATUS_SCHED_END,
			"enabled": 0,
		},
	)

	return
}

func NewSched(cmd string, endpoint string, cmdackTimeoutSecond int, preTimeoutSecond int, runTimeoutSecond int) (_id string, e error) {
	idSched := idgen.GetIdWithPref("sched")
	repo.GetSchedCtl().CreateItem(repo.Sched{
		Id:         idSched,
		Desc:       "",
		Status:     api.STATUS_SCHED_INIT,
		Endpoint:   endpoint,
		CreateAt:   time.Now().UnixMilli(),
		ActiveAt:   time.Now().UnixMilli(),
		Enabled:    1,
		PreTimeout: preTimeoutSecond,
		RunTimeout: runTimeoutSecond,
		Code:       api.INT_INVALId,
	})

	code, e := link.SendDataToEndpoint(
		endpoint,
		link.GetPackageBytes(
			time.Now().UnixMilli(),
			"v1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				Id:         idSched,
				Code:       0,
				HbInterval: config.SCHED_HEARTBEAT_INTERVAL,
				PreTimeout: preTimeoutSecond,
				RunTimeout: runTimeoutSecond,
				Msg:        cmd,
			}))

	if e != nil || code != 0 {
		log.Println("link.SendDataToEndpoint failed ")

		repo.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_LOCAL_FAIL,
		})
		return idSched, nil
	}

	idCmdackTimer, _ := service_time.NewTimer(cmdackTimeoutSecond, api.STATUS_SCHED_CMD_ACKED, idSched, "cmd_ack_timeout")
	repo.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
		"active_at":    time.Now().UnixMilli(),
		"status":       api.STATUS_SCHED_SENT,
		"cmdack_timer": idCmdackTimer,
	})
	return idSched, nil
}

func main() {
	link.SetConnChangeCallback(OnConnChange)
	link.SetBizDataCallback(OnBizDataFromRegisterEndpointWrapper)
	go link.NewServer()

	log.Println("[main] waiting select{}")
	select {}
}
