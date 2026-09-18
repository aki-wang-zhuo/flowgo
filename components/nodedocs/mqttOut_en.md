# MQTT Out

Publish the message payload to an MQTT topic (action node).

## Purpose

- Mid-flow action: left in, Success / Failure out
- topic / payload support ${msg.x} / ${metadata.x} templates
- Empty payload uses upstream msg.Data

## Client modes

| Mode | Notes |
| --- | --- |
| Reuse MQTT In | Pick an MQTT In on the canvas; share its connection (flow must be published) |
| Persistent | Without reuse: connect when the flow goes online, auto-reconnect, release on offline |
| Temporary | Without reuse: connect per publish, disconnect after (default) |

Draft debug: only Temporary works without publishing; reuse/persistent need the flow online.
Toolbar “Test connection”: briefly connect and publish one message (payload template as-is, or `{}` if empty), then disconnect and report the result.

## Configuration

| Field | Notes |
| --- | --- |
| reuseFrom | Optional mqttIn node id to reuse |
| sessionMode | persistent / temporary (when not reusing) |
| server etc. | Broker fields when not reusing |
| topic / qos / retained | Publish topic, QoS, retain flag |
| payload | Template; empty = msg.Data |
| timeoutSec | Publish wait timeout, default 10 |

## Wiring

Success / Failure share one visual port; failures take the Failure edge.
