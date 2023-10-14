package control_task

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/task/api_task"
	"collab-net-v2/task/service_task"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func InitGinService(addr string) (ee error) {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	task := r.Group("/api/v1/task")
	{
		task.POST("", PostTask)
		task.GET("", GetTask)
	}

	return r.Run(addr)
}

func PostTask(c *gin.Context) {
	var dto api_task.PostTaskReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT: " + err.Error(),
		})
		return
	}

	log.Println("PostTask:  ", dto)

	id, e := service_task.NewTask(dto.Name, dto.Cmd, dto.LinkId, config_sched.CMD_ACK_TIMEOUT, dto.TimeoutPre, dto.TimeoutRun)
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "service_task.NewTask:  " + e.Error(),
		})
		return
	}

	log.Println("[main] service_task.NewTask id :  ", id)
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: api_task.PostTaskResp{Id: id},
	})
	return
}

func GetTask(c *gin.Context) {
	log.Println("GetTask:  ")

	arr, e := service_task.GetTask()
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "service_task.NewTask:  " + e.Error(),
		})
		return
	}

	log.Println("[main] service_task.GetTask . ")
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: arr,
	})
	return
}
