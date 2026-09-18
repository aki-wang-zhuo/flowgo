package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flowgo/flowgo/api/types"
)

// ConcurrentGroupDef 面板元数据。
var ConcurrentGroupDef = types.ComponentDef{
	Type:           types.TypeConcurrentGroup,
	Label:          "并发分组",
	Labels:         map[string]string{types.LocaleEnUS: "Concurrent Group"},
	Category:       "branch",
	CategoryLabel:  "分支",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Branch"},
	Order:          30,
	Color:          "#c5cae9",
	Icon:           "∥",
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "组内按命名线路并发执行，按全部/任意完成机制汇合后出组。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Run named branches concurrently inside a group; join by all/any complete.",
	},
	Usage: `configuration.branches：并发线路 [{name}]。
组框锚点：
- 左入：外部进入本分组（执行 OnMsg）
- 公共并发入（fork）：扇出边 relation=线路名 → 各线路首节点
- 成功汇聚 / 失败汇聚：组内线路连回本框（relation=Success / Failure）
- 右出：汇合后 Success / Failure 出组（视觉单端点，边标签区分）
completeMode：all=全部完成；any=任意完成。
cancelOthersOnAny：任意完成时是否协作取消其余线路（默认 true）。
timeoutSec：超时秒数；0=不超时。
出组 msg.Data 必含 branches 映射。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "branches", Type: "array", Required: true,
			Widget:   types.WidgetBranchList,
			Default:  `[{"name":"branch1"},{"name":"branch2"}]`,
			Description: "并发线路",
			Descriptions: map[string]string{types.LocaleEnUS: "Concurrent branches"},
			Hint: "从「公共并发入」拉出同名边到各线路首节点；线路成功连到「成功汇聚」、失败连到「失败汇聚」。",
			Hints: map[string]string{
				types.LocaleEnUS: "From the fork anchor, draw same-named edges to each branch start; connect success paths to Success join and failures to Fail join.",
			},
		},
		{
			Name: "completeMode", Type: "string", Required: true, Default: "all",
			Widget: types.WidgetSelect,
			Options: []types.ConfigFieldOption{
				{Value: "all", Label: "全部完成", Labels: map[string]string{types.LocaleEnUS: "All complete"}},
				{Value: "any", Label: "任意完成", Labels: map[string]string{types.LocaleEnUS: "Any complete"}},
			},
			Description: "完成机制",
			Descriptions: map[string]string{types.LocaleEnUS: "Completion mode"},
			Hint: "全部完成：所有线路都结束后，全部成功才走 Success，任一失败则 Failure。任意完成：最先到达汇聚点的线路决定组结果；可配合下方开关取消其余线路。",
			Hints: map[string]string{
				types.LocaleEnUS: "All: wait for every branch; Success only if all succeed. Any: first branch to join decides; optionally cancel the rest.",
			},
		},
		{
			Name: "cancelOthersOnAny", Type: "boolean", Default: "true",
			Widget:  types.WidgetSwitch,
			ShowIf:  "completeMode=any",
			Description: "任意完成时取消其他线程",
			Descriptions: map[string]string{types.LocaleEnUS: "Cancel other branches on any-complete"},
			Hint: "打开后，成功或失败满足任意完成时立即协作取消其余线路（当前节点跑完再停）。",
			Hints: map[string]string{
				types.LocaleEnUS: "When on, cancel other branches cooperatively once any-complete is met (after current OnMsg finishes).",
			},
		},
		{
			Name: "timeoutSec", Type: "number", Default: "10",
			Widget: types.WidgetNumber,
			Description: "超时（秒）",
			Descriptions: map[string]string{types.LocaleEnUS: "Timeout (seconds)"},
			Hint: "0 表示不超时。超时后取消未完成线路，组走 Failure。",
			Hints: map[string]string{types.LocaleEnUS: "0 = no timeout. On timeout unfinished branches are cancelled and the group fails."},
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true},
}

// ConcurrentGroupNode 并发分组：扇出命名子链并汇合。
type ConcurrentGroupNode struct {
	cfg    GroupConfig
	nodeID string
}

// NewConcurrentGroup 工厂。
func NewConcurrentGroup() types.Node {
	return &ConcurrentGroupNode{}
}

func (n *ConcurrentGroupNode) Type() string { return types.TypeConcurrentGroup }

func (n *ConcurrentGroupNode) Init(config map[string]interface{}) error {
	n.cfg = parseGroupConfig(config)
	if len(n.cfg.Branches) == 0 {
		return fmt.Errorf("concurrentGroup: branches required")
	}
	return nil
}

// SetNodeID 由引擎编译时写入本节点 id，用于查出边与虚拟汇合点。
func (n *ConcurrentGroupNode) SetNodeID(id string) {
	n.nodeID = id
	// 固定使用组框自身的虚拟汇合 id（不再依赖 groupEnd/groupFail 节点）
	n.cfg.JoinSuccessID = types.VirtualJoinOK(id)
	n.cfg.JoinFailID = types.VirtualJoinFail(id)
}

func (n *ConcurrentGroupNode) Destroy() {}

func (n *ConcurrentGroupNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	runner := types.BranchRunnerFrom(ctx)
	if runner == nil {
		return msg, types.RelationFailure, fmt.Errorf("concurrentGroup: branch runner missing")
	}
	if n.nodeID == "" {
		return msg, types.RelationFailure, fmt.Errorf("concurrentGroup: node id unset")
	}

	starts := make(map[string]string, len(n.cfg.Branches))
	for _, br := range n.cfg.Branches {
		to, ok := runner.Next(n.nodeID, br.Name)
		if !ok || to == "" {
			return msg, types.RelationFailure, fmt.Errorf("concurrentGroup: no outgoing edge for branch %q", br.Name)
		}
		starts[br.Name] = to
	}

	runCtx, cancel := context.WithCancel(ctx)
	if n.cfg.TimeoutSec > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(n.cfg.TimeoutSec)*time.Second)
	}
	defer cancel()

	type result struct {
		name string
		out  types.BranchOutcome
	}
	results := make(chan result, len(n.cfg.Branches))
	var wg sync.WaitGroup
	for _, br := range n.cfg.Branches {
		br := br
		startID := starts[br.Name]
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := runner.RunBranch(runCtx, startID, n.cfg.JoinSuccessID, n.cfg.JoinFailID, msg)
			out.Name = br.Name
			if err != nil && out.Error == "" {
				out.Error = err.Error()
				out.OK = false
			}
			if runCtx.Err() == context.DeadlineExceeded && !out.OK {
				out.Timeout = true
				out.Cancelled = true
			}
			results <- result{name: br.Name, out: out}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	byName := make(map[string]types.BranchOutcome, len(n.cfg.Branches))
	decided := false
	groupOK := false
	timedOut := false

	for res := range results {
		byName[res.name] = res.out
		if res.out.Timeout {
			timedOut = true
		}
		if n.cfg.CompleteMode != types.CompleteModeAny || decided {
			continue
		}
		if res.out.OK {
			decided = true
			groupOK = true
			if n.cfg.CancelOthersOnAny {
				cancel()
			}
			continue
		}
		// 到达失败汇合或明确失败
		if res.out.Relation == types.RelationFailure || (res.out.Error != "" && !res.out.Cancelled) {
			decided = true
			groupOK = false
			if n.cfg.CancelOthersOnAny {
				cancel()
			}
		}
	}

	// 补齐未出现的线路（极端）
	for _, br := range n.cfg.Branches {
		if _, ok := byName[br.Name]; !ok {
			byName[br.Name] = types.BranchOutcome{Name: br.Name, Pending: true}
		}
	}

	if n.cfg.CompleteMode == types.CompleteModeAll {
		if timedOut || runCtx.Err() == context.DeadlineExceeded {
			groupOK = false
			timedOut = true
			for name, o := range byName {
				if !o.OK {
					o.Timeout = true
					o.Cancelled = true
					o.Pending = false
					byName[name] = o
				}
			}
		} else {
			groupOK = true
			for _, br := range n.cfg.Branches {
				if !byName[br.Name].OK {
					groupOK = false
					break
				}
			}
		}
	} else if !decided {
		groupOK = false
		if runCtx.Err() == context.DeadlineExceeded {
			timedOut = true
		}
	} else if groupOK && !n.cfg.CancelOthersOnAny {
		for name, o := range byName {
			if o.OK || o.Cancelled || o.Relation != "" || o.Error != "" {
				continue
			}
			o.Pending = true
			byName[name] = o
		}
	}

	outMsg, err := buildBranchesMsg(msg, byName, timedOut)
	if err != nil {
		return msg, types.RelationFailure, err
	}
	if timedOut {
		if outMsg.Meta == nil {
			outMsg.Meta = types.Metadata{}
		}
		outMsg.Meta[types.KeyErrorMsg] = "concurrent group timeout"
	}
	if groupOK {
		return outMsg, types.RelationSuccess, nil
	}
	return outMsg, types.RelationFailure, nil
}

func buildBranchesMsg(in types.Msg, byName map[string]types.BranchOutcome, timedOut bool) (types.Msg, error) {
	branches := make(map[string]interface{}, len(byName))
	for name, o := range byName {
		entry := map[string]interface{}{"ok": o.OK}
		if o.Relation != "" {
			entry["relation"] = o.Relation
		}
		if o.Error != "" {
			entry["error"] = o.Error
		}
		if o.Cancelled {
			entry["cancelled"] = true
		}
		if o.Timeout || (timedOut && !o.OK) {
			entry["timeout"] = true
		}
		if o.Pending {
			entry["pending"] = true
		}
		if strings.TrimSpace(o.Msg.Data) != "" {
			var body interface{}
			if o.Msg.DataType == types.JSON || json.Valid([]byte(o.Msg.Data)) {
				if err := json.Unmarshal([]byte(o.Msg.Data), &body); err != nil {
					body = o.Msg.Data
				}
			} else {
				body = o.Msg.Data
			}
			entry["msg"] = body
		}
		branches[name] = entry
	}
	raw, err := json.Marshal(map[string]interface{}{"branches": branches})
	if err != nil {
		return in, err
	}
	out := in
	out.Data = string(raw)
	out.DataType = types.JSON
	return out, nil
}
