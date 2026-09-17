package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flowgo/flowgo/api/types"
)

const maxHops = 256

// Engine 最小流程引擎：按 DSL 拓扑串行推进消息。
// 已 Init 的节点按 flowID + DSL 指纹缓存复用，避免每条消息重新编译 goja。
type Engine struct {
	registry *Registry
	mu       sync.RWMutex
	cache    map[string]*compiledFlow
}

// New 使用默认注册表创建引擎。
func New() *Engine {
	return &Engine{
		registry: DefaultRegistry,
		cache:    map[string]*compiledFlow{},
	}
}

// NewWithRegistry 使用自定义注册表。
func NewWithRegistry(r *Registry) *Engine {
	return &Engine{
		registry: r,
		cache:    map[string]*compiledFlow{},
	}
}

// Execute 加载流程并从 entryNode 执行一条消息，返回最终消息。
func (e *Engine) Execute(ctx context.Context, dsl *types.FlowDSL, msg types.Msg) (types.Msg, error) {
	if dsl == nil {
		return msg, fmt.Errorf("flow dsl is nil")
	}
	if dsl.EntryNode == "" {
		return msg, fmt.Errorf("entryNode is required")
	}
	out, _, err := e.ExecuteFromWithLogs(ctx, dsl, dsl.EntryNode, msg)
	return out, err
}

// ExecuteFrom 从指定节点开始执行（用于 HTTP 入口触发后从下一节点推进）。
func (e *Engine) ExecuteFrom(ctx context.Context, dsl *types.FlowDSL, startNode string, msg types.Msg) (types.Msg, error) {
	out, _, err := e.ExecuteFromWithLogs(ctx, dsl, startNode, msg)
	return out, err
}

// ExecuteFromWithLogs 从指定节点执行，并对 Debug=true 的节点采集 IN/OUT 日志。
func (e *Engine) ExecuteFromWithLogs(ctx context.Context, dsl *types.FlowDSL, startNode string, msg types.Msg) (types.Msg, []types.DebugLog, error) {
	var logs []types.DebugLog
	if dsl == nil {
		return msg, logs, fmt.Errorf("flow dsl is nil")
	}
	if startNode == "" {
		return msg, logs, fmt.Errorf("startNode is required")
	}

	compiled, err := e.getOrCompile(dsl)
	if err != nil {
		return msg, logs, err
	}

	curID := startNode
	curMsg := msg
	for hop := 0; hop < maxHops; hop++ {
		node, ok := compiled.nodes[curID]
		if !ok {
			return curMsg, logs, fmt.Errorf("node not found: %s", curID)
		}
		def := compiled.defs[curID]
		name := def.Name
		if name == "" {
			name = def.Type
		}
		if def.Debug {
			logs = append(logs, types.DebugLog{
				Ts:       time.Now().UnixMilli(),
				FlowType: "IN",
				NodeID:   curID,
				NodeName: name,
				Data:     curMsg.Data,
			})
		}
		started := time.Now()
		out, relation, err := node.OnMsg(ctx, curMsg)
		elapsed := time.Since(started).Milliseconds()
		if def.Debug {
			entry := types.DebugLog{
				Ts:           time.Now().UnixMilli(),
				FlowType:     "OUT",
				NodeID:       curID,
				NodeName:     name,
				RelationType: relation,
				DurationMs:   elapsed,
			}
			if err != nil {
				entry.Err = err.Error()
				entry.Data = curMsg.Data
			} else {
				entry.Data = out.Data
			}
			logs = append(logs, entry)
		}
		if err != nil {
			to, ok := compiled.next[curID][relation]
			if ok && relation == types.RelationFailure {
				curMsg = out
				curID = to
				continue
			}
			return curMsg, logs, fmt.Errorf("node %s: %w", curID, err)
		}
		curMsg = out
		to, ok := compiled.next[curID][relation]
		if !ok {
			return curMsg, logs, nil
		}
		curID = to
	}
	return curMsg, logs, fmt.Errorf("flow exceeded max hops (%d), possible cycle", maxHops)
}
