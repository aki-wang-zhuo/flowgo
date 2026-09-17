package nodedocs

import (
	"testing"

	"github.com/flowgo/flowgo/api/types"
)

func TestParseDocFileName(t *testing.T) {
	typeName, locale, ok := ParseDocFileName("httpEndpoint_zh.md")
	if !ok || typeName != "httpEndpoint" || locale != types.LocaleZhCN {
		t.Fatalf("zh: got %q %q %v", typeName, locale, ok)
	}
	typeName, locale, ok = ParseDocFileName("pluginEcho_en.md")
	if !ok || typeName != "pluginEcho" || locale != types.LocaleEnUS {
		t.Fatalf("en: got %q %q %v", typeName, locale, ok)
	}
	if _, _, ok = ParseDocFileName("readme.md"); ok {
		t.Fatal("readme.md should not match")
	}
}

func TestBuiltinMap(t *testing.T) {
	m := BuiltinMap()
	want := []string{"inject", "httpEndpoint", "if", "switch", "httpResponse", "jsTransform", "httpClient"}
	for _, typ := range want {
		loc := m[typ]
		if loc == nil {
			t.Fatalf("missing type %s", typ)
		}
		if loc[types.LocaleZhCN] == "" || loc[types.LocaleEnUS] == "" {
			t.Fatalf("%s missing zh/en body", typ)
		}
	}
}
