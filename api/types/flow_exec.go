/**
 * 流程执行上下文：当前 flowID / nodeID，供 mqttOut 等查找托管客户端。
 */
package types

import "context"

type flowExecKey struct{}

// FlowExec 单次节点执行时的流程定位信息。
type FlowExec struct {
	FlowID string
	NodeID string
}

// WithFlowExec 写入执行定位（flowID 在流程级设置，nodeID 可在每节点覆盖）。
func WithFlowExec(ctx context.Context, flowID, nodeID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	prev, _ := ctx.Value(flowExecKey{}).(FlowExec)
	if flowID == "" {
		flowID = prev.FlowID
	}
	if nodeID == "" {
		nodeID = prev.NodeID
	}
	return context.WithValue(ctx, flowExecKey{}, FlowExec{FlowID: flowID, NodeID: nodeID})
}

// FlowExecFrom 读取执行定位；无则返回零值。
func FlowExecFrom(ctx context.Context) FlowExec {
	if ctx == nil {
		return FlowExec{}
	}
	v, _ := ctx.Value(flowExecKey{}).(FlowExec)
	return v
}
