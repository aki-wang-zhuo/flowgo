package types

import "time"

// DataType 表示消息载荷的数据类型。
type DataType string

const (
	JSON   DataType = "JSON"
	TEXT   DataType = "TEXT"
	BINARY DataType = "BINARY"
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
