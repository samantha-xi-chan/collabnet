package link

import (
	"collab-net-v2/internal/config"
	"collab-net-v2/time/service_time"
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
	mapChan2Endpoint = make(map[*websocket.Conn]string)
)

func SendDataToEndpoint(endpoint string, bytes []byte) (e error) {
	chanDestin := mapEndpoint2Conn[endpoint]
	if chanDestin != nil {
		mapConn2ChanWrite[chanDestin] <- bytes
	}

	return nil
}

func OnBizDataFromEndpoint(endpoint string, bytes []byte) {
	log.Printf("endpoint: %s, bytes: %s", endpoint, bytes)

}

func OnNewConn(conn *websocket.Conn) {
	mapConn2ChanWrite[conn] = make(chan []byte)
	mapConn2ChanRead[conn] = make(chan []byte)
}

func OnConnDelete(conn *websocket.Conn) {
	delete(mapConn2ChanWrite, conn)
	delete(mapConn2ChanRead, conn)
}

func DispatchCmdToHostAndWait(hostName string, cmd string) (e error) {

	return nil
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
				mType, bytes, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					return
				}

				log.Printf("mType: %d receive: %s\n", mType, string(bytes))

				endpoint := mapChan2Endpoint[conn]
				if endpoint == "" { // 未注册的链接
					bytesResp, exit, e := OnMessage(conn, bytes)
					if e == nil && exit != true {
						mapConn2ChanWrite[conn] <- bytesResp
					}
				} else { // 已注册的链接
					OnBizDataFromEndpoint(endpoint, bytes)
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
					log.Println("bytes: ", string(bytes))
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
func OnMessage(conn *websocket.Conn, bytes []byte) ([]byte, bool, error) { // it is stateless
	var pack Package
	err := json.Unmarshal(bytes, &pack)
	if err != nil {
		log.Println("Error decoding JSON:", err)
		return nil, false, nil
	}

	if pack.Type == PACKAGE_TYPE_AUTH {
		var body AuthReq
		bytes, _ := json.Marshal(pack.Body)
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			log.Println("Error decoding JSON:", err)
			return nil, false, nil
		}
		log.Println("OnMessage  PACKAGE_TYPE_AUTH body : ", body)

		if body.Token == config.AuthTokenForDev {
			mapEndpoint2Conn[body.Host] = conn

			return GetPackageBytes(
				time.Now().UnixMilli(),
				"1.0",
				PACKAGE_TYPE_AUTH,
				AuthResp{
					Code:     0,
					Msg:      "",
					ExpireAt: time.Now().UnixMilli() + 3600,
				},
			), false, nil
		}
	} else if pack.Type == PACKAGE_TYPE_GOODBYE {
		log.Println("[OnMessage] body : ", "PACKAGE_TYPE_GOODBYE")
		return nil, true, nil
	} else {
		log.Println("Unknown OnMessage ")
	}

	return nil, false, nil
}

// 主机管理

//
func init() {
	service_time.Init()
}

func test() {
	idTimer, _ := service_time.NewTimer(2, "20s timer")
	service_time.DisableTimer(idTimer)
}

func NewServer() {
	//test()

	done := make(chan struct{})
	sigint := make(chan os.Signal, 1)

	//goto wait4end

	go func() {
		http.HandleFunc("/", handleWebSocket)
		log.Fatal(http.ListenAndServe(config.SCHEDULER_LISTEN_PORT, nil))
	}()

	go func() {
		time.Sleep(time.Second * 10)
		DispatchCmdToHostAndWait("M1", "touch ~/CmdFromServer")
	}()

	//wait4end:
	log.Println("waiting select{}")
	select {
	case <-sigint:
		log.Println("Interrupted by user")
		time.Sleep(time.Millisecond * 1000)
		log.Println("Interrupted by user 222")
	case <-done:
		log.Println(" chan done")
	}
}
