/**
 * MQTT 收：订阅 Broker 主题，消息到达后触发流程（入口节点）。
 * 实际订阅由 flowgo-server 管理；引擎内 OnMsg 透传。
 */
package iot

import (
	"context"
	"fmt"
	"strings"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/mqtt"
)

const (
	// TypeMqttIn MQTT 收节点类型。
	TypeMqttIn = "mqttIn"
)

// MqttInDef 面板元数据：分组「物联网」。
var MqttInDef = types.ComponentDef{
	Type:           TypeMqttIn,
	Label:          "MQTT 收",
	Labels:         map[string]string{types.LocaleEnUS: "MQTT In"},
	Category:       "iot",
	CategoryLabel:  "物联网",
	CategoryLabels: map[string]string{types.LocaleEnUS: "IoT"},
	Order:          10,
	Color:          "#80cbc4",
	Icon:           "↓",
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "订阅 MQTT 主题，收到消息后触发流程。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Subscribe to an MQTT topic and trigger the flow on messages.",
	},
	Usage: `入口节点（无入边，右侧 Success 出边）。
configuration：
- connectAndRespond：编辑态自动连接并响应；发布后始终自动连接
- server / username / password / clientId / cleanSession / maxReconnectIntervalSec
- caFile / certFile / certKeyFile（可选 TLS）
- topic：订阅主题（支持 + / #）
- qos：0 / 1 / 2
载荷写入 msg.Data，metadata 含 mqttTopic、mqttQos、mqttRetained。
从 Success 出边下游开始执行（跳过本节点）。`,
	ConfigFields: append([]types.ConfigField{
		{
			Name: "connectAndRespond", Type: "boolean", Default: "false", Widget: types.WidgetSwitch,
			Description: "连接并响应",
			Descriptions: map[string]string{types.LocaleEnUS: "Connect and respond"},
			Hint: "开启后，草稿（未发布）也会自动连接并订阅；发布上线后始终自动连接。",
			Hints: map[string]string{
				types.LocaleEnUS: "When on, auto-connect while editing (draft); always auto-connect after publish.",
			},
		},
	}, append(BrokerConfigFields(), []types.ConfigField{
		{
			Name: "topic", Type: "string", Required: true, Default: "device/+/telemetry",
			Widget: types.WidgetText,
			Description: "订阅主题",
			Descriptions: map[string]string{types.LocaleEnUS: "Subscribe topic"},
			Hint: "支持单级 + 与多级 # 通配",
			Hints: map[string]string{types.LocaleEnUS: "Supports + and # wildcards"},
		},
		QosConfigField(),
	}...)...),
	Actions: types.NodeActions{Edit: true, Delete: true, Test: true},
}

// MqttInConfig 已解析的订阅配置（供服务端读取）。
type MqttInConfig struct {
	Broker             mqtt.Config
	Topic              string
	QoS                byte
	ConnectAndRespond  bool
}

// ParseMqttInConfig 解析 mqttIn configuration。
func ParseMqttInConfig(config map[string]interface{}) (MqttInConfig, error) {
	var out MqttInConfig
	broker, err := BrokerFields(config)
	if err != nil {
		return out, err
	}
	out.Broker = broker
	out.Topic = strings.TrimSpace(asString(config["topic"]))
	if out.Topic == "" {
		return out, fmt.Errorf("mqttIn: topic required")
	}
	out.QoS = mqtt.ClampQoS(asInt(config["qos"], 1))
	out.ConnectAndRespond = asBool(config["connectAndRespond"], false)
	return out, nil
}

// MqttInNode MQTT 收节点（引擎内透传）。
type MqttInNode struct {
	cfg MqttInConfig
}

// NewMqttIn 工厂。
func NewMqttIn() types.Node { return &MqttInNode{} }

// Type 实现 types.Node。
func (n *MqttInNode) Type() string { return TypeMqttIn }

// Init 解析配置。
func (n *MqttInNode) Init(config map[string]interface{}) error {
	cfg, err := ParseMqttInConfig(config)
	if err != nil {
		return err
	}
	n.cfg = cfg
	return nil
}

// Config 返回已解析配置。
func (n *MqttInNode) Config() MqttInConfig { return n.cfg }

// OnMsg 透传；实际触发由服务端订阅回调从出边下游执行。
func (n *MqttInNode) OnMsg(_ context.Context, msg types.Msg) (types.Msg, string, error) {
	return msg, types.RelationSuccess, nil
}

// Destroy 无本地资源。
func (n *MqttInNode) Destroy() {}
