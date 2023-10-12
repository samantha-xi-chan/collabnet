package control_link

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/link/repo_link"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitGinService(addr string) (ee error) {
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	link := r.Group("/api/v1/link")
	{
		link.GET("", GetLinks)
	}

	return r.Run(addr)
}

func GetLinks(c *gin.Context) {
	items, e := repo_link.GetLinkCtl().GetItems()
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "ERR_OTHER: " + e.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: items,
	})
	return
}
