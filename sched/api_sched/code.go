package api_sched

const (
	INT_ENABLED  = 1
	INT_DISABLED = 0
)

const (
	INT_INVALID = -1
	STR_INVALID = "none"
)

const ( // 一旦到达  86431099，认为经过 1-N 次尝试，最终 失败/成功。不翻身。
	SCHED_FWK_CODE_END = 86431099 // 整个生命周期已结束
)

const (
	SCHED_EVT_TASK_END_OK = 70017701
)
