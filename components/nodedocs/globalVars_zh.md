# 全局变量

为本流程声明可在任意节点访问的变量（`global.xx`）。无入边、无出边。

## 用途

- 集中配置 API 地址、开关、阈值等
- 模板 `${global.name}`、表达式 / JS 中的 `global.name`

## 配置要点

| 字段 | 说明 |
| --- | --- |
| 变量列表 | 动态添加；每项 name / type / value |
| type | string / number / boolean / json |

## 限制

全流程**最多一个**本节点；不可作为入口，不可连线。
