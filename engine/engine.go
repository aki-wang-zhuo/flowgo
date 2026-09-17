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

// ExecuteFromWithLogs 从指定节点执行并沿出边继续，对 Debug=true 的节点采集 IN/OUT 日志。
func (e *Engine) ExecuteFromWithLogs(ctx context.Context, dsl *types.FlowDSL, startNode string, msg types.Msg) (types.Msg, []types.DebugLog, error) {
	return e.ExecuteFromWithLogsOpts(ctx, dsl, startNode, msg, ExecuteOptions{})
}

// 编译缓存轨：同一 flowID 的草稿与已发布互不覆盖。
const (
	CacheTrackDefault   = "default"
	CacheTrackDraft     = "draft"
	CacheTrackPublished = "published"
)

// ExecuteOptions 控制从某节点执行时的行为。
type ExecuteOptions struct {
	// OnlyStart 为 true 时只执行起始节点一次，不沿出边继续（编辑器「仅运行此节点」）。
	OnlyStart bool
	// CacheTrack 编译缓存槽；空则 CacheTrackDefault。调试用 draft，线上用 published。
	CacheTrack string
}

// ExecuteFromWithLogsOpts 与 ExecuteFromWithLogs 相同，可指定是否只跑起始节点。
func (e *Engine) ExecuteFromWithLogsOpts(ctx context.Context, dsl *types.FlowDSL, startNode string, msg types.Msg, opts ExecuteOptions) (types.Msg, []types.DebugLog, error) {
	var logs []types.DebugLog
	if dsl == nil {
		return msg, logs, fmt.Errorf("flow dsl is nil")
	}
	if startNode == "" {
		return msg, logs, fmt.Errorf("startNode is required")
	}

	compiled, err := e.getOrCompile(dsl, opts.CacheTrack)
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
		} else if err != nil {
			// 未开调试也记录失败 OUT：Failure 边接住后整体无 error，编辑器需靠 logs 标红
			logs = append(logs, types.DebugLog{
				Ts:           time.Now().UnixMilli(),
				FlowType:     "OUT",
				NodeID:       curID,
				NodeName:     name,
				RelationType: relation,
				Err:          err.Error(),
				DurationMs:   elapsed,
			})
		}
		// 仅运行起始节点：执行一次后立即返回，不走下游
		if opts.OnlyStart {
			if err != nil {
				return out, logs, fmt.Errorf("node %s: %w", curID, err)
			}
			return out, logs, nil
		}
		if err != nil {
			to, ok := compiled.next[curID][relation]
			if ok && relation == types.RelationFailure {
				// Failure 下游可取 errorNode（安全：仅节点名）与 errorMsg（详细，勿对外暴露）
				if out.Meta == nil {
					out.Meta = types.Metadata{}
				}
				out.Meta[types.KeyErrorMsg] = err.Error()
				out.Meta[types.KeyErrorNode] = name
				out.Meta[types.KeyErrorNodeID] = curID
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
