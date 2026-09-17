package action

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

// TestOnMsg_debugRunSkipsBodyTemplate 节点调试运行应发送 debugValue，而不是渲染后的 body 模板。
func TestOnMsg_debugRunSkipsBodyTemplate(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	n := New().(*HttpClientNode)
	if err := n.Init(map[string]interface{}{
		"method": "POST",
		"url":    srv.URL,
		"body":   `{"phone":"${msg.phone}","status":"end"}`,
	}); err != nil {
		t.Fatal(err)
	}

	debugBody := `{"phone":"13657622220","limit":5}`
	msg := types.NewMsg("DEBUG", types.JSON, debugBody, types.Metadata{
		"debug":      "true",
		"httpClient": "true",
	})
	if _, rel, err := n.OnMsg(context.Background(), msg); err != nil {
		t.Fatal(err)
	} else if rel != types.RelationSuccess {
		t.Fatalf("relation=%s", rel)
	}
	if got != debugBody {
		t.Fatalf("debug run body=%s want=%s", got, debugBody)
	}
}

// TestOnMsg_normalRunUsesBodyTemplate 非本节点调试运行仍渲染 body 模板。
func TestOnMsg_normalRunUsesBodyTemplate(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	n := New().(*HttpClientNode)
	if err := n.Init(map[string]interface{}{
		"method": "POST",
		"url":    srv.URL,
		"body":   `{"phone":"${msg.phone}","status":"end"}`,
	}); err != nil {
		t.Fatal(err)
	}

	msg := types.NewMsg("HTTP", types.JSON, `{"phone":"13012345678"}`, types.Metadata{
		"debug": "true",
	})
	if _, rel, err := n.OnMsg(context.Background(), msg); err != nil {
		t.Fatal(err)
	} else if rel != types.RelationSuccess {
		t.Fatalf("relation=%s", rel)
	}
	want := `{"phone":"13012345678","status":"end"}`
	if got != want {
		t.Fatalf("normal body=%s want=%s", got, want)
	}
}
