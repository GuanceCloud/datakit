{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

接收符合 ddtrace 协议格式的链路数据，并把数据经过统一转换成观测云的链路格式，然后上报观测云。

## 前置条件

### 不同语言平台 Referenc

- [Java](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers)
- [Python](https://docs.datadoghq.com/tracing/setup_overview/setup/python?tab=containers)
- [Ruby](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby)
- [Golang](https://docs.datadoghq.com/tracing/setup_overview/setup/go?tab=containers)
- [NodeJS](https://docs.datadoghq.com/tracing/setup_overview/setup/nodejs?tab=containers)
- [PHP](https://docs.datadoghq.com/tracing/setup_overview/setup/php?tab=containers)
- [C++](https://docs.datadoghq.com/tracing/setup_overview/setup/cpp?tab=containers)
- [.Net Core](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-core?tab=windows)
- [.Net Framework](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-framework?tab=windows)

### 不同语言平台 Source Code

- [Java](https://github.com/DataDog/dd-trace-java)
- [Python](https://github.com/DataDog/dd-trace-py)
- [Ruby](https://github.com/DataDog/dd-trace-rb)
- [Golang](https://github.com/DataDog/dd-trace-go)
- [NodeJS](https://github.com/DataDog/dd-trace-js)
- [PHP](https://github.com/DataDog/dd-trace-php)
- [C++](https://github.com/opentracing/opentracing-cpp)
- [.Net](https://github.com/DataDog/dd-trace-dotnet)

> Java： DataKit 安装目录 `data` 目录下，有预先准备好的 `dd-java-agent.jar`（推荐使用）。也可以直接去 [Maven 下载](https://mvnrepository.com/artifact/com.datadoghq/dd-java-agent)

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

> 注意：不要修改这里的 `endpoints` 列表。

```toml
endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
```

编辑 `conf.d/datakit.conf`，将 `listen` 改为 `0.0.0.0:9529`（此处目的是开放外网访问，端口可选）。此时 ddtrace 的访问地址就是 `http://<datakit-ip>:9529`。如果 trace 数据来源就是 DataKit 本机，可不用修改 `listen` 配置，直接使用 `http://localhost:9529` 即可。

如果有 trace 数据发送给 DataKit，那么在 DataKit 的 `gin.log` 上能看到：

```shell
tail -f /var/log/datakit/gin.log
[GIN] 2021/08/02 - 17:16:31 | 200 |     386.256µs |       127.0.0.1 | POST     "/v0.4/traces"
[GIN] 2021/08/02 - 17:17:30 | 200 |     116.109µs |       127.0.0.1 | POST     "/v0.4/traces"
[GIN] 2021/08/02 - 17:17:30 | 200 |     489.428µs |       127.0.0.1 | POST     "/v0.4/traces"
...
```

> 注意：如果没有 trace 发送过来，在 [monitor 页面](datakit-tools-how-to#44462aae)是看不到 ddtrace 的采集信息的。

## ddtrace 环境变量设置

### 基本环境变量

- DD_TRACE_ENABLED: 开启 global tracer (部分语言平台支持)
- DD_AGENT_HOST: ddtrace agent host address
- DD_TRACE_AGENT_PORT: ddtrace agent host port
- DD_SERVICE: service name
- DD_TRACE_SAMPLE_RATE: set sampling rate
- DD_VERSION: application version (optional)
- DD_TRACE_STARTUP_LOGS: ddtrace logger
- DD_TRACE_DEBUG: ddtrace debug mode
- DD_ENV: application env values
- DD_TAGS: application

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过命令行注入环境变量

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- 在 ddtrace.conf 中直接配置自定义标签。这种方式会影响**所有**发送给 DataKit tracing 服务的数据，需慎重考虑：

```toml
## tags is ddtrace configed key value pairs
# [inputs.ddtrace.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	## ...
```

## 关于 Tags

### 在代码中添加业务 tag

在应用代码中，可通过诸如 `span.SetTag(some-tag-key, some-tag-value)`（不同语言方式不同） 这样的方式来设置业务自定义 tag。对于这些业务自定义 tag，可通过配置 `customer_tags` 来识别并提取：

```toml
customer_tags = []
```

注意，这些 tag-key 中不能包含英文字符 '.'，带 `.` 的 tag-key 会忽略掉，示例：

```toml
customer_tags = [
	"order_id",
	"task_id",
	"some.invalid.key",  # 无效的 tag-key，DataKit 选择将其忽略
]
```

### 应用代码中添加业务 tag 注意事项

- 务必在 `customer_tags` 中添加 tag-key 列表，否则 DataKit 不会进行业务 tag 的提取
- 在开启了采样的情况下，部分添加了 tag 的 span 有可能被舍弃

## Tracing 数据

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
