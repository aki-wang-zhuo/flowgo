package endpoint

import (
	"context"

	"github.com/flowgo/flowgo/api/types"
)

const (
	// Type HTTP 入口节点类型。
	Type = "httpEndpoint"
)

// Def 面板元数据：分组「入口」。
var Def = types.ComponentDef{
	Type:           Type,
	Label:          "HTTP 请求",
	Labels:         map[string]string{types.LocaleEnUS: "HTTP Request"},
	Category:       "endpoint",
	CategoryLabel:  "入口",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Endpoint"},
	Order:          10,
	Color:          "#a6bbcf",
	Icon:           "H",
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "启动 HTTP 服务，按配置的多条路径接收请求并触发流程。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Start an HTTP server and trigger the flow for configured routes.",
	},
	Usage: `configuration 字段：
- server: 监听地址，如 ":8088" 或 "0.0.0.0:8088"
- allowCors: 是否允许跨域
- https: 是否启用 HTTPS（TLS）
- certPem / keyPem: 开启 https 时必填的证书与私钥 PEM 文本
- routers: 数组，每项 { "method":"POST", "path":"/api/demo", "name":"可选展示名" }
本节点无入边，仅右侧出线；连线后选择路径。每个路由对应一条出边，relation 为 "METHOD /path"（如 "POST /api/demo"），连线文案为 name，空则显示 METHOD + path。
保存后服务端按路径匹配，从对应出边目标节点开始执行。请求体写入 msg，方法/路径/Header 写入 metadata。同一 method 下 path 不可重复。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "server", Type: "string", Required: true, Default: ":8088", Widget: types.WidgetText,
			Description: "监听地址", Descriptions: map[string]string{types.LocaleEnUS: "Listen address"},
		},
		{
			Name: "allowCors", Type: "boolean", Default: "true", Widget: types.WidgetSwitch,
			Description: "允许 CORS", Descriptions: map[string]string{types.LocaleEnUS: "Allow CORS"},
		},
		{
			Name: "https", Type: "boolean", Default: "false", Widget: types.WidgetSwitch,
			Description: "启用 HTTPS", Descriptions: map[string]string{types.LocaleEnUS: "Enable HTTPS"},
			Hint: "开启后以 TLS 监听，须填写证书与私钥 PEM",
			Hints: map[string]string{types.LocaleEnUS: "TLS listen; certificate and private key PEM required"},
		},
		{
			Name: "certPem", Type: "string", Widget: types.WidgetTextarea, Rows: 5, ShowIf: "https=true",
			Description: "TLS 证书 PEM", Descriptions: map[string]string{types.LocaleEnUS: "TLS certificate PEM"},
		},
		{
			Name: "keyPem", Type: "string", Widget: types.WidgetTextarea, Rows: 5, ShowIf: "https=true",
			Description: "TLS 私钥 PEM", Descriptions: map[string]string{types.LocaleEnUS: "TLS private key PEM"},
		},
		{
			Name: "routers", Type: "array", Required: true, Widget: types.WidgetCodeJSON,
			Default: `[{"method":"POST","path":"/api/demo"}]`,
			Description: "路由列表 [{method,path,name?,debugValue?}]",
			Descriptions: map[string]string{types.LocaleEnUS: "Routes [{method,path,name?,debugValue?}]"},
		},
	},
	// 入口由路径连线触发，仅保留编辑 / 删除
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// HttpEndpointNode HTTP 入口节点。
// 实际监听由 flowgo-server 的 EndpointManager 根据配置拉起；
// 本节点在链路中仅做透传（便于作为入口节点被 Execute）。
type HttpEndpointNode struct {
	cfg HttpConfig
}

// New 创建未初始化实例。
func New() types.Node {
	return &HttpEndpointNode{}
}

// Type 实现 types.Node。
func (n *HttpEndpointNode) Type() string { return Type }

// Init 解析 configuration。
func (n *HttpEndpointNode) Init(config map[string]interface{}) error {
	cfg, err := ParseHttpConfig(config)
	if err != nil {
		return err
	}
	cfg.Server = NormalizeServer(cfg.Server)
	n.cfg = cfg
	return nil
}

// Config 返回已解析配置（供运行时读取）。
func (n *HttpEndpointNode) Config() HttpConfig { return n.cfg }

// OnMsg 透传消息，关系 Success。
func (n *HttpEndpointNode) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	return msg, types.RelationSuccess, nil
}

// Destroy 无本地资源。
func (n *HttpEndpointNode) Destroy() {}
