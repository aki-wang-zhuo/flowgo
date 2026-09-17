# FlowGo

[中文文档](./README_ZH.md)

**FlowGo** is an embeddable low-code **flow engine** written in Go. It parses a proprietary **FlowDSL**, executes built-in (and registerable) nodes in a serial / directed fashion, and exposes bilingual component-catalog metadata for editors and MCP tooling.

> This repository is a **library only**. HTTP APIs, auth, MCP, persistence, and the visual editor live in sibling projects.

| Related project | Role |
| --- | --- |
| [flowgo-server](https://github.com/aki-wang-zhuo/flowgo-server) | Deployable host (REST / WebSocket / MCP / storage / plugins) |
| [flowgo-editor](https://github.com/aki-wang-zhuo/flowgo-editor) | Visual flow editor (Vue 3 + LogicFlow) |
| [flowgo-node](https://github.com/aki-wang-zhuo/flowgo-node) | Out-of-process plugin SDK and example nodes |

---

## Features

- **FlowDSL** — nodes, edges, and relation-based routing (`Success` / `Failure` / `True` / `False` / `Default`, plus switch cases)
- **Serial execution engine** — follows edges from an entry (or arbitrary start) node; cycle guard (max 256 hops)
- **Built-in nodes** — inject, HTTP endpoint/client/response, IF / SWITCH, JS transform
- **Registry** — factory + `ComponentDef` metadata in one place (execution and palette share the same definitions)
- **Compile cache** — fingerprint-based reuse of initialized nodes; draft / published cache tracks
- **Debug logs** — optional per-node I/O capture (`ExecuteFromWithLogs*`)
- **Expressions** — [expr-lang](https://github.com/expr-lang/expr) for branch conditions (`msg` / `metadata` / …)
- **JavaScript** — pooled [goja](https://github.com/dop251/goja) VMs for `jsTransform`
- **Templates** — `${msg}`, `${msg.a.b}`, `${metadata.k}`, etc. for HTTP bodies and similar fields
- **i18n metadata** — Chinese defaults + English tables (`zh-CN` / `en-US`) for labels, descriptions, and docs
- **Embedded node docs** — Markdown under `components/nodedocs/` (`go:embed`)

**Not compatible with RuleGo DSL.** FlowGo uses its own DSL and runtime contracts.

---

## Requirements

- Go **1.22+**

---

## Install

```bash
git clone https://github.com/aki-wang-zhuo/flowgo.git
cd flowgo
go test ./...
```

Module path (as declared in `go.mod`):

```text
github.com/flowgo/flowgo
```

In the FlowGo monorepo, `flowgo-server` typically uses a `replace` directive pointing at `../flowgo`.

```go
import (
    "github.com/flowgo/flowgo/api/types"
    "github.com/flowgo/flowgo/engine"
)
```

---

## Quick start

```go
package main

import (
    "context"
    "fmt"

    "github.com/flowgo/flowgo/api/types"
    "github.com/flowgo/flowgo/engine"
)

func main() {
    eng := engine.New() // uses DefaultRegistry (all built-in nodes)

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

There is no separate “load flow” API: pass a `*types.FlowDSL` to `Execute` / `ExecuteFrom`. The engine compiles and caches by flow ID + track + DSL fingerprint.

---

## Package layout

```text
flowgo/
├── api/types/       # FlowDSL, Msg, Node, ComponentDef, i18n helpers, DebugLog
├── engine/          # Executor, Registry, compile cache
├── components/      # Built-in nodes + category catalog + nodedocs
│   ├── endpoint/    # inject, httpEndpoint
│   ├── branch/      # if, switch
│   ├── transform/   # jsTransform
│   ├── action/      # httpClient
│   ├── exit/        # httpResponse
│   └── nodedocs/    # zh/en Markdown docs (embedded)
└── utils/
    ├── exprx/       # expression evaluation
    ├── js/          # goja pool
    └── templatex/   # ${...} templates
```

---

## Built-in components

| Type | Category | Relations (typical) |
| --- | --- | --- |
| `inject` | endpoint | Success |
| `httpEndpoint` | endpoint | (ingress / route config; listen orchestrated by server) |
| `if` | branch | True / False |
| `switch` | branch | case labels + Default |
| `jsTransform` | transform | Success / Failure |
| `httpClient` | action | Success / Failure |
| `httpResponse` | exit | (terminal response) |

Panel categories also reserve slots such as `filter` / `other` for future or plugin nodes.

---

## Core APIs

### Engine

```go
eng := engine.New()
eng := engine.NewWithRegistry(customRegistry)

out, err := eng.Execute(ctx, dsl, msg)
out, err := eng.ExecuteFrom(ctx, dsl, startNodeID, msg)
out, logs, err := eng.ExecuteFromWithLogs(ctx, dsl, startNodeID, msg)
out, logs, err := eng.ExecuteFromWithLogsOpts(ctx, dsl, startNodeID, msg, engine.ExecuteOptions{
    OnlyStart:  false,
    CacheTrack: engine.CacheTrackDefault, // or draft / published
})

eng.Invalidate(flowID)
eng.InvalidateTrack(flowID, track)
eng.InvalidateAll()
```

### Registry

```go
r := engine.NewRegistry()
r.Register(def, factory)
node, ok := r.Create("myType")
defs := r.ListDefs()
groups := r.ListGroupsLocale("en-US")
```

`engine.DefaultRegistry` is populated at init with all built-in components.

### Node contract

```go
type Node interface {
    Type() string
    Init(config map[string]interface{}) error
    OnMsg(ctx context.Context, msg Msg) (out Msg, relation string, err error)
    Destroy()
}
```

---

## Expressions, JS, and templates

| Capability | Package | Typical use |
| --- | --- | --- |
| Expressions | `utils/exprx` | `if` / `switch` conditions; env includes `msg`, `metadata`, `msgType`, `dataType` |
| JavaScript | `utils/js` | `jsTransform`; pooled VMs, timeout, `$ctx` |
| Templates | `utils/templatex` | HTTP request/response bodies: `${msg}`, `${msg.field}`, `${metadata.key}` |

---

## Testing

```bash
go test ./...
go build ./...
```

---

## Architecture note

```text
flowgo-editor  --HTTP/WS-->  flowgo-server  --import-->  flowgo (this repo)
                                   |
                                   +-- plugins --> flowgo-node
```

- **Do** put new built-in node logic here.
- **Do not** add HTTP servers, JWT/OAuth, MCP gateways, or Vue UI to this module.

---

## License

License file is not yet published in this repository. Contact the maintainers if you need redistributable terms.

---

## Contributing

1. Keep public UI/MCP-facing strings bilingual (`zh-CN` + `en-US`) via Labels / Descriptions / Docs tables.
2. Prefer focused packages; avoid growing single files beyond a maintainable size.
3. Add or update `components/nodedocs/*_{zh,en}.md` when introducing built-in nodes.
4. Run `go test ./...` before opening a PR.
