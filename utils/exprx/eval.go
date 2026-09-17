/**
 * 表达式求值：把 Msg 暴露为 msg / metadata / msgType / dataType。
 * 编译时不绑定 Env，以便 msg 在 JSON 对象与字符串间可变。
 */
package exprx

import (
	"encoding/json"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/flowgo/flowgo/api/types"
)

// Env 从消息构造表达式环境。
func Env(msg types.Msg) map[string]interface{} {
	var msgVal interface{} = msg.Data
	if msg.DataType == types.JSON && msg.Data != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(msg.Data), &parsed); err == nil {
			msgVal = parsed
		}
	}
	meta := map[string]string{}
	for k, v := range msg.Meta {
		meta[k] = v
	}
	return map[string]interface{}{
		"msg":      msgVal,
		"metadata": meta,
		"msgType":  msg.Type,
		"dataType": string(msg.DataType),
	}
}

// CompileBool 编译布尔表达式。
func CompileBool(expression string) (*vm.Program, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression is empty")
	}
	return expr.Compile(expression, expr.AsBool())
}

// CompileAny 编译任意结果表达式。
func CompileAny(expression string) (*vm.Program, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression is empty")
	}
	return expr.Compile(expression)
}

// RunBool 执行布尔表达式。
func RunBool(program *vm.Program, msg types.Msg) (bool, error) {
	if program == nil {
		return false, fmt.Errorf("program is nil")
	}
	out, err := expr.Run(program, Env(msg))
	if err != nil {
		return false, err
	}
	b, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("expression result is not bool: %T", out)
	}
	return b, nil
}

// RunToString 执行表达式并把结果转为字符串（用于 SWITCH 匹配）。
func RunToString(program *vm.Program, msg types.Msg) (string, error) {
	if program == nil {
		return "", fmt.Errorf("program is nil")
	}
	out, err := expr.Run(program, Env(msg))
	if err != nil {
		return "", err
	}
	if out == nil {
		return "", nil
	}
	switch v := out.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprint(v), nil
	}
}
