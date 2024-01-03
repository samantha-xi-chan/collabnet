package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Event struct {
	ObjType int         `json:"obj_type"`
	ObjID   string      `json:"obj_id"`
	Data    interface{} `json:"data"`
}

func main() {
	router := gin.Default()

	router.POST("/api/v1/event", handleEvent)

	err := router.Run(":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func handleEvent(c *gin.Context) {
	var event Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle the event as needed, for now just printing the received data
	fmt.Printf("Received Event: %+v\n", event)

	c.JSON(http.StatusOK, gin.H{"message": "Event received successfully"})
}
