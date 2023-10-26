package control_link

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/link/repo_link"
	"collab-net-v2/link/service_link"
	"collab-net-v2/util/logrus_wrap"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func InitGinService(ctx context.Context, addr string) (ee error) {
	logger := logrus_wrap.GetContextLogger(ctx)
	log := logger.WithFields(logrus.Fields{
		"method": "InitGinService",
	})

	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	link := r.Group("/api/v1/link")
	{
		link.GET("", func(c *gin.Context) {
			//items, e := repo_link.GetLinkCtl().GetItems(ctx) //("online", 1)
			items, e := service_link.GetNonFirstPartyNodeLinks(ctx)
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
		})

		link.GET(":id", func(c *gin.Context) {
			id := c.Param("id")
			items, e := repo_link.GetLinkCtl().GetItemById(ctx, id) //("online", 1)
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
		})
	}

	log.Println("going to listen on : ", addr)
	return r.Run(addr)
}
