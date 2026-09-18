# MQTT In

Subscribe to a broker topic and trigger the flow (entry node).

## Purpose

- Flow entry: **no incoming edges**, Success outgoing only
- Payload goes to msg.Data; metadata includes mqttTopic, mqttQos, mqttRetained

## Connect and respond

| State | Behavior |
| --- | --- |
| Switch on + draft | Auto-connect after save; messages run the draft flow |
| Flow published | **Always** auto-connect (ignores the switch) |
| Switch off + draft | No connection |

Toolbar “Test connection”: briefly connect and subscribe with current settings, then disconnect and report the result.

## Configuration

| Field | Notes |
| --- | --- |
| connectAndRespond | Auto-connect while editing |
| server | Broker address |
| topic / qos | Subscribe topic and QoS |

## Wiring

Connect from the right Success port to downstream nodes; no incoming edges.
