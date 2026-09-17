package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/flowgo/flowgo/api/types"
)

// compiledFlow 已 Init 的流程运行时（节点实例 + 出边表），可跨消息复用。
type compiledFlow struct {
	fingerprint string
	defs        map[string]types.FlowNode
	nodes       map[string]types.Node
	next        map[string]map[string]string
}

// destroy 释放全部节点资源。
func (c *compiledFlow) destroy() {
	if c == nil {
		return
	}
	for _, n := range c.nodes {
		n.Destroy()
	}
}

// dslFingerprint 对执行相关字段做稳定哈希；忽略画布坐标与贝塞尔点。
func dslFingerprint(dsl *types.FlowDSL) (string, error) {
	if dsl == nil {
		return "", fmt.Errorf("flow dsl is nil")
	}
	type edgeFP struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Relation string `json:"relation,omitempty"`
	}
	type nodeFP struct {
		ID            string                 `json:"id"`
		Type          string                 `json:"type"`
		Debug         bool                   `json:"debug,omitempty"`
		Configuration map[string]interface{} `json:"configuration,omitempty"`
	}
	payload := struct {
		ID        string   `json:"id"`
		EntryNode string   `json:"entryNode"`
		Nodes     []nodeFP `json:"nodes"`
		Edges     []edgeFP `json:"edges"`
	}{
		ID:        dsl.ID,
		EntryNode: dsl.EntryNode,
		Nodes:     make([]nodeFP, 0, len(dsl.Nodes)),
		Edges:     make([]edgeFP, 0, len(dsl.Edges)),
	}
	for _, n := range dsl.Nodes {
		payload.Nodes = append(payload.Nodes, nodeFP{
			ID:            n.ID,
			Type:          n.Type,
			Debug:         n.Debug,
			Configuration: n.Configuration,
		})
	}
	for _, e := range dsl.Edges {
		payload.Edges = append(payload.Edges, edgeFP{
			From:     e.From,
			To:       e.To,
			Relation: e.Relation,
		})
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// buildCompiledFlow 创建并 Init 全部节点。
func (e *Engine) buildCompiledFlow(dsl *types.FlowDSL, fingerprint string) (*compiledFlow, error) {
	defs := make(map[string]types.FlowNode, len(dsl.Nodes))
	nodes := make(map[string]types.Node, len(dsl.Nodes))
	for _, def := range dsl.Nodes {
		defs[def.ID] = def
		factoryNode, ok := e.registry.Create(def.Type)
		if !ok {
			for _, n := range nodes {
				n.Destroy()
			}
			return nil, fmt.Errorf("unknown node type: %s", def.Type)
		}
		if err := factoryNode.Init(def.Configuration); err != nil {
			for _, n := range nodes {
				n.Destroy()
			}
			return nil, fmt.Errorf("init node %s: %w", def.ID, err)
		}
		nodes[def.ID] = factoryNode
	}

	next := map[string]map[string]string{}
	for _, edge := range dsl.Edges {
		rel := edge.Relation
		if rel == "" {
			rel = types.RelationSuccess
		}
		if next[edge.From] == nil {
			next[edge.From] = map[string]string{}
		}
		next[edge.From][rel] = edge.To
	}

	return &compiledFlow{
		fingerprint: fingerprint,
		defs:        defs,
		nodes:       nodes,
		next:        next,
	}, nil
}

// getOrCompile 按 flowID + 指纹复用已 Init 流程；指纹变化时替换并 Destroy 旧实例。
func (e *Engine) getOrCompile(dsl *types.FlowDSL) (*compiledFlow, error) {
	fp, err := dslFingerprint(dsl)
	if err != nil {
		return nil, err
	}
	key := dsl.ID
	if key == "" {
		key = fp
	}

	e.mu.RLock()
	cached := e.cache[key]
	e.mu.RUnlock()
	if cached != nil && cached.fingerprint == fp {
		return cached, nil
	}

	compiled, err := e.buildCompiledFlow(dsl, fp)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if old := e.cache[key]; old != nil {
		if old.fingerprint == fp {
			// 并发下另一 goroutine 已装入相同指纹，丢弃刚建的副本
			compiled.destroy()
			return old, nil
		}
		old.destroy()
	}
	if e.cache == nil {
		e.cache = map[string]*compiledFlow{}
	}
	e.cache[key] = compiled
	return compiled, nil
}

// Invalidate 丢弃指定流程的已编译缓存（保存/删除 DSL 后调用）。
func (e *Engine) Invalidate(flowID string) {
	if flowID == "" {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if old := e.cache[flowID]; old != nil {
		old.destroy()
		delete(e.cache, flowID)
	}
}

// InvalidateAll 清空全部已编译流程缓存。
func (e *Engine) InvalidateAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	for id, old := range e.cache {
		old.destroy()
		delete(e.cache, id)
	}
}

// cacheLen 测试用：当前缓存条目数。
func (e *Engine) cacheLen() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.cache)
}
