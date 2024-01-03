package repo_workflow

import (
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strings"
)

type WorkflowCtl struct{}

var workflowCtl WorkflowCtl

func GetWorkflowCtl() *WorkflowCtl {
	return &workflowCtl
}

func (Workflow) TableName() string {
	return "c_workflow"
}

type Workflow struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"unique"`
	CreateAt int64  `json:"create_at"`
	EndAt    int64  `json:"end_at"` // 包含 正常结束 异常结束
	CreateBy int64  `json:"create_by"`
	Desc     string `json:"desc"`
	//Enabled  int    `json:"enabled"` // 上层业务角度希望 有效/无效
	//Status   int    `json:"status"  gorm:"default:60021001"`
	ExitCode int    `json:"exit_code"`
	Define   string `json:"define"`
	Error    string `json:"error" `
	Iterate  int    `json:"iterate" `

	ShareDirArrStr string `json:"share_dir_arr_str"`
	Timeout        int    `json:"timeout"`      // - "timeout" 可以不填写, 默认为 0,  值 0 表示不设置超时
	AutoIterate    bool   `json:"auto_iterate"` // - "auto_iterate" 可以不填写, 默认为false , false 表示workflow 内部 不会无限循环执行
}

func (ctl *WorkflowCtl) CreateItem(item Workflow) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "WorkflowCtl.CreateItem: ")
	}

	return nil
}

func (ctl *WorkflowCtl) DeleteItemByID(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Workflow{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "WorkflowCtl.DeleteItemByID: ")
	}

	return nil
}

func (ctl *WorkflowCtl) GetItemByID(id string) (item Workflow, e error) { // todo: optimize
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Workflow{}, errors.Wrap(err, "WorkflowCtl GetItemByID ErrRecordNotFound: ")
	} else if err != nil {
		return Workflow{}, errors.Wrap(err, "WorkflowCtl GetItemByID err not nil: ")
	}

	return item, nil
}

func (ctl *WorkflowCtl) UpdateItemByID(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Workflow{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemByID e: ", e)
	}

	return nil
}

func (ctl *WorkflowCtl) UpdateItemByIDAndIterate(id string, iterate int, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Workflow{}).Where("`id` = ? AND `iterate` = ?", id, iterate).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemByID e: ", e)
	}

	return nil
}

func (ctl *WorkflowCtl) IncreaseIterate(id string) (e error) {
	err = db.Model(&Workflow{}).Where("id = ?", id).Update("iterate", gorm.Expr("iterate + ?", 1)).Error
	if err != nil {
		return errors.Wrap(err, "db.Model(&Workflow{}).Where(\"id = ?\", id).Update(\"iterate\", gorm.Expr(\"iterate + ?\", 1)).Error: ")
	}

	return nil
}

func (ctl *WorkflowCtl) GetItemsBySearch(
	hasSortBy bool, sortBy string,
	hasCreateAtGte bool, createAtGte int64,
	hasCreateAtLte bool, createAtLte int64,
	hasPageID bool, pageID int,
	hasPageSize bool, pageSize int,
) (items []Workflow, total int64, e error) {
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
