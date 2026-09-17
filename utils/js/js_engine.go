package js

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

const (
	// DefaultMaxExecTime 脚本默认最大执行时间。
	DefaultMaxExecTime = 2 * time.Second
	ctxKey             = "$ctx"
)

// Engine 基于 goja 的 JS 执行引擎（思路参考 RuleGo，独立实现）。
// 主脚本编译一次，Runtime 通过 sync.Pool 复用。
type Engine struct {
	program *goja.Program
	pool    sync.Pool
	maxExec time.Duration
}

type vmSlot struct {
	vm    *goja.Runtime
	timer *time.Timer
}

// NewEngine 编译脚本并创建引擎。script 应为完整函数定义，例如：
//
//	function Transform(msg, metadata, msgType, dataType) { ... }
func NewEngine(script string, maxExec time.Duration) (*Engine, error) {
	if maxExec <= 0 {
		maxExec = DefaultMaxExecTime
	}
	program, err := goja.Compile("", script, true)
	if err != nil {
		return nil, fmt.Errorf("compile js: %w", err)
	}
	e := &Engine{program: program, maxExec: maxExec}
	e.pool = sync.Pool{
		New: func() interface{} {
			vm := goja.New()
			if _, err := vm.RunProgram(e.program); err != nil {
				// 池创建失败时仍返回 VM，Execute 会再报错
				_ = err
			}
			slot := &vmSlot{vm: vm}
			slot.timer = time.AfterFunc(e.maxExec, func() {
				vm.Interrupt("execution timeout")
			})
			slot.timer.Stop()
			return slot
		},
	}
	return e, nil
}

// Execute 调用已编译脚本中的命名函数，并导出返回值。
func (e *Engine) Execute(ctx context.Context, funcName string, args ...interface{}) (interface{}, error) {
	defer func() {
		_ = recover()
	}()

	slot := e.pool.Get().(*vmSlot)
	vm := slot.vm
	defer func() {
		_ = vm.Set(ctxKey, nil)
		e.pool.Put(slot)
	}()

	if ctx != nil {
		_ = vm.Set(ctxKey, ctx)
	}

	slot.timer.Reset(e.maxExec)
	defer slot.timer.Stop()

	f, ok := goja.AssertFunction(vm.Get(funcName))
	if !ok {
		return nil, errors.New(funcName + " is not a function")
	}

	params := make([]goja.Value, len(args))
	for i, a := range args {
		params[i] = vm.ToValue(a)
	}
	res, err := f(goja.Undefined(), params...)
	if err != nil {
		return nil, err
	}
	return res.Export(), nil
}

// Stop 预留资源释放入口（当前池无全局句柄需关闭）。
func (e *Engine) Stop() {}
