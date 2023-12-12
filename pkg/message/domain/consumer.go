package domain

// 消息平台层协议
// 第一层
type WsReq struct {
	Type    int         `json:"type"`
	Version int         `json:"version"`
	Payload interface{} `json:"payload"`
}

// 第二层
type Subscribe struct {
	Topic string `json:"topic"` // 订阅的主题
}
