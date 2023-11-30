package repo_sched

import (
	"fmt"
	"github.com/pkg/errors"
	"log"

	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type SchedCtl struct{}

var schedCtl SchedCtl

func GetSchedCtl() *SchedCtl {
	return &schedCtl
}

func (Sched) TableName() string {
	return "sched"
}

type Sched struct {
	Id          string `json:"id" gorm:"primaryKey"`
	TaskId      string `json:"task_id"  gorm:"index:idx_task_id" `
	TaskType    int    `json:"task_type"`
	TaskEnabled int    `json:"task_enabled"` /* 上层任务是否仍然Enabled */

	LinkId     string `json:"link_id"`
	Carrier    string `json:"carrier"`
	CreateAt   int64  `json:"create_at"`
	CmdackAt   int64  `json:"cmdack_at"`
	PreparedAt int64  `json:"prepared_at"`
	FinishAt   int64  `json:"finish_at"`
	ActiveAt   int64  `json:"active_at"`

	Enabled  int `json:"enabled"`   /* 是否已和task脱钩, 包含重试场景和业务角度放弃某个task导致脱钩 */
	BestProg int `json:"best_prog"` /* 生命周期的最好阶段, 仅仅用于 debug使用 */
	BizCode  int `json:"biz_code"`  /* 业务角度的ExitCode */
	FwkCode  int `json:"fwk_code"`  /* 调度框架角度的ExitCode */

	CmdackTimeout int `json:"cmdack_timeout"` /* second */
	PreTimeout    int `json:"pre_timeout"`    /* second */
	HbTimeout     int `json:"hb_timeout"`     /* second */
	RunTimeout    int `json:"run_timeout"`    /* second */

	CmdackTimer string `json:"cmdack_timer"`
	PreTimer    string `json:"pre_timer"`
	HbTimer     string `json:"hb_timer"`
	RunTimer    string `json:"run_timer"`
	Error       string `json:"error" `
}

func (ctl *SchedCtl) CreateItem(item Sched) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "SchedCtl.CreateItem: ")
	}

	return nil
}

func (ctl *SchedCtl) UpdateItemById(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Sched{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemById e: ", e)
	}

	return nil
}

func (ctl *SchedCtl) DeleteItemById(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Sched{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "SchedCtl.DeleteItemById: ")
	}

	return nil
}

func (ctl *SchedCtl) GetItemByKeyValue(key string, val interface{}) (i Sched, e error) { // todo: optimize
	var item Sched
	err := db.Where(fmt.Sprintf("%s = ?", key), val).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Sched{}, err
	} else if err != nil {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

type QueryKeyValue struct {
	ColName  string
	ColValue interface{}
}

func (ctl *SchedCtl) GetItemByKeyValueArr(arr []QueryKeyValue) (item Sched, e error) { // todo: optimize
	if len(arr) < 1 {
		return Sched{}, errors.New("len(arr) < 1 ")
	}

	result := db.Where(arr[0].ColName, arr[0].ColValue)
	for i := 1; i < len(arr); i++ {
		result = result.Where(arr[i].ColName, arr[i].ColValue)
	}
	err := result.Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Sched{}, err
	} else if err != nil {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

func (ctl *SchedCtl) GetItemById(id string) (i Sched, e error) { // todo: optimize
	log.Println("SchedCtl GetItemById: id = ", id) // debug only

	var item Sched
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemById ErrRecordNotFound: ")
	} else if err != nil {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemById err not nil: ")
	}

	return item, nil
}
