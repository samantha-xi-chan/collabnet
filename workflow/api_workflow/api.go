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

	Name    string   `json:"name"`
	Image   string   `json:"image"`
	CmdStr  []string `json:"cmd_str"`
	SrcDir  string   `json:"src_dir,omitempty"`
	SinkDir string   `json:"sink_dir,omitempty"`
	Timeout int      `json:"timeout"`

	CheckExitCode bool `json:"check_exit_code"` /* 1 as true, 0 as false */

	ExpExitCode   int           `json:"exp_exit_code"`
	ResourceQuota ResourceQuota `json:"resource_quota"`
}

type PostWorkflowDagReq struct {
	Task []TaskInput `json:"task"`
	Edge []Edge      `json:"edge"`
}
type PostWorkflowDagResp struct {
	Id string `json:"id"`
}

type PostWorkflowReq struct {
	Task []TaskInput `json:"task"`
	Edge []Edge      `json:"edge"`
}
type PostWorkflowResp struct {
	Id           string     `json:"id"`
	QueryGetTask []TaskResp `json:"task"`
}

//type PatchTaskReq struct {
//	TaskId   string `json:"task_id"`
//	Status   int    `json:"status"`
//	ExitCode int    `json:"exit_code"`
//}
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

	//NodeId      string `json:"node_id"`
	//ContainerId string `json:"container_id"`
	HostName string `json:"host_name"`
	Carrier  string `json:"carrier"`
}

type QueryGetTaskResp struct {
	QueryGetTask []TaskResp `json:"task"`
	Total        int64      `json:"total"` // 分页之前的总数
}

type VolItem struct {
	ObjIdOut string
	Url      string
}
