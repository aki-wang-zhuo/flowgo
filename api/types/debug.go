package types

// DebugLog 调试运行产生的一条控制台日志（节点开启 debug 时写入）。
type DebugLog struct {
	Ts           int64  `json:"ts"`
	FlowType     string `json:"flowType"` // IN | OUT
	NodeID       string `json:"nodeId"`
	NodeName     string `json:"nodeName,omitempty"`
	RelationType string `json:"relationType,omitempty"`
	Data         string `json:"data,omitempty"`
	Err          string `json:"err,omitempty"`
	DurationMs   int64  `json:"durationMs,omitempty"`
}
