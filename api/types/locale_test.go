package types

import "testing"

func TestParseAcceptLanguage(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", DefaultLocale},
		{"zh-CN", LocaleZhCN},
		{"zh", LocaleZhCN},
		{"en-US,en;q=0.9", LocaleEnUS},
		{"en;q=0.8,zh-CN;q=0.9", LocaleZhCN},
		{"fr-FR", DefaultLocale},
	}
	for _, c := range cases {
		if got := ParseAcceptLanguage(c.in); got != c.want {
			t.Fatalf("ParseAcceptLanguage(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestLocalizeComponentDef(t *testing.T) {
	d := ComponentDef{
		Label:       "HTTP请求",
		Labels:      map[string]string{LocaleEnUS: "HTTP Request"},
		Description: "中文说明",
		Descriptions: map[string]string{
			LocaleEnUS: "English desc",
		},
	}
	en := LocalizeComponentDef(d, LocaleEnUS)
	if en.Label != "HTTP Request" || en.Description != "English desc" {
		t.Fatalf("en localize failed: %+v", en)
	}
	zh := LocalizeComponentDef(d, LocaleZhCN)
	if zh.Label != "HTTP请求" || zh.Description != "中文说明" {
		t.Fatalf("zh localize failed: %+v", zh)
	}
	// 原件不被修改
	if d.Label != "HTTP请求" {
		t.Fatal("original mutated")
	}
}
