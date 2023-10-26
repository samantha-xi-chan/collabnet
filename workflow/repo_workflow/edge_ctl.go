package repo_workflow

import (
	"collab-net-v2/api"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strings"
)

type EdgeCtl struct{}

var edgeCtl EdgeCtl

func GetEdgeCtl() *EdgeCtl {
	return &edgeCtl
}

func (Edge) TableName() string {
	return "compute_edge"
}

type Edge struct {
	ID          string `json:"id"  gorm:"primaryKey"`
	CreateAt    int64  `json:"create_at"`
	Name        string `json:"name"`
	StartTaskId string `json:"start_task_id" gorm:"index:idx_start_task_id"  `
	EndTaskId   string `json:"end_task_id" gorm:"index:idx_end_task_id"  `
	Resc        string `json:"resc"`
	ObjId       string `json:"obj_id"`
	Status      int    `json:"status"`
	//WorkflowId  string `json:"workflow_id" gorm:"index:idx_workflow_id"  `
}

func (ctl *EdgeCtl) CreateItem(item Edge) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "EdgeCtl.CreateItem: ")
	}

	return nil
}

//

//func (ctl *EdgeCtl) CreateItemFromWorkFlowAndFromTo(id string, name string, workflowId string,
//	startName string, endName string,
//	resource string, objId string,
//) (err error) {
//	sql := "INSERT INTO compute_edge (id, start, end, create_at, status) \n" +
//		""
//
//	subqueryResult := 0
//	db.Raw(sql).Scan(&subqueryResult)
//	log.Println(subqueryResult)
//	return nil
//
//	db.Raw(`
//	INSERT INTO compute_edge (id, start, end, create_at, status)
//	SELECT ? AS id,
//	(SELECT id FROM compute_task WHERE workflow_id = ? AND name = ?) AS start,
//		(SELECT id FROM compute_task WHERE workflow_id = ? AND name = ?) AS end,
//		1 AS create_at,
//		1 AS status
//    `, id, workflowId, startName, workflowId, endName).Scan(&subqueryResult)
//}

func (ctl *EdgeCtl) GetItemsByStartTaskId(taskId string) ([]Edge, error) {
	var items []Edge
	e := db.Where("start_task_id = ?", taskId).Find(&items).Error
	if e != nil {
		return nil, errors.Wrap(e, "EdgeCtl.GetItemsByStartTaskId")
	}

	//log.Println("In GetItemsByStartTaskId x: ", items)
	return items, nil
}

func (ctl *EdgeCtl) GetItemsByEndTaskId(taskId string) ([]Edge, error) {
	var items []Edge
	e := db.Where("end_task_id = ?", taskId).Find(&items).Error
	if e != nil {
		return nil, errors.Wrap(e, "EdgeCtl.GetItemsByEndTaskId")
	}

	//log.Println("In GetItemsByEndTaskId x: ", items)
	return items, nil
}

func (ctl *EdgeCtl) GetUnfinishedUpstremTaskId(endTaskId string) (unfinishedSize int64, err error) {
	var count int64

	db.Table("compute_edge").
		Joins("JOIN compute_task ON compute_edge.start_task_id = compute_task.id").
		Where("compute_task.status != ? AND compute_edge.end_task_id = ?", api.TASK_STATUS_END, endTaskId).
		Count(&count)

	return count, nil
}

func (ctl *EdgeCtl) DeleteItemByID(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Edge{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "EdgeCtl.DeleteItemByID: ")
	}

	return nil
}

func (ctl *EdgeCtl) GetItemByID(id string) (item Edge, e error) { // todo: optimize
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Edge{}, errors.Wrap(err, "EdgeCtl GetItemByID ErrRecordNotFound: ")
	} else if err != nil {
		return Edge{}, errors.Wrap(err, "EdgeCtl GetItemByID err not nil: ")
	}

	return item, nil
}

func (ctl *EdgeCtl) GetItemsBySearch(
	hasSortBy bool, sortBy string,
	hasCreateAtGte bool, createAtGte int64,
	hasCreateAtLte bool, createAtLte int64,
	hasPageID bool, pageID int,
	hasPageSize bool, pageSize int,
) (items []Edge, total int64, e error) {
	step := db.Where(" TRUE ")

	//
	if hasCreateAtGte {
		step = step.Where("create_at >= ?", createAtGte)
	}
	if hasCreateAtLte {
		step = step.Where("create_at <= ?", createAtLte)
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
