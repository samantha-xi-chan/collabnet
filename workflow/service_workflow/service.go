package service_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/link/service_link"
	"collab-net-v2/pkg/external/message"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/setting/service_setting"
	"collab-net-v2/time/service_time"
	"collab-net-v2/util/grammar"
	"collab-net-v2/util/idgen"
	"collab-net-v2/util/stringutil"
	"collab-net-v2/util/util_mq"
	"collab-net-v2/workflow/api_workflow"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/repo_workflow"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"log"
	"math/rand"
	"time"
)

func createNewWorkflowIterate(ctx context.Context, workflowId string, latestIterate int, req api_workflow.PostWorkflowDagReq) (ee error) {
	log.Printf("createNewWorkflowIterate: workflowId = %s, latestIterate = %d \n", workflowId, latestIterate)
	localTaskId := ""

	for idx, task := range req.Task {
		log.Println("idx: ", idx, ", val: ", task)
		jsonData, err := json.Marshal(task.CmdStr)
		if err != nil {
			fmt.Println("Error:", err)
		}

		if task.Concurrent == 0 { // 如果不设置 则默认为 1
			task.Concurrent = 1
		}

		for i := 0; i < task.Concurrent; i++ {
			taskId := idgen.GetIdWithPref("t")

			if idx == 0 && task.Concurrent == 1 && i == 0 { // waiting to queue
				localTaskId = taskId
			}

			repo_workflow.GetTaskCtl().CreateItem(repo_workflow.Task{
				ID:         taskId,
				Name:       task.Name,
				CreateAt:   time.Now().UnixMilli(),
				CreateBy:   0,
				WorkflowId: workflowId,
				Iterate:    latestIterate,

				Image:  task.Image,
				CmdStr: string(jsonData),

				StartAt: 0,
				EndAt:   0,
				Timeout: task.Timeout,
				//ExpExitCode: task.ExpExitCode,
				ExitCode: api.EXIT_CODE_INIT,
				Remain:   task.Remain,

				CheckExitCode:        grammar.GetCodeFromBool(task.CheckExitCode),
				ExitOnAnySiblingExit: grammar.GetCodeFromBool(task.ExitOnAnySiblingExit),

				Define: "",
				Status: api.TASK_STATUS_INIT,

				ImportObjId: task.ImportObjId,
				ImportObjAs: task.ImportObjAs,
			})
		}

	}

	for idx, edge := range req.Edge {
		log.Println("idx: ", idx, ", edge: ", edge)

		startTaskArr, e := repo_workflow.GetTaskCtl().GetItemsFromWorkflowAndNameAndIterate(workflowId, edge.Start, latestIterate)
		if e != nil {
			log.Println("GetItemsFromWorkflowAndName err: ", e)
			continue
		}

		endTaskArr := []string{api.RAMDOM_NAME_TASK_END}
		if edge.End != "" {
			endTaskArr, e = repo_workflow.GetTaskCtl().GetItemIdsFromWorkflowAndNameAndIterate(workflowId, edge.End, latestIterate)
			if e != nil {
				log.Println("GetItemIdsFromWorkflowAndName err: ", e)
				continue
			}
		}

		for i := 0; i < len(startTaskArr); i++ {
			for j := 0; j < len(endTaskArr); j++ {
				id := idgen.GetIdWithPref("e")
				repo_workflow.GetEdgeCtl().CreateItem(repo_workflow.Edge{
					ID:          id,
					CreateAt:    time.Now().UnixMilli(),
					Name:        fmt.Sprintf("%s -> %s", edge.Start, edge.End),
					StartTaskId: startTaskArr[i].ID,
					EndTaskId:   endTaskArr[j],
					Resc:        edge.Resc,
					ObjId:       startTaskArr[i].ID,
					Status:      0,

					Iterate: latestIterate,
				})
			}
		}

	}

	if localTaskId != "" {
		GetMqInstance().PostMsgToQueue(config.QUEUE_NAME, localTaskId, config.PRIORITY_4)
		message.GetMsgCtl().UpdateTaskWrapper(workflowId, api.SESSION_STATUS_INIT, fmt.Sprintf(" - Queueing TaskId: %s ", localTaskId)) // for debug only
	}

	return nil
}

func checkIfWorkflowIterateEnd(ctx context.Context, workflowId string, iterate int) (ifEnded bool, ee error) {

	items, e := repo_workflow.GetTaskCtl().GetItemsStatusNotEndByWfIdAndIterate(workflowId, iterate)
	if e != nil {
		log.Println("GetItemIdsFromWorkflowAndName err: ", e)
		return false, errors.Wrap(e, "repo_workflow.GetTaskCtl().GetItemsStatusNotEndByWfIdAndIterate")
	}
	if len(items) == 0 {
		return true, nil
	}

	return false, nil
}

func PostWorkflow(ctx context.Context, req api_workflow.PostWorkflowDagReq) (api_workflow.PostWorkflowResp, error) {
	log.Println("PostWorkflowReq: ", req)

	currentIterate := 1

	workflowId := idgen.GetIdWithPref("wf")
	jsonStr, _ := json.Marshal(req)

	record := repo_workflow.Workflow{
		ID:       workflowId,
		Name:     req.Name,
		Desc:     req.Desc,
		CreateAt: time.Now().UnixMilli(),
		StartAt:  time.Now().UnixMilli(),
		CreateBy: 0,
		Define:   string(jsonStr),

		ShareDirArrStr: stringutil.StringArrayToString(req.ShareDir, stringutil.DEFAULT_SEPARATOR), //req.ShareDir,
		Timeout:        req.Timeout,
		AutoIterate:    req.AutoIterate,
		Iterate:        currentIterate,

		ExitCode: api_workflow.ExitCodeWorkflowDefault,
	}
	ee := repo_workflow.GetWorkflowCtl().CreateItem(record)
	if ee != nil {
		return api_workflow.PostWorkflowResp{}, errors.Wrap(ee, "repo_workflow.GetWorkflowCtl().CreateItem: ")
	}

	e := createNewWorkflowIterate(ctx, workflowId, currentIterate, req)
	if e != nil {
		return api_workflow.PostWorkflowResp{}, errors.Wrap(e, "CreateNewWorkflowIterate: ")
	}

	addr := fmt.Sprintf("http://%s%s%s/%s/timer", config.ServiceName, config_workflow.LISTEN_PORT, config_workflow.UrlPathWorkflow, workflowId)
	if req.Timeout > 0 {
		_, ee := service_time.NewTimer(req.Timeout, config_workflow.EVT_TIMEOUT_WORKFLOW, workflowId, "wf_timer", addr)
		if ee != nil {
			return api_workflow.PostWorkflowResp{}, errors.Wrap(e, "service_time.NewTimer ")
		}
	}

	// check and return
	items, total, e := repo_workflow.GetTaskCtl().GetItemsByWorkflowIdV18(
		workflowId,
	)
	if e != nil {
		return api_workflow.PostWorkflowResp{}, errors.Wrap(e, "repo.GetTaskCtl().GetItemsByWorkflowId: ")
	}
	log.Println("items, total: ", items, total)
	return api_workflow.PostWorkflowResp{
		Id:           workflowId,
		QueryGetTask: items,
	}, nil
}

func StopWorkflowWrapper(ctx context.Context, workflowId string, exitCode int) (ee error) {
	go func() {
		evt := api.Event{
			ObjType:   api.OBJ_TYPE_WORKFLOW,
			ObjID:     workflowId,
			Timestamp: time.Now().UnixMilli(),
			Data: struct {
				Status   int `json:"status"`
				ExitCode int `json:"exit_code"`
			}{
				Status:   api.WORKFLOW_STATUS_END,
				ExitCode: exitCode,
			},
		}

		url, e := service_setting.GetSettingUrl(config.SettingCallback)
		if url != "" && e == nil {
			api.SendObjEvtRequest(url, evt)
		}
	}()

	if exitCode == api_workflow.ExitCodeWorkflowStoppedByBizTimeout {
		return stopWorkflow(ctx, workflowId, exitCode)
	} else if exitCode == api_workflow.ExitCodeWorkflowStoppedByDagEnd {
		service_time.DisableTimerByHolder(workflowId) // todo: add error handler
		return stopWorkflow(ctx, workflowId, exitCode)
	} else if exitCode == api_workflow.ExitCodeWorkflowStoppedByBizCmd {
		service_time.DisableTimerByHolder(workflowId) // todo: add error handler
		return stopWorkflow(ctx, workflowId, exitCode)
	} else if exitCode == api_workflow.ExitCodeWorkflowStoppedByUnknown {
		return stopWorkflow(ctx, workflowId, exitCode)
	} else {
		return errors.New("StopWorkflowWrapper exitCode not expected ")
	}
}

func stopWorkflow(ctx context.Context, workflowId string, exitCode int) (ee error) { // only 1 task supported

	itemWorkflow, ee := repo_workflow.GetWorkflowCtl().GetItemByID(workflowId)
	if ee != nil {
		return errors.Wrap(ee, "repo_workflow.GetWorkflowCtl().GetItemByID: ")
	}

	if itemWorkflow.ExitCode > 0 {
		return errors.New("itemWorkflow.ExitCode > 0 ")
	}

	e := repo_workflow.GetWorkflowCtl().UpdateItemByID(workflowId, map[string]interface{}{
		"status":    api.WORKFLOW_STATUS_END,
		"end_at":    time.Now().UnixMilli(),
		"exit_code": exitCode,
	})
	if e != nil {
		return errors.Wrap(e, "repo_workflow.GetWorkflowCtl().UpdateItemByID: ")
	}

	items, total, e := repo_workflow.GetTaskCtl().GetItemsByWorkflowIdV18(
		workflowId,
	)
	if e != nil {
		return errors.Wrap(e, "repo.GetTaskCtl().GetItemsByWorkflowId: ")
	}

	log.Println("workflowId , items, total: ", workflowId, items, total)

	for i := 0; i < int(total); i++ {
		item := items[i]
		log.Printf("task in workflow: taskId = %s, status=%d \n", item.Id, item.Status)

		if item.Status == api.TASK_STATUS_RUNNING {
			StopTaskByBiz(item.Id)

			// update status
			repo_workflow.GetTaskCtl().UpdateItemByID(item.Id, map[string]interface{}{
				"status": api.TASK_STATUS_PAUSED,
			})
		}
		/*
			if item.Status == api.TASK_STATUS_RUNNING && item.ContainerId != "" {
					nodeItem, e := repo.GetNodeCtl().GetItemByID(item.NodeId)
					if e != nil {
						return errors.Wrap(e, "repo.GetNodeCtl().GetItemByID: ")
					}

					ee := rpc.HttpStopContainer(nodeItem.Url, item.ContainerId)
					if ee != nil {
						log.Println("HttpStopContainer e:", ee)
						continue
					}

					// update status
					repo.GetTaskCtl().UpdateItemByID(item.Id, map[string]interface{}{
						"status": api.TASK_STATUS_PAUSED,
					})
		*/
	}

	return nil
}

func StopWfTaskById(ctx context.Context, taskId string) (ee error) {
	item, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		ee = errors.Wrap(e, "repo_workflow.GetTaskCtl().GetItemByID : ")
		return
	}

	log.Printf("StopWfTaskById item: %#v\n", item)
	service_sched.StopSchedByTaskId(taskId, "")

	return nil
}

func OnTaskStatusChange(ctx context.Context, taskId string, status int, exitCode int) (ee error) {
	log.Println("OnTaskStatusChange,  taskId: ", taskId, "  status: ", status, "  exitCode: ", exitCode)
	defer log.Println("OnTaskStatusChange ee= ", ee)

	message.GetMsgCtl().UpdateTaskWrapper(taskId, api.SESSION_STATUS_END, fmt.Sprintf("status: %d, exitCode: %d", status, exitCode)) // demo

	itemWorkflow, e := GetWorkflowByTaskId(taskId)
	if e != nil {
		ee = errors.Wrap(e, "GetWorkflowStatusByTaskId : ")
		return
	}

	if //itemWorkflow.Enabled == api.FALSE ||
	itemWorkflow.ExitCode != 0 && itemWorkflow.ExitCode != api_workflow.ExitCodeWorkflowStoppedByDagEnd /* 0 as default */ {
		log.Println("OnTaskStatusChange: itemWorkflow.ExitCode != 0 && itemWorkflow.ExitCode != api_workflow.ExitCodeWorkflowStoppedByDagEnd ")
		return
	}

	exitCodeWf := itemWorkflow.ExitCode
	log.Println("GetWorkflowStatusByTaskId: exitCodeWf = ", exitCodeWf)

	if exitCodeWf != api_workflow.ExitCodeWorkflowDefault {
		log.Println("workflow finished already ")
		return
	}

	item, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		ee = errors.Wrap(e, "repo.GetTaskCtl().GetItemByID : ")
		return
	}

	if status != api.TASK_STATUS_END {
		log.Println("status != api.TASK_STATUS_END, taskId", taskId)
		return nil
	}

	if itemWorkflow.AutoIterate {
		ifEnd, e := checkIfWorkflowIterateEnd(ctx, itemWorkflow.ID, itemWorkflow.Iterate)
		if e != nil {
			ee = errors.Wrap(e, "checkIfWorkflowIterateEnd: ")
			return
		}
		if ifEnd {
			var workflowDefine api_workflow.PostWorkflowDagReq
			err := json.Unmarshal([]byte(itemWorkflow.Define), &workflowDefine)
			if err != nil {
				fmt.Println("Unmarshal error:", err)
				return
			}

			itemWf, e := repo_workflow.GetWorkflowCtl().GetItemByID(itemWorkflow.ID)
			log.Printf("start itemWorkflow.ID: %s,  itemWf: %#v\n", itemWorkflow.ID, itemWf)

			e = repo_workflow.GetWorkflowCtl().UpdateItemByID(itemWorkflow.ID, map[string]interface{}{
				"iterate": itemWf.Iterate + 1,
			})
			if e != nil {
				ee = errors.Wrap(e, "repo_workflow.GetWorkflowCtl().UpdateItemByID: ")
				return
			}

			//e = repo_workflow.GetWorkflowCtl().IncreaseIterate(itemWorkflow.ID)
			//if e != nil {
			//	fmt.Println("repo_workflow.GetWorkflowCtl().IncreaseIterate:", e)
			//	return
			//}
			itemWf, e = repo_workflow.GetWorkflowCtl().GetItemByID(itemWorkflow.ID)
			log.Printf("2222 end itemWorkflow.ID: %s,  itemWf: %#v\n", itemWorkflow.ID, itemWf)
			e = createNewWorkflowIterate(ctx, itemWorkflow.ID, itemWf.Iterate, workflowDefine) // todo: optimize
			if e != nil {
				ee = errors.Wrap(e, "createNewWorkflowIterate: ")
				return
			}
		}
	}

	// find all sibling with attribute 'exit_on_any_sibling_exit' and stop then
	items, cnt, e := repo_workflow.GetTaskCtl().GetSiblingExitTasksByTaskId(taskId)
	if e != nil {
		ee = errors.Wrap(e, "repo_workflow.GetTaskCtl().GetSiblingExitTasksByTaskId : ")
		return
	}
	debugInfo := fmt.Sprintf("GetSiblingExitTasksByTaskId: taskId = %s , cnt = %d,  items = %#v", taskId, cnt, items)
	logrus.Debug(debugInfo)
	message.GetMsgCtl().UpdateTaskWrapper(item.WorkflowId, api.SESSION_DEBUG, debugInfo)
	for _, val := range items {
		StopTaskByBiz(val.ID)
	}

	if exitCode == 0 || (exitCode != 0 && item.CheckExitCode == api.FALSE) { // 任务成功 或 任务虽不成功 但是不介意
		log.Println("exitCode == 0 || (exitCode != 0 && item.CheckExitCode == api.FALSE) ,taskId =  ", taskId)
		// 找到可触发的后续0-N个任务
		items, _, e := repo_workflow.GetEdgeCtl().GetItemsByStartTaskIdAndIterate(taskId, itemWorkflow.Iterate)
		if e != nil {
			log.Println("repo_workflow.GetEdgeCtl().GetItemsByStartTaskIdAndIterate e: ", e)
		}
		log.Printf("taskId: %s, GetEndFromStart：%#v \n", taskId, items)

		for _, edge := range items {
			log.Println("OnTaskStatusChange edge: ", edge)

			if edge.EndTaskId == api.RAMDOM_NAME_TASK_END { // todo : set workflow as finished
				log.Println("current edge has no next, edge = ", edge)
				continue
			}

			size, e := repo_workflow.GetEdgeCtl().GetUnfinishedUpstremTaskId(edge.EndTaskId)
			if e != nil {
				return errors.Wrap(e, "GetUnfinishedUpstremTaskId e: ")
			}

			if size == 0 {
				item, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
				if e != nil {
					return errors.Wrap(e, "repo.GetTaskCtl().GetItemByID : ")
				}

				taskIdToEnq := edge.EndTaskId
				_, err := repo_workflow.GetRedisMgr().AcquireEnQueue(context.Background(), taskIdToEnq, func(id string) int {
					affected, e := repo_workflow.GetTaskCtl().UpdateItemEnqueue(id)
					if e != nil {
						return 0
					}
					if affected == 1 {
						GetMqInstance().PostMsgToQueue(config.QUEUE_NAME, id, config.PRIORITY_9)
						message.GetMsgCtl().UpdateTaskWrapper(item.WorkflowId, api.SESSION_STATUS_INIT, fmt.Sprintf("Queueing TaskId: %s ", id))
					} else {
						message.GetMsgCtl().UpdateTaskWrapper(item.WorkflowId, api.SESSION_STATUS_INIT, fmt.Sprintf("Queueing TaskId: %s Conflict", id))
					}
					return 0
				})
				if err != nil {
					return errors.Wrap(err, "repo_workflow.GetRedisMgr().AcquireEnQueue")
				}
			}
		}
	} else if exitCode != 0 && item.CheckExitCode == api.TRUE {
		log.Println("exitCode != 0 && item.CheckExitCode == api.TRUE ,taskId =  ", taskId)
		wfId := item.WorkflowId
		log.Println("task leaf end wfId: ", wfId)
	} else {
		log.Println("unknown ...")
	}

	return nil
}

func GetWorkflowStatusByTaskId(taskId string) (status int, x error) {
	log.Println("GetWorkflowStatusByTaskId taskId = ", taskId)
	itemTask, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		return 0, errors.Wrap(e, "repo.GetTaskCtl().GetItemByID ")
	}

	wfId := itemTask.WorkflowId
	itemWf, e := repo_workflow.GetWorkflowCtl().GetItemByID(wfId)
	if e != nil {
		return 0, errors.Wrap(e, "repo.GetWorkflowCtl().GetItemByID ")
	}

	return itemWf.ExitCode, nil
}

func GetWorkflowByTaskId(taskId string) (item repo_workflow.Workflow, x error) {
	log.Println("GetWorkflowStatusByTaskId taskId = ", taskId)
	itemTask, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		return item, errors.Wrap(e, "repo.GetTaskCtl().GetItemByID ")
	}

	wfId := itemTask.WorkflowId
	itemWf, e := repo_workflow.GetWorkflowCtl().GetItemByID(wfId)
	if e != nil {
		return item, errors.Wrap(e, "repo.GetWorkflowCtl().GetItemByID ")
	}

	return itemWf, nil
}

func PlayAsConsumerBlock(mqUrl string, consumerCnt int) {
	mq := util_mq.RabbitMQManager{}
	defer mq.Release()

	ctx := context.TODO()

	if err := mq.InitQ(mqUrl, consumerCnt, true); err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	log.Println("PlayAsConsumerBlock init ok")

	ackInChan := make([]int, consumerCnt)
	nackInChan := make([]int, consumerCnt)

	go func() { // debug
		for true {
			time.Sleep(time.Second * 60)

			ackInChanStr := ""
			for i := 0; i < mq.GetSize(); i++ {
				ackInChanStr = fmt.Sprintf("%s , %d ", ackInChanStr, ackInChan[i])
			}
			nackInChanStr := ""
			for i := 0; i < mq.GetSize(); i++ {
				nackInChanStr = fmt.Sprintf("%s , %d ", nackInChanStr, nackInChan[i])
			}

			log.Println("ackInChanStr: ", ackInChanStr)
			log.Println("nackInChanStr: ", nackInChanStr)
		}
	}()

	log.Println("mq.GetSize(): ", mq.GetSize())

	for i := 0; i < mq.GetSize(); i++ {
		log.Println("Consume ...")
		go mq.Consume(
			config.QUEUE_NAME,
			i,
			ackInChan, nackInChan,
			func(body []byte) bool {
				taskId := string(body)

				itemTask, e := repo_workflow.GetTaskCtl().GetItemByID(taskId)
				if e != nil { // todo: 若记录未找到 则处理掉， 理应作为 warning
					log.Println("taskId = ", taskId, ", 😄 e: ", e)
					return true
				}

				// for test : time.Sleep(time.Second * 30)
				log.Println("GetItemByID task :", itemTask)

				if itemTask.Status == api.TASK_STATUS_PAUSED {
					log.Println("taskId = ", taskId, ", 😄 TASK_STATUS_PAUSED_WHEN_QUEUEING: ")
					return true
				}

				// v2:
				itemLinks, e := service_link.GetFirstPartyNodeLinks(context.Background())
				if e != nil {
					log.Println("service_link.GetFirstPartyNodeLinks: ", e)
					time.Sleep(time.Second * 1)
					return false
				}

				sizeLinks := len(itemLinks)

				if sizeLinks <= 0 {
					log.Println("len(itemLink) <= 0 , taskid ", itemTask.ID)
					time.Sleep(time.Second * 1)
					return false
				}

				//itemLink := itemLinks[0] // todo: it is a test only
				// 随机调度，如果 node数量不太小 随机对业务无不良结果
				rand.Seed(time.Now().UnixNano())
				randomNumber := rand.Intn(sizeLinks)
				itemLink := itemLinks[randomNumber]
				log.Printf("sizeLinks: %d, randomNumber: %d , itemLink = %#v , itemTask = %#v  \n", sizeLinks, randomNumber, itemLink, itemTask)

				if itemTask.Status != api.TASK_STATUS_QUEUEING { // ERROR
					log.Println("itemTask.Status != api.TASK_STATUS_QUEUEING")
					return true
				}

				// 启动 登记
				repo_workflow.GetTaskCtl().UpdateItemByID(taskId, map[string]interface{}{
					"start_at": time.Now().UnixMilli(),
					"status":   api.TASK_STATUS_RUNNING,
				})

				// 准备任务的技术实现维度的参数
				var bindIn []api.Bind
				itemsIn, e := repo_workflow.GetEdgeCtl().GetItemsByEndTaskId(taskId)
				if e != nil {
					log.Println("taskId = ", taskId, ", e: ", e) // warning
					return true
				}
				log.Println("taskId = ", taskId, ", itemsIn: ", itemsIn)
				for _, val := range itemsIn {
					if val.Resc == "" {
						log.Println("skip Resc: val.ObjId = ", val.ObjId)
						continue
					}
					bindIn = append(bindIn, api.Bind{
						VolPath: val.Resc,
						VolId:   val.ObjId,
					})
				}
				log.Println("taskId = ", taskId, ", bindIn: ", bindIn)
				if itemTask.ImportObjId != "" { // todo: check invalid input
					bindIn = append(bindIn, api.Bind{
						VolPath: itemTask.ImportObjAs,
						VolId:   itemTask.ImportObjId,
					})
				}

				var bindOut []api.Bind
				itemsOut, e := repo_workflow.GetEdgeCtl().GetItemsByStartTaskId(taskId)
				if e != nil {
					log.Println("taskId = ", taskId, ", e: ", e) // warning
					return true
				}
				log.Println("taskId = ", taskId, ", itemsOut: ", itemsOut)
				for idx, val := range itemsOut {
					if idx >= 1 {
						break // feature: 当前仅仅支持一路 任务的文件夹输出
					}
					if val.Resc == "" {
						log.Println("skip Resc: val.ObjId = ", val.ObjId)
						continue
					}
					bindOut = append(bindOut, api.Bind{
						VolPath: val.Resc,
						VolId:   val.ObjId,
					})
				}
				log.Println("taskId : ", taskId, ", bindIn : ", bindIn)

				// todo: 任务信息 到 字符串传输 可以单独抽为 一个函数
				var stringArray []string
				err := json.Unmarshal([]byte(itemTask.CmdStr), &stringArray)
				if err != nil {
					fmt.Println("taskId = ", taskId, ", 😭 Error:", err)
					time.Sleep(time.Second * 1)
					return false
				}

				workflowId := itemTask.WorkflowId

				itemWorkflow, e := repo_workflow.GetWorkflowCtl().GetItemByID(workflowId)
				if e != nil {
					fmt.Println("workflowId = ", workflowId, ", 😭 Error:", e)
					return false
				}
				groupPath := fmt.Sprintf("%s/%s", config_workflow.DockerGroupPref, workflowId)

				env := []string{
					fmt.Sprintf("TASK_ID=%s", taskId), // WARNING: going to be deprecated
					fmt.Sprintf("TASK_ID_IN_WORKFLOW=%s", taskId),
					fmt.Sprintf("IMAGE_IN_WORKFLOW=%s", itemTask.Image),
				}
				newContainer := api.PostContainerReq{
					TaskId:         taskId,
					BucketName:     config_workflow.MINIO_BUCKET_NAME_INTERTASK,
					CbAddr:         "",
					LogRt:          true,
					CleanContainer: !itemTask.Remain,
					Image:          itemTask.Image,
					CmdStr:         stringArray,
					BindIn:         bindIn,
					BindOut:        bindOut,
					GroupPath:      groupPath,
					ShareDir:       stringutil.StringToStringArray(itemWorkflow.ShareDirArrStr, stringutil.DEFAULT_SEPARATOR),
					Env:            env,
				}
				log.Printf("newContainer: #%v\n", newContainer)

				jsonData, err := json.Marshal(newContainer)
				if err != nil {
					fmt.Println("JSON serialization error:", err) // todo: xxx
				}

				message.GetMsgCtl().UpdateTaskWrapper(workflowId, api.SESSION_STATUS_INIT, fmt.Sprintf(" - Starting TaskId: %s ", itemTask.ID)) // for debug only
				// item.CmdStr
				idSched, e := service_sched.NewSched(itemTask.ID,
					link.ACTION_TYPE_NEWTASK, link.TASK_TYPE_DOCKER,
					string(jsonData), itemLink.Id,
					config_sched.DEFAULT_CMDACK_TIMEOUT,
					config_sched.DEFAULT_PREACK_TIMEOUT,
					itemTask.Timeout, "")
				if e != nil {
					log.Printf("service_sched.NewTask: e=", e)
					time.Sleep(time.Second * 1)
					return false
				}
				log.Println("[NewTask] idSched=", idSched)

				itemSched, _ := service_sched.WaitSchedEnd(idSched) // 临时用轮询方案, 遗留bug sched 执行完成 不代表任务完成

				repo_workflow.GetTaskCtl().UpdateItemByID(taskId, map[string]interface{}{
					"exit_code": itemSched.BizCode,
					"end_at":    time.Now().UnixMilli(),
					"status":    api.TASK_STATUS_END,
				})

				// v2.0
				go func() {
					evt := api.Event{
						ObjType:   api.OBJ_TYPE_CONTAINER_TASK,
						ObjID:     taskId,
						Timestamp: time.Now().UnixMilli(),
						Data: struct {
							Status   int `json:"status"`
							ExitCode int `json:"exit_code"`
						}{
							Status:   api.TASK_STATUS_END,
							ExitCode: itemSched.BizCode,
						},
					}

					url, e := service_setting.GetSettingUrl(config.SettingCallback)
					if url != "" && e == nil {
						api.SendObjEvtRequest(url, evt)
					}
				}()

				ee := OnTaskStatusChange(context.Background(), taskId, api.TASK_STATUS_END, itemSched.BizCode)
				if ee != nil {
					log.Printf("OnTaskStatusChange: taskId= %s, itemSched.BizCode=%d\n", taskId, itemSched.BizCode)
					return false
				}

				message.GetMsgCtl().UpdateTaskWrapper(workflowId, api.SESSION_STATUS_INIT, fmt.Sprintf(" - End Watching TaskId: %s ", itemTask.ID)) // for debug only

				wfDagIterateEnd, e := checkIfWorkflowEnd(workflowId, itemTask.Iterate)
				if e != nil {
					log.Printf("OnTaskStatusChange: taskId= %s, itemSched.BizCode=%d\n", taskId, itemSched.BizCode)
					return false
				}
				if wfDagIterateEnd && !itemWorkflow.AutoIterate {
					StopWorkflowWrapper(ctx, workflowId, api_workflow.ExitCodeWorkflowStoppedByDagEnd)
					//service_time.DisableTimerByHolder(workflowId) // todo: add error handler
					//e := tagWorkflowEnd(workflowId, itemTask.Iterate, api_workflow.ExitCodeWorkflowStoppedByDagEnd)
					//if e != nil {
					//	log.Printf("tagWorkflowEnd: workflowId= %s, itemTask.Iterate=%d\n", workflowId, itemTask.Iterate)
					//	return false
					//}
				}

				return true
			})
	}

	log.Println("waiting select")
	select {}
}

func checkIfWorkflowEnd(wfId string, iterate int) (bool, error) {
	items, e := repo_workflow.GetTaskCtl().GetItemsStatusNotEndByWfIdAndIterate(wfId, iterate)
	if e != nil {
		return false, errors.Wrap(e, "repo_workflow.GetTaskCtl().GetItemsStatusNotEndByWfIdAndIterate: ")
	}

	if len(items) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func tagWorkflowEnd(wfId string, iterate int, exitCode int) error {
	e := repo_workflow.GetWorkflowCtl().UpdateItemByIDAndIterate(
		wfId,
		iterate,
		map[string]interface{}{
			"exit_code": exitCode,
		},
	)

	if e != nil {
		return errors.Wrap(e, "repo_workflow.GetWorkflowCtl().UpdateItemByIDAndIterate: ")
	}

	return nil
}
