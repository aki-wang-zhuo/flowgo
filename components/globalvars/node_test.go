package globalvars

import (
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

func TestParseVariables(t *testing.T) {
	m, err := ParseVariables(map[string]interface{}{
		"variables": []interface{}{
			map[string]interface{}{"name": "s", "type": "string", "value": "hi"},
			map[string]interface{}{"name": "n", "type": "number", "value": "3.5"},
			map[string]interface{}{"name": "b", "type": "boolean", "value": true},
			map[string]interface{}{"name": "j", "type": "json", "value": `{"a":1}`},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if m["s"] != "hi" {
		t.Fatalf("s=%v", m["s"])
	}
	if m["n"].(float64) != 3.5 {
		t.Fatalf("n=%v", m["n"])
	}
	if m["b"] != true {
		t.Fatalf("b=%v", m["b"])
	}
	obj := m["j"].(map[string]interface{})
	if obj["a"].(float64) != 1 {
		t.Fatalf("j=%v", m["j"])
	}
}

func TestFromDSL_onlyOne(t *testing.T) {
	dsl := &types.FlowDSL{
		Nodes: []types.FlowNode{
			{ID: "g1", Type: Type, Configuration: map[string]interface{}{
				"variables": []interface{}{
					map[string]interface{}{"name": "x", "type": "string", "value": "1"},
				},
			}},
			{ID: "g2", Type: Type},
		},
	}
	if _, err := FromDSL(dsl); err == nil {
		t.Fatal("want error for two globalVars")
	}
}
