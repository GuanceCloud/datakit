---
title     : 'Pinpoint'
summary   : 'Pinpoint Tracing 数据接入'
__int_icon      : 'icon/pinpoint'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Pinpoint
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

[:octicons-tag-24: Version-1.6.0](changelog.md#cl-1.6.0) · [:octicons-beaker-24: Experimental](index.md#experimental)

---

Datakit 内置的 Pinpoint Agent 用于接收，运算，分析 Pinpoint Tracing 协议数据。

## 配置 {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Datakit Pinpoint Agent 监听地址配置项为：

    ```toml
    # Pinpoint GRPC service endpoint for
    # - Span Server
    # - Agent Server(unimplemented, for service intactness and compatibility)
    # - Metadata Server(unimplemented, for service intactness and compatibility)
    # - Profiler Server(unimplemented, for service intactness and compatibility)
    address = "127.0.0.1:9991"
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

    在 Kubernetes 中支持的环境变量如下表：

    | 环境变量名                              | 类型        | 示例                                                                             |
    | --------------------------------------- | ----------- | -------------------------------------------------------------------------------- |
    | `ENV_INPUT_PINPOINT_ADDRESS`            | string      | "127.0.0.1:9991"                                                                 |
    | `ENV_INPUT_PINPOINT_KEEP_RARE_RESOURCE` | bool        | true                                                                             |
    | `ENV_INPUT_PINPOINT_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}` |
    | `ENV_INPUT_PINPOINT_SAMPLER`            | float       | 0.3                                                                              |
    | `ENV_INPUT_PINPOINT_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                              |
    | `ENV_INPUT_PINPOINT_STORAGE`            | JSON string | `{"storage":"./pinpoint_storage", "capacity": 5120}`                             |

???+ warning "Datakit 中的 Pinpoint Agent 存在以下限制"

    - 目前只支持 gRPC 协议
    - 多服务（Agent/Metadata/Stat/Span）合一的服务使用同一个端口
    - Pinpoint 链路与 Datakit 链路存在差异，详见[下文](pinpoint.md#opentracing-vs-pinpoint)

<!-- markdownlint-enable -->

### Pinpoint Collector 配置 {#collector-config}

- 下载所需的 Pinpoint APM Collector

Pinpoint 支持实现了多语言的 APM Collector 本文档使用 JAVA Collector 进行配置。[下载](https://github.com/pinpoint-apm/pinpoint/releases){:target="_blank"} JAVA APM Collector。

- 配置 Pinpoint APM Collector，打开 */path_to_pinpoint_collector/pinpoint-root.config* 配置相应的多服务端口

    - 配置 `profiler.transport.module = GRPC`
    - 配置 `profiler.transport.grpc.agent.collector.port = 9991`   （即 Datakit Pinpoint Agent 中配置的端口）
    - 配置 `profiler.transport.grpc.metadata.collector.port = 9991`（即 Datakit Pinpoint Agent 中配置的端口）
    - 配置 `profiler.transport.grpc.stat.collector.port = 9991`    （即 Datakit Pinpoint Agent 中配置的端口）
    - 配置 `profiler.transport.grpc.span.collector.port = 9991`    （即 Datakit Pinpoint Agent 中配置的端口）

- 启动 Pinpoint APM Collector 启动命令

```shell
$ java -javaagent:/path_to_pinpoint/pinpoint-bootstrap.jar \
    -Dpinpoint.agentId=agent-id \
    -Dpinpoint.applicationName=app-name \
    -Dpinpoint.config=/path_to_pinpoint/pinpoint-root.config \
    -jar /path_to_your_app.jar
```

Datakit 链路数据遵循 OpenTracing 协议，Datakit 中一条链路是通过简单的父子（子 span 中存放父 span 的 id）结构串联起来且每个 span 对应一次函数调用

<figure markdown>
  ![OpenTracing](https://static.guance.com/images/datakit/datakit-opentracing.png){ width="600" }
  <figcaption>OpenTracing</figcaption>
</figure>

Pinpoint APM 链路数据较为复杂：

- 父 span 负责产生子 span 的 ID
- 子 span 中也要存放父 span 的 ID
- 使用 span event 替代 OpenTracing 中的 span
- 一个 span 为一个服务的一次应答过程

<figure markdown>
  ![Pinpoint](https://static.guance.com/images/datakit/datakit-pinpoint.png){ width="600" }
  <figcaption>Pinpoint</figcaption>
</figure>

## 链路字段 {#tracing}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## Pinpoint 参考资料 {#references}

- [Pinpoint 官方文档](https://pinpoint-apm.gitbook.io/pinpoint/){:target="_blank"}
- [Pinpoint 版本文档库](https://pinpoint-apm.github.io/pinpoint/index.html){:target="_blank"}
- [Pinpoint 官方仓库](https://github.com/pinpoint-apm){:target="_blank"}
- [Pinpoint 线上实例](http://125.209.240.10:10123/main){:target="_blank"}
