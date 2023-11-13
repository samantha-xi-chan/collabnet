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
	"math/rand"
	"net/http"
	"os/exec"
	"time"
)

func onNewDto(dto api.PluginTask) {
	log.Println("  任务内容是： ", dto.Cmd, ", 任务准备的超时时间(秒)是： ", dto.TimeoutPre, ", 任务运行的超时时间(秒)是： ", dto.TimeoutRun)

	notifyTaskStatus(dto.Id, api.TASK_EVT_START, 0)

	quit := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				notifyTaskStatus(dto.Id, api.TASK_EVT_HEARTBEAT, 0)

			case <-quit:
				return
			}
		}
	}()

	// run
	cmd := exec.Command("sh", "-c", dto.Cmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Start(); err != nil {
		log.Printf("cmd.Start() error : %v\n", err)
		return
	}

	pid := cmd.Process.Pid
	log.Printf("dto.Id: %s, cmd.Process.Pid：%d\n", dto.Id, pid)
	err := cmd.Wait()
	if err != nil {
		fmt.Printf("Error waiting for the command to finish: %v\n", err) // todo: warning
		//os.Exit(1)
	}

	log.Printf("Command output:\n%s", stdout.String())

	quit <- true

	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(10)
	notifyTaskStatus(dto.Id, api.TASK_EVT_END, randomNumber) // 返回 [0,9] 随机值 作为 exitCode
}

func main() {

	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("BuildTime: %s\n", BuildTime)
	fmt.Printf("GitCommit: %s\n", GitCommit)

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
		log.Println("当前收到的任务编号是:  ", dto.Id, ", TaskId = ", dto.TaskId, ", 任务是否仍然有效： ", dto.Valid)
		if dto.Valid {
			log.Printf("后台处理任务 %s,  继续接收新任务\n", dto.Id)
			go onNewDto(dto)

			time.Sleep(time.Second)
			continue
		} else {
			log.Println("业务代码此时应该 关闭如果正在运行的编号为  ", dto.Id, "的任务，并发送任务结束的通知")
			notifyTaskStatus(dto.Id, api.TASK_EVT_END, 0)

			log.Printf("关闭任务 %s,  继续接收新任务\n", dto.Id)
			continue
		}
	}
}

func notifyTaskStatus(id string, status int, statusPara01 int) (x string, e error) {
	url := fmt.Sprintf("http://%s%s%s/%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER, id)
	log.Println("notifyTaskStatus url : ", url, ", status:  ", status, ", statusPara01:  ", statusPara01)

	requestBody, err := json.Marshal(api.PostPluginTaskStatusReq{
		Msg:    "PostPluginTaskStatusReq",
		Status: status,
		Para01: statusPara01,
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
		return api.PluginTask{}, err
	}

	return dto, nil
}
