package types

import (
	"context"
	"strings"
)

// 并发分组节点类型常量。
const (
	TypeConcurrentGroup = "concurrentGroup"
)

// 组内虚拟汇合点后缀（编译期写入 next 表，不是真实画布节点）。
const (
	JoinOKSuffix   = "__joinOk"
	JoinFailSuffix = "__joinFail"
)

// VirtualJoinOK 并发分组成功汇合虚拟 id。
func VirtualJoinOK(groupID string) string { return groupID + JoinOKSuffix }

// VirtualJoinFail 并发分组失败汇合虚拟 id。
func VirtualJoinFail(groupID string) string { return groupID + JoinFailSuffix }

// IsVirtualJoinOK 是否为成功汇合虚拟 id。
func IsVirtualJoinOK(id string) bool {
	return strings.HasSuffix(id, JoinOKSuffix)
}

// IsVirtualJoinFail 是否为失败汇合虚拟 id。
func IsVirtualJoinFail(id string) bool {
	return strings.HasSuffix(id, JoinFailSuffix)
}

// GroupIDFromVirtualJoin 从虚拟汇合 id 还原分组节点 id；非虚拟则返回空。
func GroupIDFromVirtualJoin(id string) string {
	if IsVirtualJoinOK(id) {
		return strings.TrimSuffix(id, JoinOKSuffix)
	}
	if IsVirtualJoinFail(id) {
		return strings.TrimSuffix(id, JoinFailSuffix)
	}
	return ""
}

// 完成机制。
const (
	CompleteModeAll = "all" // 全部完成
	CompleteModeAny = "any" // 任意完成
)

// BranchOutcome 单条并发线路的汇合结果。
type BranchOutcome struct {
	Name      string `json:"name"`
	OK        bool   `json:"ok"`
	Relation  string `json:"relation,omitempty"`
	Msg       Msg    `json:"msg,omitempty"`
	Error     string `json:"error,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
	Timeout   bool   `json:"timeout,omitempty"`
	Pending   bool   `json:"pending,omitempty"`
}

// BranchRunner 由引擎注入 context，供 concurrentGroup 并发执行组内子链。
type BranchRunner interface {
	// Next 查编译期出边：from + relation → to。
	Next(fromID, relation string) (toID string, ok bool)
	// RunBranch 从 startID 执行直到即将进入 joinOK 或 joinFail（不执行汇合节点本身）。
	// 协作取消：在节点边界检查 ctx；正在执行的 OnMsg 会跑完。
	RunBranch(ctx context.Context, startID, joinOK, joinFail string, msg Msg) (BranchOutcome, error)
}

type branchRunnerCtxKey struct{}

// WithBranchRunner 写入分支执行器。
func WithBranchRunner(ctx context.Context, r BranchRunner) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, branchRunnerCtxKey{}, r)
}

// BranchRunnerFrom 读取分支执行器；无则 nil。
func BranchRunnerFrom(ctx context.Context) BranchRunner {
	if ctx == nil {
		return nil
	}
	r, _ := ctx.Value(branchRunnerCtxKey{}).(BranchRunner)
	return r
}
