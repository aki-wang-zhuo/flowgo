# Current Time

Injects the current time (in a chosen timezone) into `msg.__dataTime`.

## Purpose

- Timestamping, date branching, building API payloads
- Optional timezone (default `Asia/Shanghai`, UTC+8)

## `__dataTime` shape

~~~json
{
  "year": 2026,
  "month": 9,
  "day": 18,
  "hour": 12,
  "minute": 30,
  "second": 45,
  "millisecond": 123,
  "timestamp": 1726638645,
  "timestampMs": 1726638645123,
  "iso": "2026-09-18T12:30:45.123+08:00",
  "timezone": "Asia/Shanghai"
}
~~~

## Configuration

| Field | Notes |
| --- | --- |
| timezone | Dropdown of IANA zones; default Asia/Shanghai |

## Wiring

Left in / right out, at most one outgoing edge; no Success label on the canvas (engine still uses Success).
