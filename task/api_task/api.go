package api_task

const (
//HTTP_CODE_UNKNOWN_ERROR = 1999
)

const (
	TRUE  = 1
	FALSE = 0
)

//
type PostTaskReq struct {
	Name       string `json:"name"`
	Cmd        string `json:"cmd"`
	HostName   string `json:"host_name"`
	LinkId     string `json:"link_id"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}
type PostTaskResp struct {
	Id string `json:"id"`
}

const (
	EXIT_CODE_INIT    = 1000
	EXIT_CODE_UNKNOWN = 1001
)

const (
	RAMDOM_NAME_TASK_END = "random_41Wq2yeoY2cjNHMK"
)

const (
	TASK_STATUS_INIT     = 60021001
	TASK_STATUS_QUEUEING = 60021041
	TASK_STATUS_STARTING = 60021051
	TASK_STATUS_RUNNING  = 60021071
	TASK_STATUS_PAUSED   = 60021072

	//TASK_STATUS_PAUSED_WHEN_QUEUEING = 60021042
	//TASK_STATUS_PAUSED_WHEN_RUNNING  = 60021072

	TASK_STATUS_DISABLED = 60021998
	TASK_STATUS_END      = 60021999
)

const (
	CONTAINER_STATUS_NONE     = 70021000
	CONTAINER_STATUS_PULL_IMG = 70021001
	CONTAINER_STATUS_RUN_ING  = 70021006
	CONTAINER_STATUS_STOPPED  = 70021989
	CONTAINER_STATUS_REMOVED  = 70021999
)

const (
	SESSION_STATUS_INIT    = 1001
	SESSION_STATUS_RUNNING = 1006
	SESSION_STATUS_END     = 1999
	SESSION_STATUS_UNKNOWN = 2000
)

const (
	EXIT_CODE_NONE = -1
)

const (
//TOPIC_ALL = "topic_all"
)
