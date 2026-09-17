# JS 转换

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
