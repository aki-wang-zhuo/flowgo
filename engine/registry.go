package engine

import (
	"sync"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/components"
	"github.com/flowgo/flowgo/components/action"
	"github.com/flowgo/flowgo/components/branch"
	"github.com/flowgo/flowgo/components/endpoint"
	"github.com/flowgo/flowgo/components/exit"
	"github.com/flowgo/flowgo/components/globalvars"
	"github.com/flowgo/flowgo/components/transform"
)

// Registry 全局节点工厂注册表（含编辑器面板元数据）。
type Registry struct {
	mu        sync.RWMutex
	factories map[string]types.NodeFactory
	metas     map[string]types.ComponentDef
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]types.NodeFactory),
		metas:     make(map[string]types.ComponentDef),
	}
}

// Register 注册节点工厂及其面板元数据（二者必须同源，避免前后端各写一份）。
func (r *Registry) Register(def types.ComponentDef, factory types.NodeFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if def.Type == "" {
		return
	}
	r.factories[def.Type] = factory
	r.metas[def.Type] = def
}

// Create 按类型创建节点。
func (r *Registry) Create(typeName string) (types.Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[typeName]
	if !ok {
		return nil, false
	}
	return f(), true
}

// ListDefs 返回已注册组件元数据副本。
func (r *Registry) ListDefs() []types.ComponentDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]types.ComponentDef, 0, len(r.metas))
	for _, d := range r.metas {
		out = append(out, d)
	}
	return out
}

// GetDef 按类型取元数据副本；未注册返回 false。
func (r *Registry) GetDef(typeName string) (types.ComponentDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.metas[typeName]
	if !ok {
		return types.ComponentDef{}, false
	}
	return d, true
}

// Unregister 移除已注册的节点类型（插件卸载时使用）。
func (r *Registry) Unregister(typeName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, typeName)
	delete(r.metas, typeName)
}

// Has 判断类型是否已注册。
func (r *Registry) Has(typeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[typeName]
	return ok
}

// ListGroups 按内置分类组装面板分组（含空分组）；使用默认语言。
func (r *Registry) ListGroups() []types.ComponentGroup {
	return r.ListGroupsLocale(types.DefaultLocale)
}

// ListGroupsLocale 按指定语言组装面板分组。
func (r *Registry) ListGroupsLocale(locale string) []types.ComponentGroup {
	return components.BuildGroups(r.ListDefs(), locale)
}

// DefaultRegistry 内置节点注册表。
var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.Register(endpoint.InjectDef, endpoint.NewInject)
	DefaultRegistry.Register(endpoint.Def, endpoint.New)
	DefaultRegistry.Register(globalvars.Def, globalvars.New)
	DefaultRegistry.Register(branch.IfDef, branch.NewIf)
	DefaultRegistry.Register(branch.SwitchDef, branch.NewSwitch)
	DefaultRegistry.Register(exit.Def, exit.New)
	DefaultRegistry.Register(transform.Def, transform.New)
	DefaultRegistry.Register(transform.CurrentTimeDef, transform.NewCurrentTime)
	DefaultRegistry.Register(action.Def, action.New)
}
