package types

import "time"

// DataType 表示消息载荷的数据类型。
type DataType string

const (
	JSON   DataType = "JSON"
	TEXT   DataType = "TEXT"
	BINARY DataType = "BINARY"
)

// 常用 metadata 键。
const (
	// KeyErrorMsg 节点走 Failure 且返回 error 时，引擎写入的详细错误（可能含内部地址/堆栈）。
	// 仅供调试或内部节点消费；勿直接回写给外部 HTTP 客户端。
	KeyErrorMsg = "errorMsg"
	// KeyErrorNode Failure 时写入的失败节点名称（面板显示名，无内部细节）。
	// 对外响应建议用「${metadata.errorNode}节点失败」一类文案。
	KeyErrorNode = "errorNode"
	// KeyErrorNodeID Failure 时写入的失败节点 id，供编辑器标红定位。
	KeyErrorNodeID = "errorNodeId"
)

// Metadata 消息元数据（字符串键值）。
type Metadata map[string]string

// Msg 是流程引擎中流转的基本消息单元。
type Msg struct {
	ID       string    `json:"id"`
	Ts       time.Time `json:"ts"`
	Type     string    `json:"type"`
	DataType DataType  `json:"dataType"`
	Data     string    `json:"data"`
	Meta     Metadata  `json:"metadata"`
}

// NewMsg 创建一条新消息。
func NewMsg(msgType string, dataType DataType, data string, meta Metadata) Msg {
	if meta == nil {
		meta = Metadata{}
	}
	return Msg{
		ID:       newID(),
		Ts:       time.Now(),
		Type:     msgType,
		DataType: dataType,
		Data:     data,
		Meta:     meta,
	}
}
