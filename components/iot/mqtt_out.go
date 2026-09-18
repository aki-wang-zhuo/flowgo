/**
 * MQTT 发：将消息发布到 Broker 主题（动作节点）。
 *
 * 客户端模式：
 * - reuseFrom：复用画布中某 mqttIn 的托管连接
 * - sessionMode=persistent：流程上线时常驻连接（自动重连），下线释放
 * - sessionMode=temporary：每次发送短连，发完断开
 */
package iot

import (
	"context"
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/flowgo/flowgo/api/types"
	"github.com/flowgo/flowgo/utils/mqtt"
	"github.com/flowgo/flowgo/utils/templatex"
)

const (
	// TypeMqttOut MQTT 发节点类型。
	TypeMqttOut = "mqttOut"
	// SessionPersistent 常驻：流程在线期间保持连接。
	SessionPersistent = "persistent"
	// SessionTemporary 临时：每次发布建连，发完断开。
	SessionTemporary = "temporary"
)

// MqttOutDef 面板元数据：分组「物联网」。
var MqttOutDef = types.ComponentDef{
	Type:           TypeMqttOut,
	Label:          "MQTT 发",
	Labels:         map[string]string{types.LocaleEnUS: "MQTT Out"},
	Category:       "iot",
	CategoryLabel:  "物联网",
	CategoryLabels: map[string]string{types.LocaleEnUS: "IoT"},
	Order:          20,
	Color:          "#80cbc4",
	Icon:           "↑",
	RelationTypes:  []string{types.RelationSuccess, types.RelationFailure},
	Source:         types.ComponentSourceBuiltin,
	Description:    "将消息载荷发布到 MQTT 主题。",
	Descriptions: map[string]string{
		types.LocaleEnUS: "Publish the message payload to an MQTT topic.",
	},
	Usage: `中间动作：左入、右出 Success / Failure。
configuration：
- reuseFrom：可选，复用画布中 mqttIn 节点的客户端（流程须已发布）
- sessionMode：未复用时 persistent（常驻）或 temporary（临时短连），默认 temporary
- server 等 Broker 字段：仅未复用时需要
- topic / qos / retained / payload / timeoutSec
成功时 metadata.mqttTopic / mqttQos 写入实际值。`,
	ConfigFields: mqttOutConfigFields(),
	Actions:      types.NodeActions{Edit: true, Delete: true, Test: true},
}

func mqttOutConfigFields() []types.ConfigField {
	fields := []types.ConfigField{
		{
			Name: "reuseFrom", Type: "string", Default: "", Widget: types.WidgetMqttInRef,
			Description: "复用 MQTT 收客户端",
			Descriptions: map[string]string{types.LocaleEnUS: "Reuse MQTT In client"},
			Hint: "选择画布中的 MQTT 收；留空则使用下方连接模式与 Broker 配置",
			Hints: map[string]string{
				types.LocaleEnUS: "Pick an MQTT In on the canvas; leave empty to use session mode + broker settings",
			},
		},
		{
			Name: "sessionMode", Type: "string", Default: SessionTemporary, Widget: types.WidgetSelect,
			ShowIf: "reuseFrom=",
			Options: []types.ConfigFieldOption{
				{
					Value: SessionTemporary, Label: "临时（每次发送后断开）",
					Labels: map[string]string{types.LocaleEnUS: "Temporary (disconnect after each publish)"},
				},
				{
					Value: SessionPersistent, Label: "常驻（流程在线保持连接）",
					Labels: map[string]string{types.LocaleEnUS: "Persistent (keep alive while flow is online)"},
				},
			},
			Description: "连接模式",
			Descriptions: map[string]string{types.LocaleEnUS: "Session mode"},
		},
	}
	fields = append(fields, BrokerConfigFields("reuseFrom=")...)
	fields = append(fields,
		types.ConfigField{
			Name: "topic", Type: "string", Required: true, Default: "device/cmd",
			Widget: types.WidgetText,
			Description: "发布主题（可模板）",
			Descriptions: map[string]string{types.LocaleEnUS: "Publish topic (templates allowed)"},
		},
		QosConfigField(),
		types.ConfigField{
			Name: "retained", Type: "boolean", Default: "false", Widget: types.WidgetSwitch,
			Description: "保留消息",
			Descriptions: map[string]string{types.LocaleEnUS: "Retained message"},
		},
		types.ConfigField{
			Name: "payload", Type: "string", Default: "", Widget: types.WidgetCodeJSON,
			Description: "载荷模板（空=用 msg.Data）",
			Descriptions: map[string]string{
				types.LocaleEnUS: "Payload template (empty = use msg.Data)",
			},
		},
		types.ConfigField{
			Name: "timeoutSec", Type: "number", Default: "10", Widget: types.WidgetNumber,
			Description: "发布超时（秒）",
			Descriptions: map[string]string{types.LocaleEnUS: "Publish timeout (seconds)"},
		},
	)
	return fields
}

// MqttOutConfig 已解析的发布配置（供服务端拉起常驻/复用）。
type MqttOutConfig struct {
	ReuseFrom   string
	SessionMode string
	Broker      mqtt.Config
	Topic       string
	QoS         byte
	Retain      bool
	Payload     string
	Timeout     time.Duration
}

// ParseMqttOutConfig 解析 mqttOut configuration。
func ParseMqttOutConfig(config map[string]interface{}) (MqttOutConfig, error) {
	var out MqttOutConfig
	out.ReuseFrom = strings.TrimSpace(asString(config["reuseFrom"]))
	out.SessionMode = strings.TrimSpace(asString(config["sessionMode"]))
	if out.SessionMode == "" {
		out.SessionMode = SessionTemporary
	}
	if out.SessionMode != SessionTemporary && out.SessionMode != SessionPersistent {
		return out, fmt.Errorf("mqttOut: invalid sessionMode %q", out.SessionMode)
	}
	out.Topic = strings.TrimSpace(asString(config["topic"]))
	if out.Topic == "" {
		return out, fmt.Errorf("mqttOut: topic required")
	}
	out.QoS = mqtt.ClampQoS(asInt(config["qos"], 1))
	out.Retain = asBool(config["retained"], false)
	out.Payload = asString(config["payload"])
	sec := asInt(config["timeoutSec"], 10)
	if sec <= 0 {
		sec = 10
	}
	out.Timeout = time.Duration(sec) * time.Second

	// 复用收节点时不校验本节点 Broker
	if out.ReuseFrom != "" {
		return out, nil
	}
	broker, err := BrokerFields(config)
	if err != nil {
		return out, err
	}
	out.Broker = broker
	return out, nil
}

// NeedsManagedClient 是否需要服务端托管连接（复用或常驻）。
func (c MqttOutConfig) NeedsManagedClient() bool {
	return c.ReuseFrom != "" || c.SessionMode == SessionPersistent
}

// MqttOutNode MQTT 发布节点。
type MqttOutNode struct {
	cfg MqttOutConfig
}

// NewMqttOut 工厂。
func NewMqttOut() types.Node { return &MqttOutNode{} }

// Type 实现 types.Node。
func (n *MqttOutNode) Type() string { return TypeMqttOut }

// Init 解析配置（托管连接由服务端在发布时建立）。
func (n *MqttOutNode) Init(config map[string]interface{}) error {
	cfg, err := ParseMqttOutConfig(config)
	if err != nil {
		return err
	}
	n.cfg = cfg
	return nil
}

// Config 返回已解析配置。
func (n *MqttOutNode) Config() MqttOutConfig { return n.cfg }

// OnMsg 按模式取客户端并发布。
func (n *MqttOutNode) OnMsg(ctx context.Context, msg types.Msg) (types.Msg, string, error) {
	client, closeAfter, err := n.resolveClient(ctx)
	if err != nil {
		return msg, types.RelationFailure, err
	}
	if closeAfter && client != nil {
		defer client.Disconnect(250)
	}

	global := types.FlowGlobalFrom(ctx)
	topic, err := templatex.RenderEnv(n.cfg.Topic, msg, global)
	if err != nil {
		return msg, types.RelationFailure, fmt.Errorf("mqttOut topic: %w", err)
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return msg, types.RelationFailure, fmt.Errorf("mqttOut: empty topic after render")
	}

	body := msg.Data
	if strings.TrimSpace(n.cfg.Payload) != "" {
		body, err = templatex.RenderEnv(n.cfg.Payload, msg, global)
		if err != nil {
			return msg, types.RelationFailure, fmt.Errorf("mqttOut payload: %w", err)
		}
	}

	token := client.Publish(topic, n.cfg.QoS, n.cfg.Retain, []byte(body))
	deadline := n.cfg.Timeout
	if deadline <= 0 {
		deadline = 10 * time.Second
	}
	if ctx != nil {
		if d, ok := ctx.Deadline(); ok {
			if left := time.Until(d); left > 0 && left < deadline {
				deadline = left
			}
		}
	}
	if !token.WaitTimeout(deadline) {
		return msg, types.RelationFailure, fmt.Errorf("mqttOut: publish timeout topic=%s", topic)
	}
	if err := token.Error(); err != nil {
		return msg, types.RelationFailure, fmt.Errorf("mqttOut: publish: %w", err)
	}

	if msg.Meta == nil {
		msg.Meta = types.Metadata{}
	}
	msg.Meta["mqttTopic"] = topic
	msg.Meta["mqttQos"] = fmt.Sprintf("%d", n.cfg.QoS)
	if n.cfg.Retain {
		msg.Meta["mqttRetained"] = "true"
	}
	return msg, types.RelationSuccess, nil
}

func (n *MqttOutNode) resolveClient(ctx context.Context) (client paho.Client, closeAfter bool, err error) {
	exec := types.FlowExecFrom(ctx)
	if n.cfg.NeedsManagedClient() {
		c, ok := LookupPublisher(exec.FlowID, exec.NodeID)
		if !ok || c == nil || !c.IsConnected() {
			return nil, false, ErrNoManagedPublisher
		}
		return c, false, nil
	}
	// 临时短连
	c, err := mqtt.Connect(n.cfg.Broker, 15*time.Second)
	if err != nil {
		return nil, false, err
	}
	return c, true, nil
}

// Destroy 无本地常驻连接（托管连接由服务端释放）。
func (n *MqttOutNode) Destroy() {}
