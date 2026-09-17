package transform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
必须 return 一个对象，例如：
  return {'msg':msg,'metadata':metadata,'msgType':msgType,'dataType':dataType};
右侧视觉上一个出口：首条出边默认 Success，第二条为 Failure。
脚本/编解码错误走 Failure（有失败边则继续执行，无则中止）。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "jsScript", Type: "string", Required: true, Default: DefaultScript,
			Widget: types.WidgetCodeJS, Rows: 12,
			Description: "Transform 函数体",
			Descriptions: map[string]string{types.LocaleEnUS: "Transform function body"},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true, Run: true, RunOnly: true},
}

// JsTransformNode 使用 goja 对消息做转换。
type JsTransformNode struct {
	script string
	engine *js.Engine
}

// New 创建未初始化的节点实例。
func New() types.Node {
	return &JsTransformNode{}
}

// Type 实现 types.Node。
func (n *JsTransformNode) Type() string { return Type }

// Init 读取 configuration.jsScript 并编译。
func (n *JsTransformNode) Init(config map[string]interface{}) error {
	script := DefaultScript
	if config != nil {
		if v, ok := config["jsScript"].(string); ok && v != "" {
			script = v
		}
	}
	n.script = script
	wrapped := fmt.Sprintf(funcTpl, script)
	eng, err := js.NewEngine(wrapped, 0)
	if err != nil {
		return err
	}
	n.engine = eng
	return nil
}

// OnMsg 执行 Transform，期望返回 map：msg / metadata / msgType / dataType。
func (n *JsTransformNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	if n.engine == nil {
		return msg, types.RelationFailure, errors.New("js engine not initialized")
	}

	payload, err := decodePayload(msg)
	if err != nil {
		return msg, types.RelationFailure, err
	}

	out, err := n.engine.Execute(ctx, funcName, payload, map[string]string(msg.Meta), msg.Type, string(msg.DataType))
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
	if meta, ok := m["metadata"].(map[string]interface{}); ok {
		result.Meta = toStringMap(meta)
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
	}
}

func decodePayload(msg types.Msg) (interface{}, error) {
	if msg.DataType == types.JSON {
		if msg.Data == "" {
			return map[string]interface{}{}, nil
		}
		var v interface{}
		if err := json.Unmarshal([]byte(msg.Data), &v); err != nil {
			return nil, err
		}
		return v, nil
	}
	return msg.Data, nil
}

func encodePayload(v interface{}, prefer types.DataType) (string, types.DataType, error) {
	switch t := v.(type) {
	case string:
		return t, types.TEXT, nil
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(t)
		if err != nil {
			return "", prefer, err
		}
		return string(b), types.JSON, nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return "", prefer, err
		}
		return string(b), types.JSON, nil
	}
}

func toStringMap(in map[string]interface{}) types.Metadata {
	out := types.Metadata{}
	for k, v := range in {
		out[k] = fmt.Sprint(v)
	}
	return out
}
