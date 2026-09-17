package transform

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/flowgo/flowgo/api/types"
)

const varsKey = "vars"

// getVars 从节点 configuration 提取 vars，注入 JS 引擎（对齐 RuleGo GetVars）。
func getVars(configuration map[string]interface{}) map[string]interface{} {
	if configuration == nil {
		return nil
	}
	v, ok := configuration[varsKey]
	if !ok {
		return nil
	}
	return map[string]interface{}{varsKey: v}
}

// decodePayloadForJS 为脚本准备入参副本。
// JSON 每次重新 Unmarshal，避免 goja 修改对象影响原始 msg.Data 语义；
// BINARY 复制字节切片；其它类型返回字符串。
func decodePayloadForJS(msg types.Msg) (interface{}, error) {
	switch msg.DataType {
	case types.JSON:
		if msg.Data == "" {
			return map[string]interface{}{}, nil
		}
		var v interface{}
		if err := json.Unmarshal([]byte(msg.Data), &v); err != nil {
			return nil, err
		}
		return v, nil
	case types.BINARY:
		b := []byte(msg.Data)
		cp := make([]byte, len(b))
		copy(cp, b)
		return cp, nil
	default:
		return msg.Data, nil
	}
}

// encodePayload 将脚本返回的 msg 字段写回消息载荷。
func encodePayload(v interface{}, prefer types.DataType) (string, types.DataType, error) {
	switch t := v.(type) {
	case string:
		return t, types.TEXT, nil
	case []byte:
		return string(t), types.BINARY, nil
	case []interface{}:
		if bytes, ok := tryByteArray(t); ok {
			return string(bytes), types.BINARY, nil
		}
		b, err := json.Marshal(t)
		if err != nil {
			return "", prefer, err
		}
		return string(b), types.JSON, nil
	case map[string]interface{}:
		b, err := json.Marshal(t)
		if err != nil {
			return "", prefer, err
		}
		return string(b), types.JSON, nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return "", prefer, err
		}
		return string(b), types.JSON, nil
	}
}

// tryByteArray 尝试把 JS 导出的数字数组转为 []byte（0–255 整数）。
func tryByteArray(items []interface{}) ([]byte, bool) {
	if len(items) == 0 {
		return nil, false
	}
	out := make([]byte, len(items))
	for i, v := range items {
		n, ok := toByte(v)
		if !ok {
			return nil, false
		}
		out[i] = n
	}
	return out, true
}

func toByte(v interface{}) (byte, bool) {
	switch t := v.(type) {
	case float64:
		if t < 0 || t > 255 || t != float64(int(t)) {
			return 0, false
		}
		return byte(t), true
	case int64:
		if t < 0 || t > 255 {
			return 0, false
		}
		return byte(t), true
	case int:
		if t < 0 || t > 255 {
			return 0, false
		}
		return byte(t), true
	case json.Number:
		i, err := t.Int64()
		if err != nil || i < 0 || i > 255 {
			return 0, false
		}
		return byte(i), true
	case string:
		i, err := strconv.Atoi(t)
		if err != nil || i < 0 || i > 255 {
			return 0, false
		}
		return byte(i), true
	default:
		return 0, false
	}
}

func toStringMap(in map[string]interface{}) types.Metadata {
	out := types.Metadata{}
	for k, v := range in {
		out[k] = fmt.Sprint(v)
	}
	return out
}
