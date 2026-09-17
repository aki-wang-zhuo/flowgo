package types

// 组件来源常量。
const (
	ComponentSourceBuiltin     = "builtin"     // 服务端内置注册
	ComponentSourcePlugin      = "plugin"      // 本地加载的 Go 插件（预留）
	ComponentSourceMarketplace = "marketplace" // 市场安装（预留）
)

// ConfigField 节点 configuration 字段说明（供属性面板动态渲染 / MCP 文档）。
type ConfigField struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string | number | boolean | object | array
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
	// Descriptions 字段说明多语言。
	Descriptions map[string]string `json:"descriptions,omitempty"`
	// Hint 开关等控件旁的补充说明（可选）。
	Hint string `json:"hint,omitempty"`
	// Hints 补充说明多语言。
	Hints map[string]string `json:"hints,omitempty"`
	// Widget 可选 UI 控件；空则按 Type 推断（text / textarea / code-json / code-js / switch / number）。
	Widget string `json:"widget,omitempty"`
	// Rows 多行输入行数（textarea）。
	Rows int `json:"rows,omitempty"`
	// ShowIf 条件显示，格式 "field=value"（如 https=true）；空表示始终显示。
	ShowIf string `json:"showIf,omitempty"`
}

// 常用 Widget 常量。
const (
	WidgetText     = "text"
	WidgetTextarea = "textarea"
	WidgetCodeJSON = "code-json"
	WidgetCodeJS   = "code-js"
	WidgetSwitch   = "switch"
	WidgetNumber   = "number"
)

// NodeActions 画布选中快捷栏能力。
// 默认全部关闭；仅显式 true 的字段会在前端显示对应按钮。
type NodeActions struct {
	Edit    bool `json:"edit,omitempty"`
	Delete  bool `json:"delete,omitempty"`
	Run     bool `json:"run,omitempty"`     // 从此节点运行
	RunOnly bool `json:"runOnly,omitempty"` // 仅运行此节点
}

// ComponentDef 可拖拽组件的元数据（供编辑器面板展示，与执行工厂同源注册）。
// Label / CategoryLabel / Description 为默认文案（中文）；多语言表见 Labels 等字段。
// API / MCP 响应前应调用 LocalizeComponentDef，按 Accept-Language 写入上述展示字段。
type ComponentDef struct {
	Type          string `json:"type"`
	Label         string `json:"label"`
	// Labels 多语言显示名（key: zh-CN / en-US）；不序列化，仅注册与本地化使用。
	Labels        map[string]string `json:"-"`
	Category      string            `json:"category"`      // 分组 id，如 transform
	CategoryLabel string            `json:"categoryLabel"` // 分组显示名，如 转换
	// CategoryLabels 多语言分组名；不序列化。
	CategoryLabels map[string]string `json:"-"`
	Order          int               `json:"order"` // 同组内排序，越小越靠前
	Color          string            `json:"color,omitempty"`
	Icon           string            `json:"icon,omitempty"` // 图标区字符
	DefaultScript  string            `json:"defaultScript,omitempty"`
	RelationTypes  []string          `json:"relationTypes,omitempty"`
	Description    string            `json:"description,omitempty"` // 简短说明
	// Descriptions 多语言简短说明；不序列化。
	Descriptions map[string]string `json:"-"`
	Usage        string            `json:"usage,omitempty"` // 给 AI / MCP 的用法说明（非 Markdown 面板文档）
	// Doc 编辑器「文档」页 Markdown（与 Usage 分离）；列表接口不下发。
	Doc string `json:"doc,omitempty"`
	// Docs 多语言 Markdown 文档；不序列化。
	Docs   map[string]string `json:"-"`
	Source string            `json:"source,omitempty"` // builtin | plugin | marketplace
	ConfigFields []ConfigField     `json:"configFields,omitempty"`
	// Actions 选中快捷栏能力；零值表示全部关闭，需显式打开。
	Actions NodeActions `json:"actions,omitempty"`
}

// ComponentItem 面板单项（与前端 PaletteItem 对齐）。
type ComponentItem struct {
	Type          string        `json:"type"`
	Label         string        `json:"label"`
	Color         string        `json:"color,omitempty"`
	IconText      string        `json:"iconText,omitempty"`
	DefaultScript string        `json:"defaultScript,omitempty"`
	Description   string        `json:"description,omitempty"`
	Actions       NodeActions   `json:"actions,omitempty"`
	// ConfigFields 属性面板动态表单定义。
	ConfigFields []ConfigField `json:"configFields,omitempty"`
	// InPorts / OutPorts：面板端口预览数量（与画布锚点一致；0 表示无该侧端口）。
	InPorts  int `json:"inPorts"`
	OutPorts int `json:"outPorts"`
}

// ComponentGroup 面板分组。
type ComponentGroup struct {
	ID    string          `json:"id"`
	Label string          `json:"label"`
	Items []ComponentItem `json:"items"`
}

// ComponentsResponse GET /api/components 响应体。
type ComponentsResponse struct {
	Groups []ComponentGroup `json:"groups"`
}

// ComponentDocResponse GET /api/components/{type}/doc —— 单节点编辑器文档。
type ComponentDocResponse struct {
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

// ComponentDocItem 批量文档列表中的一项。
type ComponentDocItem struct {
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

// ComponentDocsResponse GET /api/components/docs —— 全部已启用节点的编辑器文档。
type ComponentDocsResponse struct {
	Items []ComponentDocItem `json:"items"`
}

// ComponentManageItem 节点管理列表项。
type ComponentManageItem struct {
	Type          string `json:"type"`
	Label         string `json:"label"`
	Category      string `json:"category"`
	CategoryLabel string `json:"categoryLabel"`
	Description   string `json:"description,omitempty"`
	Source        string `json:"source"`
	Enabled       bool   `json:"enabled"`
	// PluginID 非空表示来自本地插件，可停用/卸载。
	PluginID string `json:"pluginId,omitempty"`
}

// ComponentManageGroup 节点管理分组。
type ComponentManageGroup struct {
	ID    string                `json:"id"`
	Label string                `json:"label"`
	Items []ComponentManageItem `json:"items"`
}

// ComponentManageResponse GET /api/settings/components。
type ComponentManageResponse struct {
	Groups []ComponentManageGroup `json:"groups"`
}
