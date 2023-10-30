package api

type PostPluginTaskStatusReq struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
	Para01 int    `json:"para_01"`
}

type PluginTask struct {
	Id         string `json:"id"`
	TaskId     string `json:"task_id"`
	Msg        string `json:"msg"`
	Cmd        string `json:"cmd"`
	Valid      bool   `json:"valid"`
	TimeoutPre int    `json:"timeout_pre"` // 秒
	TimeoutRun int    `json:"timeout_run"` // 秒
}

const (
	TASK_EVT_REJECT    = 61021021 // 指令被平台拒绝
	TASK_EVT_ACCEPT    = 61021022 // 指令被平台接受
	TASK_EVT_CMDACK    = 61021070 // node收到指令
	TASK_EVT_START     = 61021071 // 指令启动
	TASK_EVT_PREACK    = 61021072 // 指令准备工作完成
	TASK_EVT_HEARTBEAT = 61021077 // 指令运行过程中的心跳
	TASK_EVT_STOPPED   = 61021998 // 指令停止
	TASK_EVT_END       = 61021999 // 指令执行结束（包含成功与失败）

	TASK_EVT_REPORT = 61022001 // 指令执行过程中的信息汇报
)

const ( // 内部可能在此不断重试 所以状态会不断循环迁移
	STATUS_SCHED_INIT      = 86031001 // 指令初始化
	STATUS_SCHED_ANALYZE   = 86031002 // 指令在本地分析
	STATUS_SCHED_SENT      = 86031012 // 指令已发送
	STATUS_SCHED_CMD_ACKED = 86031031 // 指令已收回执
	STATUS_SCHED_PRE_ACKED = 86031041 // 准备动作完成
	STATUS_SCHED_RUNNING   = 86031051 // 运行状态的心跳
	STATUS_SCHED_RUN_END   = 86031099 // 收到运行结束通知(包含正常结束、异常结束)
)

const ( // timeout , bad senario
	SCHED_EVT_TIMEOUT_CMDACK = 70017751
	SCHED_EVT_TIMEOUT_PREACK = 70017761
	SCHED_EVT_TIMEOUT_HB     = 70017771
	SCHED_EVT_TIMEOUT_RUN    = 70017781
)

const (
	ERR_CREAT_CONTAINER = 70032001
)
