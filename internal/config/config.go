package config

const (
	PLUGIN_SERVICE_IP        = "localhost"
	PLUGIN_SERVICE_PORT      = ":8090"
	PLUGIN_SERVICE_ROUTER    = "/api/v1/task"
	PLUGIN_SERVICE_ROUTER_ID = "/api/v1/task/:id"
)

type PluginTask struct {
	Id         string `json:"id"`
	Cmd        string `json:"cmd"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}
