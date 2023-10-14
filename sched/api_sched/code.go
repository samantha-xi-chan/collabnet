package api_sched

const (
	INT_ENABLED  = 1
	INT_DISABLED = 0
)

const (
	INT_INVALID = -1
	STR_INVALID = "none"
)

const ( // 内部可能在此不断重试 所以状态会不断循环迁移
	STATUS_SCHED_INIT       = 86031001 // 指令初始化
	STATUS_SCHED_LOCAL_FAIL = 86031002 // 指令初始化
	STATUS_SCHED_SENT       = 86031012 // 指令已发送
	STATUS_SCHED_CMD_ACKED  = 86031031 // 指令已收回执
	STATUS_SCHED_PRE_ACKED  = 86031041 // 准备动作完成
	STATUS_SCHED_RUNNING    = 86031051 // 运行状态的心跳
	STATUS_SCHED_END        = 86031099 // 整个生命周期已结束
)

const ( // 一旦到达  86431099，认为经过 1-N 次尝试，最终 失败/成功。不翻身。
	SCHED_FWK_CODE_END = 86431099 // 整个生命周期已结束
)

const (
	SCHED_EVT_TIMEOUT_CMDACK = 70017751
	SCHED_EVT_TIMEOUT_PREACK = 70017761
	SCHED_EVT_TIMEOUT_HB     = 70017771
	SCHED_EVT_TIMEOUT_RUN    = 70017781
)

const (
	SCHED_EVT_TASK_END_OK = 70017701
)
