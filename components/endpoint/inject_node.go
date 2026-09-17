package endpoint

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/components/nodedocs"
)

const (
	// TypeInject 注入执行入口节点类型。
	TypeInject = "inject"
	// DefaultInjectPayload 默认注入 JSON。
	DefaultInjectPayload = "{}"
)

var (
	injectDoc, injectDocs = nodedocs.Pair(nodedocs.InjectZH, nodedocs.InjectEN)
)

// InjectDef 面板元数据：分组「入口」— 注入执行。
var InjectDef = types.ComponentDef{
	Type:          TypeInject,
	Label:         "注入执行",
	Labels:        map[string]string{types.LocaleEnUS: "Inject"},
	Category:      "endpoint",
	CategoryLabel: "入口",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Endpoint"},
	Order:         5,
	Color:         "#a6bbcf",
	Icon:          "↓",
	RelationTypes: []string{types.RelationSuccess},
	Source:        types.ComponentSourceBuiltin,
	Description:   "手动注入一条 JSON 消息并沿 Success 出边执行后续节点。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Manually inject a JSON message and continue along the Success edge.",
	},
	Usage: `configuration.payload 为要注入的 JSON 文本（对象或数组）。
本节点无入边、仅右侧出口；默认最多一条 Success 出边。节点浮动栏「运行」会用 payload 作为消息体并进入下一步（出边浮动栏不提供运行）。
也可作为流程 entryNode：Engine.Execute 时同样会写入 payload。`,
	Doc:  injectDoc,
	Docs: injectDocs,
	ConfigFields: []types.ConfigField{
		{
			Name: "payload", Type: "string", Required: true,
			Default: "{\n  \"hello\": \"world\"\n}",
			Widget: types.WidgetCodeJSON,
			Description: "注入的 JSON 消息体",
			Descriptions: map[string]string{types.LocaleEnUS: "JSON payload to inject"},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true, Run: true},
}

// InjectNode 注入执行节点。
type InjectNode struct {
	payload string
}

// NewInject 创建未初始化实例。
func NewInject() types.Node {
	return &InjectNode{}
}

// Type 实现 types.Node。
func (n *InjectNode) Type() string { return TypeInject }

// Init 读取 configuration.payload。
func (n *InjectNode) Init(config map[string]interface{}) error {
	n.payload = DefaultInjectPayload
	if config == nil {
		return nil
	}
	if v, ok := config["payload"].(string); ok {
		s := strings.TrimSpace(v)
		if s != "" {
			n.payload = s
		}
	}
	return nil
}

// Payload 返回已配置的注入内容。
func (n *InjectNode) Payload() string { return n.payload }

// OnMsg 将 payload 写入消息后走 Success。
func (n *InjectNode) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	out := msg
	body := strings.TrimSpace(n.payload)
	if body == "" {
		body = DefaultInjectPayload
	}
	// 非法 JSON 仍原样注入为 TEXT，便于调试；合法则标为 JSON
	out.Data = body
	out.DataType = types.TEXT
	if json.Valid([]byte(body)) {
		out.DataType = types.JSON
	}
	if out.Type == "" || out.Type == "DEFAULT" {
		out.Type = "INJECT"
	}
	if out.Meta == nil {
		out.Meta = types.Metadata{}
	}
	out.Meta["inject"] = "true"
	return out, types.RelationSuccess, nil
}

// Destroy 无资源。
func (n *InjectNode) Destroy() {}

// ParseInjectPayload 从节点 configuration 读取 payload。
func ParseInjectPayload(config map[string]interface{}) string {
	if config == nil {
		return DefaultInjectPayload
	}
	if v, ok := config["payload"].(string); ok {
		s := strings.TrimSpace(v)
		if s != "" {
			return s
		}
	}
	return DefaultInjectPayload
}
