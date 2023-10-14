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
