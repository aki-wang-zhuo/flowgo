package templatex

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flowgo/flowgo/api/types"
)

// ${...} 占位符，如 ${msg}、${msg.user.id}、${metadata.httpMethod}
var placeholderRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// Render 用消息内容替换模板中的 ${...}。
// 支持：
//   - ${msg}          整段消息数据
//   - ${msg.a.b}      若 msg 为 JSON，按路径取值
//   - ${metadata.k}   元数据字段
//   - ${msgType}      消息类型
//   - ${dataType}     数据类型
func Render(tpl string, msg types.Msg) (string, error) {
	if !strings.Contains(tpl, "${") {
		return tpl, nil
	}
	var jsonRoot interface{}
	_ = json.Unmarshal([]byte(msg.Data), &jsonRoot)

	var firstErr error
	out := placeholderRe.ReplaceAllStringFunc(tpl, func(m string) string {
		inner := strings.TrimSpace(m[2 : len(m)-1])
		v, err := resolve(inner, msg, jsonRoot)
		if err != nil && firstErr == nil {
			firstErr = err
			return m
		}
		return v
	})
	return out, firstErr
}

func resolve(expr string, msg types.Msg, jsonRoot interface{}) (string, error) {
	switch {
	case expr == "msg":
		return msg.Data, nil
	case expr == "msgType":
		return msg.Type, nil
	case expr == "dataType":
		return string(msg.DataType), nil
	case strings.HasPrefix(expr, "metadata."):
		key := strings.TrimPrefix(expr, "metadata.")
		if msg.Meta == nil {
			return "", nil
		}
		return msg.Meta[key], nil
	case expr == "metadata":
		b, err := json.Marshal(msg.Meta)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case strings.HasPrefix(expr, "msg."):
		path := strings.TrimPrefix(expr, "msg.")
		if jsonRoot == nil {
			return "", fmt.Errorf("msg is not JSON, cannot resolve ${msg.%s}", path)
		}
		val, err := dig(jsonRoot, strings.Split(path, "."))
		if err != nil {
			return "", err
		}
		return stringify(val), nil
	default:
		return "", fmt.Errorf("unknown placeholder: ${%s}", expr)
	}
}

func dig(root interface{}, parts []string) (interface{}, error) {
	cur := root
	for _, p := range parts {
		if p == "" {
			continue
		}
		switch node := cur.(type) {
		case map[string]interface{}:
			v, ok := node[p]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", p)
			}
			cur = v
		case []interface{}:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, fmt.Errorf("invalid array index: %s", p)
			}
			cur = node[idx]
		default:
			return nil, fmt.Errorf("cannot dig into %T at %s", cur, p)
		}
	}
	return cur, nil
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JSON 数字
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(b)
	}
}
