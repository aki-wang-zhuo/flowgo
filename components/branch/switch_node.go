package branch

import (
	"context"
	"fmt"
	"strings"

	"github.com/expr-lang/expr/vm"
	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/exprx"
)

const (
	// TypeSwitch SWITCH 分支节点类型。
	TypeSwitch = "switch"
)
// CaseDef 一条匹配分支（出边 relation = Value 的字符串形式）。
type CaseDef struct {
	// Value 与表达式结果（转字符串后）比较的值，同时作为出边 relation。
	Value string `json:"value"`
	// Type 编辑器侧值类型：string / number / boolean（可选，仅展示与录入）。
	Type string `json:"type,omitempty"`
	// Name 连线展示名（可选）；空则显示 Value。
	Name string `json:"name,omitempty"`
}

// SwitchDef 面板元数据：分组「分支」。
var SwitchDef = types.ComponentDef{
	Type:           TypeSwitch,
	Label:          "SWITCH 分支",
	Labels:         map[string]string{types.LocaleEnUS: "SWITCH Branch"},
	Category:       "branch",
	CategoryLabel:  "分支",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Branch"},
	Order:          20,
	Color:          "#c5cae9", // 与分组「分支」统一
	Icon:           "⇄",
	RelationTypes:  []string{types.RelationDefault},
	Source:         types.ComponentSourceBuiltin,
	Description:    "按表达式结果匹配 cases，未命中走 Default。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Match expression result against cases; fall through to Default.",
	},
	Usage: `configuration 字段：
- expression: 表达式，结果转为字符串后与 cases.value 比较
- cases: [{ "value":"create", "type":"string", "name":"创建" }, ...]
  type 为 string / number / boolean（编辑器用）；匹配时统一按字符串比较。
可用变量同 IF：msg、metadata、msgType、dataType、global。
示例 expression：msg.action 或 global.env 或 metadata.route
命中第一条 value 相等的分支，出边 relation 为该 value；均未命中走 Default。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "expression", Type: "string", Required: true, Default: "msgType",
			// 走代码编辑器高亮/补全；语法为 Go expr，前端对 switch 表达式禁用格式化。
			Widget: types.WidgetCodeJS, Rows: 4,
			Description: "取值表达式",
			Descriptions: map[string]string{types.LocaleEnUS: "Value expression"},
			Hint: "取值表达式使用 Go expr（非 JavaScript）。结果转字符串后与 cases.value 比较；未命中走 Default。变量：msg、metadata、msgType、dataType、global。",
			Hints: map[string]string{
				types.LocaleEnUS: "Value expression uses Go expr (not JavaScript). Result is string-compared to case values; unmatched goes to Default. Vars: msg, metadata, msgType, dataType, global.",
			},
		},
		{
			Name: "cases", Type: "array", Required: true, Widget: types.WidgetCaseList,
			Default: `[{"value":"a","type":"string","name":"分支 A"},{"value":"b","type":"string","name":"分支 B"}]`,
			Description: "匹配分支",
			Descriptions: map[string]string{types.LocaleEnUS: "Match cases"},
			Hint: "动态添加分支：填写匹配值、数据类型与分支名称；出边 relation 使用值的字符串形式。",
			Hints: map[string]string{
				types.LocaleEnUS: "Add cases dynamically: value, data type, and branch name. Edge relation uses the string form of value.",
			},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// SwitchNode SWITCH 分支节点。
type SwitchNode struct {
	expression string
	cases      []CaseDef
	program    *vm.Program
}

// NewSwitch 创建未初始化的 SWITCH 节点。
func NewSwitch() types.Node {
	return &SwitchNode{}
}

// Type 实现 types.Node。
func (n *SwitchNode) Type() string { return TypeSwitch }

// Init 编译表达式并读取 cases。
func (n *SwitchNode) Init(config map[string]interface{}) error {
	exprStr := "msgType"
	var cases []CaseDef
	if config != nil {
		if v, ok := config["expression"].(string); ok && v != "" {
			exprStr = v
		}
		cases = parseCases(config["cases"])
	}
	prog, err := exprx.CompileAny(exprStr)
	if err != nil {
		return fmt.Errorf("switch expression: %w", err)
	}
	n.expression = exprStr
	n.cases = cases
	n.program = prog
	return nil
}

func parseCases(raw interface{}) []CaseDef {
	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return nil
	}
	out := make([]CaseDef, 0, len(arr))
	seen := map[string]bool{}
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		val, _ := m["value"].(string)
		val = strings.TrimSpace(val)
		if val == "" {
			val = strings.TrimSpace(fmt.Sprint(m["value"]))
		}
		if val == "" || val == "<nil>" || val == types.RelationDefault {
			continue
		}
		if seen[val] {
			continue
		}
		seen[val] = true
		name, _ := m["name"].(string)
		name = strings.TrimSpace(name)
		typ, _ := m["type"].(string)
		typ = strings.ToLower(strings.TrimSpace(typ))
		if typ != "string" && typ != "number" && typ != "boolean" {
			typ = ""
		}
		out = append(out, CaseDef{Value: val, Type: typ, Name: name})
	}
	return out
}

// OnMsg 匹配 cases，否则 Default。
func (n *SwitchNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	got, err := exprx.RunToStringEnv(n.program, msg, types.FlowGlobalFrom(ctx))
	if err != nil {
		return msg, types.RelationFailure, err
	}
	for _, c := range n.cases {
		if c.Value == got {
			return msg, c.Value, nil
		}
	}
	return msg, types.RelationDefault, nil
}

// Destroy 实现 types.Node。
func (n *SwitchNode) Destroy() {}

// CaseLabel 连线展示：名称优先，否则 value。
func CaseLabel(c CaseDef) string {
	if n := strings.TrimSpace(c.Name); n != "" {
		return n
	}
	return c.Value
}
