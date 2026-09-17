package transform

import (
	"context"
	"strings"
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

func TestJsTransform_passThrough(t *testing.T) {
	n := New().(*JsTransformNode)
	if err := n.Init(map[string]interface{}{"jsScript": DefaultScript}); err != nil {
		t.Fatal(err)
	}
	if !n.passThrough || n.engine != nil {
		t.Fatalf("passThrough=%v engine=%v", n.passThrough, n.engine != nil)
	}
	in := types.NewMsg("T", types.JSON, `{"a":1}`, nil)
	out, rel, err := n.OnMsg(context.Background(), in)
	if err != nil || rel != types.RelationSuccess {
		t.Fatalf("err=%v rel=%s", err, rel)
	}
	if out.Data != in.Data {
		t.Fatalf("data changed: %s", out.Data)
	}
}

func TestJsTransform_extractOrderIds(t *testing.T) {
	n := New().(*JsTransformNode)
	script := `
var data = (msg && msg.data) ? msg.data : {};
var list = data.list || [];
var orderIds = [];
for (var i = 0; i < list.length; i++) {
  if (list[i] && list[i].orderId) { orderIds.push(list[i].orderId); }
}
return {
  msg: { total: data.total != null ? data.total : orderIds.length, orderIds: orderIds },
  metadata: metadata,
  msgType: msgType,
  dataType: dataType
};
`
	if err := n.Init(map[string]interface{}{"jsScript": script}); err != nil {
		t.Fatal(err)
	}
	defer n.Destroy()

	in := types.NewMsg("HTTP", types.JSON, `{
  "code":200,
  "data":{"list":[{"orderId":"a"},{"orderId":"b"}],"total":2}
}`, nil)
	out, rel, err := n.OnMsg(context.Background(), in)
	if err != nil || rel != types.RelationSuccess {
		t.Fatalf("err=%v rel=%s", err, rel)
	}
	if !strings.Contains(out.Data, `"a"`) || !strings.Contains(out.Data, `"b"`) {
		t.Fatalf("unexpected data: %s", out.Data)
	}
	if !strings.Contains(out.Data, `"total":2`) {
		t.Fatalf("missing total: %s", out.Data)
	}
}

func TestJsTransform_jsonCopyIsolation(t *testing.T) {
	n := New().(*JsTransformNode)
	script := `msg.hacked=true; return {msg:msg, metadata:metadata, msgType:msgType, dataType:dataType};`
	if err := n.Init(map[string]interface{}{"jsScript": script}); err != nil {
		t.Fatal(err)
	}
	defer n.Destroy()

	raw := `{"phone":"1"}`
	in := types.NewMsg("T", types.JSON, raw, nil)
	out, _, err := n.OnMsg(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if in.Data != raw {
		t.Fatalf("input Data mutated: %s", in.Data)
	}
	if !strings.Contains(out.Data, "hacked") {
		t.Fatalf("output should contain hacked: %s", out.Data)
	}
}
