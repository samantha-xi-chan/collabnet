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
	Id         string `json:"id" gorm:"primaryKey"`
	Desc       string `json:"desc"`
	Status     int    `json:"status"`
	Code       int    `json:"code"`
	Endpoint   string `json:"endpoint"`
	CreateAt   int64  `json:"create_at"`
	PreparedAt int64  `json:"prepared_at"`
	FinishAt   int64  `json:"finish_at"`
	ActiveAt   int64  `json:"active_at"`
	Enabled    int    `json:"enabled"`

	CmdackTimeout int `json:"cmdack_timeout"` /* second */
	PreTimeout    int `json:"pre_timeout"`    /* second */
	HbTimeout     int `json:"hb_timeout"`     /* second */
	RunTimeout    int `json:"run_timeout"`    /* second */

	CmdackTimer string `json:"cmdack_timer"`
	PreTimer    string `json:"pre_timer"`
	HbTimer     string `json:"hb_timer"`
	RunTimer    string `json:"run_timer"`
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

func (ctl *SchedCtl) GetItemById(id string) (i Sched, e error) { // todo: optimize
	var item Sched
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemById ErrRecordNotFound: ")
	} else if err != nil {
		return Sched{}, errors.Wrap(err, "SchedCtl GetItemById err not nil: ")
	}

	return item, nil
}
