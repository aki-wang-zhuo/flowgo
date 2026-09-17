package transform

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/flowgo/flowgo/api/types"
)

func TestCurrentTime_injectsDataTime(t *testing.T) {
	n := NewCurrentTime()
	if err := n.Init(map[string]interface{}{"timezone": "Asia/Shanghai"}); err != nil {
		t.Fatal(err)
	}
	msg := types.NewMsg("T", types.JSON, `{"a":1}`, nil)
	out, rel, err := n.OnMsg(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}
	if rel != types.RelationSuccess {
		t.Fatalf("rel=%s", rel)
	}
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(out.Data), &root); err != nil {
		t.Fatal(err)
	}
	if root["a"].(float64) != 1 {
		t.Fatalf("lost a: %#v", root)
	}
	raw, ok := root[FieldDataTime].(map[string]interface{})
	if !ok {
		t.Fatalf("missing %s: %#v", FieldDataTime, root)
	}
	for _, k := range []string{"year", "month", "day", "hour", "minute", "second", "millisecond", "timestamp", "timestampMs", "iso", "timezone"} {
		if _, ok := raw[k]; !ok {
			t.Fatalf("missing key %s in %#v", k, raw)
		}
	}
	if raw["timezone"] != "Asia/Shanghai" {
		t.Fatalf("timezone=%v", raw["timezone"])
	}
	// 时间戳应接近现在
	ts := int64(raw["timestamp"].(float64))
	now := time.Now().Unix()
	if ts < now-5 || ts > now+5 {
		t.Fatalf("timestamp=%d now=%d", ts, now)
	}
}

func TestCurrentTime_invalidTimezone(t *testing.T) {
	n := NewCurrentTime()
	if err := n.Init(map[string]interface{}{"timezone": "Not/AZone"}); err == nil {
		t.Fatal("want error")
	}
}

func TestCurrentTime_emptyMsg(t *testing.T) {
	n := NewCurrentTime()
	_ = n.Init(nil)
	out, _, err := n.OnMsg(context.Background(), types.NewMsg("T", types.TEXT, "", nil))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(out.Data), &root); err != nil {
		t.Fatal(err)
	}
	if root[FieldDataTime] == nil {
		t.Fatal("want __dataTime")
	}
}
