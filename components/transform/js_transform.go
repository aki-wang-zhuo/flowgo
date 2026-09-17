package transform

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/js"
)

const (
	// Type 节点类型标识。
	Type = "jsTransform"
	funcName = "Transform"
	funcTpl  = "function Transform(msg, metadata, msgType, dataType) { %s }"
	// DefaultScript 未配置 jsScript 时使用的默认脚本。
	DefaultScript = "return {'msg':msg,'metadata':metadata,'msgType':msgType,'dataType':dataType};"
)

// Def 编辑器面板元数据（与 Type / New 一同注册到 Registry）。
var Def = types.ComponentDef{
	Type:           Type,
	Label:          "JS 转换",
	Labels:         map[string]string{types.LocaleEnUS: "JS Transform"},
	Category:       "transform",
	CategoryLabel:  "转换",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Transform"},
	Order:          10,
	Color:          "#fdd0a2",
	Icon:           "ƒ",
	DefaultScript:  DefaultScript,
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "使用 JavaScript（goja）转换消息；单出口可连 Success / Failure 两条边。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Transform the message with JavaScript (goja); Success / Failure edges share one visual port.",
	},
	Usage: `在 configuration.jsScript 中编写函数体（不要写 function 外壳）。
可用参数：msg（已解析对象或字符串）、metadata（对象）、msgType、dataType。
可访问：global（进程级属性）、vars（节点 configuration.vars）、已注册 UDF。
必须 return 一个对象，例如：
  return {'msg':msg,'metadata':metadata,'msgType':msgType,'dataType':dataType};
默认脚本（或空脚本）走直通，不进入 goja。
debugValue：仅编辑器对本节点点「运行」时作为 msg 入参，真实部署 / 上游触发不读。
右侧视觉上一个出口：首条出边默认 Success，第二条为 Failure。
脚本/编解码错误走 Failure（有失败边则继续执行，无则中止）。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "jsScript", Type: "string", Required: true, Default: DefaultScript,
			Widget: types.WidgetCodeJS, Rows: 12,
			Description: "Transform 函数体",
			Descriptions: map[string]string{types.LocaleEnUS: "Transform function body"},
		},
		{
			Name: "debugValue", Type: "string", Default: "{\n  \n}", Widget: types.WidgetCodeJSON, Rows: 8,
			Description: "测试值：节点「运行」时作为脚本 msg",
			Descriptions: map[string]string{types.LocaleEnUS: "Test JSON used as script msg when Run is clicked"},
			Hint: "仅调试运行使用，真实流程不读此字段。",
			Hints: map[string]string{types.LocaleEnUS: "Editor Run only; not used in live flows."},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true, Run: true},
}

// JsTransformNode 使用 goja 对消息做转换。
type JsTransformNode struct {
	script      string
	passThrough bool
	engine      *js.Engine
}

// New 创建未初始化的节点实例。
func New() types.Node {
	return &JsTransformNode{}
}

// Type 实现 types.Node。
func (n *JsTransformNode) Type() string { return Type }

// Init 读取 configuration.jsScript；默认/空脚本启用直通，否则编译 goja 引擎。
func (n *JsTransformNode) Init(config map[string]interface{}) error {
	script := DefaultScript
	if config != nil {
		if v, ok := config["jsScript"].(string); ok {
			script = v
		}
	}
	trimmed := strings.TrimSpace(script)
	if trimmed == "" || trimmed == DefaultScript {
		n.passThrough = true
		n.script = DefaultScript
		n.engine = nil
		return nil
	}

	n.passThrough = false
	n.script = script
	wrapped := fmt.Sprintf(funcTpl, script)
	eng, err := js.NewEngine(wrapped, js.DefaultConfig(), getVars(config))
	if err != nil {
		return err
	}
	n.engine = eng
	return nil
}

// OnMsg 执行 Transform；直通模式直接转发消息。
func (n *JsTransformNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	if n.passThrough {
		return msg, types.RelationSuccess, nil
	}
	if n.engine == nil {
		return msg, types.RelationFailure, errors.New("js engine not initialized")
	}

	payload, err := decodePayloadForJS(msg)
	if err != nil {
		return msg, types.RelationFailure, err
	}

	meta := map[string]string(msg.Meta)
	if meta == nil {
		meta = map[string]string{}
	}

	out, err := n.engine.Execute(ctx, funcName, payload, meta, msg.Type, string(msg.DataType))
	if err != nil {
		return msg, types.RelationFailure, err
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		return msg, types.RelationFailure, errors.New("jsTransform must return a map")
	}

	result := msg
	if v, ok := m["msgType"].(string); ok && v != "" {
		result.Type = v
	}
	if v, ok := m["dataType"].(string); ok && v != "" {
		result.DataType = types.DataType(v)
	}
	if metaOut, ok := m["metadata"].(map[string]interface{}); ok {
		result.Meta = toStringMap(metaOut)
	} else if metaStr, ok := m["metadata"].(map[string]string); ok {
		result.Meta = types.Metadata(metaStr)
	}
	if raw, exists := m["msg"]; exists {
		data, dt, err := encodePayload(raw, result.DataType)
		if err != nil {
			return msg, types.RelationFailure, err
		}
		result.Data = data
		result.DataType = dt
	}
	return result, types.RelationSuccess, nil
}

// Destroy 释放引擎。
func (n *JsTransformNode) Destroy() {
	if n.engine != nil {
		n.engine.Stop()
		n.engine = nil
	}
}
