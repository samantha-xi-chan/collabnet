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
	Cmd        string `json:"cmd"`       // 表示启动任务对应的 Cmd字符串
	CmdStop    string `json:"cmd_stop"`  // 表示结束任务对应的 Cmd字符串, 超时自动结束以及通过接口结束时 被调用 //v2.0
	HostName   string `json:"host_name"` //
	LinkId     string `json:"link_id"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}
type PostTaskResp struct {
	Id string `json:"id"`
}

type PatchTaskReq struct { // todo: deprecate
	Name       string `json:"name"`
	Cmd        string `json:"cmd"`
	HostName   string `json:"host_name"`
	LinkId     string `json:"link_id"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}
