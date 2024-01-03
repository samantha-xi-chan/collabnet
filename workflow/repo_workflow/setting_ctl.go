package repo_workflow

import (
	"collab-net-v2/api"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strings"
)

type SettingCtl struct{}

var settingCtl SettingCtl

func GetSettingCtl() *SettingCtl {
	return &settingCtl
}

func (Setting) TableName() string {
	return "setting"
}

type Setting struct {
	Id       string `json:"id"  gorm:"primaryKey"`
	Name     string `json:"name"`
	CreateAt int64  `json:"create_at"`
	Value    string `json:"value"`
}

func (ctl *SettingCtl) CreateItem(item Setting) (err error) {
	if err := db.Create(&item).Error; err != nil {
		return errors.Wrap(err, "SettingCtl.CreateItem: ")
	}

	return nil
}

func (ctl *SettingCtl) FirstOrCreate(item Setting) (err error) {

	result := db.Where(Setting{Id: item.Id}).Assign(item).FirstOrCreate(&item)

	if result.Error != nil {
		return errors.Wrap(result.Error, "db.Where(Setting{Id: item.Id}).Assign(item).FirstOrCreate(&item): ")
	}

	return nil
}

func (ctl *SettingCtl) GetItemsByStartTaskId(taskId string) ([]Setting, error) {
	var items []Setting
	e := db.Where("start_task_id = ?", taskId).Find(&items).Error
	if e != nil {
		return nil, errors.Wrap(e, "SettingCtl.GetItemsByStartTaskId")
	}

	//log.Println("In GetItemsByStartTaskId x: ", items)
	return items, nil
}

func (ctl *SettingCtl) GetItemsByStartTaskIdAndIterate(taskId string, iterate int) ([]Setting, int64, error) {
	var items []Setting
	var cnt int64
	e := db.Where("`start_task_id` = ? AND `iterate` = ? ", taskId, iterate).Find(&items).Count(&cnt).Error
	if e != nil {
		return nil, cnt, errors.Wrap(e, "SettingCtl.GetItemsByStartTaskId")
	}

	//log.Println("In GetItemsByStartTaskId x: ", items)
	return items, cnt, nil
}

func (ctl *SettingCtl) GetItemsByEndTaskId(taskId string) ([]Setting, error) {
	var items []Setting
	e := db.Where("end_task_id = ?", taskId).Find(&items).Error
	if e != nil {
		return nil, errors.Wrap(e, "SettingCtl.GetItemsByEndTaskId")
	}

	//log.Println("In GetItemsByEndTaskId x: ", items)
	return items, nil
}

func (ctl *SettingCtl) GetUnfinishedUpstremTaskId(endTaskId string) (unfinishedSize int64, err error) {
	var count int64

	db.Table("c_edge").
		Joins("JOIN c_task ON c_edge.start_task_id = c_task.id").
		Where("c_task.status != ? AND c_edge.end_task_id = ?", api.TASK_STATUS_END, endTaskId).
		Count(&count)

	return count, nil
}

func (ctl *SettingCtl) DeleteItemByID(id string) (err error) {
	result := db.Where("id = ?", id).Delete(&Setting{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "SettingCtl.DeleteItemByID: ")
	}

	return nil
}

func (ctl *SettingCtl) GetItemByID(id string) (item Setting, e error) { // todo: optimize
	err := db.Where("id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Setting{}, errors.Wrap(err, "SettingCtl GetItemByID ErrRecordNotFound: ")
	} else if err != nil {
		return Setting{}, errors.Wrap(err, "SettingCtl GetItemByID err not nil: ")
	}

	return item, nil
}

func (ctl *SettingCtl) GetItemsBySearch(
	hasSortBy bool, sortBy string,
	hasCreateAtGte bool, createAtGte int64,
	hasCreateAtLte bool, createAtLte int64,
	hasPageID bool, pageID int,
	hasPageSize bool, pageSize int,
) (items []Setting, total int64, e error) {
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
