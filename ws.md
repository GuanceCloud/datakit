
# DataKit、DataWay Websocket 改造

为便于在网页端实现对 DataKit 的精确管理，现于 DataWay 上增加 websocket 服务器，基本通信机制如下：

```
                            DataWay                          Web ----.
    metric/log/obj/event  +---------+                                |
   .~~~~~~~~~~~~~~~~~~~~> |  HTTP   | ~~~~~~~~~~~~~~~~~~~.           |
   |                      |---------|                    v           |
DataKit                   |         | <----.         +----------+    |
   ^     websocket        | wsproxy |      |         |  DF 后端 |    |
   `--------------------> |         |      |         |----------|    |
                          +---------+      |         |   HTTP   | <--`
                                           |         |----------|
                                           `-------> |    WS    |
                                                     +----------+
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

## 机制约定

- DataWay 不应对 datakit/dataflux 过来的数据做任何验证，它只是一个 ws 代理，让 DataWay 在 WS 处理上保持简单。换言之，所有 WS 消息的验证、处理都应该在 DataKit 以及 DataFlux 上完成。
- 用于 DataKit 和 DataFlux 之间通信的消息体定义应该定义在 DataFlux 的 kodo 项目中，datakit 可以 import kodo 的代码
- 所有消息，都以如下方式来定义

```json
{
	"type"    : int 消息类型,
	"id"      : "消息 ID，形如 msg_XID",
	"dest"    : [ "dkit_xid", "dkit_xid", ...], # 表示消息应该发给谁，对于 DataKit 上报的消息，此字段可为空
	"b64data" : "xxxxxxxxxxx"           # 经 base64 编码后的消息体（可以是任何消息格式）
}

type:

- 1~99    : datakit 上报
- 100-199 : web 下发
- 200     : ok
- 400~499 : 错误

```


- web 端通过往 DataFlux 发送 HTTP 请求来控制 datakit，DataFlux 收到请求后，生成 ws 请求经由 dataway 发送给 datakit，此时 DataFlux 需等待 datakit 的 ws 请求返回（视不同消息类型而定），当 ws 请求返回后，再完成 web 端的 HTTP 请求。这里允许 ws 请求超时，这种情况下，web 端的 HTTP 请求应该返回 timeout (HTTP 504) 错误。

## 采集器消息上报

采集器消息上报指 DataKit 主动上报一些数据给 DataFlux。原则上，datakit 上报的消息，都无需 dataflux 有任何返回。

### DataKit 上线

DataKit 启动后，自动连接 DataWay 上指定的 ws 服务，并发送一个 info 指令给 DataWay，告知自己的一些基本信息（如版本、UUID 等）。DataWay 收到后，要将对应信息同步到 DataFlux，便于统一管理。

- DataKit 上线以后，其上线的一些基本信息，DataFlux 应该需要保存一份。当 web 端对 datakit 进行操作的时候，可以做一些基本的验证（如 datakit 是否存在）
- 当 DataFlux 判断某 DataKit 心跳过期以后，应该移除该 DataKit 的登陆信息

```json
{
	"type": 1,
	"id": "msg_xxxxxxxxxxxxx",
	"dest": "",
	"b64data": base64(
		{
			"id"               : "dkit_xid",
			"version"          : "dkit-version",
			"os"               : "linux",
			"arch"             : "amd64",
			"name"             : "dkit 名称",
			"heartbeat"        : "30s",
			"enabled_inputs"   : ["cpu", "mem",...] # 开启的采集器列表
			"available_inputs" : ["cpu", "mem",...] # 可用的采集器列表，非常长
			... # 此处可追加其它字段
		}
	)
}
```

### DataKit 心跳

DataKit online 成功后，应按照约定的频率（该约定频率可在 online 消息中指定）往 DataFlux 发送心跳。当 DataFlux 发现某 DataKit 的心跳间隔大于 `约定间隔 * 1.5` 时，则认为 DataKit 无效。

```json
{
	"type": 2,
	"id": "msg_xxxxxxxxxxxxx",
	"dest": "",
	"b64data": base64(dkit_xid) # 只需说明 datakit id 即可
}
```

## Web 端采集器操作

Web 通过调用 DataFlux 上的一系列 HTTP 接口，实现对 DataKit 的管理

### 获取指定采集器配置

web 端可获取某个 DataKit 上指定名称（可指定多个）的采集器配置。如果其中一个采集器尚未启用，则返回 Error。

- 请求体示例

```json
{
	"type": 100,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(["cpu", "mysqlMonitor"])
}
```

- 返回体示例

```json
{
  "type": 200,
  "id": "msg_1234", # 保持和请求体一样的 ID
  "dest": "",
  "b64data": base64({
		"cpu-<md5>": "cpu-cfg",
		"mysqlMonitor-<md5>": "mysqlMonitor-cfg",
		})
}
```

其中 `cpu-cfg 形式如下（mysql-monitor 也类似，以其 template 形式而定）：

```json
{
	"input": "cpu",
	"fields": [
		{
			"key": "percpu",
			"type": "boolean",
			"default": true,
			"value": false,
		},
		{
			"key": "totalcpu",
			"type": "boolean",
			"default": true,
		},
		{
			"key": "collect_cpu_time",
			"type": "boolean",
			"default": false,
			"value": true,
		},
		{
			"key": "report_active",
			"type": "boolean",
			"default": false,
			"value": true,
		}
	]
}
```

前端拿到这个 JSON 后，直接能以 UI 形式展现。

- Error 返回

```json
{
	"type": 400,
	"id": "msg_1234", # 保持和请求体一样的 ID
	"dest": "",
	"b64data": base64(
			"input xxx not enabled"
	)
}
```

### 更新采集器配置

web 端获取到某个采集器当前配置后，可直接进行更改并同步给 datakit

```json
{
	"type": 102,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(
		"cpu-<md5>": "cpu-cfg",
	)
}
```

此处 cpu-cfg 跟上面的形式一致。

>注意，此处 md5 是原始配置的 md5（不然 datakit 无法定位原配置），datakit 收到新的配置后，会重命名磁盘上对应的 `cpu-<md5>.conf`。

### 新增采集器配置

- 请求体

```json
{
	"type": 102,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(
		{
				"cpu": [ cpu-cfg ],
				"mem": [ mem-cfg ],
				"mysqlMonitor": [ mysqlMonitor-cfg ],
		}
	)
}
```

此处 `cpu-cfg` 形式跟上文一致，都是 JSON 形式的 UI 模板，datakit 拿到后，需转换成对应 toml 文件：`cpu-<md5>.conf`

- 返回体

```json
{
	"type": 200,
	"id": "msg_1234", # 保持和请求体一样的 ID
	"dest": "",
	"b64data": base64("ok")
}
```

- Error 返回

```json
{
	"type": 400,
	"id": "msg_1234", # 保持和请求体一样的 ID
	"dest": "",
	"b64data": base64(
			"input xxx not exists" 
			# 对某些平台而言，部分采集器是不能用的（比如 oracle-monitor 在 windows 版本的 DataKit 上就不能用）
	)
}
```

### 删除指定采集器

web 端可删除某个 DataKit 上指定名称的采集器。注意，此时的删除是物理删除，但需 reload DataKit 才能生效。

- 请求体

```json
{
	"type": 103,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(["cpu-<md5>", "mem-<md5>", "mysqlMonitor-<md5>", ...])
}
```

- 返回体（无）
- Error 返回（无）

>注：要实现该需求， datakit 可采用计算配置 md5 的方式，每个不同的采集器配置之后，datakit 都生成一个 `<input-name>-<md5>.conf` 文件。web 端下发给 datakit 的配置总能计算出一个 MD5 出来，可通过该 MD5 来查找并删除对应的 conf 文件。

> 基于此，datakit 收到每一个采集器配置，都应该单独分文件管理好，假定 web 下发了三个 mysql 配置（假定它们各不相同），那么应该在 datakit 所在机器的磁盘上，生成三个 mysql-xxx.conf 文件。

>为保持 datakit 前后版本的兼容性，升级程序需对现有的采集器配置做一下升级，任何 `xxx.conf` 文件，都应该重命名成 `<intput-name>-<md5>.conf`的形式。其中 `md5` 为 toml->obj->toml 转换之后的字符串值（即去除了注释、且字段顺序无关的 toml），此外，此处带上了采集器名称前缀，也相对便于人工查找（```find . -name 'input-name*'```）。

### reload 采集器配置

一般情况下，采集器配置变更后（如启用或删除某个采集器)，Web 端需手动下发 reload 指令。DataKit 收到 reload 之后，只应重新加载配置，不能重启服务。


- 请求体

```json
{
	"type": 104,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": "" # 此处无消息体
}
```

- 返回体：此消息无返回。DataKit reload 完成后，需重新 online（因配置变更）
- Error 返回（无）

### 获取 DataKit 指定采集器的配置模板

即使某个采集器尚未启用，web 端可通过指定的消息类型，获取某个 DataKit 上指定采集器的配置模板，便于前端表单化配置界面。

- 请求体

```json
{
	"type": 105,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(["cpu", "mysqlMonitor"])
}
```

配置模板格式在后面提及。

- 返回体：见后文
- Error 返回：见后文

### 临时测试某个采集器是否能正常工作

在 web 端设置了某个采集器之后，在实际下发配置之前，可对改配置做一个临时测试，并可在 web 端查看采集到的数据。

- 请求体

```json
{
	"type": 106,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(cpu-cfg)
}
```

此处 `cpu-cfg` 形式跟上文一致，都是 JSON 形式的 UI 模板，datakit 拿到后，需转换成临时 toml 文件，作为 telegraf 输入(测试完后，需删除临时文件)。如果是 datakit 采集器，则实例化具体的采集器对象，并调用 `Test()` 接口获取到示例数据。如果配置有误，则应返回对应错误信息（如数据库连接失败）

- 返回体

```json
{
	"type": 200,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(
			行协议的 CPU 数据。如果是对象数据，则此处是 JSON 格式
	)
}
```

- Error 返回

```json
{
	"type": 4xx,
	"id": "msg_1234",
	"dest": "dkit_xid",
	"b64data": base64(Test() 接口报错信息)
}
```

>注：要实现该需求，datakit 的每个 input 需实现一个 `Test()` 接口，对 telegraf 采集器，则统一实现一个 `Test()` 接口即可，然后命令行调用 telegraf 的 test 功能，生成一个临时的 telegraf conf 供测试，抓取命令行输出即可得到结果。

## 采集器模板

为便于 web 端提供 UI 方式来生成采集器配置，这里需要提供一个中间层来实现 UI 配置到采集器配置的转换，以 `cpu` 采集器为例，当前的 cpu 采集器配置项有 4 个：

- `percpu`
- `totalcpu`
- `collect_cpu_time`
- `report_active`

针对该四个配置项，前端的 UI 配置模板（JSON）如下：

```json
{
	"input": "cpu",
	"fields": [
		{
			"key": "percpu",
			"type": "boolean",
			"default": true,
			"desc": "Whether to report per-cpu stats or not",
		},
		{
			"key": "totalcpu",
			"type": "boolean",
			"default": true,
			"desc": "Whether to report total system cpu stats or not"
		},
		{
			"key": "collect_cpu_time",
			"type": "boolean",
			"default": false,
			"desc": "If true, collect raw CPU time metrics",
		},
		{
			"key": "report_active",
			"type": "boolean",
			"default": false,
			"desc": "If true, compute and report the sum of all non-idle CPU states"
		}
	]
}
```

如果某个采集器的 toml 配置层次比较深，对深层次的配置，直接提取到第一层，此处的 UI 配置模板无需跟 toml 配置结构一致，能一一对应上即可。原则上，一个采集器对象，就是一个如上的 JSON 形式的 UI 模板。

当 web 端请求 cpu 配置模板时，datakit 就应该返回类似上面这种配置模板，这样前端页面就能渲染出对应的表单 UI 了（具体字段，以 UI 实现时具体需求为准）。web 表单生成后，用户可更新表单内容，更新后的表单，可以 POST 给 DataFlux 用以更新采集器配置：

```json
{
	"input": "cpu",
	"fields": [
		{
			"key": "percpu",
			"type": "boolean",
			"default": true,
			"value": false, # 关闭该选项
		},
		{
			"key": "totalcpu",
			"type": "boolean",
			"default": true,
		},
		{
			"key": "collect_cpu_time",
			"type": "boolean",
			"default": false,
			"value": true, # 开启该选项
		},
		{
			"key": "report_active",
			"type": "boolean",
			"default": false,
			"value": true, # 开启该选项
		}
	]
}
```

DataFlux 收到后，直接以 ws 方式下发给 datakit，datakit 应识别该模板并将其转换成采集器的 toml 配置，覆盖之前存在的 `cpu-<md5>.conf`，此处新老 toml 文件的 md5 内容不尽相同，故原 toml 文件被重命名了。

基于以上这些设定：

- DataFlux 无需管理这个模板对象，datakit 拿到请求后，直接返回配置模板
- datakit 需提供一套机制，能将这种 json 模板转换成 toml。对每个采集器而言，提供一个全局的 json 模板即可，只用于前端 UI 渲染。
- 对每个采集器而言，此处只需挑选部分必要的配置作为 fields 即可（某些采集器有几十个选项，需精心挑选，但对 MySQL 而言，一般情况下，提供用户名、密码、连接地址即可）
- 由于这些模板都是管理在 datakit 上的，不同版本的 datakit，可能模板不尽相同，故前端 UI 也可能不尽相同。

初步的配置模板对象定义如下：

```go
type InputField struct {
	Key string `json:"key"`
	Type string `json:"type"` // boolean, int, float, string, enum, ...
	Default interface{} `json:"default"` // 默认值，视 Type 而定
	Value interface{} `json:"value"` // web 端填充的值，视 Type 而定
	...
}

type InputTemplate struct {
	Input string `json:input`
	Fields []*InputField `json:"fields"`
}
```
