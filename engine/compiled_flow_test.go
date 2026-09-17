package engine

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

// countingNode 用于断言 Init 次数的测试节点。
type countingNode struct {
	inits *int32
}

func (n *countingNode) Type() string { return "countingTest" }
func (n *countingNode) Init(config map[string]interface{}) error {
	atomic.AddInt32(n.inits, 1)
	return nil
}
func (n *countingNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	return msg, types.RelationSuccess, nil
}
func (n *countingNode) Destroy() {}

func TestEngine_reusesCompiledFlow(t *testing.T) {
	var inits int32
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "countingTest", Label: "c"}, func() types.Node {
		return &countingNode{inits: &inits}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-cache-1",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "countingTest", Configuration: map[string]interface{}{"v": 1}},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{}`, nil)
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&inits); got != 1 {
		t.Fatalf("inits=%d want 1", got)
	}
	if eng.cacheLen() != 1 {
		t.Fatalf("cacheLen=%d", eng.cacheLen())
	}
}

func TestEngine_fingerprintRebuild(t *testing.T) {
	var inits int32
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "countingTest", Label: "c"}, func() types.Node {
		return &countingNode{inits: &inits}
	})
	eng := NewWithRegistry(reg)

	dsl := &types.FlowDSL{
		ID:        "flow-cache-2",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "countingTest", Configuration: map[string]interface{}{"v": 1}},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{}`, nil)
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	dsl.Nodes[0].Configuration = map[string]interface{}{"v": 2}
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&inits); got != 2 {
		t.Fatalf("inits=%d want 2 after config change", got)
	}
}

func TestEngine_Invalidate(t *testing.T) {
	var inits int32
	reg := NewRegistry()
	reg.Register(types.ComponentDef{Type: "countingTest", Label: "c"}, func() types.Node {
		return &countingNode{inits: &inits}
	})
	eng := NewWithRegistry(reg)
	dsl := &types.FlowDSL{
		ID:        "flow-cache-3",
		EntryNode: "n1",
		Nodes: []types.FlowNode{
			{ID: "n1", Type: "countingTest"},
		},
	}
	msg := types.NewMsg("T", types.JSON, `{}`, nil)
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	eng.Invalidate("flow-cache-3")
	if eng.cacheLen() != 0 {
		t.Fatalf("cacheLen=%d after invalidate", eng.cacheLen())
	}
	if _, err := eng.Execute(context.Background(), dsl, msg); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&inits); got != 2 {
		t.Fatalf("inits=%d want 2 after invalidate+rerun", got)
	}
}

func TestDslFingerprint_ignoresLayout(t *testing.T) {
	a := &types.FlowDSL{
		ID: "f", EntryNode: "n1",
		Nodes: []types.FlowNode{{ID: "n1", Type: "t", X: 1, Y: 2}},
		Edges: []types.FlowEdge{{From: "n1", To: "n2", PointsList: []types.Point{{X: 1, Y: 2}}}},
	}
	b := &types.FlowDSL{
		ID: "f", EntryNode: "n1",
		Nodes: []types.FlowNode{{ID: "n1", Type: "t", X: 99, Y: 88}},
		Edges: []types.FlowEdge{{From: "n1", To: "n2"}},
	}
	fa, err := dslFingerprint(a)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := dslFingerprint(b)
	if err != nil {
		t.Fatal(err)
	}
	if fa != fb {
		t.Fatalf("layout should not affect fingerprint")
	}
}
