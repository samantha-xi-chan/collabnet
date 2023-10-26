package control_task

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/sched/config_sched"
	"collab-net-v2/task/api_task"
	"collab-net-v2/task/service_task"
	"collab-net-v2/util/logrus_wrap"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func InitGinService(ctx context.Context, addr string) (ee error) {
	logger := logrus_wrap.GetContextLogger(ctx)
	log := logger.WithFields(logrus.Fields{
		"method": "InitGinService",
	})

	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	task := r.Group("/api/v1/task")
	{
		task.POST("", PostTask)
		task.GET("", GetTask)
		task.GET("/:id", GetTaskById)
		task.PATCH("/:id", PatchTask)
	}

	log.Println("going to listen on : ", addr)
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

func PatchTask(c *gin.Context) {
	id := c.Param("id")
	log.Println("                         [main] service_task.PatchTask id :  ", id)

	// 判断 id 是否有效

	// 判断是否真实在运行

	e := service_task.PatchTask(id)
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "service_task.NewTask:  " + e.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
	})
	return
}

func GetTask(c *gin.Context) {
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

func GetTaskById(c *gin.Context) { //todo: opt
	id := c.Param("id")
	log.Println("                         [main] GetTaskById id :  ", id)

	// 判断 id 是否有效

	// 判断是否真实在运行
	item, e := service_task.GetTaskById(id)
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "service_task.GetTaskById:  " + e.Error(),
		})
		return
	}

	log.Println("[main] service_task.GetTask . ")
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: item,
	})
	return
}
