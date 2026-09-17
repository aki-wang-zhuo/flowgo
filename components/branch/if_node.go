package branch

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr/vm"
	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/exprx"
)

const (
	// TypeIf IF 分支节点类型。
	TypeIf = "if"
)
// IfDef 面板元数据：分组「分支」。
var IfDef = types.ComponentDef{
	Type:           TypeIf,
	Label:          "IF 分支",
	Labels:         map[string]string{types.LocaleEnUS: "IF Branch"},
	Category:       "branch",
	CategoryLabel:  "分支",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Branch"},
	Order:          10,
	Color:          "#c5cae9",
	Icon:           "?",
	RelationTypes:  []string{types.RelationTrue, types.RelationFalse},
	Source:         types.ComponentSourceBuiltin,
	Description:    "按布尔表达式选择 True / False 出边。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Choose True / False edges based on a boolean expression.",
	},
	Usage: `configuration.expression 为 Go expr 布尔表达式。
可用变量：msg（JSON 已解析对象或文本）、metadata、msgType、dataType。
示例：
  msg.status == 200
  metadata["route"] == "a"
  msg.ok == true
成立走 True，否则走 False。求值错误时返回 Failure 语义由引擎中止（带 err）。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "expression", Type: "string", Required: true, Default: "true",
			Widget: types.WidgetTextarea, Rows: 4,
			Description: "布尔表达式",
			Descriptions: map[string]string{types.LocaleEnUS: "Boolean expression"},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// IfNode IF 分支节点。
type IfNode struct {
	expression string
	program    *vm.Program
}

// NewIf 创建未初始化的 IF 节点。
func NewIf() types.Node {
	return &IfNode{}
}

// Type 实现 types.Node。
func (n *IfNode) Type() string { return TypeIf }

// Init 编译表达式。
func (n *IfNode) Init(config map[string]interface{}) error {
	exprStr := "true"
	if config != nil {
		if v, ok := config["expression"].(string); ok && v != "" {
			exprStr = v
		}
	}
	prog, err := exprx.CompileBool(exprStr)
	if err != nil {
		return fmt.Errorf("if expression: %w", err)
	}
	n.expression = exprStr
	n.program = prog
	return nil
}

// OnMsg 求值后走 True 或 False（err 必须为 nil 才能跟边）。
func (n *IfNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	_ = ctx
	ok, err := exprx.RunBool(n.program, msg)
	if err != nil {
		// 表达式错误：中止链路（与 jsTransform 脚本错误一致）
		return msg, types.RelationFailure, err
	}
	if ok {
		return msg, types.RelationTrue, nil
	}
	return msg, types.RelationFalse, nil
}

// Destroy 实现 types.Node。
func (n *IfNode) Destroy() {}
