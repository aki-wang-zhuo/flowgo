/**
 * 全局变量内置节点：无入边/出边，仅配置本流程可访问的 global.xx。
 * 全流程最多一个；执行时不参与 hop，由引擎在启动前解析注入。
 */
package globalvars

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/flowgo/flowgo/api/types"
)

const (
	// Type 节点类型。
	Type = "globalVars"
	// ConfigKey 配置字段名。
	ConfigKey = "variables"
)

// 支持的变量类型（UI / 运行时）。
const (
	VarTypeString  = "string"
	VarTypeNumber  = "number"
	VarTypeBoolean = "boolean"
	VarTypeJSON    = "json"
)

// Def 面板元数据：分组「通用」；无端口。
var Def = types.ComponentDef{
	Type:           Type,
	Label:          "全局变量",
	Labels:         map[string]string{types.LocaleEnUS: "Global Variables"},
	Category:       "common",
	CategoryLabel:  "通用",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Common"},
	Order:          5,
	Icon:           "G",
	RelationTypes:  nil, // 无出边
	Source:         types.ComponentSourceBuiltin,
	Description:    "为本流程定义全局变量，任意节点可通过 global.xx 访问；全流程仅允许一个。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Define flow-scoped globals accessible as global.xx; only one per flow.",
	},
	Usage: `configuration.variables 为数组，每项 {name,type,value}。
type: string | number | boolean | json。
本节点无入边/出边，不参与消息推进；引擎在执行前注入 global。
模板：${global.name}；表达式/JS：global.name。
全流程最多一个本类型节点。`,
	ConfigFields: []types.ConfigField{
		{
			Name: ConfigKey, Type: "array", Required: true,
			Default: `[]`,
			Widget:  types.WidgetVarList,
			Description: "变量列表",
			Descriptions: map[string]string{types.LocaleEnUS: "Variable list"},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// Node 占位节点：不处理消息。
type Node struct{}

// New 工厂。
func New() types.Node { return &Node{} }

func (n *Node) Type() string { return Type }

func (n *Node) Init(config map[string]interface{}) error {
	// 校验配置可解析即可；真正取值在 FromDSL
	_, err := ParseVariables(config)
	return err
}

func (n *Node) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	// 正常不应被执行；若误连边则透传
	return msg, types.RelationSuccess, nil
}

func (n *Node) Destroy() {}

// VarDef 单项变量定义。
type VarDef struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ParseVariables 从节点 configuration 解析并做类型转换。
func ParseVariables(config map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	if config == nil {
		return out, nil
	}
	raw, ok := config[ConfigKey]
	if !ok || raw == nil {
		return out, nil
	}
	list, err := toVarList(raw)
	if err != nil {
		return nil, err
	}
	for i, item := range list {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("variables[%d]: name is required", i)
		}
		if !isIdent(name) {
			return nil, fmt.Errorf("variables[%d]: invalid name %q", i, name)
		}
		typ := strings.ToLower(strings.TrimSpace(item.Type))
		if typ == "" {
			typ = VarTypeString
		}
		val, err := coerceValue(typ, item.Value)
		if err != nil {
			return nil, fmt.Errorf("variables[%d] %s: %w", i, name, err)
		}
		out[name] = val
	}
	return out, nil
}

// FromDSL 提取本流程全局变量；超过一个 globalVars 节点则报错。
func FromDSL(dsl *types.FlowDSL) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	if dsl == nil {
		return out, nil
	}
	var foundID string
	for _, n := range dsl.Nodes {
		if n.Type != Type {
			continue
		}
		if foundID != "" {
			return nil, fmt.Errorf("only one %s node allowed per flow (found %s and %s)", Type, foundID, n.ID)
		}
		foundID = n.ID
		m, err := ParseVariables(n.Configuration)
		if err != nil {
			return nil, fmt.Errorf("node %s: %w", n.ID, err)
		}
		out = m
	}
	return out, nil
}

func toVarList(raw interface{}) ([]VarDef, error) {
	switch v := raw.(type) {
	case []VarDef:
		return v, nil
	case []interface{}:
		out := make([]VarDef, 0, len(v))
		for i, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("variables[%d]: want object", i)
			}
			out = append(out, VarDef{
				Name:  asString(m["name"]),
				Type:  asString(m["type"]),
				Value: m["value"],
			})
		}
		return out, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil, nil
		}
		var list []VarDef
		if err := json.Unmarshal([]byte(s), &list); err != nil {
			return nil, fmt.Errorf("variables json: %w", err)
		}
		return list, nil
	default:
		b, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("variables: unsupported type %T", raw)
		}
		var list []VarDef
		if err := json.Unmarshal(b, &list); err != nil {
			return nil, fmt.Errorf("variables: %w", err)
		}
		return list, nil
	}
}

func coerceValue(typ string, raw interface{}) (interface{}, error) {
	switch typ {
	case VarTypeString:
		return asString(raw), nil
	case VarTypeNumber:
		return asNumber(raw)
	case VarTypeBoolean:
		return asBool(raw)
	case VarTypeJSON:
		return asJSON(raw)
	default:
		return nil, fmt.Errorf("unsupported type %q (want string|number|boolean|json)", typ)
	}
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case json.Number:
		return t.String()
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(b)
	}
}

func asNumber(v interface{}) (float64, error) {
	if v == nil {
		return 0, nil
	}
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case json.Number:
		return t.Float64()
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, nil
		}
		return strconv.ParseFloat(s, 64)
	case bool:
		if t {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

func asBool(v interface{}) (bool, error) {
	if v == nil {
		return false, nil
	}
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		switch s {
		case "", "0", "false", "no", "off":
			return false, nil
		case "1", "true", "yes", "on":
			return true, nil
		default:
			return false, fmt.Errorf("invalid boolean %q", t)
		}
	case float64:
		return t != 0, nil
	case int, int32, int64:
		n, _ := asNumber(t)
		return n != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to boolean", v)
	}
}

func asJSON(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, nil
		}
		var parsed interface{}
		if err := json.Unmarshal([]byte(s), &parsed); err != nil {
			return nil, fmt.Errorf("invalid json: %w", err)
		}
		return parsed, nil
	case map[string]interface{}, []interface{}:
		return t, nil
	default:
		// 数字/布尔等也允许作为 json 标量
		return t, nil
	}
}

func isIdent(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' ||
			(i > 0 && r >= '0' && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}
