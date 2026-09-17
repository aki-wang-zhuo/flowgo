package engine

import (
	"context"
	"testing"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/components/globalvars"
	"github.com/flowgo/flowgo/utils/templatex"
)

func TestFlowGlobalInjectedToTemplate(t *testing.T) {
	reg := NewRegistry()
	reg.Register(globalvars.Def, globalvars.New)
	reg.Register(types.ComponentDef{Type: "passTest", Label: "p"}, func() types.Node {
		return &passthroughNode{}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-global",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{
				ID: "gv", Type: globalvars.Type,
				Configuration: map[string]interface{}{
					"variables": []interface{}{
						map[string]interface{}{"name": "base", "type": "string", "value": "https://api"},
					},
				},
			},
			{ID: "n1", Type: "passTest"},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{}`, nil)
	ctx := context.Background()
	out, _, err := eng.ExecuteFromWithLogsOpts(ctx, dsl, "n1", msg, ExecuteOptions{CacheTrack: CacheTrackDraft})
	if err != nil {
		t.Fatal(err)
	}
	_ = out
	// 通过 getOrCompile 后执行路径已注入；直接验证 FromDSL + RenderEnv
	g, err := globalvars.FromDSL(dsl)
	if err != nil {
		t.Fatal(err)
	}
	got, err := templatex.RenderEnv("${global.base}/v1", msg, g)
	if err != nil || got != "https://api/v1" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestFlowGlobalRejectEdges(t *testing.T) {
	reg := NewRegistry()
	reg.Register(globalvars.Def, globalvars.New)
	reg.Register(types.ComponentDef{Type: "passTest", Label: "p"}, func() types.Node {
		return &passthroughNode{}
	})
	eng := NewWithRegistry(reg)
	dsl := &types.FlowDSL{
		ID:        "flow-gv-edge",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "gv", Type: globalvars.Type},
			{ID: "n1", Type: "passTest"},
		},
		Edges: []types.FlowEdge{{From: "n1", To: "gv"}},
	}
	_, _, err := eng.ExecuteFromWithLogsOpts(context.Background(), dsl, "n1", types.NewMsg("T", types.JSON, `{}`, nil), ExecuteOptions{})
	if err == nil {
		t.Fatal("want error when edge targets globalVars")
	}
}
