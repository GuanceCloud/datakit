## luascript 开发使用文档

### 说明

luascript 旨在提供一个 golang 所使用的功能完善、集成简单的 lua 沙盒环境，以适应日渐繁杂的 DataFlux 业务需求。

luascript 支持跨平台，其所使用的 lua 沙盒环境为 lua 5.1 版本。

### 创建 LuaScript 对象

基础对象为 LuaScript，适用 `NewLuaScript(workerNum int)` 创建，`workerNum` 为内部同时工作的模块数量。

在创建 LuaScript 对象后，需要向其添加需要执行的 lua 代码，函数签名如下：

```
AddLuaCodes(name string, codes[]string) error
```

函数会以 `codes` 中的 lua 代码创建一条执行流水线，`name` 为该流水线的标识，后续收到的数据会做标识匹配，`name` 不可重复。

使用 `Run()` 和 `Stop()` 函数用以启动和停止。

使用 `SendData(d LuaData)` 将数据对象发送到 LuaScript 中，**此行为是异步**。

### 构建接口对象

LuaData 为接口对象，定义如下：

```
type LuaData interface {
	DataToLua() interface{}
	Handle(value string, err error)
	Name() string
	CallbackFnName() string
	CallbackTypeName() string
}
```

- DataToLua：发送到 lua 代码中的数据，此 `interface{}` 会转成 lua 格式的数据，如果无法转换将以 `userdata` 的方式传入 lua

- Handle：执行函数，接收从 lua 中返回的数据的 json 字符串，需要将其反序列化成对象

- Name：当前对象需要执行的 lua 代码的标识名称

- CallbackFnName：lua 执行函数的名称

- CallbackTypeName：lua 执行函数的形参名称

### lua 代码

以当前 lua 代码为例，LuaData 的 `CallbackFnName` 为 `handle`，`CallbackTypeName` 是 `points`。

lua 代码中的入口执行函数，都要具有 `return` 返回值。

```
function handle(points)
        print(points)
	return points
end
```

### 示例

见 script_test.go 单元测试。
