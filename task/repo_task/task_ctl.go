package repo_task

import (
	"fmt"
	"github.com/pkg/errors"
	"log"

	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TaskCtl struct{}

var taskCtl TaskCtl

func GetTaskCtl() *TaskCtl {
	return &taskCtl
}

func (Task) TableName() string {
	return "task"
}

type Task struct {
	Id     string `json:"id" gorm:"primaryKey"`
	Desc   string `json:"desc"`
	Cmd    string `json:"cmd"`
	Status int    `json:"status"`
	//Code     int    `json:"code"`
	CreateAt int64 `json:"create_at"`
	QueueAt  int64 `json:"queue_at"`

	IdSched string `json:"id_sched"  gorm:"index:idx_id_sched"  `
}

func (ctl *TaskCtl) CreateItem(item Task) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "TaskCtl.CreateItem: ")
	}

	return nil
}

func (ctl *TaskCtl) UpdateItemById(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Task{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemById e: ", e)
	}

	return nil
}

func (ctl *TaskCtl) DeleteItemById(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Task{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "TaskCtl.DeleteItemById: ")
	}

	return nil
}

func (ctl *TaskCtl) GetItemByKeyValue(key string, val interface{}) (i Task, e error) { // todo: optimize
	var item Task
	err := db.Where(fmt.Sprintf("%s = ?", key), val).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, err
	} else if err != nil {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

func (ctl *TaskCtl) GetItemById(id string) (i Task, e error) { // todo: optimize
	var item Task
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemById ErrRecordNotFound: ")
	} else if err != nil {
		return Task{}, errors.Wrap(err, "TaskCtl GetItemById err not nil: ")
	}

	return item, nil
}
