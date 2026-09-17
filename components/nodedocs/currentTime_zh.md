# 当前时间

按指定时区生成当前时间，写入消息根字段 `__dataTime`。

## 用途

- 在流程中打时间戳、按年月日分支、拼接口参数
- 可选时区（默认东八区 `Asia/Shanghai`）

## `__dataTime` 结构

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

## 配置

| 字段 | 说明 |
| --- | --- |
| timezone | 下拉选择 IANA 时区，默认 Asia/Shanghai |

## 连线

左入右出，最多一条出边；画布不显示 Success 标签（引擎仍按 Success 推进）。
