package service_task

import (
	"collab-net-v2/link/service_link"
	"collab-net-v2/package/util/idgen"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/task/config_task"
	"collab-net-v2/task/repo_task"
	"github.com/pkg/errors"
	"log"
	"time"
)

func init() {
	// 与下层的通信 03
	repo_task.Init(config_task.RepoMySQLDsn, config_task.RepoLogLevel, config_task.RepoSlowMs)

	// 与下层的交互
	service_sched.SetCallbackFun(func(idSched string, evt int, bytes []byte) (ee error) {
		log.Println("[service_task.SetCallbackFun.anon] idSched=", idSched, " ,evt= ", evt)

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
func NewTask(name string, cmd string, linkId string, cmdackTimeoutSecond int, preTimeoutSecond int, runTimeoutSecond int) (id string, ee error) {
	idTask := idgen.GetIdWithPref("task") // "NewTask" //

	item, e := service_link.GetLinkItemFromId(linkId)
	if e != nil {
		return "", errors.Wrap(e, "service_link.GetLinkItemFromId")
	}
	if item.Online != 1 {
		return "", errors.New("item.Online != 1")
	}

	idSched, e := service_sched.NewSched(idTask, cmd, item.HostName, cmdackTimeoutSecond, preTimeoutSecond, runTimeoutSecond)
	if e != nil {
		log.Printf("service_sched.NewTask: e=", e)
		return "", e
	}
	log.Println("[NewTask] idSched=", idSched)

	repo_task.GetTaskCtl().CreateItem(repo_task.Task{
		Id:   idTask,
		Name: name,
		Cmd:  cmd,
		//Status:   api_task.TASK_STATUS_INIT,
		CreateAt: time.Now().UnixMilli(),
		//Enabled:  api_sched.INT_ENABLED,
	})

	return idTask, nil
}

func PatchTask(idTask string) (ee error) {

	service_sched.StopSched(idTask)

	return nil
}

func GetTask() (arr []repo_task.TaskInfo, ee error) {
	return repo_task.GetTaskCtl().GetTasks("")
}
