package components

import (
	"strings"

	"github.com/flowgo/flowgo/api/types"
)

// Category 面板分组定义（空分组也会返回，便于前端占位）。
type Category struct {
	ID    string
	Label string
	// Labels 多语言分组名；缺省回退 Label（中文）。
	Labels map[string]string
	Order  int
	// Color 同组节点统一填充色（面板与画布一致）。
	Color string
}

// Categories 内置分组顺序；新节点通过 Category 字段归入对应组。
// 同组节点共用 Color，勿在单个 Def 上再各自设色。
var Categories = []Category{
	{
		ID: "endpoint", Label: "入口", Order: 5, Color: "#a6bbcf",
		Labels: map[string]string{types.LocaleEnUS: "Endpoint"},
	},
	{
		ID: "branch", Label: "分支", Order: 7, Color: "#c5cae9",
		Labels: map[string]string{types.LocaleEnUS: "Branch"},
	},
	{
		ID: "exit", Label: "出口", Order: 8, Color: "#b8e0d2",
		Labels: map[string]string{types.LocaleEnUS: "Exit"},
	},
	{
		ID: "transform", Label: "转换", Order: 10, Color: "#fdd0a2",
		Labels: map[string]string{types.LocaleEnUS: "Transform"},
	},
	{
		ID: "filter", Label: "过滤", Order: 20, Color: "#e3f2c0",
		Labels: map[string]string{types.LocaleEnUS: "Filter"},
	},
	{
		ID: "action", Label: "动作", Order: 30, Color: "#f9b3a7",
		Labels: map[string]string{types.LocaleEnUS: "Action"},
	},
	{
		// 编码未指定分类、或不在标准列表中的节点统一归入此组
		ID: CategoryOther, Label: "其他", Order: 999, Color: "#d1d5db",
		Labels: map[string]string{types.LocaleEnUS: "Other"},
	},
}

// CategoryOther 非标准 / 未指定分类时的统一分组 id。
const CategoryOther = "other"

// IsKnownCategory 是否为内置标准分类（含 other）。
func IsKnownCategory(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	for _, c := range Categories {
		if c.ID == id {
			return true
		}
	}
	return false
}

// NormalizeCategory 将空或非标准分类规范为 other。
func NormalizeCategory(id string) string {
	id = strings.TrimSpace(id)
	if IsKnownCategory(id) {
		return id
	}
	return CategoryOther
}

// CategoryLabel 按 id 取默认（中文）显示名；未知分类原样返回 id。
func CategoryLabel(id string) string {
	return CategoryLabelLocale(id, types.DefaultLocale)
}

// CategoryLabelLocale 按语言取分组显示名。
func CategoryLabelLocale(id, locale string) string {
	for _, c := range Categories {
		if c.ID == id {
			return types.PickI18n(c.Labels, locale, c.Label)
		}
	}
	return id
}

// CategoryColor 按分组 id 取统一颜色。
func CategoryColor(id string) string {
	for _, c := range Categories {
		if c.ID == id {
			return c.Color
		}
	}
	return ""
}

// BuildGroups 将组件定义按 Categories 顺序组装为面板分组（含空组）。
// locale 用于分组名与节点 Label/Description 本地化（defs 原件不会被修改）。
func BuildGroups(defs []types.ComponentDef, locale string) []types.ComponentGroup {
	locale = types.NormalizeLocale(locale)
	defs = types.LocalizeComponentDefs(defs, locale)

	byCat := make(map[string][]types.ComponentDef, len(defs))
	for _, d := range defs {
		d.Category = NormalizeCategory(d.Category)
		if d.Category == CategoryOther {
			d.CategoryLabel = types.PickI18n(
				map[string]string{types.LocaleEnUS: "Other"},
				locale,
				"其他",
			)
		}
		byCat[d.Category] = append(byCat[d.Category], d)
	}
	// 同组内按 Order、Type 排序
	for cat, list := range byCat {
		for i := 1; i < len(list); i++ {
			j := i
			for j > 0 && defLess(list[j], list[j-1]) {
				list[j], list[j-1] = list[j-1], list[j]
				j--
			}
		}
		byCat[cat] = list
	}

	out := make([]types.ComponentGroup, 0, len(Categories))
	for _, c := range Categories {
		items := make([]types.ComponentItem, 0, len(byCat[c.ID]))
		for _, d := range byCat[c.ID] {
			items = append(items, toItem(d, c.Color))
		}
		out = append(out, types.ComponentGroup{
			ID:    c.ID,
			Label: types.PickI18n(c.Labels, locale, c.Label),
			Items: items,
		})
	}
	return out
}

func defLess(a, b types.ComponentDef) bool {
	if a.Order != b.Order {
		return a.Order < b.Order
	}
	return a.Type < b.Type
}

func toItem(d types.ComponentDef, catColor string) types.ComponentItem {
	color := catColor
	if color == "" {
		color = d.Color
	}
	inPorts, outPorts := portsFor(d)
	return types.ComponentItem{
		Type:          d.Type,
		Label:         d.Label,
		Color:         color,
		IconText:      d.Icon,
		DefaultScript: d.DefaultScript,
		Description:   d.Description,
		Actions:       d.Actions,
		ConfigFields:  d.ConfigFields,
		InPorts:       inPorts,
		OutPorts:      outPorts,
	}
}

// portsFor 面板/画布端口数量预览（与模型锚点一致）。
// 入口分组仅出、出口分组仅入；其余默认左右皆有。
func portsFor(d types.ComponentDef) (inPorts, outPorts int) {
	switch d.Category {
	case "endpoint":
		return 0, 1
	case "exit":
		return 1, 0
	}
	switch d.Type {
	case "jsTransform", "httpClient":
		// 视觉单出口；Success / Failure 两条出边共用
		return 1, 1
	default:
		inPorts = 1
		outPorts = 1
		if len(d.RelationTypes) == 0 {
			outPorts = 0
		}
		return inPorts, outPorts
	}
}

// FilterDefs 按类型是否启用过滤（enabled=nil 表示全部启用）。
func FilterDefs(defs []types.ComponentDef, enabled func(typeName string) bool) []types.ComponentDef {
	if enabled == nil {
		return defs
	}
	out := make([]types.ComponentDef, 0, len(defs))
	for _, d := range defs {
		if enabled(d.Type) {
			out = append(out, d)
		}
	}
	return out
}
