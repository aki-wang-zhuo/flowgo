package action

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/templatex"
)

const (
	// Type HTTP 客户端节点类型。
	Type = "httpClient"
	// defaultTimeout 默认请求超时。
	defaultTimeout = 10 * time.Second
	// maxBodyBytes 响应体读取上限，防止异常大包拖垮内存。
	maxBodyBytes = 8 << 20 // 8 MiB
)

// Def 面板元数据：分组「动作」；左右锚点，Success / Failure 出边。
var Def = types.ComponentDef{
	Type:           Type,
	Label:          "HTTP 客户端",
	Labels:         map[string]string{types.LocaleEnUS: "HTTP Client"},
	Category:       "action",
	CategoryLabel:  "动作",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Action"},
	Order:          10,
	Color:          "#f9b3a7",
	Icon:           "⇄",
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "发起 HTTP 请求并把响应写入消息；成功走 Success，失败走 Failure。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Send an HTTP request and write the response into the message; Success / Failure edges.",
	},
	Usage: `作为中间动作节点：左入、右出（视觉单出口，可连 Success / Failure）。
configuration:
- method: GET / POST / PUT / DELETE / PATCH，默认 POST
- url: 请求地址（必填），支持 ${msg}、${msg.a.b}、${metadata.xxx} 等模板
- headers: 对象，键值均可模板渲染
- body: 请求体模板；GET 等无体方法可留空；空且非 GET 时可用上游 msg.Data
- timeoutSec: 超时秒数，默认 10
- debugValue: 仅编辑器「运行」调试用的上游 JSON，真实部署/HTTP 入口触发时不会使用
成功时：msg.Data=响应体，metadata.httpStatus=状态码，metadata.httpClientUrl=最终 URL。
网络错误 / 超时 / 模板错误走 Failure。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "method", Type: "string", Default: "POST", Widget: types.WidgetText,
			Description: "HTTP 方法",
			Descriptions: map[string]string{types.LocaleEnUS: "HTTP method"},
		},
		{
			Name: "url", Type: "string", Required: true, Default: "", Widget: types.WidgetText,
			Description: "请求 URL（可模板）",
			Descriptions: map[string]string{types.LocaleEnUS: "Request URL (templates allowed)"},
		},
		{
			Name: "headers", Type: "object", Widget: types.WidgetCodeJSON, Default: "{}",
			Description: "请求头对象",
			Descriptions: map[string]string{types.LocaleEnUS: "Request headers object"},
		},
		{
			Name: "body", Type: "string", Default: "", Widget: types.WidgetCodeJSON,
			Description: "请求体模板",
			Descriptions: map[string]string{types.LocaleEnUS: "Request body template"},
		},
		{
			Name: "timeoutSec", Type: "number", Default: "10", Widget: types.WidgetNumber,
			Description: "超时秒数",
			Descriptions: map[string]string{types.LocaleEnUS: "Timeout in seconds"},
		},
		{
			Name: "debugValue", Type: "string", Default: "{\n  \n}", Widget: types.WidgetCodeJSON,
			Description: "编辑器调试用上游 JSON（部署不用）",
			Descriptions: map[string]string{types.LocaleEnUS: "Upstream JSON for editor debug runs only"},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true, Run: true, RunOnly: true},
}

// HttpClientNode 出站 HTTP 请求节点。
type HttpClientNode struct {
	method     string
	url        string
	headers    map[string]string
	body       string
	timeout    time.Duration
	httpClient *http.Client
}

// New 创建未初始化实例。
func New() types.Node {
	return &HttpClientNode{}
}

// Type 实现 types.Node。
func (n *HttpClientNode) Type() string { return Type }

// Init 解析 configuration（不读取 debugValue，调试体仅编辑器运行时注入消息）。
func (n *HttpClientNode) Init(config map[string]interface{}) error {
	n.method = http.MethodPost
	n.url = ""
	n.headers = map[string]string{}
	n.body = ""
	n.timeout = defaultTimeout

	if config != nil {
		if v, ok := config["method"].(string); ok && strings.TrimSpace(v) != "" {
			n.method = strings.ToUpper(strings.TrimSpace(v))
		}
		if v, ok := config["url"].(string); ok {
			n.url = strings.TrimSpace(v)
		}
		if v, ok := config["body"].(string); ok {
			n.body = v
		}
		if v, ok := config["timeoutSec"]; ok {
			sec, err := toPositiveFloat(v)
			if err != nil {
				return fmt.Errorf("timeoutSec: %w", err)
			}
			if sec > 0 {
				n.timeout = time.Duration(sec * float64(time.Second))
			}
		}
		if raw, ok := config["headers"]; ok && raw != nil {
			h, err := toStringMap(raw)
			if err != nil {
				return fmt.Errorf("headers: %w", err)
			}
			n.headers = h
		}
	}

	if n.url == "" {
		return fmt.Errorf("url is required")
	}
	switch n.method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead:
	default:
		return fmt.Errorf("unsupported method: %s", n.method)
	}

	n.httpClient = &http.Client{Timeout: n.timeout}
	return nil
}

// OnMsg 渲染模板并发起 HTTP 请求，将响应写入消息。
func (n *HttpClientNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	out := msg
	if out.Meta == nil {
		out.Meta = types.Metadata{}
	}

	urlStr, err := templatex.Render(n.url, msg)
	if err != nil {
		return out, types.RelationFailure, fmt.Errorf("url template: %w", err)
	}
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return out, types.RelationFailure, fmt.Errorf("url is empty after render")
	}

	var bodyReader io.Reader
	bodyText := ""
	if n.method != http.MethodGet && n.method != http.MethodHead {
		tpl := strings.TrimSpace(n.body)
		if tpl != "" {
			bodyText, err = templatex.Render(tpl, msg)
			if err != nil {
				return out, types.RelationFailure, fmt.Errorf("body template: %w", err)
			}
		} else if msg.Data != "" {
			bodyText = msg.Data
		}
		if bodyText != "" {
			bodyReader = strings.NewReader(bodyText)
		}
	}

	req, err := http.NewRequestWithContext(ctx, n.method, urlStr, bodyReader)
	if err != nil {
		return out, types.RelationFailure, fmt.Errorf("new request: %w", err)
	}
	for k, v := range n.headers {
		rk, err := templatex.Render(k, msg)
		if err != nil {
			return out, types.RelationFailure, fmt.Errorf("header key template: %w", err)
		}
		rv, err := templatex.Render(v, msg)
		if err != nil {
			return out, types.RelationFailure, fmt.Errorf("header value template: %w", err)
		}
		rk = strings.TrimSpace(rk)
		if rk == "" {
			continue
		}
		req.Header.Set(rk, rv)
	}
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		if looksJSON(bodyText) {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		} else {
			req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		}
	}

	client := n.httpClient
	if client == nil {
		client = &http.Client{Timeout: n.timeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return out, types.RelationFailure, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return out, types.RelationFailure, fmt.Errorf("read body: %w", err)
	}
	if len(raw) > maxBodyBytes {
		return out, types.RelationFailure, fmt.Errorf("response body exceeds %d bytes", maxBodyBytes)
	}

	out.Data = string(raw)
	if looksJSON(out.Data) {
		out.DataType = types.JSON
	} else {
		out.DataType = types.TEXT
	}
	out.Meta["httpStatus"] = strconv.Itoa(resp.StatusCode)
	out.Meta["httpClientUrl"] = urlStr
	out.Meta["httpClientMethod"] = n.method
	return out, types.RelationSuccess, nil
}

// Destroy 释放客户端（无长连接池需关闭）。
func (n *HttpClientNode) Destroy() {
	n.httpClient = nil
}

func toPositiveFloat(v interface{}) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(t), 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", v)
	}
}

func toStringMap(v interface{}) (map[string]string, error) {
	out := map[string]string{}
	switch t := v.(type) {
	case map[string]string:
		for k, val := range t {
			out[k] = val
		}
		return out, nil
	case map[string]interface{}:
		for k, val := range t {
			out[k] = fmt.Sprint(val)
		}
		return out, nil
	case []interface{}:
		// 兼容 [{name,value}] / [{key,value}]
		for i, item := range t {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("headers[%d] must be object", i)
			}
			key := firstString(m, "name", "key")
			val := firstString(m, "value")
			if key == "" {
				continue
			}
			out[key] = val
		}
		return out, nil
	case nil:
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}

func looksJSON(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) > 0 && (s[0] == '{' || s[0] == '[')
}

// ParseDebugValue 读取编辑器调试用上游 JSON；空则 "{}"。真实 OnMsg 不使用此字段。
func ParseDebugValue(configuration map[string]interface{}) string {
	if configuration == nil {
		return "{}"
	}
	v, ok := configuration["debugValue"].(string)
	if !ok {
		return "{}"
	}
	s := strings.TrimSpace(v)
	if s == "" {
		return "{}"
	}
	return s
}
