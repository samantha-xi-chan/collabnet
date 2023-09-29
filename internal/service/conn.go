package service

import (
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

func connectToWebSocketServer(host string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func NewConnection(HostAddr string, Ping string, readBuf chan []byte, writeBuf chan []byte) {

	go func() {
		var errConn chan struct{}

		rand.Seed(time.Now().Unix())
		maxRetryAttempts := 9999
		retryDelayBase := 1
		maxWaitTimeSecond := 30
		retryDelayMultiplier := 1

		for retry := 1; retry <= maxRetryAttempts; retry++ {
			log.Println("new errConn and closeOnce")
			errConn = make(chan struct{})
			var closeOnce sync.Once

			c, err := connectToWebSocketServer(HostAddr)
			if err != nil {
				log.Println("Failed to connect:", err)
				if c != nil {
					c.Close()
				}
				Wait(retryDelayBase, retry, retryDelayMultiplier, maxWaitTimeSecond)
				continue
			}

			go func() {
				defer log.Println("quit current goroutine A")

				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Println("read:", err)

						closeOnce.Do(func() {
							close(errConn)
						})

						break
					}

					readBuf <- message
					//OnMsgReceived(message)
				}

			}()

			go func() {
				defer log.Println("quit current goroutine B")

				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()

				for {
					select {
					case write := <-writeBuf:
						err := c.WriteMessage(websocket.TextMessage, write)
						if err != nil {
							log.Println("write:", err)

							closeOnce.Do(func() {
								close(errConn)
							})

							break
						}

					case <-ticker.C:
						err := c.WriteMessage(websocket.TextMessage, []byte(Ping))
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

			<-errConn
			log.Println("<-errConn")
			c.Close()

			retry = 1
			Wait(retryDelayBase, retry, retryDelayMultiplier, maxWaitTimeSecond)
		}
	}()
}

func Wait(retryDelayBase int, retry int, retryDelayMultiplier int, maxWaitSecond int) {
	retryDelay := retryDelayBase * (retryDelayMultiplier << uint(retry-1))
	//log.Println("retryDelay: ", retryDelay)
	jitter := rand.Intn(100)
	retryDelayMilliseconds := (retryDelay)*1000 + jitter

	if retryDelayMilliseconds > maxWaitSecond*1000 {
		retryDelayMilliseconds = maxWaitSecond * 1000
	}

	log.Printf("waiting %d ms to retry\n", retryDelayMilliseconds)
	time.Sleep(time.Duration(retryDelayMilliseconds) * time.Millisecond)
}

func OnMsgReceived(msg []byte) (e error) {

	log.Printf("Received: %s\n", msg)

	return
}

func OnMsgToWrite(x []byte) (e error) {

	return
}
