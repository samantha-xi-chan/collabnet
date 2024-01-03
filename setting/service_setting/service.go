package service_setting

import (
	"collab-net-v2/workflow/repo_workflow"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

func GetSettingUrl(id string) (url string, e error) {
	itemSetting, ee := repo_workflow.GetSettingCtl().GetItemByID(id)
	if ee != nil {
		return "", errors.Wrap(ee, "repo_workflow.GetSettingCtl().GetItemByID: ")
	} else {
		var innerData map[string]interface{}
		err := json.Unmarshal([]byte(itemSetting.Value), &innerData)
		if err != nil {
			return "", errors.Wrap(err, "json.Unmarshal: ")
		}
		url = innerData["url"].(string)
		fmt.Println("URL:", url)
	}

	return url, nil
}
