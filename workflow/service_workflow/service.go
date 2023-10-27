package service_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/link/service_link"
	"collab-net-v2/package/grammar"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/package/util/util_mq"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/workflow/api_workflow"
	"collab-net-v2/workflow/config_workflow"
	repo "collab-net-v2/workflow/repo_workflow"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"time"
)

func PostWorkflow(ctx context.Context, req api_workflow.PostWorkflowReq) (api_workflow.PostWorkflowResp, error) {
	log.Println("PostWorkflowReq: ", req)
	localTaskId := ""

	workflowId := idgen.GetIdWithPref("wf")
	jsonStr, _ := json.Marshal(req)
	repo.GetWorkflowCtl().CreateItem(repo.Workflow{
		ID:       workflowId,
		Name:     workflowId,
		Desc:     "",
		CreateAt: time.Now().UnixMilli(),
		CreateBy: 0,
		Define:   string(jsonStr),
	})

	for idx, task := range req.Task {
		log.Println("idx: ", idx, ", val: ", task)
		jsonData, err := json.Marshal(task.CmdStr)
		if err != nil {
			fmt.Println("Error:", err)
		}

		taskId := idgen.GetIdWithPref("task_")
		repo.GetTaskCtl().CreateItem(repo.Task{
			ID:         taskId,
			Name:       task.Name,
			CreateAt:   time.Now().UnixMilli(),
			CreateBy:   0,
			WorkflowId: workflowId,

			Image:  task.Image,
			CmdStr: string(jsonData),

			StartAt:       0,
			EndAt:         0,
			Timeout:       task.Timeout,
			ExpExitCode:   task.ExpExitCode,
			ExitCode:      api.EXIT_CODE_INIT,
			CheckExitCode: grammar.GetCodeFromBool(task.CheckExitCode),
			Define:        "",
			Status:        api.TASK_STATUS_INIT,
			NodeId:        "",

			ImportObjId: task.ImportObjId,
			ImportObjAs: task.ImportObjAs,
		})

		if idx == 0 {
			localTaskId = taskId
		}
	}

	for idx, edge := range req.Edge {
		id := idgen.GetIdWithPref("edge")
		log.Println("idx: ", idx, ", edge: ", edge)

		startTask, e := repo.GetTaskCtl().GetItemFromWorkflowAndName(workflowId, edge.Start)
		if e != nil {
			log.Println("GetItemFromWorkflowAndName err: ", e)
			continue
		}

		var endTaskId string
		if edge.End != "" {
			endTask, e := repo.GetTaskCtl().GetItemFromWorkflowAndName(workflowId, edge.End)
			if e != nil {
				log.Println("GetItemFromWorkflowAndName err: ", e)
				continue
			}
			endTaskId = endTask.ID
		} else {
			endTaskId = api.RAMDOM_NAME_TASK_END
		}

		repo.GetEdgeCtl().CreateItem(repo.Edge{
			ID:          id,
			CreateAt:    time.Now().UnixMilli(),
			Name:        fmt.Sprintf("%s -> %s", edge.Start, edge.End),
			StartTaskId: startTask.ID,
			EndTaskId:   endTaskId,
			Resc:        edge.Resc,
			ObjId:       startTask.ID,
			Status:      0,
		})
	}

	if localTaskId != "" {
		GetMqInstance().PostMsg(config.QUEUE_NAME, localTaskId, config.PRIORITY_4)
		// v2 scope: message.GetMsgCtl().UpdateTaskWrapper(workflowId, api.SESSION_STATUS_INIT, fmt.Sprintf(" - Queueing TaskId: %s ", localTaskId)) // todo: demo
	}

	items, total, e := repo.GetTaskCtl().GetItemsByWorkflowId(
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

func StopWorkflow(ctx context.Context, workflowId string) (ee error) { // only 1 task supported

	items, total, e := repo.GetTaskCtl().GetItemsByWorkflowId(
		workflowId,
	)
	if e != nil {
		return errors.Wrap(e, "repo.GetTaskCtl().GetItemsByWorkflowId: ")
	}

	log.Println("workflowId , items, total: ", workflowId, items, total)

	for i := 0; i < int(total); i++ {
		item := items[i]
		log.Printf("task in workflow: taskId = %s, ContainerId = %s, status=%d \n", item.Id, item.ContainerId, item.Status)

		if item.Status == api.TASK_STATUS_RUNNING && item.ContainerId != "" { // item.NodeId != "" &&
			/*
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
	}

	return nil
}

func OnTaskStatusChange(ctx context.Context, taskId string, status int, exitCode int) (ee error) {
	log.Println("OnTaskStatusChange,  taskId: ", taskId, "  status: ", status, "  exitCode: ", exitCode)
	defer log.Println("OnTaskStatusChange ee= ", ee)

	// v2 scope: message.GetMsgCtl().UpdateTaskWrapper(taskId, api.SESSION_STATUS_END, fmt.Sprintf("status: %d, exitCode: %d", status, exitCode)) // demo

	statusWf, e := GetWorkflowStatusByTaskId(taskId)
	if e != nil {
		ee = errors.Wrap(e, "GetWorkflowStatusByTaskId : ")
		return
	}
	log.Println("GetWorkflowStatusByTaskId: status = ", status)

	if statusWf == api.TASK_STATUS_END {
		log.Println("workflow finished already ")
		return
	}

	item, e := repo.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		ee = errors.Wrap(e, "repo.GetTaskCtl().GetItemByID : ")
		return
	}

	if status != api.TASK_STATUS_END {
		log.Println("status != api.TASK_STATUS_END, taskId", taskId)
		return nil
	}

	if exitCode == 0 || (exitCode != 0 && item.CheckExitCode == api.FALSE) { // 任务成功 或 任务虽不成功 但是不介意
		log.Println("exitCode == 0 || (exitCode != 0 && item.CheckExitCode == api.FALSE) ,taskId =  ", taskId)
		// 找到可触发的后续0-N个任务
		items, e := repo.GetEdgeCtl().GetItemsByStartTaskId(taskId)
		if e != nil {
			log.Println("GetEndFromStart e: ", e)
		}
		log.Println("GetEndFromStart： ", items)

		for _, edge := range items {
			log.Println("OnTaskStatusChange edge: ", edge)

			if edge.EndTaskId == api.RAMDOM_NAME_TASK_END { // todo : set workflow as finished
				log.Println("current edge has no next, edge = ", edge)
				continue
			}

			size, e := repo.GetEdgeCtl().GetUnfinishedUpstremTaskId(edge.EndTaskId)
			if e != nil {
				log.Println("GetUnfinishedUpstremTaskId e: ", e)
			}

			if size == 0 {
				GetMqInstance().PostMsg(config.QUEUE_NAME, edge.EndTaskId, config.PRIORITY_9)
				// v2 scope: message.GetMsgCtl().UpdateTaskWrapper(item.WorkflowId, api.SESSION_STATUS_INIT, fmt.Sprintf("Queueing TaskId: %s ", edge.EndTaskId)) // demo
			}
		}
	} else if exitCode != 0 && item.CheckExitCode == api.TRUE {
		log.Println("exitCode != 0 && item.CheckExitCode == api.TRUE ,taskId =  ", taskId)
		wfId := item.WorkflowId
		log.Println("wfId: ", wfId)
		repo.GetWorkflowCtl().UpdateItemByID(wfId, map[string]interface{}{
			"status": api.TASK_STATUS_END,
		})
	} else {
		log.Println("unknown ...")
	}

	return nil
}

func GetWorkflowStatusByTaskId(taskId string) (status int, x error) {
	log.Println("GetWorkflowStatusByTaskId taskId = ", taskId)
	itemTask, e := repo.GetTaskCtl().GetItemByID(taskId)
	if e != nil {
		return 0, errors.Wrap(e, "repo.GetTaskCtl().GetItemByID ")
	}

	wfId := itemTask.WorkflowId
	itemWf, e := repo.GetWorkflowCtl().GetItemByID(wfId)
	if e != nil {
		return 0, errors.Wrap(e, "repo.GetWorkflowCtl().GetItemByID ")
	}

	return itemWf.Status, nil
}

func RecordTaskAndWorkflowFinish(taskId string, wfId string, errStr string) (x error) {
	log.Println("RecordTaskAndWorkflowFinish ")
	e := repo.GetTaskCtl().UpdateItemByID(taskId, map[string]interface{}{
		"status":    api.TASK_STATUS_END,
		"exit_code": api.EXIT_CODE_UNKNOWN,
		"error":     errStr,
	})
	if e != nil {
		return errors.Wrap(e, "repo.GetTaskCtl().UpdateItemByID ")
	}

	if wfId == "" {
		itemTask, e := repo.GetTaskCtl().GetItemByID(taskId)
		if e != nil {
			return errors.Wrap(e, "repo.GetTaskCtl().GetItemByID ")
		}

		wfId = itemTask.WorkflowId
	}

	if wfId == "" {
		return errors.New("wfId empty")
	}

	errWf := fmt.Sprintf("error: taskid=%s, errStr: %s", taskId, errStr)
	repo.GetWorkflowCtl().UpdateItemByID(wfId, map[string]interface{}{
		"status": api.TASK_STATUS_END,
		"error":  errWf,
	})

	return nil
}

func PlayAsConsumerBlock(mqUrl string, consumerCnt int) {
	mq := util_mq.RabbitMQManager{}
	defer mq.Release()

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

				itemTask, e := repo.GetTaskCtl().GetItemByID(taskId)
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

				// 启动 登记
				repo.GetTaskCtl().UpdateItemByID(taskId, map[string]interface{}{
					"start_at": time.Now().UnixMilli(),
				})

				// 准备任务的技术实现维度的参数
				var bindIn []api.Bind
				itemsIn, e := repo.GetEdgeCtl().GetItemsByEndTaskId(taskId)
				if e != nil {
					log.Println("taskId = ", taskId, ", e: ", e) // warning
					return true
				}
				log.Println("taskId = ", taskId, ", itemsIn: ", itemsIn)
				for _, val := range itemsIn {
					bindIn = append(bindIn, api.Bind{
						VolPath: val.Resc,
						VolId:   val.ObjId,
					})
				}
				log.Println("taskId = ", taskId, ", bindIn: ", bindIn)
				if itemTask.ImportObjId != "" {
					bindIn = append(bindIn, api.Bind{
						VolPath: itemTask.ImportObjAs,
						VolId:   itemTask.ImportObjId,
					})
				}

				var bindOut []api.Bind
				itemsOut, e := repo.GetEdgeCtl().GetItemsByStartTaskId(taskId)
				if e != nil {
					log.Println("taskId = ", taskId, ", e: ", e) // warning
					return true
				}
				log.Println("taskId = ", taskId, ", itemsOut: ", itemsOut)
				for idx, val := range itemsOut {
					if idx >= 1 {
						break // feature: 当前仅仅支持一路 任务的文件夹输出
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
				newContainer := api.PostContainerReq{
					TaskId:         taskId,
					BucketName:     config_workflow.BUCKET_NAME,
					CbAddr:         "",
					LogRt:          true,
					CleanContainer: false,
					Name:           "cont_" + taskId,
					Image:          itemTask.Image,
					CmdStr:         stringArray,
					BindIn:         bindIn,
					BindOut:        bindOut,
				}
				log.Println("newContainer: ", newContainer)

				// 将结构体序列化为JSON
				jsonData, err := json.Marshal(newContainer)
				if err != nil {
					fmt.Println("JSON serialization error:", err) // todo: xxx
				}

				// item.CmdStr
				idSched, e := service_sched.NewSched(itemTask.ID, link.BIZ_TYPE_NEW_DOCKER_TASK, string(jsonData), itemLink.Id,
					config_sched.DEFAULT_CMDACK_TIMEOUT,
					config_sched.DEFAULT_PREACK_TIMEOUT,
					itemTask.Timeout)
				if e != nil {
					log.Printf("service_sched.NewTask: e=", e)
					time.Sleep(time.Second * 1)
					return false
				}
				log.Println("[NewTask] idSched=", idSched)

				itemSched, _ := service_sched.WaitSchedEnd(idSched) // 临时用轮询方案, 遗留bug sched 执行完成 不代表任务完成

				repo.GetTaskCtl().UpdateItemByID(taskId, map[string]interface{}{
					"exit_code": itemSched.BizCode,
					"end_at":    time.Now().UnixMilli(),
					"status":    api.TASK_STATUS_END,
				})

				OnTaskStatusChange(context.Background(), taskId, api.TASK_STATUS_END, itemSched.BizCode)

				return true
			})
	}

	log.Println("waiting select")
	select {}
}
