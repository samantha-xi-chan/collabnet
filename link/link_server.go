package link

import (
	"collab-net-v2/api"
	"collab-net-v2/internal/config"
	"collab-net-v2/link/config_link"
	"collab-net-v2/link/repo_link"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/util/idgen"
	"context"
	"encoding/json"
	mapset "github.com/deckarep/golang-set"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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

	mapConn2IdLlink = make(map[*websocket.Conn]string)
)

var callbackFuncBizData func(endpoint string, bytes []byte) (e error)
var callbackFuncConnChange func(endpoint string, _type int) (e error)

func SetBizDataCallback(_func func(endpoint string, bytes []byte) (e error)) {
	callbackFuncBizData = _func
}
func SetConnChangeCallback(_func func(endpoint string, _type int) (e error)) {
	callbackFuncConnChange = _func
}

func SendDataToLinkId(linkId string, bytes []byte) (errCode int, e error) {
	item, e := repo_link.GetLinkCtl().GetItemById(context.Background(), linkId)
	if e != nil {
		return 0, errors.Wrap(e, "repo_link.GetLinkCtl().GetItemById")
	}

	chanDestin := mapEndpoint2Conn[item.HostName]
	if chanDestin != nil {
		mapConn2ChanWrite[chanDestin] <- bytes
		return 0, nil
	}

	return INT_INVALID, nil
}

func OnNewConn(conn *websocket.Conn) {
	mapConn2ChanWrite[conn] = make(chan []byte, CHAN_BUF_SIZE)
	mapConn2ChanRead[conn] = make(chan []byte, CHAN_BUF_SIZE)
}

func OnConnLost(endpoint string, conn *websocket.Conn) { // todo: mem leak ,chan leak ?
	delete(mapConn2ChanWrite, conn)
	delete(mapConn2ChanRead, conn)

	delete(mapConn2Endpoint, conn)

	conn.Close()

	delete(mapEndpoint2Conn, endpoint)

	idLink := mapConn2IdLlink[conn]
	repo_link.GetLinkCtl().UpdateItemById(idLink, map[string]interface{}{
		"delete_at": time.Now().UnixMilli(),
		"online":    0,
	})
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
		log.Println("[handleWebSocket] function defer ....")
		endpoint := mapConn2Endpoint[conn]

		callbackFuncConnChange(endpoint, LINK_EVT_BYE)
		OnConnLost(endpoint, conn)

		conn.Close()
	}()

	go func() { // reader
		for {
			select {
			case <-chanEndNotify:
				log.Println("chanEndNotify")
				return
			default:
				log.Println("[handleWebSocket] waiting conn.ReadMessage")
				mType, bytes, err := conn.ReadMessage()
				if err != nil {
					log.Println("conn.ReadMessage: e = ", err)
					close(chanEndNotify)
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
					OnMessageOfRegisterChan(endpoint, conn, bytes)
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
					if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("WriteMessage e: ", err)
						close(chanEndNotify)
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

func OnMessageOfRegisterChan(endpoint string, conn *websocket.Conn, bytesPack []byte) ([]byte, int, error) { // it is stateless

	var pack Package
	err := json.Unmarshal(bytesPack, &pack)

	if err != nil {
		log.Println("Error decoding JSON:", err)
		return nil, LINK_EVT_NONE, nil
	}

	if pack.Type == PACKAGE_TYPE_AUTHOK_RECVED {
		callbackFuncConnChange(endpoint, LINK_EVT_HANDSHAKE_OK)

		var body BizInit
		bytes, _ := json.Marshal(pack.Body)
		err = json.Unmarshal(bytes, &body)
		if err != nil {
			log.Println("Error decoding JSON:", err, " string(bytes): ", string(bytes))
		}
		log.Println("[<-readChan] body : ", body)
		// todo: 判重

		idLink := idgen.GetIdWithPref("co")

		// 源地址处理
		//remoteAddr := conn.RemoteAddr()
		//from := conn.RemoteAddr().String()
		//tcpAddr, ok := remoteAddr.(*net.TCPAddr)
		//if ok {
		//	from = tcpAddr.IP.String()
		//}
		//log.Println("From : ", from)

		repo_link.GetLinkCtl().CreateItem(repo_link.Link{
			Id:         idLink,
			HostName:   endpoint,
			FirstParty: body.Para01,
			From:       body.Para02, //
			CreateAt:   time.Now().UnixMilli(),
			DeleteAt:   0,
			Online:     api.TRUE,
		})
		mapConn2IdLlink[conn] = idLink

		return nil, LINK_EVT_HANDSHAKE_OK, nil
	} else if pack.Type == PACKAGE_TYPE_GOODBYE {
		log.Println("[OnMessage] body : ", "PACKAGE_TYPE_GOODBYE")

		return nil, LINK_EVT_BYE, nil
	} else if pack.Type == PACKAGE_TYPE_BIZ {
		bytes, _ := json.Marshal(pack.Body)
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

		if body.Token == config_sched.AuthTokenForDev {
			mapEndpoint2Conn[body.HostName] = conn
			mapConn2Endpoint[conn] = body.HostName

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

func initRepo() {
	mySqlDsn, e := config.GetMySqlDsn()
	if e != nil {
		log.Fatal("config.GetMySqlDsn: ", e)
	}
	log.Println("mySqlDsn", mySqlDsn)

	repo_link.Init(mySqlDsn, config_link.RepoLogLevel, config_link.RepoSlowMs)
}

func NewServer() {
	initRepo()

	done := make(chan struct{})
	sigint := make(chan os.Signal, 1)

	go func() {
		http.HandleFunc("/", handleWebSocket)
		log.Fatal(http.ListenAndServe(config_sched.SCHEDULER_LISTEN_PORT, nil))
	}()

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
