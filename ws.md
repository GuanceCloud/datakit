
# DataKit、DataWay Websocket 改造

为便于在网页端实现对 DataKit 的精确管理，现于 DataWay 上增加 websocket 服务器，基本通信机制如下：

```
                                                               Web ---+
                                                                |     |
                                            cmd-with-auth       |     |
             websocket                |  <======================+     | 
DataKit <---------------> DataWay(ws) |     update-datakit-info       |
   |         info/cmd          |      |  -----------------------+     |
   |                           |                                |     |
   |                           |            check-auth          v     v
   |                           `============================> DataFlux 后端
   |                                                               ^
   |               metric/logging/object/event                     |
   `--------------------> DataWay(http) ---------------------------`

                            DataWay                          Web 
    metric/log/obj/event  +---------+                         |
   .--------------------> |  HTTP   | -----------------.      |
   |                      |---------|                  v      v 
DataKit                   |  wscli  | <------------> DF 后端(HTTP/ws server)                                                                       
   ^     websocket        |v^v^v^v^v|
   `--------------------> |  wssrv  |
                          +---------+
```

上图主要用于传递如下数据

- 用户在 Web 端可以下发、修改 DataKit 配置（cmd）
	- 获取所有可用采集器列表
	- 获取已开启的采集器运行情况
	- 获取某个采集器的 config-sample
	- 新增一个采集器的配置
	- 修改已有的某个采集器的配置
	- reload DataKit 使配置生效
	- ...
- DataKit 信息变更（info）可以通过 DataWay 上传到 DataFlux（版本、运行状态等）

## DataWay online 设计

DataWay 启动后，自动连接 DataWay 上指定的 ws 服务，并发送一个 info 指令给 DataWay，告知自己的一些基本信息（如版本、UUID 等）。DataWay 收到后，要将对应信息同步到 DataFlux，便于统一管理。

## 采集器配置下发设计

Web 通过调用 DataWay 上的一系列 HTTP 接口，实现对 DataKit 的管理。DataWay 收到请求后，构建一个 ws 指令，发送给对应 ID 且 online 的 DataKit，并等待 DataKit 返回，然后同步 HTTP 返回给 web。如果 DataKit 不存在，则 404 错误。

### 下发细节处理

- 每个 input 下发后，DataWay 生成一个 input-id，这个 ID 会带到 DataFlux，后续用户通过该 id 来管理该 input

	- input-id 建议使用 `cliutils.XID("dkinput_")`，即使用 mongodb 中的 ID 命名规范
	- 同一个 datakit 上同名 input，不允许 input 配置一样（无意义），此可避免 web 重复提交 config 的问题

- 每个 input 下发下去之后，DataKit 会在 conf.d 目录下生成一个 input-id.conf 的文件，里面就是 toml 格式的具体配置
- input-id 是有一定命名规范的，不符合该规范的 xxx.conf 生效后，其 input 不能通过 web 方式来管理（但能采集到数据）
- 如果一次下发多个同名的采集器配置，就会生成多个 input-id
- DataWay 每下发一个消息，需带上 msg-trace-id，DataWay 自行管理该 trace-id，并做过期管理
- DataWay 收到 web 的 HTTP 请求后，如需下发消息给 DataKit，则异步等待 DataKit 的处理结果，然后同步返回给 web。msg-trace-id 无需传递到 web

## 消息体设计

ws 消息体采用 json 来封装，初步定义如下结构：

### DataKit online 消息

```
{
	"type": "online"

	"msgs": [
		{
			"id": <datakit-id>,
			"version": <datakit-version>,
			"os": "linux",
			"arch": "amd64",
			"name": <datakit名称>,
			"enabled_inputs": ["cpu", "mem",...], // 开启的采集器列表
		},
		...
	]
}
```

> 注意：

- 由于 DataWay 可能重启、升级，故 DataKit 需增加重连机制。
- 由于 DataKit 连接 ws 一般都会带上 token，故 DataWay 需验证该 token 是否合法（调用 DataFlux 接口），否则拒绝 DataKit 连接 ws）
- DataFlux 需绑定 DataKit 和 DataWay 的关系，便于下发时确定指定的 DataWay 地址。不然无法知道某个确定的 DataKit 连接的是哪个 DataWay

### 配置下发消息

```
{
	"type": "config"

	"msgs": [
		{
			"id": <input-id>,
			"input": <采集器名称>,
			"input_type": <datakit|telegraf|...>, // 采集器类型
			"cfg": base64(toml)                   // base64 编码后的 toml 配置
		},
		...
	]
}
```

### 获取 input configsample

```
{
	"type": "get-config"

	"msgs": [
		{
			"input": <采集器名称>
		},
		...
	]
}
```

### reload 指令

```
{
	"type": "reload",
	"msg": null
}
```

### disable/enable 部分采集器

```
{
	"type": "disable-inputs", // 或 "enable-inputs"
	"msg": [
		"input-id-1",
		"input-id-2",
		"input-id-3",
		...
	]
}
```

> 注意：enable/disable 需判定采集器是否存在、开启，不然 4xx 错误

### DataKit 错误消息

```
{
	"code": "datakit.inputNotFound", // 此处可扩充若干 code 出来，命名规范：from.errCode
	"status": 404,
	"msg": "input xxx not exist"
}
```

### DataWay 错误消息

```
{
	"code": "dataway.datakitNotFound", // 此处可扩充若干 code 出来，命名规范：from.errCode
	"status": 404,
	"msg": "dataway xxx not found"
}
```

### 新增接口整理

- DataWay：新增若干接口，供 web 端调用
	- 开启/关闭采集器
	- 新增采集器
	- 获取 config-sample
- DataFlux: 增加若干接口，供 DataWay 上传 DataKit 信息（接口列表待定）
