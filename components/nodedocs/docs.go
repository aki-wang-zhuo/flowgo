/**
 * 内置节点编辑器文档（Markdown）。
 * 与 ComponentDef.Usage（MCP/AI）分离；仅供属性面板「文档」页。
 * 注意：本文件用原始字符串，正文内勿出现反引号；代码块用 ~~~ 围栏。
 */
package nodedocs

import "github.com/flowgo/flowgo/api/types"

// Pair 返回中文默认 Doc 与多语言 Docs 表。
func Pair(zh, en string) (doc string, docs map[string]string) {
	return zh, map[string]string{types.LocaleEnUS: en}
}

const InjectZH = `# 注入执行

定时或手动触发流程的入口节点。

## 用途

- 作为流程起点，**无入边**，仅有出边
- 可配置一次注入的消息体（JSON）
- 适合调试、定时任务、主动轮询类流程

## 配置要点

| 字段 | 说明 |
| --- | --- |
| payload / 消息 | 注入时作为流程初始 msg |
| 调试相关 | 面板「运行」可直接从此节点启动 |

## 连线

从右侧出口连到下游节点；不可接收入边。
`

const InjectEN = `# Inject

Entry node that starts a flow on a schedule or by manual run.

## Purpose

- Flow start: **no incoming edges**, outgoing only
- Configure an initial JSON message body
- Useful for debugging, cron-like jobs, and polling flows

## Configuration

| Field | Notes |
| --- | --- |
| Payload / message | Used as the initial msg |
| Debug | Panel **Run** can start from this node |

## Wiring

Connect from the right-side output; no inputs allowed.
`

const HttpEndpointZH = `# HTTP 呼入

将外部 HTTP 请求接入流程的入口节点。

## 用途

- 监听服务端 HTTP 入口（与 FlowGo Server 的 endpoint 端口配合）
- 每条请求路径对应一条出边（路径不可重复占用）
- 适合 Webhook、开放 API、表单回调

## 配置要点

| 字段 | 说明 |
| --- | --- |
| 路由列表 | 方法 + 路径；每条路径最多连一条出边 |
| TLS / 证书 | 按部署需要开启 |

## 连线

从右侧「出」拉到下游；连线时选择对应请求路径。路径标签显示在连线中点。
`

const HttpEndpointEN = `# HTTP Endpoint

Ingress node that feeds external HTTP requests into a flow.

## Purpose

- Listens on the server HTTP endpoint port
- Each route maps to one outgoing edge (a route cannot be shared)
- Typical for webhooks, public APIs, and form callbacks

## Configuration

| Field | Notes |
| --- | --- |
| Routes | Method + path; at most one edge per path |
| TLS | Enable when required by deployment |

## Wiring

Drag from the right output; pick the route when connecting. The path label sits on the edge midpoint.
`

const IfZH = `# IF 条件

按表达式将消息分流到 True / False（或自定义）分支。

## 用途

- 简单二路判断
- 每个分支出口通常只连一条边

## 配置要点

在属性中编写条件表达式；结果决定走哪条出边。

## 连线

右侧出口连下游；新建连线时选择分支标签。
`

const IfEN = `# IF

Branches the message by evaluating an expression (True / False or custom outlets).

## Purpose

- Simple binary (or few-way) decisions
- Each outlet usually has at most one edge

## Configuration

Set the condition expression in the property panel; the result picks the outlet.

## Wiring

Connect from the right output and choose the branch label when creating the edge.
`

const SwitchZH = `# SWITCH 多路分支

按匹配规则将消息分到多个命名出口（含默认分支）。

## 用途

- 多路路由（状态码、类型字段等）
- 每个出口一条边；可配置 default

## 配置要点

在属性中维护 case 列表与默认出口。

## 连线

连线时选择对应 case / default 标签。
`

const SwitchEN = `# SWITCH

Routes the message to named outlets by match rules (including a default).

## Purpose

- Multi-way routing (status codes, type fields, etc.)
- One edge per outlet; optional default

## Configuration

Maintain the case list and default outlet in the property panel.

## Wiring

Pick the matching case / default label when connecting.
`

const HttpResponseZH = `# HTTP 响应

流程终点：把当前消息写回 HTTP 呼入请求的响应。

## 用途

- 与 **HTTP 呼入** 配对，结束一次请求-响应
- **无出边**，仅有入边

## 配置要点

可配置状态码、响应体（支持模板）。空体时按节点约定回写。

## 连线

只接收来自上游的入边；不要再向外连线。
`

const HttpResponseEN = `# HTTP Response

Flow exit that writes the current message back as the HTTP response.

## Purpose

- Pair with **HTTP Endpoint** to finish a request/response
- **No outgoing edges**; incoming only

## Configuration

Status code and body (templates supported). Empty body follows node defaults.

## Wiring

Incoming edges only; do not connect outwards.
`

const JsTransformZH = `# JS 转换

使用 JavaScript（goja）转换消息；视觉单出口，可连 Success / Failure。

## 用途

- 改写 msg / metadata / msgType
- 脚本错误可走 Failure 边

## 配置要点

| 字段 | 说明 |
| --- | --- |
| jsScript | Transform **函数体**（不要写 function 外壳） |
| debugValue | 仅面板「运行」时作为 msg 测试值 |

### 脚本约定

~~~js
return { msg, metadata, msgType, dataType };
~~~

可用：msg、metadata、msgType、dataType、global、vars。

## 连线

首条出边默认 Success，第二条为 Failure；仅一条出边时可在连线上切换结果类型。
`

const JsTransformEN = `# JS Transform

Transforms the message with JavaScript (goja). One visual port; Success / Failure edges.

## Purpose

- Mutate msg / metadata / msgType
- Script errors can follow the Failure edge

## Configuration

| Field | Notes |
| --- | --- |
| jsScript | Transform **function body** (no function wrapper) |
| debugValue | Test msg for panel **Run** only |

### Script contract

~~~js
return { msg, metadata, msgType, dataType };
~~~

Available: msg, metadata, msgType, dataType, global, vars.

## Wiring

First outgoing edge defaults to Success, second to Failure; with one edge you can toggle the result on the edge.
`

const HttpClientZH = `# HTTP 客户端

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
`

const HttpClientEN = `# HTTP Client

Sends an outbound HTTP request; results can follow Success / Failure.

## Purpose

- Call external REST APIs / webhooks
- Failures or unexpected status can use Failure

## Configuration

| Field | Notes |
| --- | --- |
| URL / method | Target and verb |
| Headers / body | Request headers and body (templates OK) |
| Timeouts | Tune per deployment |

## Wiring

Same pattern as JS Transform: Success / Failure share one visual port.
`
