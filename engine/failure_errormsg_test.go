package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

// failNode 始终返回 Failure + error，用于验证引擎写入 metadata.errorMsg。
type failNode struct{}

func (n *failNode) Type() string                                      { return "failTest" }
func (n *failNode) Init(config map[string]interface{}) error           { return nil }
func (n *failNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	return msg, types.RelationFailure, errors.New("upstream boom")
}
func (n *failNode) Destroy() {}

// passthroughNode 透传消息，用于断言下游能读到 errorMsg。
type passthroughNode struct{}

func (n *passthroughNode) Type() string                            { return "passTest" }
func (n *passthroughNode) Init(config map[string]interface{}) error { return nil }
func (n *passthroughNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	return msg, types.RelationSuccess, nil
}
func (n *passthroughNode) Destroy() {}

func TestEngine_failureWritesErrorMsg(t *testing.T) {
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "failTest", Label: "f"}, func() types.Node {
		return &failNode{}
	})
	reg.Register(types.ComponentDef{Type: "passTest", Label: "p"}, func() types.Node {
		return &passthroughNode{}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-fail-errormsg",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "failTest", Name: "fail-node"},
			{ID: "n2", Type: "passTest"},
		},
		Edges: []types.FlowEdge{
			{From: "n1", To: "n2", Relation: types.RelationFailure},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{"phone":"1"}`, nil)
	out, logs, err := eng.ExecuteFromWithLogs(context.Background(), dsl, dsl.EntryNode, msg)
	if err != nil {
		t.Fatal(err)
	}
	if out.Meta == nil || out.Meta[types.KeyErrorMsg] != "upstream boom" {
		t.Fatalf("want metadata.errorMsg=upstream boom, got %#v", out.Meta)
	}
	if out.Meta[types.KeyErrorNode] != "fail-node" {
		t.Fatalf("want metadata.errorNode=fail-node, got %#v", out.Meta)
	}
	if out.Meta[types.KeyErrorNodeID] != "n1" {
		t.Fatalf("want metadata.errorNodeId=n1, got %#v", out.Meta)
	}
	// 未开 Debug 也应有带 Err 的 OUT 日志，供编辑器标红
	if len(logs) == 0 || logs[0].Err == "" || logs[0].NodeID != "n1" {
		t.Fatalf("want failure debug log, got %#v", logs)
	}
	// 原消息 Data 仍保留（节点未改写）
	if out.Data != `{"phone":"1"}` {
		t.Fatalf("Data=%q", out.Data)
	}
}
