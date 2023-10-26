package config

const (
	PLUGIN_SERVICE_IP        = "localhost"
	PLUGIN_SERVICE_PORT      = ":8090"
	PLUGIN_SERVICE_ROUTER    = "/api/v1/task"
	PLUGIN_SERVICE_ROUTER_ID = "/api/v1/task/:id"

	QUEUE_NAME   = "tasks"
	PRIORITY_MAX = 10
	PRIORITY_9   = 9
	PRIORITY_4   = 4
)
