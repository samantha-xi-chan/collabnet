package domain

// 业务逻辑层协议
// 外层结构
type BizMsg struct {
	Type      int         `json:"type"`      // 类型ID： 包含 报文变动、用例变动 2种
	TimeStamp int64       `json:"timestamp"` // 毫秒级时间戳
	Payload   interface{} `json:"payload"`   // 对应 PackageChange或CaseChange 类型
}

const (
	CASE_CHANGE    = 8001
	PACKAGE_CHANGE = 8002
)

// 内层结构
type CaseChange struct { // 【用例变动】结构定义
	CaseID         string
	CaseChangeType int
	CaseAttr       string
}

type PackageChange struct { // 【报文变动】结构定义
	Level     int
	Direction int
	Text      string
}

type SessionProperty struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}
