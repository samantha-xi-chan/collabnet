package control_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/workflow/api_workflow"
	repo "collab-net-v2/workflow/repo_workflow"
	"collab-net-v2/workflow/service_workflow"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

func StartHttpServer(listenAddr string) {
	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	task := r.Group("/api/v1/task")
	{
		task.GET("", GetTask)
		task.PATCH("", PatchTask)
	}
	workflow := r.Group("/api/v1/workflow")
	{
		workflow.GET("/:workflow_id", GetWorkflow)
		workflow.PATCH("/:workflow_id", PatchWorkflow)
		workflow.POST("", PostWorkflow)
	}

	log.Println("listenAddr : ", listenAddr)
	if e := r.Run(listenAddr); e != nil {
		log.Fatal("gin Run e: ", e)
	}
}

func GetTask(c *gin.Context) {
	var query api_workflow.QueryGetTaskReq
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	items, total, e := repo.GetTaskCtl().GetItemsByWorkflowId(
		query.WorkflowId,
	)
	if e != nil {
		log.Println("GetItemsByWorkflowId e: ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	log.Println("GetItemsByWorkflowId items, total: ", items, total)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
		Data: api_workflow.QueryGetTaskResp{
			QueryGetTask: items,
			Total:        total,
		},
	})
	return
}

func PatchTask(c *gin.Context) {
	ctx := context.Background()

	var dto api_workflow.PatchTaskReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT",
		})
		return
	}

	log.Println("PatchTaskReq: ", dto)

	e := service_workflow.OnTaskStatusChange(ctx, dto.TaskId, dto.Status, dto.ExitCode)
	if e != nil {
		logrus.Error("bad request in Post(): ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	//
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: api_workflow.PatchTaskResp{},
	})
	return
}

func PatchWorkflow(c *gin.Context) {
	workflowId := c.Param("workflow_id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	e := service_workflow.StopWorkflow(context.Background(), workflowId)
	if e != nil {
		logrus.Error("ervice.StopWorkflow(context.Background(): ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	//log.Println("items, total: ", items, total)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
	})
	return
}
func GetWorkflow(c *gin.Context) {
	workflowId := c.Param("workflow_id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	items, total, e := repo.GetTaskCtl().GetItemsByWorkflowId(
		workflowId,
	)
	if e != nil {
		logrus.Error("GetItemsByWorkflowId: ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	if total == 0 {
		logrus.Error("total == 0: ")
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	log.Println("items, total: ", items, total)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
		Data: api_workflow.QueryGetTaskResp{
			QueryGetTask: items,
			Total:        total,
		},
	})
	return
}

func PostWorkflow(c *gin.Context) {
	var dto api_workflow.PostWorkflowReq
	if err := c.BindJSON(&dto); err != nil {
		logrus.Error("bad request in Post(): ", err)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_FORMAT,
			Msg:  "ERR_FORMAT",
		})
		return
	}

	ctx := context.Background()
	postWorkflowResp, e := service_workflow.PostWorkflow(ctx, dto)
	if e != nil {
		logrus.Error("bad request in Post(): ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
		Data: postWorkflowResp,
	})
	return
}
