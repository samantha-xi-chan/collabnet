package service_task

// raw task only , in v1.8

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link"
	"collab-net-v2/link/service_link"
	"collab-net-v2/sched/repo_sched"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/setting/service_setting"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/repo_task"
	"collab-net-v2/util/idgen"
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/pkg/errors"
	"log"
	"time"
)

func Init() {
	mySqlDsn, e := config.GetMySqlDsn()
	if e != nil {
		log.Fatal("config.GetMySqlDsn: ", e)
	}
	log.Println("mySqlDsn", mySqlDsn)

	// 与下层的通信 03
	repo_task.Init(mySqlDsn, config_task.RepoLogLevel, config_task.RepoSlowMs)

	// 与下层的交互
	service_sched.SetCallbackFun(func(idSched string, evt int, bytes []byte) (ee error) {
		log.Println("[service_task.SetCallbackFun.anon] idSched=", idSched, " ,evt= ", evt)

		itemSched, e := repo_sched.GetSchedCtl().GetItemById(idSched)
		if e != nil {

			return errors.Wrap(e, "repo_sched.GetSchedCtl().GetItemById(: ")
		}

		callbackFunc(itemSched.TaskId, evt, nil)

		// v2.0
		go func() {
			event := api.RawTaskEvent{
				ObjType:   api.OBJ_TYPE_RAW_TASK,
				ObjID:     itemSched.TaskId,
				Timestamp: time.Now().UnixMilli(),
				Data: struct {
					Evt int `json:"evt"`
				}{
					Evt: evt,
				},
			}

			url, e := service_setting.GetSettingUrl(config.SettingCallback)
			if url != "" && e == nil {
				api.SendObjEvtRequest(url, event)
			}
		}()

		return nil
	})
}

// 对上通知
type CALLBACK func(idTask string, evt int, bytes []byte) (x error)

var callbackFunc CALLBACK

func SetTaskCallback(tmp CALLBACK) {
	callbackFunc = tmp
}

// 对上 接口
func NewTask(name string, cmd string, cmdStop string, linkId string, cmdackTimeoutSecond int, preTimeoutSecond int, runTimeoutSecond int) (id string, ee error) {
	idTask := idgen.GetIdWithPref("ta")

	item, e := service_link.GetLinkItemFromId(context.Background(), linkId)
	if e != nil {
		return "", errors.Wrap(e, "service_link.GetLinkItemFromId")
	}
	if item.Online != 1 {
		return "", errors.New("item.Online != 1")
	}

	idSched, e := service_sched.NewSched(idTask,
		link.ACTION_TYPE_NEWTASK, link.TASK_TYPE_RAW,
		cmd, linkId, cmdackTimeoutSecond, preTimeoutSecond, runTimeoutSecond, cmdStop)
	if e != nil {
		log.Printf("service_sched.NewTask: e=", e)
		return "", e
	}
	log.Println("[NewTask] idSched=", idSched)

	repo_task.GetTaskCtl().CreateItem(repo_task.Task{
		Id:      idTask,
		Name:    name,
		Cmd:     cmd,
		CmdStop: cmdStop,
		//Status:   api_task.TASK_STATUS_INIT,
		CreateAt: time.Now().UnixMilli(),
		//Enabled:  api_sched.INT_ENABLED,
	})

	return idTask, nil
}

func PatchTask(idTask string) (ee error) {
	itemTask, e := repo_task.GetTaskCtl().GetItemById(idTask)
	if e != nil {
		return errors.Wrap(e, "repo_task.GetTaskCtl().GetItemById")
	}

	service_sched.StopSchedByTaskId(idTask, itemTask.CmdStop)

	return nil
}

func GetTask() (arr []repo_task.TaskInfo, ee error) {
	return repo_task.GetTaskCtl().GetTasks()
}

func GetTaskById(id string) (arr repo_task.TaskInfo, ee error) {
	tasks, e := repo_task.GetTaskCtl().GetTaskById(id)
	if e != nil {
		return repo_task.TaskInfo{}, errors.Wrap(e, "repo_task.GetTaskCtl().GetTaskById: ")
	}

	if len(tasks) == 1 {
		return tasks[0], nil
	}

	return repo_task.TaskInfo{}, errors.Wrap(e, fmt.Sprint("repo_task.GetTaskCtl().GetTaskById: len(tasks) = ", len(tasks)))
}
