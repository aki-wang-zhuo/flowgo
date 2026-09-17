package js

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

// preCompileUdf 预编译配置中的 JS UDF，避免每个 VM 重复解析字符串。
func (e *Engine) preCompileUdf() error {
	cache := make(map[string]*goja.Program)
	for k, v := range e.config.Udf {
		if jsFuncStr, ok := v.(string); ok {
			p, err := goja.Compile(k, jsFuncStr, true)
			if err != nil {
				return fmt.Errorf("compile udf %s: %w", k, err)
			}
			cache[k] = p
			continue
		}
		script, ok := v.(Script)
		if !ok {
			continue
		}
		if script.Type != "" && script.Type != ScriptTypeJS {
			continue
		}
		switch c := script.Content.(type) {
		case string:
			p, err := goja.Compile(k, c, true)
			if err != nil {
				return fmt.Errorf("compile udf %s: %w", k, err)
			}
			cache[k] = p
		case *goja.Program:
			if c != nil {
				cache[k] = c
			}
		}
	}
	e.udfProgramCache = cache
	return nil
}

// newVm 创建已注入 global / vars / UDF 并加载主脚本的 Runtime。
func (e *Engine) newVm(fromVars map[string]interface{}) *goja.Runtime {
	vm := goja.New()

	if fromVars != nil {
		for k, v := range fromVars {
			if err := vm.Set(k, v); err != nil {
				// 注入失败不中断创建；Execute 时若缺变量由脚本自行报错
				_ = err
			}
		}
	}

	if len(e.config.Properties) > 0 {
		_ = vm.Set(GlobalKey, e.config.Properties)
	}

	for k, v := range e.config.Udf {
		var err error
		if _, ok := v.(string); ok {
			if p, exists := e.udfProgramCache[k]; exists {
				_, err = vm.RunProgram(p)
			}
		} else if script, scriptOk := v.(Script); scriptOk {
			if script.Type == "" || script.Type == ScriptTypeJS {
				switch script.Content.(type) {
				case string, *goja.Program:
					if p, exists := e.udfProgramCache[k]; exists {
						_, err = vm.RunProgram(p)
					}
				default:
					if script.Content != nil {
						funcName := strings.Replace(k, ScriptTypeJS+ScriptFuncSeparator, "", 1)
						err = vm.Set(funcName, script.Content)
					}
				}
			}
		} else {
			err = vm.Set(k, v)
		}
		_ = err
	}

	if e.program != nil {
		_, _ = vm.RunProgram(e.program)
	}
	return vm
}
