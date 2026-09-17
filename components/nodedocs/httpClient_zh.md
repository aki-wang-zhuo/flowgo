# HTTP 客户端

主动发起 HTTP 请求；结果可走 Success / Failure。

## 用途

- 调用外部 REST / Webhook
- 请求失败或非预期状态可走 Failure

## 配置要点

| 字段 | 说明 |
| --- | --- |
| URL / 方法 | 目标地址与动词 |
| Headers / Body | 请求头与正文（可模板） |
| 超时等 | 按部署调优 |

## 连线

与 JS 转换类似：Success / Failure 共用视觉出口。
