package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	OBJ_TYPE_CONTAINER_TASK = 60021031
	OBJ_TYPE_WORKFLOW       = 60021032
	OBJ_TYPE_RAW_TASK       = 60021033
)

// Event represents the structure of the event data
type Event struct {
	ObjType   int    `json:"obj_type"`
	ObjID     string `json:"obj_id"`
	Timestamp int64  `json:"ts"` // ms
	Data      struct {
		Status   int `json:"status"`
		ExitCode int `json:"exit_code"`
	} `json:"data"`
}

type RawTaskEvent struct {
	ObjType   int    `json:"obj_type"`
	ObjID     string `json:"obj_id"`
	Timestamp int64  `json:"ts"` // ms
	Data      struct {
		Evt int `json:"evt"`
	} `json:"data"`
}

func SendObjEvtRequest(url string, event interface{}) {
	log.Printf("SendObjEvtRequest: url = %s, event = #%v\n", url, event)

	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}
