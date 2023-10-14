package api

type PostPluginTaskStatusReq struct {
	//Id     string `json:"id"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
	Para01 int    `json:"para_01"`
}

type PluginTask struct {
	Id         string `json:"id"`
	Msg        string `json:"msg"`
	Cmd        string `json:"cmd"`
	Valid      bool   `json:"valid"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}

const (
	TASK_EVT_REJECT    = 61021022
	TASK_EVT_CMDACK    = 61021070
	TASK_EVT_START     = 61021071
	TASK_EVT_PREACK    = 61021072
	TASK_EVT_HEARTBEAT = 61021077
	TASK_EVT_STOPPED   = 61021998
	TASK_EVT_END       = 61021999
)
