package types

// FlowDSL 是 FlowGo 自有的流程图定义格式（与 RuleGo DSL 不兼容）。
// 首期仅支持串行 / 有向无环的简单拓扑：从入口节点沿 edges 前进。
type FlowDSL struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Version     int        `json:"version,omitempty"`
	EntryNode   string     `json:"entryNode"`
	Nodes       []FlowNode `json:"nodes"`
	Edges       []FlowEdge `json:"edges"`
}

// FlowNode 流程图中的一个节点。
type FlowNode struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name,omitempty"`
	// Debug 开启后，草稿调试运行会将该节点的入/出消息写入编辑器控制台。
	// 已发布运行（HTTP 入口 / execute API）一律忽略此开关，不采集调试日志。
	Debug bool `json:"debug,omitempty"`
	// X / Y 为编辑器画布坐标，引擎执行时忽略。
	X             float64                `json:"x,omitempty"`
	Y             float64                `json:"y,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

// FlowEdge 节点之间的连线。relation 用于分支（如 Success / Failure）。
type FlowEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation,omitempty"`
	// PointsList 贝塞尔控制点（编辑器布局用，引擎执行忽略）。
	PointsList []Point `json:"pointsList,omitempty"`
}

// Point 画布坐标点。
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Relation 常用关系常量。
const (
	RelationSuccess = "Success"
	RelationFailure = "Failure"
	RelationTrue    = "True"
	RelationFalse   = "False"
	RelationDefault = "Default"
)
