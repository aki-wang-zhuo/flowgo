/**
 * 流程级全局变量（仅本流程可见）。
 * 由 DSL 中唯一的 globalVars 节点解析，经 context 注入模板 / 表达式 / JS。
 */
package types

import "context"

type flowGlobalKey struct{}

// WithFlowGlobal 将本流程全局变量表写入 context（勿传 nil，可用空 map）。
func WithFlowGlobal(ctx context.Context, global map[string]interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if global == nil {
		global = map[string]interface{}{}
	}
	return context.WithValue(ctx, flowGlobalKey{}, global)
}

// HasFlowGlobal 判断 context 是否显式携带了本流程全局变量表。
func HasFlowGlobal(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	return ctx.Value(flowGlobalKey{}) != nil
}

// FlowGlobalFrom 读取本流程全局变量；无则返回空 map（非 nil）。
func FlowGlobalFrom(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return map[string]interface{}{}
	}
	v := ctx.Value(flowGlobalKey{})
	if v == nil {
		return map[string]interface{}{}
	}
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return map[string]interface{}{}
	}
	return m
}
