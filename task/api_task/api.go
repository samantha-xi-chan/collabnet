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
