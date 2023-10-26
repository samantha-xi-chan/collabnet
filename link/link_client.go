package link

import (
	"collab-net-v2/util/statemachine"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	err    error
	done   = make(chan struct{})
	sigint = make(chan os.Signal, 1)

	readChanX  = make(chan []byte, CHAN_BUF_SIZE)
	writeChanX = make(chan []byte, CHAN_BUF_SIZE)

	transitions = []statemachine.Transition{
		// 升级
		{CurrentState: STATE_INIT, Event: EVT_CONNECT_SUCC, NextState: STATE_CONNECT_Ok_BIZ_NONE},
		{CurrentState: STATE_CONNECT_NOK, Event: EVT_CONNECT_SUCC, NextState: STATE_CONNECT_Ok_BIZ_NONE},
		{CurrentState: STATE_CONNECT_Ok_BIZ_NONE, Event: EVT_CONNECT_AUTH_OK, NextState: STATE_CONNECT_Ok_AUTH_Ok},

		// 维持 0
		{CurrentState: STATE_INIT, Event: EVT_CONNECT_FAIL, NextState: STATE_CONNECT_NOK},
		{CurrentState: STATE_CONNECT_NOK, Event: EVT_CONNECT_FAIL, NextState: STATE_CONNECT_NOK},

		// 维持 1
		{CurrentState: STATE_CONNECT_Ok_BIZ_NONE, Event: EVT_HEARTBEAT, NextState: STATE_CONNECT_Ok_BIZ_NONE},
		{CurrentState: STATE_CONNECT_Ok_AUTH_Ok, Event: EVT_HEARTBEAT, NextState: STATE_CONNECT_Ok_AUTH_Ok},
		{CurrentState: STATE_CONNECT_Ok_AUTH_NOk, Event: EVT_HEARTBEAT, NextState: STATE_CONNECT_Ok_AUTH_NOk},

		// 降级
		{CurrentState: STATE_CONNECT_Ok_AUTH_Ok, Event: EVT_CONNECT_FAIL, NextState: STATE_CONNECT_NOK},
		{CurrentState: STATE_CONNECT_Ok_AUTH_NOk, Event: EVT_CONNECT_FAIL, NextState: STATE_CONNECT_NOK},
		{CurrentState: STATE_CONNECT_Ok_BIZ_NONE, Event: EVT_CONNECT_FAIL, NextState: STATE_CONNECT_NOK},
	}

	// 创建状态机
	sm = statemachine.NewStateMachine(transitions, STATE_INIT)
)

type Config struct {
	Ver      string
	Auth     string
	HostName string
	HostAddr string
	Para01   int
}

func connectToWebSocketServer(host string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func NewClientConnection(
	conf Config,
	//notify chan int,
	readChanEx chan []byte,
	writeChanEx chan []byte,
) {
	go func() {
		var errConn chan struct{}
		rand.Seed(time.Now().Unix())
		maxRetryAttempts := 9999
		retryDelayBase := 1
		maxWaitTimeSecond := 30
		retryDelayMultiplier := 1

		for retry := 1; retry <= maxRetryAttempts; retry++ {
			log.Println("[NewConnection] new errConn and closeOnce")
			errConn = make(chan struct{})
			var closeOnce sync.Once

			//notify <- EVT_CONNECT_FAIL
			sm.HandleEvent(EVT_CONNECT_FAIL)
			c, err := connectToWebSocketServer(conf.HostAddr)
			if err != nil {
				log.Println("Failed to connect:", err)
				if c != nil {
					c.Close()
				}
				Wait(retryDelayBase, retry, retryDelayMultiplier, maxWaitTimeSecond)
				continue
			}

			sm.HandleEvent(EVT_CONNECT_SUCC)

			go func() {
				defer log.Println("quit current goroutine ReadMessage")

				for {
					_, bytes, err := c.ReadMessage()
					if err != nil {
						log.Println("read err:", err)

						closeOnce.Do(func() {
							close(errConn)
						})

						break
					}

					if len(bytes) == 0 {
						log.Println("len(bytes) == 0")
						continue
					}

					log.Println("[NewConnection goroutine ReadMessage] ReadMessage message:  ", string(bytes))
					readChanX <- bytes
				}

			}()

			heartbeatTicker := time.NewTicker(30 * time.Second)
			go func() {
				defer log.Println("quit current goroutine WriteMessage")
				for {
					select {
					case <-heartbeatTicker.C:
						if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
							log.Println("Failed to send ping:", err)
							return
						}
						log.Println("WriteMessage PingMessage ok")
					case write := <-writeChanX:
						err := c.WriteMessage(websocket.TextMessage, write)
						if err != nil {
							log.Println("write:", err)

							closeOnce.Do(func() {
								close(errConn)
							})

							break
						}
					case write := <-writeChanEx:
						err := c.WriteMessage(websocket.TextMessage, write)
						if err != nil {
							log.Println("write:", err)

							closeOnce.Do(func() {
								close(errConn)
							})

							break
						}
					case <-errConn:
						log.Println("func B: case <-errConn:")
						return
					} // end of select
				} // end of for
			}()

			writeChanX <- GetPackageBytes(
				time.Now().UnixMilli(),
				"1.0",
				PACKAGE_TYPE_AUTH,
				//AuthReq{Token: config_task.AuthTokenForDev},
				AuthReq{
					Token:    conf.Auth,
					HostName: conf.HostName,
				},
			) // []byte(conf.Auth)

			<-errConn
			log.Println("<-errConn")
			heartbeatTicker.Stop()
			c.Close()

			sm.HandleEvent(EVT_CONNECT_FAIL)

			retry = 1
			Wait(retryDelayBase, retry, retryDelayMultiplier, maxWaitTimeSecond)
		}
	}()

	go func() { // handle read
		for true {
			bytes, ok := <-readChanX
			if ok != true {
				log.Println("ok != true")
				return
			}

			log.Println("[NewClientConnection] (routine read) string(bytes): ", string(bytes))
			{
				var pack Package
				err := json.Unmarshal(bytes, &pack)
				if err != nil {
					log.Println("Error decoding JSON:", err, " string(bytes): ", string(bytes))
					return
				}

				if pack.Type == PACKAGE_TYPE_AUTH {
					var body AuthResp
					bytes, _ := json.Marshal(pack.Body)
					err = json.Unmarshal(bytes, &body)
					if err != nil {
						log.Println("Error decoding JSON:", err, " string(bytes): ", string(bytes))
						return
					}
					log.Println("[<-readChan] body : ", body)

					writeChanX <- GetPackageBytes(
						time.Now().UnixMilli(),
						"1.0",
						PACKAGE_TYPE_AUTHOK_RECVED,
						BizInit{Para01: conf.Para01},
					)

					sm.HandleEvent(EVT_CONNECT_AUTH_OK)

				} else if pack.Type == PACKAGE_TYPE_BIZ {
					var body BizData
					bytes, _ := json.Marshal(pack.Body)
					err = json.Unmarshal(bytes, &body)
					if err != nil {
						log.Println("Error decoding JSON:", err, " string(bytes): ", string(bytes))
						return
					}
					log.Println("[<-readChan] body : ", body)

					readChanEx <- bytes
				}
			}
		}
	}()
}

func Wait(retryDelayBase int, retry int, retryDelayMultiplier int, maxWaitSecond int) {
	retryDelay := retryDelayBase * (retryDelayMultiplier << uint(retry-1))

	jitter := rand.Intn(100)
	retryDelayMilliseconds := (retryDelay)*1000 + jitter

	if retryDelayMilliseconds > maxWaitSecond*1000 {
		retryDelayMilliseconds = maxWaitSecond * 1000
	}

	log.Printf("waiting %d ms to retry\n", retryDelayMilliseconds)
	time.Sleep(time.Duration(retryDelayMilliseconds) * time.Millisecond)
}
