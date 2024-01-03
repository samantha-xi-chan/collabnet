package control_workflow

import (
	"collab-net-v2/api"
	"collab-net-v2/link/middleware"
	"collab-net-v2/workflow/api_workflow"
	"collab-net-v2/workflow/config_workflow"
	"collab-net-v2/workflow/repo_workflow"
	"collab-net-v2/workflow/service_workflow"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

func StartHttpServer(listenAddr string) {
	r := gin.Default()
	r.Use(middleware.GetLoggerMiddleware())

	setting := r.Group(config_workflow.UrlPathSetting)
	{
		setting.GET("/:id", GetSettingById)
		setting.PUT("/:id", PutSettingById)
	}
	task := r.Group(config_workflow.UrlPathTask)
	{
		task.GET("", GetTasks)
		task.GET("/:id", GetTaskById)
		task.PATCH("/:id", PatchWfTaskById) //v2.0
	}
	workflow := r.Group(config_workflow.UrlPathWorkflow)
	{
		workflow.GET("/:id", GetWorkflowById)
		workflow.PATCH("/:id", PatchWorkflowById)
		workflow.POST("", PostWorkflow)
		workflow.DELETE("/:id", DeleteWorkflowById)

		workflow.PATCH("/:id/timer", PatchWorkflowTimer)
	}

	workflowV2 := r.Group(config_workflow.UrlPathWorkflowV2)
	{
		workflowV2.GET("/:id", GetWorkflowByIdV2)
	}

	log.Println("listenAddr : ", listenAddr)
	if e := r.Run(listenAddr); e != nil {
		log.Fatal("gin Run e: ", e)
	}
}

func GetSettingById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	item, e := repo_workflow.GetSettingCtl().GetItemByID(id)
	if e != nil {
		log.Println("GetItemsByWorkflowId e: ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	log.Println(".GetSettingCtl().GetItemByID ", item)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
		Data: item,
	})
	return
}

func PutSettingById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_VALUE,
			Msg:  "api.ERR_VALUE: " + err.Error(),
		})
		return
	}

	item := repo_workflow.Setting{
		Id:       id,
		Name:     id,
		CreateAt: time.Now().UnixMilli(),
		Value:    string(body),
	}
	err = repo_workflow.GetSettingCtl().FirstOrCreate(item)
	if err != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_VALUE,
			Msg:  "api.ERR_VALUE: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
	})
	return
}

func GetTasks(c *gin.Context) {
	var query api_workflow.QueryGetTaskReq
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	items, total, e := repo_workflow.GetTaskCtl().GetItemsByWorkflowIdV18(
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
		Data: api_workflow.QueryGetTasksResp{
			QueryGetTasks: items,
			Total:         total,
		},
	})
	return
}

func PatchWfTaskById(c *gin.Context) {
	taskId := c.Param("id")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	log.Println("PatchWfTaskById : ", taskId, "")

	ctx := context.Background()
	e := service_workflow.StopWfTaskById(ctx, taskId)
	if e != nil {
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "error: " + e.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
	})
	return
}

func GetTaskById(c *gin.Context) {
	taskId := c.Param("id")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	item, count, e := repo_workflow.GetTaskCtl().GetTaskRespItemByTaskId(taskId)
	if e != nil {
		log.Println("GetItemsByWorkflowId e: ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL",
		})
		return
	}

	if count != 1 {
		log.Println("count != 1 ")
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "count != 1",
		})
		return
	}

	log.Println(".GetTaskCtl().GetItemByID ", item)

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
		Data: api_workflow.QueryGetTaskResp{
			QueryGetTask: item,
		},
	})
	return
}

func PatchWorkflowById(c *gin.Context) {
	workflowId := c.Param("id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	e := service_workflow.StopWorkflowWrapper(context.Background(), workflowId, api_workflow.ExitCodeWorkflowStoppedByBizCmd)
	if e != nil {
		logrus.Error("service.StopWorkflow(context.Background(): ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_OTHER,
			Msg:  "ERR_OTHER: " + e.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
	})
	return
}

func GetWorkflowById(c *gin.Context) {
	workflowId := c.Param("id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	items, total, e := repo_workflow.GetTaskCtl().GetItemsByWorkflowIdV18(
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
		Data: api_workflow.QueryGetTasksResp{
			QueryGetTasks: items,
			Total:         total,
		},
	})
	return
}

func GetWorkflowByIdV2(c *gin.Context) {
	workflowId := c.Param("id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	items, total, e := repo_workflow.GetTaskCtl().GetItemsByWorkflowIdV18(
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

	itemWorkflow, e := repo_workflow.GetWorkflowCtl().GetItemByID(workflowId)
	dto := api_workflow.WorkflowResp{
		Id:       workflowId,
		Name:     itemWorkflow.Name,
		CreateAt: itemWorkflow.CreateAt,
		StartAt:  itemWorkflow.StartAt,
		EndAt:    itemWorkflow.EndAt,
		Status:   itemWorkflow.Status,
		ExitCode: itemWorkflow.ExitCode,
		Error:    itemWorkflow.Error,
	}

	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "",
		Data: api_workflow.QueryGetWorkflowResp{
			WorkflowResp: dto,
			QueryGetTasksResp: api_workflow.QueryGetTasksResp{
				QueryGetTasks: items,
				Total:         total},
		},
	})
	return
}

func DeleteWorkflowById(c *gin.Context) {
	workflowId := c.Param("id")
	if workflowId == "" {
		c.JSON(http.StatusBadRequest, api.HttpRespBody{
			Code: api.ERR_URL_ID,
			Msg:  "ERR_URL_ID",
		})
		return
	}

	// todo: 删除任务记录， 制品数据

	c.JSON(http.StatusBadRequest, api.HttpRespBody{
		Code: api.ERR_OTHER,
		Msg:  "ERR_OTHER",
	})
	return
}

func PostWorkflow(c *gin.Context) {
	var dto api_workflow.PostWorkflowDagReq
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

func PatchWorkflowTimer(c *gin.Context) { // todo: change to RPC
	workflowId := c.Param("id")
	log.Printf("PatchWorkflowTimer: workflowId = %s \n", workflowId)
	//var dto api_time.CallbackReq
	//if err := c.BindJSON(&dto); err != nil {
	//	logrus.Error("bad request in Post(): ", err)
	//	c.JSON(http.StatusOK, api.HttpRespBody{
	//		Code: api.ERR_FORMAT,
	//		Msg:  "ERR_FORMAT: " + err.Error(),
	//	})
	//	return
	//}
	//log.Printf("PatchTimer: %#v \n", dto)

	e := service_workflow.StopWorkflowWrapper(context.Background(), workflowId, api_workflow.ExitCodeWorkflowStoppedByBizTimeout)
	if e != nil {
		logrus.Error("service_workflow.StopWorkflowWrapper , e:  ", e)
		c.JSON(http.StatusOK, api.HttpRespBody{
			Code: api.ERR_INTERNAL,
			Msg:  "ERR_INTERNAL: " + e.Error(),
		})
		return
	}

	//log.Printf("PatchTimer: %#v \n", dto)
	c.JSON(http.StatusOK, api.HttpRespBody{
		Code: 0,
		Msg:  "ok",
	})
	return
}
