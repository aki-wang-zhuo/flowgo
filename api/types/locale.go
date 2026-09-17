package types

import (
	"strconv"
	"strings"
)

// 支持的界面 / API 语言码（与编辑器 vue-i18n 对齐）。
const (
	LocaleZhCN      = "zh-CN"
	LocaleEnUS      = "en-US"
	DefaultLocale   = LocaleZhCN
)

// NormalizeLocale 将短码或完整码规范为受支持的语言；未知则回退默认。
func NormalizeLocale(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return DefaultLocale
	}
	// 只取首个 tag（若误传入整段 Accept-Language）
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if i := strings.IndexByte(s, ';'); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	switch {
	case s == "zh-cn" || s == "zh" || strings.HasPrefix(s, "zh-"):
		return LocaleZhCN
	case s == "en-us" || s == "en" || strings.HasPrefix(s, "en"):
		return LocaleEnUS
	default:
		return DefaultLocale
	}
}

// ParseAcceptLanguage 解析 Accept-Language，选取 q 最高的受支持语言。
func ParseAcceptLanguage(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return DefaultLocale
	}
	best := DefaultLocale
	bestQ := -1.0
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := part
		q := 1.0
		if i := strings.IndexByte(part, ';'); i >= 0 {
			tag = strings.TrimSpace(part[:i])
			rest := strings.TrimSpace(part[i+1:])
			if j := strings.Index(rest, "q="); j >= 0 {
				if v, err := strconv.ParseFloat(strings.TrimSpace(rest[j+2:]), 64); err == nil {
					q = v
				}
			}
		}
		loc := NormalizeLocale(tag)
		if q > bestQ {
			bestQ = q
			best = loc
		}
	}
	return best
}

// PickI18n 从多语言表取文案；无对应项时回退 fallback（通常为中文默认字段）。
func PickI18n(table map[string]string, locale, fallback string) string {
	locale = NormalizeLocale(locale)
	if table != nil {
		if v := strings.TrimSpace(table[locale]); v != "" {
			return v
		}
	}
	return fallback
}

// LocalizeComponentDef 按语言写入 Label / CategoryLabel / Description / ConfigFields（返回副本）。
func LocalizeComponentDef(d ComponentDef, locale string) ComponentDef {
	locale = NormalizeLocale(locale)
	d.Label = PickI18n(d.Labels, locale, d.Label)
	d.CategoryLabel = PickI18n(d.CategoryLabels, locale, d.CategoryLabel)
	d.Description = PickI18n(d.Descriptions, locale, d.Description)
	if len(d.ConfigFields) > 0 {
		fields := make([]ConfigField, len(d.ConfigFields))
		copy(fields, d.ConfigFields)
		for i := range fields {
			fields[i].Description = PickI18n(fields[i].Descriptions, locale, fields[i].Description)
			fields[i].Hint = PickI18n(fields[i].Hints, locale, fields[i].Hint)
		}
		d.ConfigFields = fields
	}
	return d
}

// LocalizeComponentDefs 批量本地化组件定义。
func LocalizeComponentDefs(defs []ComponentDef, locale string) []ComponentDef {
	out := make([]ComponentDef, len(defs))
	for i := range defs {
		out[i] = LocalizeComponentDef(defs[i], locale)
	}
	return out
}
