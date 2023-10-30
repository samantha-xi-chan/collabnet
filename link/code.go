package link

import (
	"encoding/json"
	"log"
)

const (
	INT_INVALID = -1
)

const (
	STATE_INIT                = 1001
	STATE_CONNECT_NOK         = 1002
	STATE_CONNECT_Ok_BIZ_NONE = 1003
	STATE_CONNECT_Ok_AUTH_Ok  = 1004
	STATE_CONNECT_Ok_AUTH_NOk = 1005
)

const (
	EVT_CONNECT_SUCC     = 1001
	EVT_CONNECT_FAIL     = 1002
	EVT_CONNECT_AUTH_OK  = 1003
	EVT_CONNECT_AUTH_NOK = 1004
	EVT_HEARTBEAT        = 1005
)

const (
	PACKAGE_TYPE_HELLO         = 1001
	PACKAGE_TYPE_AUTH          = 1011
	PACKAGE_TYPE_AUTHOK_RECVED = 1012
	PACKAGE_TYPE_HEARTBEAT     = 1032
	PACKAGE_TYPE_GOODBYE       = 1043
	PACKAGE_TYPE_REPORT        = 1061
	PACKAGE_TYPE_BIZ           = 1083
)

const (
	BIZ_TYPE_NEWTASK  = 13891011 //  启动任务在用，任务结束通知也在用
	BIZ_TYPE_STOPTASK = 13891021

	BIZ_TYPE_NEW_DOCKER_TASK  = 13891111
	BIZ_TYPE_STOP_DOCKER_TASK = 13891121
)

type Package struct {
	Id   int64       `json:"id"`
	Ver  string      `json:"ver"`
	Type int         `json:"type"` /* 1, hello, 2 Auth 3. HeartBeat 4 GoodBye 5. ReportStatus 6.  */
	Body interface{} `json:"body"`
}

type AuthReq struct {
	Token    string `json:"token"`
	HostName string `json:"host_name"`
}
type AuthResp struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	ExpireAt int64  `json:"expire_at"`
}

type BizInit struct { // 暂时做语意角度的开放
	Para01 int `json:"para01"`
}

type BizData struct { // 业务角度： 任务新建、任务停止、
	TypeId  int    `json:"type_id"`
	SchedId string `json:"sched_id"`
	TaskId  string `json:"task_id"`

	Para01 int    `json:"para01"`
	Para02 int    `json:"para02"`
	Para03 int    `json:"para03"`
	Para11 string `json:"para11"`
	// 洋葱进去一层
	Para0101 int    `json:"para0101"`
	Para0102 string `json:"para0102"`

	//HbInterval int `json:"hb_interval"` /* second */
	//PreTimeout int `json:"pre_timeout"` /* second */
	//RunTimeout int `json:"run_timeout"` /* second */
}

//type HelloReq struct {
//	Host string
//}

func GetPackageBytes(id int64, ver string, typee int, body interface{}) (x []byte) {
	pack := Package{
		Id:   id,
		Ver:  ver,
		Type: typee,
		Body: body,
	}

	jsonData, _ := json.Marshal(pack)
	log.Println("[GetPackageBytes] ", string(jsonData))

	return jsonData
}
