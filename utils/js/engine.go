package js

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// Engine 基于 goja 的 JS 执行引擎（对齐 RuleGo GojaJsEngine 的池化与预编译思路）。
// 主脚本与 UDF 编译一次；Runtime 经 sync.Pool 复用；超时 timer 绑定槽位 Reset 复用。
type Engine struct {
	config          ScriptConfig
	program         *goja.Program
	udfProgramCache map[string]*goja.Program
	pool            sync.Pool
}

// vmSlot 池化 VM 槽位，绑定可复用超时定时器。
type vmSlot struct {
	vm    *goja.Runtime
	timer *time.Timer
}

// NewEngine 编译主脚本并创建引擎。
// script 应为完整函数定义，例如：function Transform(...) { ... }
// cfg 由调用方传入（常用 DefaultConfig()）；MaxExecutionTime<=0 表示不启用超时。
// fromVars 注入到每个新建 VM（如 {"vars": ...}）。
func NewEngine(script string, cfg ScriptConfig, fromVars map[string]interface{}) (*Engine, error) {
	if script == "" {
		return nil, errors.New("js script is empty")
	}
	cfg = cloneConfig(cfg)

	program, err := goja.Compile("", script, true)
	if err != nil {
		return nil, fmt.Errorf("compile js: %w", err)
	}

	e := &Engine{
		config:  cfg,
		program: program,
	}
	if err = e.preCompileUdf(); err != nil {
		return nil, err
	}

	maxExec := cfg.MaxExecutionTime
	e.pool = sync.Pool{
		New: func() interface{} {
			vm := e.newVm(fromVars)
			slot := &vmSlot{vm: vm}
			if maxExec > 0 {
				slot.timer = time.AfterFunc(maxExec, func() {
					vm.Interrupt("execution timeout")
				})
				slot.timer.Stop()
			}
			return slot
		},
	}
	return e, nil
}

// NewEngineSimple 兼容旧调用：仅指定超时；其余使用 DefaultConfig。
// maxExec<=0 时采用 DefaultConfig 中的超时（通常为 DefaultMaxExecTime）。
func NewEngineSimple(script string, maxExec time.Duration) (*Engine, error) {
	cfg := DefaultConfig()
	if maxExec > 0 {
		cfg.MaxExecutionTime = maxExec
	}
	return NewEngine(script, cfg, nil)
}

// Execute 调用已编译脚本中的命名函数，并导出返回值。
// panic 转为 error；归还池前清空 $ctx，避免脏上下文拖住 GC。
func (e *Engine) Execute(ctx context.Context, funcName string, args ...interface{}) (out interface{}, err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("%v", caught)
			out = nil
		}
	}()

	slot := e.pool.Get().(*vmSlot)
	vm := slot.vm
	defer func() {
		_ = vm.Set(CtxKey, nil)
		e.pool.Put(slot)
	}()

	if ctx != nil {
		_ = vm.Set(CtxKey, ctx)
	}

	if slot.timer != nil {
		slot.timer.Reset(e.config.MaxExecutionTime)
		defer slot.timer.Stop()
	}

	f, ok := goja.AssertFunction(vm.Get(funcName))
	if !ok {
		return nil, errors.New(funcName + " is not a function")
	}

	var params []goja.Value
	if len(args) > 0 {
		params = make([]goja.Value, len(args))
		for i, a := range args {
			params[i] = vm.ToValue(a)
		}
	}

	res, err := f(goja.Undefined(), params...)
	if err != nil {
		return nil, err
	}
	return res.Export(), nil
}

// Stop 预留资源释放入口（当前池无全局句柄需关闭）。
func (e *Engine) Stop() {}
