package templatex

import (
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

func TestRenderMsgAndMeta(t *testing.T) {
	msg := types.NewMsg("HTTP", types.JSON, `{"id":1,"name":"a"}`, types.Metadata{
		"httpMethod": "POST",
	})
	out, err := Render(`{"id":${msg.id},"m":"${metadata.httpMethod}","raw":${msg}}`, msg)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":1,"m":"POST","raw":{"id":1,"name":"a"}}`
	if out != want {
		t.Fatalf("got %s want %s", out, want)
	}
}

func TestRenderPlainPassthrough(t *testing.T) {
	msg := types.NewMsg("HTTP", types.TEXT, "hello", nil)
	out, err := Render("no placeholders", msg)
	if err != nil {
		t.Fatal(err)
	}
	if out != "no placeholders" {
		t.Fatalf("got %q", out)
	}
}
