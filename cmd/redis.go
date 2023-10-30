package main

import (
	"context"
)

func main() {
	go func() {
		AcquireEnQueue(context.Background(), "taskid_001")
	}()
	go func() {
		AcquireEnQueue(context.Background(), "taskid_001")
	}()

	select {}
}
