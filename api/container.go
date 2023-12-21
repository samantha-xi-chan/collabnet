package api

// bind
type Bind struct {
	VolPath string `json:"vol_path"`
	VolId   string `json:"vol_id"`
}

type PostContainerReq struct {
	TaskId         string `json:"task_id"`
	BucketName     string `json:"bucket_name"`
	CbAddr         string `json:"cb_addr"`
	LogRt          bool   `json:"log_rt"`
	CleanContainer bool   `json:"clean_container"`

	//Name   string   `json:"name"`
	Image  string   `json:"image"`
	CmdStr []string `json:"cmd_str"`

	BindIn    []Bind `json:"bind_in"`
	BindOut   []Bind `json:"bind_out"`
	GroupPath string `json:"group_path"`
	ShareDir  string `json:"share_dir"`
}
