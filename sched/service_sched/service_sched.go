package service_sched

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/api_sched"
	"collab-net-v2/sched/config_sched"
	repo_sched "collab-net-v2/sched/repo_sched"
	"collab-net-v2/time/service_time"
	"encoding/json"
	"github.com/pkg/errors"
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
	log.Println("service_sched [init] : ")

	mqDsn, e := config.GetMqDsn()
	if e != nil {
		log.Fatal("config.GetMqDsn() e=", e)
	}
	log.Println("mqDsn: ", mqDsn)

	// 与下层的通信 01
	link.SetConnChangeCallback(OnConnChange)
	link.SetBizDataCallback(OnBizDataFromRegisterEndpointWrapper)
	go link.NewServer()

	// 与下层的通信 02
	service_time.SetCallback(OnTimerWrapper)
	service_time.Init(mqDsn, config_sched.AMQP_EXCH)

	// 与下层的通信 03
	repo_sched.Init(config_sched.RepoMySQLDsn, config_sched.RepoLogLevel, config_sched.RepoSlowMs)
}

// 接收下层调用
func OnConnChange(endpoint string, _type int) (e error) {
	log.Println("[OnConnChange]: endpoint ", endpoint, ", _type ", _type)

	return nil
}

func OnTimerWrapper(idTimer string, _type int, holder string, bytes []byte) (ee error) {
	log.Printf("[OnTimerWrapper] holder %s,  idTimer: %s, bytes: %s", holder, idTimer, string(bytes))
	go func() {
		OnTimer(idTimer, _type, holder, bytes)
	}()
	return nil
}

func OnTimer(idTimer string, evtType int, holder string, bytes []byte) (ee error) { // 定时器 事件
	log.Printf("[OnTimer]holder %s,  idTimer: %s, evtType: %d, bytes: %s\n", holder, idTimer, evtType, string(bytes))
	itemTask, e := repo_sched.GetSchedCtl().GetItemById(holder)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById e :", e)
		return nil
	}
	if itemTask.FwkCode == api_sched.SCHED_FWK_CODE_END {
		log.Println("itemTask.FwkCode == api_sched.SCHED_FWK_CODE_END")
		return nil
	}
	if itemTask.TaskEnabled == api_sched.INT_DISABLED { // 业务角度抛弃
		log.Println("itemTask.TaskEnabled == api_sched.INT_DISABLED")
		return nil
	}
	if itemTask.Enabled == api_sched.INT_DISABLED { // 技术角度抛弃
		log.Println("itemTask.Enabled == api_sched.INT_DISABLED")
		return nil
	}

	// 更新状态 并关闭 已经不再需要的定时器
	if evtType == api.SCHED_EVT_TIMEOUT_CMDACK {
		log.Printf("[OnTimer]  STATUS_SCHED_CMD_ACKED ")
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"fwk_code": api_sched.SCHED_FWK_CODE_END,
		})
	} else if evtType == api.SCHED_EVT_TIMEOUT_PREACK {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"fwk_code": api_sched.SCHED_FWK_CODE_END,
		})
	} else if evtType == api.SCHED_EVT_TIMEOUT_HB {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"fwk_code": api_sched.SCHED_FWK_CODE_END,
		})
	} else if evtType == api.SCHED_EVT_TIMEOUT_RUN {
		repo_sched.GetSchedCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
			"fwk_code": api_sched.SCHED_FWK_CODE_END,
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
	idSched := body.SchedId

	itemTask, e := repo_sched.GetSchedCtl().GetItemById(idSched)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById: ", e)
		return
	}

	if itemTask.FwkCode == api_sched.SCHED_FWK_CODE_END {
		log.Println("iitemTask.FwkCode == api_sched.SCHED_FWK_CODE_END")
		return nil
	}
	if itemTask.TaskEnabled == api_sched.INT_DISABLED { // 业务角度抛弃
		log.Println("itemTask.TaskEnabled == api_sched.INT_DISABLED")
		return nil
	}
	if itemTask.Enabled == api_sched.INT_DISABLED { // 技术角度抛弃
		log.Println("itemTask.Enabled == api_sched.INT_DISABLED")
		return nil
	}

	// api.TASK_EVT_PREACK , api.SCHED_EVT_TIMEOUT_PREACK
	if body.Para01 == api.TASK_EVT_CMDACK {
		service_time.DisableTimer(itemTask.CmdackTimer)
		idPreTimer, e := service_time.NewTimer(itemTask.PreTimeout, api.SCHED_EVT_TIMEOUT_PREACK, idSched, "pre_ack_timeout")
		if e != nil {
			return errors.Wrap(e, "service_time.NewTimer: ")
		}

		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"cmdack_at": time.Now().UnixMilli(),
			"best_prog": api.STATUS_SCHED_CMD_ACKED,
			"pre_timer": idPreTimer,
		})

		callback(idSched, body.Para01, nil)
	} else if body.Para01 == api.TASK_EVT_PREACK {
		service_time.DisableTimer(itemTask.PreTimer)
		idRunTimer, _ := service_time.NewTimer(itemTask.RunTimeout, api.SCHED_EVT_TIMEOUT_RUN, idSched, "run_finish_timeout")
		idHbTimer, _ := service_time.NewTimer(config_sched.SCHED_HEARTBEAT_TIMEOUT, api.SCHED_EVT_TIMEOUT_HB, idSched, "hb_timeout")

		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"active_at":   time.Now().UnixMilli(),
			"prepared_at": time.Now().UnixMilli(),
			"best_prog":   api.STATUS_SCHED_PRE_ACKED,
			"run_timer":   idRunTimer,
			"hb_timer":    idHbTimer,
		})

		callback(idSched, body.Para01, nil)
	} else if body.Para01 == api.TASK_EVT_HEARTBEAT {
		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"best_prog": api.STATUS_SCHED_RUNNING,
		})

		// reset timer
		service_time.RenewTimer(itemTask.HbTimer, config_sched.SCHED_HEARTBEAT_TIMEOUT)

		callback(idSched, body.Para01, nil)
	} else if body.Para01 == api.TASK_EVT_END {
		//idTimer, _ := service_time.NewTimer(itemTask.PreTimeout, api.STATUS_SCHED_PRE_ACKED, idSched, "prepare_timeout")

		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"active_at": time.Now().UnixMilli(),
			"finish_at": time.Now().UnixMilli(),
			"best_prog": api.STATUS_SCHED_RUN_END,
			"fwk_code":  api_sched.SCHED_FWK_CODE_END,
			"biz_code":  body.Para0101,
		})

		service_time.DisableTimer(itemTask.HbTimer)
		service_time.DisableTimer(itemTask.RunTimer)

		callback(idSched, api.TASK_EVT_END, nil)
	} else {
		log.Println("OnBizDataFromRegisterEndpoint unknown else")
	}

	return
}

func StopSched(taskId string) (ee error) { // todo: send stop cmd to excutors
	var arr []repo_sched.QueryKeyValue
	arr = append(arr, repo_sched.QueryKeyValue{
		"task_id",
		taskId,
	})
	arr = append(arr, repo_sched.QueryKeyValue{
		"enabled",
		api_sched.INT_ENABLED,
	})
	item, e := repo_sched.GetSchedCtl().GetItemByKeyValueArr(arr)
	if e != nil {
		log.Println("repo_sched.GetSchedCtl().GetItemById, e=", e)
		return
	}

	log.Println("[StopSched]  GetItemByKeyValueArr", item)

	// todo: 发送消息到 node 节点
	code, e := link.SendDataToLinkId(
		item.LinkId,
		link.GetPackageBytes(
			time.Now().UnixMilli(),
			"v1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_STOPTASK,
				SchedId: item.Id,
			}))
	if e != nil {
		// 记录关键错误
		log.Println("StopSched e ", e)
	}
	log.Println("StopSched code", code)

	if item.FwkCode == api_sched.SCHED_FWK_CODE_END {
		log.Println("item.FwkCode == api_sched.SCHED_FWK_CODE_END")
		return
	}

	// todo: add timer ...

	repo_sched.GetSchedCtl().UpdateItemById(
		item.Id,
		map[string]interface{}{
			"task_enabled": api_sched.INT_DISABLED,
		},
	)

	callback(item.Id, api.TASK_EVT_STOPPED, nil)

	return
}

func NewSched(taskId string, cmd string, linkId string, cmdackTimeoutSecond int, preTimeoutSecond int, runTimeoutSecond int) (_id string, e error) {
	idSched := idgen.GetIdWithPref("sched")
	repo_sched.GetSchedCtl().CreateItem(repo_sched.Sched{
		Id:            idSched,
		TaskId:        taskId,
		TaskEnabled:   api_sched.INT_ENABLED,
		BestProg:      api.STATUS_SCHED_INIT,
		LinkId:        linkId,
		CreateAt:      time.Now().UnixMilli(),
		ActiveAt:      time.Now().UnixMilli(),
		Enabled:       api_sched.INT_ENABLED,
		CmdackTimeout: cmdackTimeoutSecond,
		PreTimeout:    preTimeoutSecond,
		HbTimeout:     config_sched.SCHED_HEARTBEAT_TIMEOUT,
		RunTimeout:    runTimeoutSecond,
		BizCode:       api_sched.INT_INVALID,
		FwkCode:       api_sched.INT_INVALID,
	})

	code, e := link.SendDataToLinkId(
		linkId,
		link.GetPackageBytes(
			time.Now().UnixMilli(),
			"v1.0",
			link.PACKAGE_TYPE_BIZ,
			link.BizData{
				TypeId:  link.BIZ_TYPE_NEWTASK,
				SchedId: idSched,
				Para01:  config_sched.SCHED_HEARTBEAT_INTERVAL,
				Para02:  preTimeoutSecond,
				Para03:  runTimeoutSecond,
				Para11:  cmd,
			}))
	if e != nil || code != 0 {
		log.Println("link.SendDataToEndpoint failed ")

		repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
			"best_prog": api.STATUS_SCHED_ANALYZE,
			"fwk_code":  api_sched.SCHED_FWK_CODE_END,
		})
		callback(taskId, api.TASK_EVT_REJECT, nil)
		return idSched, nil
	}

	idCmdackTimer, _ := service_time.NewTimer(cmdackTimeoutSecond, api.SCHED_EVT_TIMEOUT_CMDACK, idSched, "cmd_ack_timeout")
	repo_sched.GetSchedCtl().UpdateItemById(idSched, map[string]interface{}{
		"active_at":    time.Now().UnixMilli(),
		"best_prog":    api.STATUS_SCHED_SENT,
		"cmdack_timer": idCmdackTimer,
	})

	item, e := repo_sched.GetSchedCtl().GetItemById(idSched)
	if e != nil {
		log.Println("item: ", item)
		return
	}
	log.Println("[NewTask] item: ", item)

	callback(taskId, api.TASK_EVT_ACCEPT, nil)
	return idSched, nil
}

func GetSched(idSched string) (repo_sched.Sched, error) {
	item, e := repo_sched.GetSchedCtl().GetItemById(idSched)
	if e != nil {
		return repo_sched.Sched{}, errors.Wrap(e, "repo_sched.GetSchedCtl().GetItemById: ")
	}

	return item, nil
}
