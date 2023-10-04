package link

import (
	"collab-net-v2/internal/config"
	"encoding/json"
	mapset "github.com/deckarep/golang-set"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	WildcardAsterisk = "*"
	mapTopicChanSet  = make(map[string](mapset.Set))

	mapConn2ChanWrite = make(map[*websocket.Conn]chan []byte)
	mapConn2ChanRead  = make(map[*websocket.Conn]chan []byte)

	mapEndpoint2Conn = make(map[string]*websocket.Conn)
	mapConn2Endpoint = make(map[*websocket.Conn]string)
)

var callbackFuncBizData func(endpoint string, bytes []byte) (e error)
var callbackFuncConnChange func(endpoint string, _type int) (e error)

func SetBizDataCallback(_func func(endpoint string, bytes []byte) (e error)) {
	callbackFuncBizData = _func
}
func SetConnChangeCallback(_func func(endpoint string, _type int) (e error)) {
	callbackFuncConnChange = _func
}

func SendDataToEndpoint(endpoint string, bytes []byte) (errCode int, e error) {
	chanDestin := mapEndpoint2Conn[endpoint]
	if chanDestin != nil {
		mapConn2ChanWrite[chanDestin] <- bytes
		return 0, nil
	}

	return -1, nil
}

func OnNewConn(conn *websocket.Conn) {
	mapConn2ChanWrite[conn] = make(chan []byte)
	mapConn2ChanRead[conn] = make(chan []byte)
}

func OnConnDelete(conn *websocket.Conn) {
	delete(mapConn2ChanWrite, conn)
	delete(mapConn2ChanRead, conn)
}

const (
	SOCK_BUF_SIZE = 1024
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  SOCK_BUF_SIZE,
	WriteBufferSize: SOCK_BUF_SIZE,
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) { // it is stateless
	chanEndNotify := make(chan []byte)

	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	OnNewConn(conn)

	defer func() {
		conn.Close()
		OnConnDelete(conn)
		log.Println("[handleWebSocket] function defer ....")
	}()

	go func() { // reader
		for {
			select {
			case <-chanEndNotify:
				log.Println("chanEndNotify")
				return
			default:
				log.Println("waiting conn.ReadMessage")
				mType, bytes, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					return
				}

				log.Printf("mType: %d receive: %s\n", mType, string(bytes))

				endpoint := mapConn2Endpoint[conn]
				if endpoint == "" { // 未注册的链接
					bytesResp, evtType, e := OnMessageOfUnregisterChan(conn, bytes)
					if e == nil && evtType != LINK_EVT_BYE {
						mapConn2ChanWrite[conn] <- bytesResp
					}
				} else { // 已注册的链接
					OnMessageOfRegisterChan(endpoint, bytes)
					//bytesResp, evtType, e := OnMessageOfRegisterChan(endpoint, bytes)
				}
			}

		}
	}()

	go func() { // writer
		for true {
			select {
			case <-chanEndNotify:
				log.Println("chanEndNotify")
				return
			case bytes, ok := <-mapConn2ChanWrite[conn]:
				if ok {
					log.Println("routine writer bytes: ", string(bytes))
					if err := conn.WriteMessage(0x01, bytes); err != nil {
						log.Println("WriteMessage e: ", err)
						return
					}
				}
			}
		}
	}()

	select {
	case <-chanEndNotify:
		return
	}
}

// 连接管理

const (
	LINK_EVT_NONE         = 1001
	LINK_EVT_HANDSHAKE_OK = 1011
	LINK_EVT_BYE          = 1099
)

func OnMessageOfRegisterChan(endpoint string, bytesPack []byte) ([]byte, int, error) { // it is stateless

	var pack Package
	err := json.Unmarshal(bytesPack, &pack)

	if err != nil {
		log.Println("Error decoding JSON:", err)
		return nil, LINK_EVT_NONE, nil
	}

	if pack.Type == PACKAGE_TYPE_AUTHOK_RECVED {
		callbackFuncConnChange(endpoint, LINK_EVT_HANDSHAKE_OK)
		return nil, LINK_EVT_HANDSHAKE_OK, nil
	} else if pack.Type == PACKAGE_TYPE_GOODBYE {
		log.Println("[OnMessage] body : ", "PACKAGE_TYPE_GOODBYE")
		callbackFuncConnChange(endpoint, LINK_EVT_BYE)
		return nil, LINK_EVT_BYE, nil
	} else if pack.Type == PACKAGE_TYPE_BIZ {
		bytes, _ := json.Marshal(pack.Body)
		//OnBizDataFromRegisterEndpoint(endpoint, bytes)
		callbackFuncBizData(endpoint, bytes)

		return nil, LINK_EVT_NONE, nil
	} else {
		log.Println("[OnMessageOfRegisterChan] unknown")
		return nil, LINK_EVT_NONE, nil
	}

}

func OnMessageOfUnregisterChan(conn *websocket.Conn, bytes []byte) ([]byte, int, error) { // it is stateless
	var pack Package
	err := json.Unmarshal(bytes, &pack)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return nil, LINK_EVT_NONE, nil
	}

	if pack.Type == PACKAGE_TYPE_AUTH {
		var body AuthReq
		bytes, _ := json.Marshal(pack.Body)
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			log.Println("Error decoding JSON:", err)
			return nil, LINK_EVT_NONE, nil
		}
		log.Println("OnMessage  PACKAGE_TYPE_AUTH body : ", body)

		if body.Token == config.AuthTokenForDev {
			mapEndpoint2Conn[body.Host] = conn
			mapConn2Endpoint[conn] = body.Host

			return GetPackageBytes(
				time.Now().UnixMilli(),
				"1.0",
				PACKAGE_TYPE_AUTH,
				AuthResp{
					Code:     0,
					Msg:      "",
					ExpireAt: time.Now().UnixMilli() + 3600,
				},
			), LINK_EVT_NONE, nil
		}
	} else if pack.Type == PACKAGE_TYPE_AUTHOK_RECVED { // deprecated
		host := mapConn2Endpoint[conn]
		if host != "" {
			callbackFuncConnChange(host, LINK_EVT_HANDSHAKE_OK)
		}

		return nil, LINK_EVT_HANDSHAKE_OK, nil
	} else if pack.Type == PACKAGE_TYPE_GOODBYE { // deprecated
		log.Println("[OnMessage] body : ", "PACKAGE_TYPE_GOODBYE")
		return nil, LINK_EVT_BYE, nil
	} else {
		log.Println("Unknown OnMessage ")
	}

	return nil, LINK_EVT_NONE, nil
}

func init() {
}

func NewServer() {
	done := make(chan struct{})
	sigint := make(chan os.Signal, 1)

	go func() {
		http.HandleFunc("/", handleWebSocket)
		log.Fatal(http.ListenAndServe(config.SCHEDULER_LISTEN_PORT, nil))
	}()

	//go func() {
	//	time.Sleep(time.Second * 10)
	//	DispatchCmdToHostAndWait("M1", "touch ~/CmdFromServer")
	//}()

	log.Println("[NewServer] waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
		time.Sleep(time.Millisecond * 1000)
		log.Println("Interrupted by user 222")
	case <-done:
		log.Println(" chan done")
	}
}
