package control_time

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/time/api_time"
	"collab-net-v2/time/service_time"
	"context"
	"github.com/gin-gonic/gin"
	//"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func InitTimeHttpService(ctx context.Context, addr string) (ee error) {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	time := r.Group("/api/v1/time")
	{
		time.POST("", PostTime)
		time.PATCH("/:id", PatchTime)
		time.DELETE("/:id", DeleteTime)
	}
	test := r.Group("/api/v1/test")
	{ // curl -X POST http://192.168.36.5:30088/api/v1/time -d '{"holder":"h1","desc":"d11","timeout":20,"type":1,"callback_addr":"http://192.168.18.201:8088/api/v1/test"}'
		test.POST("", PostTest)
	}

	return r.Run(addr)
}

func PatchTime(c *gin.Context) {
	id := c.Param("id")
	log.Println("id:  ", id)

	var dto api_time.PatchTimeReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT: " + err.Error(),
		})
		return
	}

	service_time.RenewTimer(id, dto.Timeout)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: api_time.PostTimeResp{Id: id},
	})
	return
}
func DeleteTime(c *gin.Context) {
	id := c.Param("id")
	log.Println("id:  ", id)

	service_time.DisableTimer(id)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: api_time.PostTimeResp{Id: id},
	})
	return
}

func PostTime(c *gin.Context) {
	var dto api_time.PostTimeReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT: " + err.Error(),
		})
		return
	}

	log.Printf("PostTime: %#v \n", dto)

	id, e := service_time.NewTimer(dto.Timeout, dto.Type, dto.Holder, dto.Desc, dto.CallbackAddr)
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "service_time.NewTimer:  " + e.Error(),
		})
		return
	}

	log.Println("[main] service_time.NewTimer id :  ", id)
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: api_time.PostTimeResp{Id: id},
	})
	return
}

func PostTest(c *gin.Context) {
	var dto api_time.CallbackReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT: " + err.Error(),
		})
		return
	}

	log.Printf("PostTest: %#v \n", dto)
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
	})
	return
}
