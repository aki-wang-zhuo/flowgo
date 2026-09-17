/**
 * 当前时间节点：按配置时区生成时间字段，写入 msg 根对象的 __dataTime。
 */
package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flowgo/flowgo/api/types"
)

const (
	// TypeCurrentTime 当前时间节点类型。
	TypeCurrentTime = "currentTime"
	// FieldDataTime 注入到消息 JSON 根上的字段名。
	FieldDataTime = "__dataTime"
	// DefaultTimezone 默认东八区（IANA）。
	DefaultTimezone = "Asia/Shanghai"
)

// CurrentTimeDef 面板元数据：分组「转换」。
var CurrentTimeDef = types.ComponentDef{
	Type:           TypeCurrentTime,
	Label:          "当前时间",
	Labels:         map[string]string{types.LocaleEnUS: "Current Time"},
	Category:       "transform",
	CategoryLabel:  "转换",
	CategoryLabels: map[string]string{types.LocaleEnUS: "Transform"},
	Order:          20,
	Icon:           "⏱",
	RelationTypes:  []string{types.RelationSuccess},
	Source:         types.ComponentSourceBuiltin,
	Description:    "按指定时区生成当前时间对象，写入消息字段 __dataTime。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Inject current time fields into msg.__dataTime using a chosen timezone.",
	},
	Usage: `configuration.timezone：IANA 时区名，默认 Asia/Shanghai（东八区）。
执行时在 msg JSON 根对象写入 __dataTime：
  year, month, day, hour, minute, second, millisecond,
  timestamp（秒）, timestampMs（毫秒）, iso（RFC3339Nano）。
若上游 msg 不是 JSON 对象，则以 {} 为根再写入。
出边 RelationSuccess。`,
	ConfigFields: []types.ConfigField{
		{
			Name: "timezone", Type: "string", Default: DefaultTimezone, Widget: types.WidgetSelect,
			Description: "时区",
			Descriptions: map[string]string{
				types.LocaleEnUS: "Timezone",
			},
			Hint: "下拉选择 IANA 时区；默认东八区 Asia/Shanghai",
			Hints: map[string]string{
				types.LocaleEnUS: "Pick an IANA timezone; default Asia/Shanghai (UTC+8)",
			},
			Options: timezoneSelectOptions(),
		},
	},
	Actions: types.NodeActions{Edit: true, Delete: true, Run: true, RunOnly: true},
}

// CurrentTimeNode 向消息注入 __dataTime。
type CurrentTimeNode struct {
	loc *time.Location
}

// NewCurrentTime 工厂。
func NewCurrentTime() types.Node {
	return &CurrentTimeNode{loc: time.FixedZone("CST", 8*3600)}
}

func (n *CurrentTimeNode) Type() string { return TypeCurrentTime }

// Init 解析时区；非法时区返回错误。
func (n *CurrentTimeNode) Init(config map[string]interface{}) error {
	tz := DefaultTimezone
	if config != nil {
		if v, ok := config["timezone"].(string); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				tz = v
			}
		}
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	n.loc = loc
	return nil
}

// OnMsg 生成 __dataTime 并写回消息 Data（JSON）。
func (n *CurrentTimeNode) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	loc := n.loc
	if loc == nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	now := time.Now().In(loc)
	dt := buildDataTime(now, loc.String())

	root, err := decodeJSONObject(msg.Data)
	if err != nil {
		return msg, types.RelationSuccess, fmt.Errorf("msg data: %w", err)
	}
	root[FieldDataTime] = dt
	raw, err := json.Marshal(root)
	if err != nil {
		return msg, types.RelationSuccess, err
	}
	out := msg
	out.Data = string(raw)
	out.DataType = types.JSON
	return out, types.RelationSuccess, nil
}

func (n *CurrentTimeNode) Destroy() {}

// dataTimePayload 写入消息的时间对象结构。
type dataTimePayload struct {
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	Day         int    `json:"day"`
	Hour        int    `json:"hour"`
	Minute      int    `json:"minute"`
	Second      int    `json:"second"`
	Millisecond int    `json:"millisecond"`
	Timestamp   int64  `json:"timestamp"`   // 秒
	TimestampMs int64  `json:"timestampMs"` // 毫秒
	ISO         string `json:"iso"`
	Timezone    string `json:"timezone"`
}

func buildDataTime(now time.Time, tzName string) dataTimePayload {
	return dataTimePayload{
		Year:        now.Year(),
		Month:       int(now.Month()),
		Day:         now.Day(),
		Hour:        now.Hour(),
		Minute:      now.Minute(),
		Second:      now.Second(),
		Millisecond: now.Nanosecond() / 1e6,
		Timestamp:   now.Unix(),
		TimestampMs: now.UnixMilli(),
		ISO:         now.Format(time.RFC3339Nano),
		Timezone:    tzName,
	}
}

// decodeJSONObject 将消息体解析为对象；空或非对象则返回新 map。
func decodeJSONObject(data string) (map[string]interface{}, error) {
	s := strings.TrimSpace(data)
	if s == "" {
		return map[string]interface{}{}, nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}
	// 非对象（数组/标量）：包一层保留原值，再写 __dataTime
	return map[string]interface{}{"_payload": v}, nil
}
