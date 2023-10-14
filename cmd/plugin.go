package main

import (
	"bytes"
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	for {
		log.Println("-    -    -    -    -    ")
		dto, e := waitTaskCmd()
		if e != nil {
			log.Println("performLongPoll e :", e)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("dto: %#v \n", dto)

		// 业务接入提示：
		log.Println("当前收到的任务编号是:  ", dto.Id, ", 任务是否仍然有效： ", dto.Valid, ", 任务内容是： ", dto.Cmd, ", 任务准备的超时时间(秒)是： ", dto.TimeoutPre, ", 任务运行的超时时间(秒)是： ", dto.TimeoutRun)

		if !dto.Valid { //  如果 !dto.Valid 且此任务正在运行
			log.Println("业务代码此时应该 关闭正在运行的编号为  ", dto.Id, "的任务，并发送任务结束的通知")
			notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_END_SUCC)
			continue
		} else if false { //  如果 !dto.Valid 且此任务未处理过 (这里明显代码逻辑不对 就不要照着抄写了)

			continue
		}

		// 判断内容  如果当前任务的属性为 有效 则发 任务开始执行的http, 解析出 任务的执行时长条件要求，
		notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_START)

		// 此处是任务执行 用 sleep 代替, 执行过程中需要发送心跳, 这个demo表示 任务执行耗时 3秒
		for i := 0; i < 3; i++ {
			// 这里是任务执行（例如 执行安全测试）
			time.Sleep(time.Millisecond * 500)
			notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_HEARTBEAT)
			time.Sleep(time.Millisecond * 500)
		}

		notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_END_SUCC)
	}
}

func notifyTaskStatus(id string, status int) (x string, e error) {
	url := fmt.Sprintf("http://%s%s%s/%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER, id)
	log.Println("notifyTaskStatus url : ", url)

	requestBody, err := json.Marshal(api.PostPluginTaskStatusReq{
		//Id:     "",
		Msg:    "PostPluginTaskStatusReq",
		Status: status,
		Para01: 0,
	})
	if err != nil {
		fmt.Println("转换为JSON时发生错误:", err)
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("HTTP请求发送失败:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Println("HTTP请求成功")
	} else {
		log.Println("HTTP请求失败:", resp.Status)
	}

	return "", nil
}

func waitTaskCmd() (pluginTask api.PluginTask, e error) { // 包含任务启动、任务终止
	url := fmt.Sprintf("http://%s%s%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER)
	log.Println("waitTaskCmd URL = ", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Long poll request failed:", err)
		return api.PluginTask{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Long poll request returned status:", resp.Status)
		return api.PluginTask{}, errors.New("!= http.StatusOK")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return api.PluginTask{}, err
	}

	//log.Println(" bodyBytes: ", bodyBytes, "\n string(bodyBytes): ", string(bodyBytes))
	log.Println(" string(bodyBytes): ", string(bodyBytes))

	var httpRespBody api.HttpRespBody
	err = json.Unmarshal(bodyBytes, &httpRespBody)
	if err != nil {
		log.Println("httpRespBody JSON unmarshal error:", err)
		//notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_REJECT)
		return api.PluginTask{}, err
	}

	if httpRespBody.Code != 0 {
		return api.PluginTask{}, errors.New("httpRespBody.Code != 0")
	}

	dataBytes, err := json.Marshal(httpRespBody.Data)
	if err != nil {
		log.Println("json.Marshal(httpRespBody.Data): ", err)
		return api.PluginTask{}, err
	}

	//log.Println("resp: ", respBytes)
	var dto api.PluginTask
	err = json.Unmarshal(dataBytes, &dto)
	if err != nil {
		log.Println("JSON unmarshal error:", err)
		//notifyTaskStatus(dto.Id, config.PLUGIN_TASK_EVT_REJECT)
		return api.PluginTask{}, err
	}

	return dto, nil
}
