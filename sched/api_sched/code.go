package api_sched

const (
	INT_ENABLED  = 1
	INT_DISABLED = 0
)

const (
	BIZ_CODE_INVALID = -1001
	FWK_CODE_INVALID = -1011
)

const (
	SCHED_FWK_CODE_END = 86431099 // 整个生命周期已结束 // 一旦到达  86431099，认为经过 1-N 次尝试，最终 失败/成功。不翻身。
)
