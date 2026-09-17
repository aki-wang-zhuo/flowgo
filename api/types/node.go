package types

import "context"

// Node 是流程节点接口。所有内置 / 扩展节点均实现该接口。
type Node interface {
	// Type 返回节点类型标识，需与 FlowNode.Type 一致。
	Type() string
	// Init 使用节点配置初始化（引擎加载流程时调用一次）。
	Init(config map[string]interface{}) error
	// OnMsg 处理消息，返回下一条消息与出边关系名（如 Success）。
	OnMsg(ctx context.Context, msg Msg) (out Msg, relation string, err error)
	// Destroy 释放资源。
	Destroy()
}

// NodeFactory 用于按类型创建节点实例。
type NodeFactory func() Node
