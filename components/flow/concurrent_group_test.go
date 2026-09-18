package flow_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/engine"
)

func TestConcurrentGroup_AllComplete(t *testing.T) {
	dsl := &types.FlowDSL{
		ID:        "t-cg",
		EntryNode: "in",
		Nodes: []types.FlowNode{
			{ID: "in", Type: "inject", Configuration: map[string]interface{}{}},
			{
				ID: "cg", Type: types.TypeConcurrentGroup,
				Configuration: map[string]interface{}{
					"branches":     []map[string]interface{}{{"name": "a"}, {"name": "b"}},
					"completeMode": "all",
					"timeoutSec":   5,
				},
			},
			{ID: "wa", Type: "jsTransform", ParentID: "cg", Configuration: map[string]interface{}{
				"jsScript": "return {msg:{v:'a'},metadata:metadata,msgType:msgType,dataType:dataType};",
			}},
			{ID: "wb", Type: "jsTransform", ParentID: "cg", Configuration: map[string]interface{}{
				"jsScript": "return {msg:{v:'b'},metadata:metadata,msgType:msgType,dataType:dataType};",
			}},
			{ID: "ok", Type: "httpResponse", Configuration: map[string]interface{}{
				"statusCode": 200, "body": `${msg}`,
			}},
		},
		Edges: []types.FlowEdge{
			{From: "in", To: "cg", Relation: types.RelationSuccess},
			{From: "cg", To: "wa", Relation: "a"},
			{From: "cg", To: "wb", Relation: "b"},
			{From: "wa", To: "cg", Relation: types.RelationSuccess},
			{From: "wb", To: "cg", Relation: types.RelationSuccess},
			{From: "cg", To: "ok", Relation: types.RelationSuccess},
		},
	}
	eng := engine.New()
	out, err := eng.Execute(context.Background(), dsl, types.NewMsg("T", types.JSON, `{}`, nil))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(out.Data), &root); err != nil {
		t.Fatal(err)
	}
	branches, _ := root["branches"].(map[string]interface{})
	if branches == nil {
		t.Fatalf("missing branches: %s", out.Data)
	}
	for _, name := range []string{"a", "b"} {
		b, _ := branches[name].(map[string]interface{})
		if b == nil || b["ok"] != true {
			t.Fatalf("branch %s not ok: %#v", name, branches[name])
		}
	}
}
