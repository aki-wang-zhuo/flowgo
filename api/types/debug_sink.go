/**
 * 节点调试扩展日志：引擎在开启节点 Debug 时注入 sink，节点可追加 REQUEST 等条目。
 */
package types

import "context"

type debugSinkCtxKey struct{}

// DebugFlow 常用 FlowType 常量（控制台展示）。
const (
	DebugFlowIN       = "IN"
	DebugFlowOUT      = "OUT"
	DebugFlowRequest  = "REQUEST"
	DebugFlowResponse = "RESPONSE"
)

// WithDebugSink 将可写的调试日志切片挂到 context（仅草稿调试轨使用）。
func WithDebugSink(ctx context.Context, sink *[]DebugLog) context.Context {
	if sink == nil {
		return ctx
	}
	return context.WithValue(ctx, debugSinkCtxKey{}, sink)
}

// AppendDebugLog 向当前 context 的调试槽追加一条日志；无槽则忽略。
func AppendDebugLog(ctx context.Context, entry DebugLog) {
	v := ctx.Value(debugSinkCtxKey{})
	if v == nil {
		return
	}
	sink, ok := v.(*[]DebugLog)
	if !ok || sink == nil {
		return
	}
	*sink = append(*sink, entry)
}

// HasDebugSink 当前是否正在采集节点扩展调试日志。
func HasDebugSink(ctx context.Context) bool {
	v := ctx.Value(debugSinkCtxKey{})
	if v == nil {
		return false
	}
	_, ok := v.(*[]DebugLog)
	return ok
}
