package exit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/templatex"
)

const (
	// Type HTTP 响应节点类型。
	Type = "httpResponse"
)

// Def 面板元数据：分组「出口」。
var Def = types.ComponentDef{
	Type:           Type,
	Label:          "HTTP 响应",
	Labels:         map[string]string{types.LocaleEnUS: "HTTP Response"},
	Category:       "exit",
	CategoryLabel:  "出口",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Exit"},
	Order:          10,
	Color:          "#b8e0d2",
	Icon:           "↩",
	RelationTypes:  []string{},
	Source:         types.ComponentSourceBuiltin,
	Description:    "向 HTTP 请求客户端返回响应（状态码与消息体）。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Return an HTTP response (status code and body) to the client.",
	},
	Usage: `作为 HTTP 链路的出口节点：仅有入边、无出边。
- statusCode: HTTP 状态码，默认 200
- body: 可选响应体模板；空则直接返回当前消息数据
  可用占位符：${msg}、${msg.a.b}、${metadata.xxx}、${msgType}、${dataType}
写入 metadata.httpStatus；服务端在整链结束后用最终消息写回客户端。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "statusCode", Type: "number", Default: "200", Widget: types.WidgetNumber,
			Description: "HTTP 状态码",
			Descriptions: map[string]string{types.LocaleEnUS: "HTTP status code"},
		},
		{
			Name: "body", Type: "string", Default: "", Widget: types.WidgetCodeJSON,
			Description: "响应体模板，空则用消息数据",
			Descriptions: map[string]string{types.LocaleEnUS: "Response body template; empty = use message data"},
		},
	},
	// 出口节点仅保留编辑 / 删除
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// HttpResponseNode 将当前消息作为 HTTP 响应返回给客户端。
type HttpResponseNode struct {
	statusCode int
	body       string
}

// New 创建未初始化实例。
func New() types.Node {
	return &HttpResponseNode{statusCode: 200}
}

// Type 实现 types.Node。
func (n *HttpResponseNode) Type() string { return Type }

// Init 解析 configuration。
func (n *HttpResponseNode) Init(config map[string]interface{}) error {
	n.statusCode = 200
	n.body = ""
	if config == nil {
		return nil
	}
	if v, ok := config["statusCode"]; ok {
		code, err := toInt(v)
		if err != nil {
			return fmt.Errorf("statusCode: %w", err)
		}
		if code < 100 || code > 599 {
			return fmt.Errorf("statusCode out of range: %d", code)
		}
		n.statusCode = code
	}
	if v, ok := config["body"].(string); ok {
		n.body = v
	}
	return nil
}

// OnMsg 渲染响应体模板、写入状态码元数据并结束链路。
func (n *HttpResponseNode) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	out := msg
	if out.Meta == nil {
		out.Meta = types.Metadata{}
	}
	out.Meta["httpStatus"] = strconv.Itoa(n.statusCode)

	tpl := strings.TrimSpace(n.body)
	if tpl != "" {
		rendered, err := templatex.Render(tpl, msg)
		if err != nil {
			return out, types.RelationSuccess, fmt.Errorf("body template: %w", err)
		}
		out.Data = rendered
		if looksJSON(out.Data) {
			out.DataType = types.JSON
		} else {
			out.DataType = types.TEXT
		}
	}
	return out, types.RelationSuccess, nil
}

// Destroy 无本地资源。
func (n *HttpResponseNode) Destroy() {}

func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		return strconv.Atoi(strings.TrimSpace(t))
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

func looksJSON(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 0 && (s[0] == '{' || s[0] == '[')
}
