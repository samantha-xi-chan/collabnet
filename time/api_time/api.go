package api_time

const (
	STATUS_TIMER_INITED    = 86001002
	STATUS_TIMER_RUNNING   = 86001003
	STATUS_TIMER_TRIGGERED = 86001004

	STATUS_TIMER_DONE     = 86001899
	STATUS_TIMER_DISABLED = 86001999
)

type PostTimeReq struct {
	Type         int    `json:"type"`
	Holder       string `json:"holder"`
	Desc         string `json:"desc"`
	Timeout      int    `json:"timeout"` // 秒单位
	CallbackAddr string `json:"callback_addr"`
}

type PostTimeResp struct {
	Id string `json:"id"`
}

//
type PatchTimeReq struct {
	Timeout int `json:"timeout"` // 秒单位
}

//
type CallbackReq struct {
	Id      string `json:"id"`
	Type    int    `json:"type"`
	Holder  string `json:"holder"`
	Desc    string `json:"desc"`
	Timeout int    `json:"timeout"` // 秒单位
}
