/**
 * 物联网分组：MQTT 收 / 发共用的 Broker 配置解析与面板字段。
 */
package iot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/mqtt"
)

// BrokerFields 从 configuration 解析共用 Broker 连接字段。
func BrokerFields(config map[string]interface{}) (mqtt.Config, error) {
	cfg := mqtt.Config{
		CleanSession:         true,
		MaxReconnectInterval: 60 * time.Second,
	}
	if config == nil {
		return cfg, fmt.Errorf("mqtt: configuration required")
	}
	cfg.Server = strings.TrimSpace(asString(config["server"]))
	if cfg.Server == "" {
		return cfg, fmt.Errorf("mqtt: server required")
	}
	cfg.Username = strings.TrimSpace(asString(config["username"]))
	cfg.Password = asString(config["password"])
	cfg.ClientID = strings.TrimSpace(asString(config["clientId"]))
	if v, ok := config["cleanSession"]; ok {
		cfg.CleanSession = asBool(v, true)
	}
	if sec := asInt(config["maxReconnectIntervalSec"], 60); sec > 0 {
		cfg.MaxReconnectInterval = time.Duration(sec) * time.Second
	}
	cfg.CAFile = strings.TrimSpace(asString(config["caFile"]))
	cfg.CertFile = strings.TrimSpace(asString(config["certFile"]))
	cfg.CertKeyFile = strings.TrimSpace(asString(config["certKeyFile"]))
	return cfg, nil
}

// BrokerConfigFields 面板共用 Broker 字段（中英）。
// showIf 非空时写入每个字段（如 reuseFrom= 表示未复用收节点时才显示）。
func BrokerConfigFields(showIf ...string) []types.ConfigField {
	cond := ""
	if len(showIf) > 0 {
		cond = showIf[0]
	}
	fields := []types.ConfigField{
		{
			Name: "server", Type: "string", Required: true,
			Default: "127.0.0.1:1883", Widget: types.WidgetText,
			Description: "Broker 地址",
			Descriptions: map[string]string{types.LocaleEnUS: "Broker address"},
			Hint: "host:port 或 tcp://host:port；TLS 可用 ssl://host:port",
			Hints: map[string]string{
				types.LocaleEnUS: "host:port or tcp://host:port; use ssl:// for TLS",
			},
		},
		{
			Name: "username", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "用户名",
			Descriptions: map[string]string{types.LocaleEnUS: "Username"},
		},
		{
			Name: "password", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "密码",
			Descriptions: map[string]string{types.LocaleEnUS: "Password"},
		},
		{
			Name: "clientId", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "客户端 ID",
			Descriptions: map[string]string{types.LocaleEnUS: "Client ID"},
			Hint: "留空则由库自动生成；同 Broker 下需唯一",
			Hints: map[string]string{
				types.LocaleEnUS: "Leave empty to auto-generate; must be unique per broker",
			},
		},
		{
			Name: "cleanSession", Type: "boolean", Default: "true", Widget: types.WidgetSwitch,
			Description: "清除会话",
			Descriptions: map[string]string{types.LocaleEnUS: "Clean session"},
		},
		{
			Name: "maxReconnectIntervalSec", Type: "number", Default: "60", Widget: types.WidgetNumber,
			Description: "最大重连间隔（秒）",
			Descriptions: map[string]string{types.LocaleEnUS: "Max reconnect interval (seconds)"},
		},
		{
			Name: "caFile", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "CA 证书路径",
			Descriptions: map[string]string{types.LocaleEnUS: "CA certificate file path"},
		},
		{
			Name: "certFile", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "客户端证书路径",
			Descriptions: map[string]string{types.LocaleEnUS: "Client certificate file path"},
		},
		{
			Name: "certKeyFile", Type: "string", Default: "", Widget: types.WidgetText,
			Description: "客户端私钥路径",
			Descriptions: map[string]string{types.LocaleEnUS: "Client private key file path"},
		},
	}
	if cond != "" {
		for i := range fields {
			fields[i].ShowIf = cond
		}
	}
	return fields
}

// QosConfigField QoS 单选（仅 0 / 1 / 2）。
func QosConfigField() types.ConfigField {
	return types.ConfigField{
		Name: "qos", Type: "string", Default: "1", Widget: types.WidgetSelect,
		Options: []types.ConfigFieldOption{
			{
				Value: "0", Label: "0 · 最多一次",
				Labels: map[string]string{types.LocaleEnUS: "0 · At most once"},
			},
			{
				Value: "1", Label: "1 · 至少一次",
				Labels: map[string]string{types.LocaleEnUS: "1 · At least once"},
			},
			{
				Value: "2", Label: "2 · 恰好一次",
				Labels: map[string]string{types.LocaleEnUS: "2 · Exactly once"},
			},
		},
		Description: "QoS",
		Descriptions: map[string]string{types.LocaleEnUS: "QoS"},
	}
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func asBool(v interface{}, def bool) bool {
	if v == nil {
		return def
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		if s == "true" || s == "1" {
			return true
		}
		if s == "false" || s == "0" {
			return false
		}
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return def
}

func asInt(v interface{}, def int) int {
	if v == nil {
		return def
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err == nil {
			return n
		}
	}
	return def
}
