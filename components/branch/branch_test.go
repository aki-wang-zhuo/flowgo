package branch_test

import (
	"context"
	"testing"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/components/branch"
)

func TestIfNode(t *testing.T) {
	n := branch.NewIf()
	if err := n.Init(map[string]interface{}{
		"expression": `msg.status == 200`,
	}); err != nil {
		t.Fatal(err)
	}
	msg := types.NewMsg("t", types.JSON, `{"status":200}`, nil)
	_, rel, err := n.OnMsg(context.Background(), msg)
	if err != nil || rel != types.RelationTrue {
		t.Fatalf("want True, got %q err=%v", rel, err)
	}
	msg2 := types.NewMsg("t", types.JSON, `{"status":500}`, nil)
	_, rel, err = n.OnMsg(context.Background(), msg2)
	if err != nil || rel != types.RelationFalse {
		t.Fatalf("want False, got %q err=%v", rel, err)
	}
}

func TestSwitchNode(t *testing.T) {
	n := branch.NewSwitch()
	if err := n.Init(map[string]interface{}{
		"expression": "msg.action",
		"cases": []interface{}{
			map[string]interface{}{"value": "create", "name": "创建"},
			map[string]interface{}{"value": "update"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	msg := types.NewMsg("t", types.JSON, `{"action":"update"}`, nil)
	_, rel, err := n.OnMsg(context.Background(), msg)
	if err != nil || rel != "update" {
		t.Fatalf("want update, got %q err=%v", rel, err)
	}
	msg2 := types.NewMsg("t", types.JSON, `{"action":"other"}`, nil)
	_, rel, err = n.OnMsg(context.Background(), msg2)
	if err != nil || rel != types.RelationDefault {
		t.Fatalf("want Default, got %q err=%v", rel, err)
	}
}
