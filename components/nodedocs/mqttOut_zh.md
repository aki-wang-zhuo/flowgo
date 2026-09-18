# MQTT 发

将消息载荷发布到 MQTT 主题（动作节点）。

## 用途

- 中间动作：左入、右出 Success / Failure
- topic / payload 支持 ${msg.x} / ${metadata.x} 模板
- payload 为空时使用上游 msg.Data

## 客户端模式

| 模式 | 说明 |
| --- | --- |
| 复用 MQTT 收 | 下拉选择画布中的 MQTT 收，共用其连接（须流程已发布） |
| 常驻 | 未复用时：流程上线即建连并自动重连，下线释放 |
| 临时 | 未复用时：每次发送建连，发完断开（默认） |

草稿调试：仅「临时」可直接发；复用/常驻需先发布上线。
浮动栏「测试连接」：用当前属性短暂连接并发布一条（载荷模板原文，空则 `{}`），然后断开并提示结果。

## 配置要点

| 字段 | 说明 |
| --- | --- |
| reuseFrom | 可选，复用的 mqttIn 节点 id |
| sessionMode | persistent / temporary（未复用时） |
| server 等 | 未复用时的 Broker 连接项 |
| topic / qos / retained | 发布主题、QoS、是否保留 |
| payload | 载荷模板；空=msg.Data |
| timeoutSec | 发布等待超时，默认 10 |

## 连线

Success / Failure 共用视觉出口；失败走 Failure。
