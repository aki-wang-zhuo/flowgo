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
// CaseDef 一条匹配分支（出边 relation = Value）。
type CaseDef struct {
	// Value 与表达式结果比较的值，同时作为出边 relation。
	Value string `json:"value"`
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
- cases: [{ "value":"create", "name":"创建" }, ...]
可用变量同 IF：msg、metadata、msgType、dataType。
示例 expression：msg.action 或 metadata.route
命中第一条 value 相等的分支，出边 relation 为该 value；均未命中走 Default。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "expression", Type: "string", Required: true, Default: "msgType",
			Widget: types.WidgetText,
			Description: "取值表达式",
			Descriptions: map[string]string{types.LocaleEnUS: "Value expression"},
		},
		{
			Name: "cases", Type: "array", Required: true, Widget: types.WidgetCodeJSON,
			Default: `[{"value":"a","name":"分支 A"},{"value":"b","name":"分支 B"}]`,
			Description: "匹配分支 [{value,name?}]",
			Descriptions: map[string]string{types.LocaleEnUS: "Cases [{value,name?}]"},
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
		out = append(out, CaseDef{Value: val, Name: name})
	}
	return out
}

// OnMsg 匹配 cases，否则 Default。
func (n *SwitchNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	_ = ctx
	got, err := exprx.RunToString(n.program, msg)
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
