package main

import (
	"bytes"
	"collab-net-v2/internal/config"
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
		resp, e := waitTaskCmd()
		if e != nil {
			log.Println("performLongPoll e :", e)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Println("resp: ", resp)

		// 判断内容  如果当前任务的属性为 有效 则发 任务开始执行的http, 解析出 任务的执行时长条件要求，
		ackTaskStart("task_id")

		// 此处是任务执行 用 sleep 1秒代替
		time.Sleep(time.Second * 1)

		ackTaskEnd("task_id", 1)
	}
}

func ackTaskStart(id string) (x string, e error) {
	url := fmt.Sprintf("http://%s%s%s/%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER, id)
	log.Println("ackTaskStart url : ", url)

	requestBody := []byte(`{"state":5}`)
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

func ackTaskEnd(id string, result int) (x string, e error) {
	url := fmt.Sprintf("http://%s%s%s/%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER, id)
	log.Println("ackTaskEnd url : ", url)

	requestBody := []byte(`{"state":9}`)
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

func waitTaskCmd() (x string, e error) {
	url := fmt.Sprintf("http://%s%s%s", config.PLUGIN_SERVICE_IP, config.PLUGIN_SERVICE_PORT, config.PLUGIN_SERVICE_ROUTER)
	log.Println("URL = ", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Long poll request failed:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Long poll request returned status:", resp.Status)
		return "", errors.New("!= http.StatusOK")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return "", err
	}

	//fmt.Println("Received data from server:", string(body))
	return string(body), nil
}
