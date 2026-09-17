package engine

import (
	"context"
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

// TestPublishedRunIgnoresDebugSwitch 发布轨即使节点 Debug=true 也不采集日志。
func TestPublishedRunIgnoresDebugSwitch(t *testing.T) {
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "passTest", Label: "p"}, func() types.Node {
		return &passthroughNode{}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-pub-ignore-debug",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "passTest", Name: "n1", Debug: true},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{"a":1}`, nil)

	_, draftLogs, err := eng.ExecuteFromWithLogsOpts(context.Background(), dsl, "n1", msg, ExecuteOptions{
		CacheTrack: CacheTrackDraft,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(draftLogs) == 0 {
		t.Fatal("draft track should collect debug logs when Debug=true")
	}

	_, pubLogs, err := eng.ExecuteFromWithLogsOpts(context.Background(), dsl, "n1", msg, ExecuteOptions{
		CacheTrack: CacheTrackPublished,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pubLogs) != 0 {
		t.Fatalf("published track must ignore Debug, got logs %#v", pubLogs)
	}
}

// TestPublishedRunSkipsFailureDebugLog 发布轨也不写失败 OUT 调试日志。
func TestPublishedRunSkipsFailureDebugLog(t *testing.T) {
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "failTest", Label: "f"}, func() types.Node {
		return &failNode{}
	})
	reg.Register(types.ComponentDef{Type: "passTest", Label: "p"}, func() types.Node {
		return &passthroughNode{}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-pub-skip-fail-log",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "failTest", Name: "fail-node"},
			{ID: "n2", Type: "passTest"},
		},
		Edges: []types.FlowEdge{
			{From: "n1", To: "n2", Relation: types.RelationFailure},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{}`, nil)
	out, logs, err := eng.ExecuteFromWithLogsOpts(context.Background(), dsl, "n1", msg, ExecuteOptions{
		CacheTrack: CacheTrackPublished,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 0 {
		t.Fatalf("published should not emit failure debug logs, got %#v", logs)
	}
	// metadata.errorMsg 仍应写入，供业务节点使用
	if out.Meta == nil || out.Meta[types.KeyErrorMsg] == "" {
		t.Fatalf("want errorMsg in metadata, got %#v", out.Meta)
	}
}
