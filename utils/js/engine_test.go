package js

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestExecute_udfAndGlobal(t *testing.T) {
	cfg := ScriptConfig{
		MaxExecutionTime: DefaultMaxExecTime,
		Properties:       map[string]string{"name": "flowgo"},
		Udf: map[string]interface{}{
			"add": func(a, b int) int { return a + b },
			"isNumberScript": `function isNumber(value){
				return typeof value === "number";
			}`,
		},
	}
	script := `
function Transform(msg, metadata, msgType, dataType) {
	return {
		msg: {
			sum: add(2, 3),
			ok: isNumber(1),
			name: global.name
		},
		metadata: metadata,
		msgType: msgType,
		dataType: dataType
	};
}`
	eng, err := NewEngine(script, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Stop()

	out, err := eng.Execute(context.Background(), "Transform", map[string]interface{}{}, map[string]string{}, "T", "JSON")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("out type %T", out)
	}
	msg, _ := m["msg"].(map[string]interface{})
	if msg["sum"] != int64(5) && msg["sum"] != 5 && msg["sum"] != float64(5) {
		t.Fatalf("sum=%v", msg["sum"])
	}
	if msg["ok"] != true {
		t.Fatalf("ok=%v", msg["ok"])
	}
	if msg["name"] != "flowgo" {
		t.Fatalf("name=%v", msg["name"])
	}
}

func TestExecute_timeout(t *testing.T) {
	cfg := ScriptConfig{MaxExecutionTime: 50 * time.Millisecond}
	script := `function Hang(){ while(true){} }`
	eng, err := NewEngine(script, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Stop()

	_, err = eng.Execute(context.Background(), "Hang")
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("want timeout error, got %v", err)
	}
}

func TestExecute_panicToError(t *testing.T) {
	cfg := ScriptConfig{MaxExecutionTime: DefaultMaxExecTime}
	// 调用不存在的函数名会返回普通 error；用非法方式触发 goja 异常即可
	script := `function Boom(){ throw new Error("boom"); }`
	eng, err := NewEngine(script, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Stop()
	_, err = eng.Execute(context.Background(), "Boom")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecute_clearCtxOnReturn(t *testing.T) {
	cfg := ScriptConfig{MaxExecutionTime: DefaultMaxExecTime}
	script := `function Transform(){ return $ctx ? "has" : "nil"; }`
	eng, err := NewEngine(script, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Stop()

	out, err := eng.Execute(context.Background(), "Transform")
	if err != nil {
		t.Fatal(err)
	}
	if out != "has" {
		t.Fatalf("first run out=%v", out)
	}

	// 第二次不传 ctx，池化 VM 上 $ctx 应已被清空为 null/undefined
	out, err = eng.Execute(nil, "Transform")
	if err != nil {
		t.Fatal(err)
	}
	if out != "nil" {
		t.Fatalf("second run should see cleared ctx, out=%v", out)
	}
}

func TestExecute_fromVars(t *testing.T) {
	cfg := ScriptConfig{MaxExecutionTime: DefaultMaxExecTime}
	script := `function Transform(){ return vars.x; }`
	eng, err := NewEngine(script, cfg, map[string]interface{}{
		"vars": map[string]interface{}{"x": "ok"},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := eng.Execute(context.Background(), "Transform")
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("out=%v", out)
	}
}
