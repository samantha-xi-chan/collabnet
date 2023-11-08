package repo_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/workflow/api_workflow"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strings"
)

type TaskCtl struct{}

var taskCtl TaskCtl

func GetTaskCtl() *TaskCtl {
	return &taskCtl
}

func (Task) TableName() string {
	return "compute_task"
}

type Task struct {
	ID         string `json:"id" gorm:"primaryKey"`
	Name       string `json:"name"`
	CreateAt   int64  `json:"create_at"`
	CreateBy   int64  `json:"create_by"`
	WorkflowId string `json:"workflow_id" gorm:"index:idx_workflow_id"  `

	Image  string `json:"image"`
	CmdStr string `json:"cmd_str"`

	StartAt     int64 `json:"start_at"`
	EndAt       int64 `json:"end_at"`
	Timeout     int   `json:"timeout"` // Second
	ExpExitCode int   `json:"exp_exit_code"`
	ExitCode    int   `json:"exit_code" gorm:"default:1000"` // EXIT_CODE_INIT    = 1000

	CheckExitCode int `json:"check_exit_code"` /* 1 as true, 0 as false */

	Define string `json:"define" `
	Status int    `json:"status" gorm:"index:idx_status"   gorm:"default:60021001"` // TASK_STATUS_INIT     = 60021001
	Error  string `json:"error" `
	//NodeId string `json:"node_id" `
	//ContainerId string `json:"container_id"`

	ImportObjId string `json:"import_obj_id"`
	ImportObjAs string `json:"import_obj_as"`
}

func (ctl *TaskCtl) CreateItem(item Task) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "TaskCtl.CreateItem: ")
	}

	return nil
}

func (ctl *TaskCtl) DeleteItemByID(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Task{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "TaskCtl.DeleteItemByID: ")
	}

	return nil
}

func (ctl *TaskCtl) GetItemByContainerId(containerId string) (i Task, e error) { // todo: optimize
	var item Task
	err := db.Where("container_id = ?", containerId).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, err
	} else if err != nil {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

func (ctl *TaskCtl) GetItemByID(id string) (i Task, e error) { // todo: optimize
	var item Task

	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByID ErrRecordNotFound: ")
	} else if err != nil {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByID err not nil: ")
	}

	return item, nil
}

func (ctl *TaskCtl) GetItemFromWorkflowAndName(wfId string, name string) (i Task, e error) {
	var item Task
	err := db.Where("workflow_id = ? AND name = ?", wfId, name).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByID ErrRecordNotFound: ")
	} else if err != nil {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByID err not nil: ")
	}

	return item, nil
}

func (ctl *TaskCtl) GetWorkflowObjIdsFromTaskId(taskID string) (_ []api_workflow.VolItem, e error) { // todo: optimize
	var result []api_workflow.VolItem

	sql := "SELECT ct.obj_id_out, cn.url \nFROM compute_task ct\nJOIN compute_node cn ON ct.node_id = cn.id\nWHERE ct.workflow_id IN (\n    SELECT workflow_id\n    FROM compute_task\n    WHERE id = ?)"

	// SELECT ct.obj_id_out , ct.node_id, cn.url  FROM biz.compute_task ct LEFT JOIN compute_node cn ON ct.node_id = cn.id
	if err := db.Raw(sql, taskID).Scan(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (ctl *TaskCtl) UpdateItemByID(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Task{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemByID e: ", e)
	}

	return nil
}

func (ctl *TaskCtl) UpdateItemEnqueue(id string) (rowsAffected int64, e error) {
	updateQuery := "UPDATE compute_task SET status = ? WHERE id = ? AND status = ? "
	result := db.Exec(updateQuery, api.TASK_STATUS_QUEUEING, id, api.TASK_STATUS_INIT)

	if result.Error != nil {
		return 0, errors.Wrap(result.Error, "db.Exec: ")
	}

	rowsAffected = result.RowsAffected
	//log.Printf("影响的行数：%d\n", rowsAffected)
	if rowsAffected == 0 {
		log.Println("rarely happen: UpdateItemEnqueue, id = ", id)
	}

	return
}

func (ctl *TaskCtl) GetNextTasksByTaskId(taskId string) (items []Task, total int64, e error) {
	//step := db.Model(&Task{})
	step := db.Where(" TRUE ")
	//log.Println("taskId: ", taskId)

	step = step.Where("task_id_pre = ?", taskId).Find(&items)
	step.Count(&total)

	//step.Limit(-1).
	//	Offset(-1).
	//	Count(&total)

	return items, total, nil
}

func (ctl *TaskCtl) GetItemsByWorkflowIdDeprecated(wfId string) (x []api_workflow.TaskResp, total int64, e error) {

	var tasks []api_workflow.TaskResp
	db.Table("compute_task").
		Select("DISTINCT compute_task.id, compute_task.name, compute_task.start_at, compute_task.end_at, compute_task.status, compute_task.node_id, compute_task.container_id, compute_task.exit_code, ce.obj_id").
		Joins("JOIN compute_edge AS ce ON compute_task.id = ce.start_task_id").
		Where("compute_task.workflow_id = ?", wfId).
		Find(&tasks).Limit(-1).
		Offset(-1).
		Count(&total)
	/*
			sql := `
		        SELECT DISTINCT (ct_sub.id), ct_sub.name, ct_sub.start_at, ct_sub.end_at, ct_sub.status, ct_sub.node_id, ct_sub.check_exit_code, ce.obj_id
		        FROM (SELECT * FROM compute_task WHERE workflow_id = ?) AS ct_sub
		        JOIN compute_edge AS ce
		        ON ct_sub.id = ce.start_task_id
		    `
			err = db.Raw(sql, wfId).Scan(&tasks).Error
	*/

	/*
	   SELECT DISTINCT (ct_sub.id), ct_sub.name, ct_sub.start_at, ct_sub.end_at, ct_sub.status, ct_sub.check_exit_code, ce.obj_id
	   FROM (SELECT * FROM compute_task WHERE workflow_id = 'wf_1698824209416jbie') AS ct_sub
	   JOIN compute_edge AS ce
	   ON ct_sub.id = ce.start_task_id

	*/

	return tasks, total, nil
}

func (ctl *TaskCtl) GetItemsByWorkflowIdV18(wfId string) (x []api_workflow.TaskResp, total int64, e error) { // todo: ORDER BY create_at
	var tasks []api_workflow.TaskResp

	db.Table("(SELECT DISTINCT ct.id, ct.name, ct.create_at,ct.start_at, ct.end_at, ct.status, ct.exit_code, ce.obj_id FROM compute_task AS ct JOIN compute_edge AS ce ON ct.id = ce.start_task_id WHERE ct.workflow_id = ?) AS task_sub", wfId).
		Select("task_sub.*, link.host_name, link.from, sched.carrier, sched.reason, sched.error ").
		Joins("LEFT JOIN sched ON sched.task_id = task_sub.id").
		Joins("LEFT JOIN link ON link.id = sched.link_id").
		Order("task_sub.create_at ASC").
		Scan(&tasks).Limit(-1).
		Offset(-1).
		Count(&total)

	/*
			SELECT task_sub.* ,link.host_name, sched.carrier
		    FROM ( SELECT DISTINCT ct.id, ct.name, ct.start_at, ct.end_at, ct.status, ct.exit_code, ce.obj_id
			    FROM `compute_task` AS ct JOIN compute_edge AS ce
			    ON ct.id = ce.start_task_id
			    WHERE ct.workflow_id = 'wf_1699007538633eeyr') AS task_sub
			LEFT JOIN sched
			ON sched.task_id  = task_sub.id
			LEFT JOIN link
			ON link.id = sched.link_id
	*/

	/*
		db.Table("compute_task").
			Select("DISTINCT compute_task.id, compute_task.name, compute_task.start_at, compute_task.end_at, compute_task.status, link.host_name , sched.carrier , compute_task.exit_code, ce.obj_id").
			Joins("JOIN compute_edge AS ce JOIN sched JOIN link  ON compute_task.id = ce.start_task_id  AND sched.task_id = compute_task.id AND sched.link_id = link.id ").
			Where("compute_task.workflow_id = ?", wfId).
			Find(&tasks).Limit(-1).
			Offset(-1).
			Count(&total)

		SELECT DISTINCT ct.id, ct.name, ct.start_at, ct.end_at, ct.status, link.host_name , sched.carrier , ct.exit_code, ce.obj_id
		FROM `compute_task` AS ct JOIN compute_edge AS ce JOIN sched JOIN link
		ON ct.id = ce.start_task_id  AND sched.task_id = ct.id AND sched.link_id = link.id
		WHERE ct.workflow_id = 'wf_1698825735413xzkd'
	*/

	return tasks, total, nil
}

func (ctl *TaskCtl) GetItemsBySearch(
	hasWorkflowId bool, workflowId string,
	hasSortBy bool, sortBy string,
	hasPageID bool, pageID int,
	hasPageSize bool, pageSize int,
) (items []Task, total int64, e error) {
	step := db.Where(" TRUE ")

	//
	if hasWorkflowId {
		step = step.Where("workflow_id = ?", workflowId)
	}

	if hasSortBy {
		flagASC := "ASC"
		if strings.HasPrefix(sortBy, "-") {
			flagASC = "DESC"
			sortBy = sortBy[1:]
		}
		log.Println("sortBy:", sortBy, ", flagASC:", flagASC)
		step = step.Order(fmt.Sprintf("%s %s", sortBy, flagASC))
	}

	if hasPageID && hasPageSize {
		step = step.Offset((pageID - 1) * pageSize).Limit(pageSize)
	}

	step.Find(&items)

	step.Limit(-1).
		Offset(-1).
		Count(&total)

	return items, total, nil
}

func (ctl *TaskCtl) GetMaxLeftQuotaNodeId() (nodeId string, leftQuota int, e error) {
	var result struct {
		NodeID      string
		Quota       int
		RunningTask int
		LeftQuota   int
	}

	subQuery := db.Table("compute_task").
		Select("node_id, COUNT(*) AS running_task_count").
		//Where("exit_code = ?", api.TASK_ERROR_CODE_INIT).
		Where("status = ?", api.TASK_STATUS_RUNNING).
		Group("node_id")

	db.Table("compute_node cn").
		Select("cn.id AS node_id, cn.task_quota AS quota, COALESCE(rt.running_task_count, 0) AS running_task, (cn.task_quota - COALESCE(rt.running_task_count, 0)) AS left_quota").
		Joins("LEFT JOIN (?) AS rt ON cn.id = rt.node_id", subQuery).
		Order("left_quota DESC").
		Limit(1).
		Scan(&result)

	log.Printf("Node ID: %s, Quota: %d, Running Task: %d, Left Quota: %d\n", result.NodeID, result.Quota, result.RunningTask, result.LeftQuota)

	return result.NodeID, result.LeftQuota, nil
}

var subqueryResult struct {
	ID uint
}

func (ctl *TaskCtl) UpdatePreTaskIdFromWorkflowAndName(workflowId string, preName string, nextName string) (nodeId string, e error) {
	db.Raw(`
        UPDATE compute_task AS ct1
        JOIN (
            SELECT id
            FROM compute_task
            WHERE workflow_id = ? AND name = ?
            LIMIT 1
        ) AS subquery
        ON ct1.workflow_id = ? AND ct1.name = ?
        SET ct1.task_id_pre = subquery.id
    `, workflowId, preName, workflowId, nextName).Scan(&subqueryResult)

	log.Println(subqueryResult)

	return "", nil
}
