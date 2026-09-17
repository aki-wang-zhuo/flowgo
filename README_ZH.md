# FlowGo

[English](./README.md)

**FlowGo** 是用 Go 编写的可嵌入**低代码流程引擎库**。它解析自有 **FlowDSL**，以串行 / 有向方式执行内置（及可注册）节点，并为编辑器与 MCP 提供中英双语的组件目录元数据。

> 本仓库**仅为引擎库**。HTTP API、鉴权、MCP、持久化与可视化编辑器位于同级其它项目。

| 相关项目 | 定位 |
| --- | --- |
| [flowgo-server](https://github.com/aki-wang-zhuo/flowgo-server) | 可部署服务端（REST / WebSocket / MCP / 存储 / 插件） |
| [flowgo-editor](https://github.com/aki-wang-zhuo/flowgo-editor) | 可视化流程编辑器（Vue 3 + LogicFlow） |
| [flowgo-node](https://github.com/aki-wang-zhuo/flowgo-node) | 进程外插件 SDK 与示例节点 |

---

## 特性

- **FlowDSL** — 节点、连线与关系路由（`Success` / `Failure` / `True` / `False` / `Default`，以及 switch case）
- **串行执行引擎** — 从入口（或任意起始）节点沿边推进；防环（最大 256 跳）
- **内置节点** — inject、HTTP 入口/客户端/响应、IF / SWITCH、JS 转换
- **注册表** — 工厂与 `ComponentDef` 元数据同源（执行与面板共用定义）
- **编译缓存** — 按指纹复用已 Init 节点；草稿 / 发布分轨缓存
- **调试日志** — 可选节点级 I/O 采集（`ExecuteFromWithLogs*`）
- **表达式** — 基于 [expr-lang](https://github.com/expr-lang/expr) 的分支条件
- **JavaScript** — 池化 [goja](https://github.com/dop251/goja) 虚拟机（`jsTransform`）
- **模板** — `${msg}`、`${msg.a.b}`、`${metadata.k}` 等，用于 HTTP 体等字段
- **国际化元数据** — 中文默认 + 英文表（`zh-CN` / `en-US`）
- **嵌入式节点文档** — `components/nodedocs/` 下中英 Markdown（`go:embed`）

**与 RuleGo DSL 不兼容。** FlowGo 使用自有 DSL 与运行时约定。

---

## 环境要求

- Go **1.22+**

---

## 安装

```bash
git clone https://github.com/aki-wang-zhuo/flowgo.git
cd flowgo
go test ./...
```

`go.mod` 中的模块路径：

```text
github.com/flowgo/flowgo
```

在 FlowGo monorepo 中，`flowgo-server` 通常通过 `replace` 指向 `../flowgo`。

```go
import (
    "github.com/flowgo/flowgo/api/types"
    "github.com/flowgo/flowgo/engine"
)
```

---

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/flowgo/flowgo/api/types"
    "github.com/flowgo/flowgo/engine"
)

func main() {
    eng := engine.New() // 使用 DefaultRegistry（全部内置节点）

    dsl := &types.FlowDSL{
        ID:        "demo",
        EntryNode: "n1",
        Nodes: []types.FlowNode{
            {
                ID:   "n1",
                Type: "inject",
                Configuration: map[string]interface{}{
                    "payload": `{"hello":"flowgo"}`,
                },
            },
            {
                ID:   "n2",
                Type: "jsTransform",
                Configuration: map[string]interface{}{
                    "jsScript": `msg.hello = msg.hello + "!"; return {msg: msg, metadata: metadata, msgType: msgType};`,
                },
            },
        },
        Edges: []types.FlowEdge{
            {From: "n1", To: "n2", Relation: types.RelationSuccess},
        },
    }

    in := types.NewMsg("DEFAULT", types.JSON, "{}", nil)
    out, err := eng.Execute(context.Background(), dsl, in)
    if err != nil {
        panic(err)
    }
    fmt.Println(out.Data)
}
```

没有单独的「加载流程」API：向 `Execute` / `ExecuteFrom` 传入 `*types.FlowDSL` 即可。引擎按流程 ID + 缓存轨 + DSL 指纹编译并缓存。

---

## 目录结构

```text
flowgo/
├── api/types/       # FlowDSL、Msg、Node、ComponentDef、i18n、DebugLog
├── engine/          # 执行器、注册表、编译缓存
├── components/      # 内置节点 + 分类 catalog + nodedocs
│   ├── endpoint/    # inject、httpEndpoint
│   ├── branch/      # if、switch
│   ├── transform/   # jsTransform
│   ├── action/      # httpClient
│   ├── exit/        # httpResponse
│   └── nodedocs/    # 中英 Markdown 文档（嵌入）
└── utils/
    ├── exprx/       # 表达式求值
    ├── js/          # goja 池
    └── templatex/   # ${...} 模板
```

---

## 内置组件

| Type | 分类 | 典型出边 |
| --- | --- | --- |
| `inject` | endpoint | Success |
| `httpEndpoint` | endpoint |（入口/路由配置；实际监听由 server 编排）|
| `if` | branch | True / False |
| `switch` | branch | case + Default |
| `jsTransform` | transform | Success / Failure |
| `httpClient` | action | Success / Failure |
| `httpResponse` | exit |（终端响应）|

面板分类中还预留了 `filter` / `other` 等，供后续或插件节点使用。

---

## 核心 API

### 引擎

```go
eng := engine.New()
eng := engine.NewWithRegistry(customRegistry)

out, err := eng.Execute(ctx, dsl, msg)
out, err := eng.ExecuteFrom(ctx, dsl, startNodeID, msg)
out, logs, err := eng.ExecuteFromWithLogs(ctx, dsl, startNodeID, msg)
out, logs, err := eng.ExecuteFromWithLogsOpts(ctx, dsl, startNodeID, msg, engine.ExecuteOptions{
    OnlyStart:  false,
    CacheTrack: engine.CacheTrackDefault, // 或 draft / published
})

eng.Invalidate(flowID)
eng.InvalidateTrack(flowID, track)
eng.InvalidateAll()
```

### 注册表

```go
r := engine.NewRegistry()
r.Register(def, factory)
node, ok := r.Create("myType")
defs := r.ListDefs()
groups := r.ListGroupsLocale("zh-CN")
```

`engine.DefaultRegistry` 在 init 时注册全部内置组件。

### 节点契约

```go
type Node interface {
    Type() string
    Init(config map[string]interface{}) error
    OnMsg(ctx context.Context, msg Msg) (out Msg, relation string, err error)
    Destroy()
}
```

---

## 表达式、JS 与模板

| 能力 | 包 | 典型用途 |
| --- | --- | --- |
| 表达式 | `utils/exprx` | `if` / `switch`；环境含 `msg`、`metadata`、`msgType`、`dataType` |
| JavaScript | `utils/js` | `jsTransform`；VM 池、超时、`$ctx` |
| 模板 | `utils/templatex` | HTTP 请求/响应体：`${msg}`、`${msg.field}`、`${metadata.key}` |

---

## 测试

```bash
go test ./...
go build ./...
```

---

## 架构说明

```text
flowgo-editor  --HTTP/WS-->  flowgo-server  --import-->  flowgo（本仓库）
                                   |
                                   +-- 插件 --> flowgo-node
```

- **应**把新内置节点逻辑放在本仓库。
- **勿**在本模块中加入 HTTP 服务、JWT/OAuth、MCP 网关或 Vue UI。

---

## 许可证

本仓库尚未发布 LICENSE 文件。如需再分发条款，请联系维护者。

---

## 贡献指南

1. 面向编辑器 / MCP 的文案须中英双语（`Labels` / `Descriptions` / `Docs` 等）。
2. 保持包职责清晰；单文件不宜过大。
3. 新增内置节点时同步补充 `components/nodedocs/*_{zh,en}.md`。
4. 提交前执行 `go test ./...`。
