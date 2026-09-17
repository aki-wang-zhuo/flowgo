package js

import (
	"sync"
	"time"
)

const (
	// DefaultMaxExecTime 脚本默认最大执行时间。
	DefaultMaxExecTime = 2 * time.Second
	// GlobalKey 脚本内通过 global.xxx 访问的全局属性对象名。
	GlobalKey = "global"
	// CtxKey 脚本内注入的上下文键；归还 VM 池前必须清空。
	CtxKey = "$ctx"
	// ScriptTypeJS JS 脚本类型标记（与 RuleGo Script.Type 对齐的轻量子集）。
	ScriptTypeJS = "Js"
	// ScriptFuncSeparator UDF 名称中脚本类型前缀分隔符（如 "Js#myFunc"）。
	ScriptFuncSeparator = "#"
)

// Script 包装一段可注入 VM 的脚本或已编译程序 / Go 函数。
type Script struct {
	// Type 脚本类型；空或 ScriptTypeJS 表示 JavaScript。
	Type string
	// Content 可为 string、*goja.Program 或任意 Go 可注入函数。
	Content interface{}
}

// ScriptConfig JS 引擎运行配置（对齐 RuleGo Config 中与脚本相关的子集）。
type ScriptConfig struct {
	// MaxExecutionTime 单次 Execute 超时；<=0 表示不启用超时中断。
	MaxExecutionTime time.Duration
	// Properties 注入为脚本全局对象 global（map 键值）。
	Properties map[string]string
	// Udf 用户自定义函数：值为 JS 字符串、Script、或 Go 函数。
	Udf map[string]interface{}
}

var (
	defaultCfgMu sync.RWMutex
	defaultCfg   = ScriptConfig{
		MaxExecutionTime: DefaultMaxExecTime,
		Properties:       map[string]string{},
		Udf:              map[string]interface{}{},
	}
)

// DefaultConfig 返回当前进程级默认脚本配置的浅拷贝（Udf/Properties 独立 map）。
func DefaultConfig() ScriptConfig {
	defaultCfgMu.RLock()
	defer defaultCfgMu.RUnlock()
	return cloneConfig(defaultCfg)
}

// SetDefaultConfig 替换进程级默认脚本配置（供服务启动时注册 UDF / 调整超时）。
func SetDefaultConfig(cfg ScriptConfig) {
	defaultCfgMu.Lock()
	defer defaultCfgMu.Unlock()
	defaultCfg = cloneConfig(cfg)
}

// RegisterUdf 向默认配置注册一个 UDF（线程安全）。
func RegisterUdf(name string, value interface{}) {
	if name == "" || value == nil {
		return
	}
	defaultCfgMu.Lock()
	defer defaultCfgMu.Unlock()
	if defaultCfg.Udf == nil {
		defaultCfg.Udf = map[string]interface{}{}
	}
	defaultCfg.Udf[name] = value
}

// SetProperty 向默认配置写入一条 global 属性。
func SetProperty(key, value string) {
	if key == "" {
		return
	}
	defaultCfgMu.Lock()
	defer defaultCfgMu.Unlock()
	if defaultCfg.Properties == nil {
		defaultCfg.Properties = map[string]string{}
	}
	defaultCfg.Properties[key] = value
}

func cloneConfig(in ScriptConfig) ScriptConfig {
	out := ScriptConfig{MaxExecutionTime: in.MaxExecutionTime}
	if len(in.Properties) > 0 {
		out.Properties = make(map[string]string, len(in.Properties))
		for k, v := range in.Properties {
			out.Properties[k] = v
		}
	} else {
		out.Properties = map[string]string{}
	}
	if len(in.Udf) > 0 {
		out.Udf = make(map[string]interface{}, len(in.Udf))
		for k, v := range in.Udf {
			out.Udf[k] = v
		}
	} else {
		out.Udf = map[string]interface{}{}
	}
	return out
}
