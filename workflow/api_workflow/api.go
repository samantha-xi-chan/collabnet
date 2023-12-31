package api_workflow

type Edge struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Resc  string `json:"resc"`
}

type ResourceQuota struct {
	CPUPerc int `json:"cpu_perc"`
	MemMB   int `json:"mem_mb"`
	DiskMB  int `json:"disk_mb"`
}

type TaskInput struct {
	ImportObjId string `json:"import_obj_id"`
	ImportObjAs string `json:"import_obj_as"`

	Name   string   `json:"name"`
	Image  string   `json:"image"`
	CmdStr []string `json:"cmd_str"`

	SrcDir  string `json:"src_dir,omitempty"`
	SinkDir string `json:"sink_dir,omitempty"`

	Timeout int  `json:"timeout"`
	Remain  bool `json:"remain"`

	CheckExitCode        bool `json:"check_exit_code"`
	ExitOnAnySiblingExit bool `json:"exit_on_any_sibling_exit"`

	//ExpExitCode   int           `json:"exp_exit_code"`
	ResourceQuota ResourceQuota `json:"resource_quota"`

	Concurrent int `json:"concurrent"`
}

type PostWorkflowDagReq struct {
	Name string `json:"name"`
	Desc string `json:"desc"`

	Task     []TaskInput `json:"task"`
	Edge     []Edge      `json:"edge"`
	ShareDir []string    `json:"share_dir"`

	AutoIterate bool `json:"auto_iterate"`
	Timeout     int  `json:"timeout"`
}
type PostWorkflowDagResp struct {
	Id string `json:"id"`
}

type PostWorkflowResp struct {
	Id           string     `json:"id"`
	QueryGetTask []TaskResp `json:"task"`
}

type PatchTaskResp struct {
}

type PatchContainerReq struct {
	Status int `json:"status"`
}
type PatchContainerResp struct {
	Status int `json:"status"`
}

type HttpPatchContainerResp struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data PatchContainerResp `json:"data"`
}

type PostNodeReq struct {
	Url       string `json:"url"`
	TaskQuota int    `json:"task_quota"`
}

type PostNodeResp struct {
	Id string `json:"id"`
}

type QueryGetTaskReq struct {
	WorkflowId string `form:"workflow_id"`
}

type TaskResp struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
	Status  int    `json:"status"`

	ExitCode int `json:"exit_code"`

	ObjId string `json:"obj_id"`

	HostName string `json:"host_name"`
	HostIp   string `json:"host_ip"`
	Carrier  string `json:"carrier"`
	Error    string `json:"error"`
}

type WorkflowResp struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	CreateAt int64  `json:"create_at"`
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`

	Status   int    `json:"status"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
}

type QueryGetTaskResp struct {
	QueryGetTask TaskResp `json:"task"`
}

type QueryGetTasksResp struct {
	QueryGetTasks []TaskResp `json:"task"`
	Total         int64      `json:"total"` // 分页之前的总数
}

type QueryGetWorkflowResp struct {
	WorkflowResp      WorkflowResp      `json:"workflow"`
	QueryGetTasksResp QueryGetTasksResp `json:"task"`
}

type VolItem struct {
	ObjIdOut string
	Url      string
}
