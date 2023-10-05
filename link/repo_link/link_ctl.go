package repo_link

import (
	"fmt"
	"github.com/pkg/errors"
	"log"

	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type LinkCtl struct{}

var linkCtl LinkCtl

func GetLinkCtl() *LinkCtl {
	return &linkCtl
}

func (Link) TableName() string {
	return "link"
}

type Link struct {
	Id       string `json:"id" gorm:"primaryKey"`
	Host     string `json:"host" ` //  gorm:"unique"
	CreateAt int64  `json:"create_at"`
	DeleteAt int64  `json:"delete_at"`
	Online   int    `json:"online"`
}

func (ctl *LinkCtl) CreateItem(item Link) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "LinkCtl.CreateItem: ")
	}

	return nil
}

func (ctl *LinkCtl) UpdateItemById(id string, fieldsToUpdate map[string]interface{}) (e error) {

	result := db.Model(&Link{}).Where("id = ?", id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("UpdateItemById e: ", e)
	}

	return nil
}

func (ctl *LinkCtl) DeleteItemById(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Link{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "LinkCtl.DeleteItemById: ")
	}

	return nil
}

func (ctl *LinkCtl) GetItemByKeyValue(key string, val interface{}) (i Link, e error) { // todo: optimize
	var item Link
	err := db.Where(fmt.Sprintf("%s = ?", key), val).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Link{}, err
	} else if err != nil {
		return Link{}, errors.Wrap(err, "LinkCtl GetItemByContainerId err not nil: ")
	}

	return item, nil
}

func (ctl *LinkCtl) GetItemById(id string) (i Link, e error) { // todo: optimize
	var item Link
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Link{}, errors.Wrap(err, "LinkCtl GetItemById ErrRecordNotFound: ")
	} else if err != nil {
		return Link{}, errors.Wrap(err, "LinkCtl GetItemById err not nil: ")
	}

	return item, nil
}

func (ctl *LinkCtl) UpdateItemByKeyValue(key string, val interface{}, fieldsToUpdate map[string]interface{}) (i Link, e error) { // todo: optimize
	var item Link
	err := db.Where(fmt.Sprintf("%s = ?", key), val).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Link{}, err
	} else if err != nil {
		return Link{}, errors.Wrap(err, "LinkCtl GetItemByContainerId err not nil: ")
	}

	result := db.Model(&Link{}).Where("id = ?", item.Id).Updates(fieldsToUpdate)
	if result.Error != nil {
		log.Println("Updates e: ", e)
		return Link{}, errors.Wrap(err, "LinkCtl Updates fieldsToUpdate err ")
	}

	return item, nil
}
