package service_workflow

// task.go

import (
	"collab-net-v2/api"
	"collab-net-v2/sched/repo_sched"
	"collab-net-v2/sched/service_sched"
	"collab-net-v2/workflow/repo_workflow"
)

func StopTaskByBiz(idTask string) (ee error) {
	service_sched.StopSchedByTaskId(idTask)

	repo_sched.GetSchedCtl().UpdateItemById(
		idTask,
		map[string]interface{}{
			"task_enabled": api.FALSE,
		},
	)

	repo_workflow.GetTaskCtl().UpdateItemByID(idTask, map[string]interface{}{
		"enable": api.FALSE,
	})

	return nil
}
