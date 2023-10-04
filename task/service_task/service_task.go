package service_task

import (
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/repo_sched"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/task/api_task"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/repo_task"
	"log"
	"time"
)

func init() {
	// 与下层的通信 03
	repo_task.Init(config_task.RepoMySQLDsn, config_task.RepoLogLevel, config_task.RepoSlowMs)

	// 与下层的交互交互
	service_sched.SetCallbackFun(func(idSched string, evt int, bytes []byte) (ee error) {
		log.Println("[service_task.SetCallbackFun.anon] idSched=", idSched, " ,evt= ", evt)

		go func() {
			if evt == 0 {
				// if success
				item, e := repo_task.GetTaskCtl().GetItemByKeyValue("id_sched", idSched)
				if e != nil {
					return
				}
				repo_task.GetTaskCtl().UpdateItemById(item.Id, map[string]interface{}{
					"status": api_task.TASK_STATUS_END,
				})

				callbackFunc(item.Id, 0, nil)
			} else {
				itemTask, e := repo_task.GetTaskCtl().GetItemByKeyValue("id_sched", idSched)
				if e != nil {
					return
				}

				itemSched, e := repo_sched.GetSchedCtl().GetItemById(itemTask.IdSched)

				idSched, e := service_sched.NewSched(itemTask.Cmd, itemSched.Endpoint, itemSched.CmdackTimeout, itemSched.PreTimeout, itemSched.RunTimeout)
				if e != nil {
					log.Printf("service_sched.NewTask: e=", e)
					return
				}

				repo_task.GetTaskCtl().UpdateItemById(itemTask.Id, map[string]interface{}{
					"id_sched": idSched,
				})
			}

			// if fail, reschedual

			// update new sched_id

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
func NewTask(cmd string, endpoint string, cmdackTimeoutSecond int, preTimeoutSecond int, runTimeoutSecond int) (id string, ee error) {
	idTask := idgen.GetIdWithPref("task")

	idSched, e := service_sched.NewSched(cmd, endpoint, cmdackTimeoutSecond, preTimeoutSecond, runTimeoutSecond)
	if e != nil {
		log.Printf("service_sched.NewTask: e=", e)
		return
	}

	repo_task.GetTaskCtl().CreateItem(repo_task.Task{
		Id:       idTask,
		Desc:     "",
		Cmd:      cmd,
		Status:   api_task.TASK_STATUS_INIT,
		CreateAt: time.Now().UnixMilli(),
		QueueAt:  0,
		IdSched:  idSched,
	})

	return idTask, nil
}

func StopTask(idTask string) (ee error) {

	return
}
