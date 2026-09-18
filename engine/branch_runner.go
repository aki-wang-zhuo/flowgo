package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/flowgo/flowgo/api/types"
)

// compiledBranchRunner 在已编译流程上跑单条并发子链，直到汇合点。
type compiledBranchRunner struct {
	compiled    *compiledFlow
	collectLogs bool
	appendLog   func(types.DebugLog)
}

func (r *compiledBranchRunner) Next(fromID, relation string) (string, bool) {
	if r == nil || r.compiled == nil {
		return "", false
	}
	m := r.compiled.next[fromID]
	if m == nil {
		return "", false
	}
	to, ok := m[relation]
	return to, ok
}

func (r *compiledBranchRunner) RunBranch(
	ctx context.Context,
	startID, joinOK, joinFail string,
	msg types.Msg,
) (types.BranchOutcome, error) {
	out := types.BranchOutcome{Msg: msg}
	if startID == "" {
		return out, fmt.Errorf("branch start is empty")
	}
	if joinOK == "" || joinFail == "" {
		return out, fmt.Errorf("group join nodes are required")
	}
	curID := startID
	curMsg := msg
	for hop := 0; hop < maxHops; hop++ {
		// 协作取消：仅在节点边界生效，不打断正在执行的 OnMsg
		if err := ctx.Err(); err != nil {
			out.Cancelled = true
			out.OK = false
			out.Error = err.Error()
			out.Msg = curMsg
			return out, nil
		}
		// 下一步若是汇合点：不执行汇合节点，直接结束本分支
		if curID == joinOK {
			out.OK = true
			out.Relation = types.RelationSuccess
			out.Msg = curMsg
			return out, nil
		}
		if curID == joinFail {
			out.OK = false
			out.Relation = types.RelationFailure
			out.Msg = curMsg
			if out.Error == "" && curMsg.Meta != nil {
				out.Error = curMsg.Meta[types.KeyErrorMsg]
			}
			return out, nil
		}

		node, ok := r.compiled.nodes[curID]
		if !ok {
			return out, fmt.Errorf("node not found: %s", curID)
		}
		def := r.compiled.defs[curID]
		name := def.Name
		if name == "" {
			name = def.Type
		}
		if r.collectLogs && def.Debug && r.appendLog != nil {
			r.appendLog(types.DebugLog{
				Ts:       time.Now().UnixMilli(),
				FlowType: types.DebugFlowIN,
				NodeID:   curID,
				NodeName: name,
				Data:     curMsg.Data,
			})
		}
		started := time.Now()
		var nodeExtras []types.DebugLog
		runCtx := ctx
		if r.collectLogs && def.Debug {
			runCtx = types.WithDebugSink(ctx, &nodeExtras)
		}
		result, relation, err := node.OnMsg(runCtx, curMsg)
		elapsed := time.Since(started).Milliseconds()
		if r.collectLogs && def.Debug && r.appendLog != nil {
			now := time.Now().UnixMilli()
			for i := range nodeExtras {
				if nodeExtras[i].Ts == 0 {
					nodeExtras[i].Ts = now
				}
				nodeExtras[i].NodeID = curID
				if nodeExtras[i].NodeName == "" {
					nodeExtras[i].NodeName = name
				}
				r.appendLog(nodeExtras[i])
			}
			entry := types.DebugLog{
				Ts:           now,
				FlowType:     types.DebugFlowOUT,
				NodeID:       curID,
				NodeName:     name,
				RelationType: relation,
				DurationMs:   elapsed,
			}
			if err != nil {
				entry.Err = err.Error()
				entry.Data = curMsg.Data
			} else {
				entry.Data = result.Data
			}
			r.appendLog(entry)
		} else if r.collectLogs && err != nil && r.appendLog != nil {
			r.appendLog(types.DebugLog{
				Ts:           time.Now().UnixMilli(),
				FlowType:     types.DebugFlowOUT,
				NodeID:       curID,
				NodeName:     name,
				RelationType: relation,
				Err:          err.Error(),
				DurationMs:   elapsed,
			})
		}

		if err != nil {
			to, has := r.compiled.next[curID][relation]
			if has && relation == types.RelationFailure {
				if result.Meta == nil {
					result.Meta = types.Metadata{}
				}
				result.Meta[types.KeyErrorMsg] = err.Error()
				result.Meta[types.KeyErrorNode] = name
				result.Meta[types.KeyErrorNodeID] = curID
				curMsg = result
				curID = to
				continue
			}
			out.OK = false
			out.Relation = types.RelationFailure
			out.Error = err.Error()
			out.Msg = curMsg
			return out, nil
		}
		curMsg = result
		to, has := r.compiled.next[curID][relation]
		if !has {
			// 子链断头且未进汇合：视为失败
			out.OK = false
			out.Relation = types.RelationFailure
			out.Error = fmt.Sprintf("branch ended at %s without reaching group join", curID)
			out.Msg = curMsg
			return out, nil
		}
		// 若下一跳是汇合点，下一轮循环开头会捕获
		curID = to
	}
	return out, fmt.Errorf("branch exceeded max hops (%d)", maxHops)
}
