package link

import (
	"encoding/json"
	"log"
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
	PACKAGE_TYPE_HELLO     = 1001
	PACKAGE_TYPE_AUTH      = 1011
	PACKAGE_TYPE_HEARTBEAT = 1012
	PACKAGE_TYPE_GOODBYE   = 1013
	PACKAGE_TYPE_REPORT    = 1021
	PACKAGE_TYPE_BIZ       = 1013
)

type Package struct {
	Id   int64
	Ver  string
	Type int /* 1, hello, 2 Auth 3. HeartBeat 4 GoodBye 5. ReportStatus 6.  */
	Body interface{}
}

type AuthReq struct {
	Token string
	Host  string
}
type AuthResp struct {
	Code     int
	Msg      string
	ExpireAt int64
}
type BizData struct {
	Code int
	Msg  string
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
