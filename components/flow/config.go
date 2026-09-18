package flow

import (
	"encoding/json"
	"strings"

	"github.com/flowgo/flowgo/api/types"
)

// BranchDef 并发线路定义。
type BranchDef struct {
	Name string `json:"name"`
}

// GroupConfig 并发分组配置。
type GroupConfig struct {
	Branches           []BranchDef `json:"branches"`
	CompleteMode       string      `json:"completeMode"`       // all | any
	CancelOthersOnAny  bool        `json:"cancelOthersOnAny"`  // 任意完成时是否取消其他线程
	TimeoutSec         int         `json:"timeoutSec"`         // 0=不超时
	JoinSuccessID      string      `json:"joinSuccessId,omitempty"` // 运行时填入虚拟汇合 id
	JoinFailID         string      `json:"joinFailId,omitempty"`    // 运行时填入虚拟汇合 id
	Width              float64     `json:"width,omitempty"`         // 画布框宽（引擎忽略）
	Height             float64     `json:"height,omitempty"`        // 画布框高（引擎忽略）
}

func parseGroupConfig(raw map[string]interface{}) GroupConfig {
	cfg := GroupConfig{
		CompleteMode:      types.CompleteModeAll,
		CancelOthersOnAny: true,
		TimeoutSec:        10,
	}
	if raw == nil {
		return cfg
	}
	b, _ := json.Marshal(raw)
	_ = json.Unmarshal(b, &cfg)
	cfg.CompleteMode = strings.TrimSpace(cfg.CompleteMode)
	if cfg.CompleteMode != types.CompleteModeAny {
		cfg.CompleteMode = types.CompleteModeAll
	}
	if cfg.TimeoutSec < 0 {
		cfg.TimeoutSec = 0
	}
	// 规范化线路名
	out := make([]BranchDef, 0, len(cfg.Branches))
	seen := map[string]struct{}{}
	for _, br := range cfg.Branches {
		name := strings.TrimSpace(br.Name)
		if name == "" {
			continue
		}
		// 禁止与出组 Success/Failure 撞名
		if name == types.RelationSuccess || name == types.RelationFailure {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, BranchDef{Name: name})
	}
	cfg.Branches = out
	return cfg
}
