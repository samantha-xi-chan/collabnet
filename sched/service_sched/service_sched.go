package service_sched

import (
	"collab-net-v2/link"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/api"
	"collab-net-v2/sched/config_sched"
	repo_sched "collab-net-v2/sched/repo_sched"
	"collab-net-v2/time/service_time"
	"encoding/json"
	"log"
	"time"
)

var (
//map
)

// 对上层

type FUN_CALLBACK func(idSched string, evt int, bytes []byte) (ee error)

var callback FUN_CALLBACK

func SetCallbackFun(f FUN_CALLBACK) {
	callback = f
}

func init() {
	log.Println("service_sched init: ")

	// 与下层的通信 01
	link.SetConnChangeCallback(OnConnChange)
	link.SetBizDataCallback(OnBizDataFromRegisterEndpointWrapper)
	go link.NewServer()

	// 与下层的通信 02
	service_time.SetCallback(OnTimerWrapper)
	service_time.Init(config_sched.AMQP_URL, config_sched.AMQP_EXCH)

	// 与下层的通信 03
	repo_sched.Init(config_sched.RepoMySQLDsn, config_sched.RepoLogLevel, config_sched.RepoSlowMs)
}

// 对下层
func OnConnChange(endpoint string, _type int) (e error) {
	log.Println("[OnConnChange]: endpoint ", endpoint, ", _type ", _type)

	//go func() {
	//	if _type == link.LINK_EVT_HANDSHAKE_OK {
	//		for i := 0; i < config_task.TESTCASE_CNT; i++ {
	//
	//			//go func() {
	//
	//			log.Println("test starting ...")
	//			id, e := NewSched("ls -alh ", endpoint, config_task.CMD_ACK_TIMEOUT, config_task.TEST_TIMEOUT_PREPARE, config_task.TEST_TIMEOUT_RUN)
	//			if e != nil {
	//				log.Println("NewTask e: ", e)
	//			}
	//			log.Println("NewTask succ: id ", id)
	//			//}()
	//		}
	//	} else if _type == link.LINK_EVT_BYE {
	//		log.Println("bye: ", endpoint)
	//	} else {
	//		log.Println("unknown")
	//	}
	//
	//}()

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
	itemTask, e := repo_sched.GetSchedCtl().GetItemById(holder)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById e :", e)
		return nil
	}
	if itemTask.Status == api.STATUS_SCHED_END {
		log.Println("itemTask.Status == api.STATUS_SCHED_END")
		return nil
	}
	if itemTask.Enabled == api.INT_DISABLED {
		log.Println("itemTask.Enabled == api.INT_DISABLED")
		return nil
	}

	// 更新状态 并关闭 已经不再需要的定时器
	if _type == api.STATUS_SCHED_CMD_ACKED {
		log.Printf("[OnTimer]  STATUS_SCHED_CMD_ACKED ")
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_CMDACK_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_PRE_ACKED {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_PRE_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_HEARTBEAT {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_KILLED_HB_TIMEOUT,
		})
	} else if _type == api.STATUS_SCHED_END {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   0,
		})

		service_time.DisableTimer(itemTask.HbTimer)

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

	itemTask, e := repo_sched.GetSchedCtl().GetItemById(idSched)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById: ", e)
		return
	}

	if itemTask.Status == api.STATUS_SCHED_END {
		log.Println("itemTask.Status == api.STATUS_SCHED_END")
		return nil
	}
	if itemTask.Enabled == api.INT_DISABLED {
		log.Println("itemTask.Enabled == api.INT_DISABLED")
		return nil
	}

	status := api.INT_INVALId
	if body.Msg == "STATUS_SCHED_CMD_ACKED" {
		status = api.STATUS_SCHED_CMD_ACKED

		service_time.DisableTimer(itemTask.CmdackTimer)
		idPreTimer, _ := service_time.NewTimer(itemTask.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "pre_ack_timeout")

		repo_sched.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
			"pre_timer": idPreTimer,
		})
	} else if body.Msg == "STATUS_SCHED_PRE_ACKED" {
		status = api.STATUS_SCHED_PRE_ACKED

		service_time.DisableTimer(itemTask.PreTimer)
		idRunTimer, _ := service_time.NewTimer(itemTask.RunTimeout, api.STATUS_SCHED_END, idSched, "run_finish_timeout")
		idHbTimer, _ := service_time.NewTimer(config_sched.SCHED_HEARTBEAT_TIMEOUT, api.STATUS_SCHED_HEARTBEAT, idSched, "hb_timeout")

		repo_sched.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at":   time.Now().UnixMilli(),
			"prepared_at": time.Now().UnixMilli(),
			"status":      status,
			"run_timer":   idRunTimer,
			"hb_timer":    idHbTimer,
		})
	} else if body.Msg == "STATUS_SCHED_HEARTBEAT" {
		status = api.STATUS_SCHED_HEARTBEAT

		repo_sched.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"status":    status,
		})

		// reset timer
		service_time.RenewTimer(itemTask.HbTimer, config_sched.SCHED_HEARTBEAT_TIMEOUT)
	} else if body.Msg == "STATUS_SCHED_END" {
		status = api.STATUS_SCHED_END
		//idTimer, _ := service_time.NewTimer(itemTask.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "prepare_timeout")

		repo_sched.GetSchedCtl().UpdateItemById(body.Id, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"finish_at": time.Now().UnixMilli(),
			"status":    status,
			"code":      api.RESULT_SCHED_OK,
		})

		service_time.DisableTimer(itemTask.HbTimer)
		service_time.DisableTimer(itemTask.RunTimer)

		callback(idSched, 1, nil)
	} else {

	}

	return
}

func StopSched(id string) (ee error) { // todo: send stop cmd to excutors

	item, e := repo_sched.GetSchedCtl().GetItemById(id)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById, e=", e)
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

	repo_sched.GetSchedCtl().UpdateItemById(
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
	repo_sched.GetSchedCtl().CreateItem(repo_sched.Sched{
		Id:            idSched,
		Desc:          "",
		Status:        api.STATUS_SCHED_INIT,
		Endpoint:      endpoint,
		CreateAt:      time.Now().UnixMilli(),
		ActiveAt:      time.Now().UnixMilli(),
		Enabled:       1,
		CmdackTimeout: cmdackTimeoutSecond,
		PreTimeout:    preTimeoutSecond,
		RunTimeout:    runTimeoutSecond,
		Code:          api.INT_INVALId,
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
				HbInterval: config_sched.SCHED_HEARTBEAT_INTERVAL,
				PreTimeout: preTimeoutSecond,
				RunTimeout: runTimeoutSecond,
				Msg:        cmd,
			}))

	if e != nil || code != 0 {
		log.Println("link.SendDataToEndpoint failed ")

		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"status": api.STATUS_SCHED_END,
			"code":   api.RESULT_SCHED_LOCAL_FAIL,
		})
		return idSched, nil
	}

	idCmdackTimer, _ := service_time.NewTimer(cmdackTimeoutSecond, api.STATUS_SCHED_CMD_ACKED, idSched, "cmd_ack_timeout")
	repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
		"active_at":    time.Now().UnixMilli(),
		"status":       api.STATUS_SCHED_SENT,
		"cmdack_timer": idCmdackTimer,
	})

	item, e := repo_sched.GetSchedCtl().GetItemById(idSched)
	if e != nil {
		log.Println("item: ", item)
		return
	}
	log.Println("[NewTask] item: ", item)

	return idSched, nil
}
