package repo_time

import (
	"fmt"
	"github.com/pkg/errors"
	"log"

	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type TimeCtl struct{}

var timeCtl TimeCtl

func GetTimeCtl() *TimeCtl {
	return &timeCtl
}

func (Time) TableName() string {
	return "time"
}

type Time struct {
	Id           string `json:"id" gorm:"primaryKey"`
	Type         int    `json:"type"`
	Holder       string `json:"holder"`
	Desc         string `json:"desc"`
	Status       int    `json:"status"`
	CreateAt     int64  `json:"create_at"`
	IdOnce       string `json:"id_once"  gorm:"index:idx_id_once"  `
	CreateBy     int64  `json:"create_by"`
	Timeout      int    `json:"timeout"`
	CallbackAddr string `json:"callback_addr"`
}

func (ctl *TimeCtl) CreateItem(item Time) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "TimeCtl.CreateItem: ")
	}

	return nil
}

func (ctl *TimeCtl) UpdateItemById(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Time{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemById e: ", e)
	}

	return nil
}

func (ctl *TimeCtl) DeleteItemById(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Time{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "TimeCtl.DeleteItemById: ")
	}

	return nil
}

func (ctl *TimeCtl) GetItemByKeyValue(key string, val interface{}) (i Time, e error) { // todo: optimize
	var item Time
	err := db.Where(fmt.Sprintf("%s = ?", key), val).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Time{}, err
	} else if err != nil {
		return Time{}, errors.Wrap(err, "TimeCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

func (ctl *TimeCtl) GetItemById(id string) (i Time, e error) { // todo: optimize
	var item Time
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Time{}, errors.Wrap(err, "TimeCtl GetItemById ErrRecordNotFound: ")
	} else if err != nil {
		return Time{}, errors.Wrap(err, "TimeCtl GetItemById err not nil: ")
	}

	return item, nil
}
