---
title     : 'DDTrace PHP'
summary   : 'DDTrace PHP 集成'
tags      :
  - 'DDTRACE'
  - 'PHP'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---


## 安装依赖 {#dependence}

PHP tracer 以扩展形式在用户代码执行前加载。请按 [Datadog PHP 接入文档](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/php/){:target="_blank"} 安装与当前 PHP、运行方式（CLI、PHP-FPM 或 Apache）兼容的扩展；不要只在业务 PHP 文件中动态设置配置，因为那时自动插桩通常已经初始化完成。

## 配置 {#config}

根据 PHP 的运行方式（CLI、PHP-FPM、Apache 模块）配置位置不同：环境变量应在容器、FPM pool 或 Web 服务器进程级设置，INI 配置应在 `php.ini` 或对应运行方式的配置中设置。修改后必须重启相应的 FPM/Apache/容器进程。完整配置见 [Datadog PHP tracer 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}。

接入 DataKit 时，显式覆盖上游默认的 `8126` trace 端口。以下环境变量示例适用于容器或进程启动环境：

```shell
DD_SERVICE=my-php-service \
DD_ENV=production \
DD_VERSION=1.0.0 \
DD_AGENT_HOST=datakit-service \
DD_TRACE_AGENT_PORT=9529 \
php -S 0.0.0.0:8080 -t public
```

也可以使用 `datadog.agent_host` 和 `datadog.trace.agent_port` 这两个 INI 项。若设置 `DD_TRACE_AGENT_URL` / `datadog.trace.agent_url`，它优先于 host 与 port；请只保留一种目标配置。

高并发 Web 应用会频繁创建 PHP 请求上下文。如果当前 tracer 版本支持遥测开关，且不需要 SDK 遥测数据，可在进程启动环境关闭它，以减少额外诊断上报：

```shell
export DD_INSTRUMENTATION_TELEMETRY_ENABLED=false
```

启动后请求一个已插桩的页面，并在 DataKit monitor 中确认 trace 请求。PHP-FPM 环境的 tracer 日志通常写入实际生效的 `error_log`；不要仅以 `php -i` 的输出判断运行时配置。

## 环境变量支持 {#envs}

下面是常用的 PHP APM 参数。完整参数与版本差异见 [Datadog PHP 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}。

- **`DD_AGENT_HOST`**

    **INI**：`datadog.agent_host`

    **默认值**：`localhost`

    Trace 接收端主机地址。接入 DataKit 时填写 DataKit 主机或 Kubernetes Service。

- **`DD_TRACE_AGENT_PORT`**

    **INI**：`datadog.trace.agent_port`

    **上游默认值**：`8126`

    Trace 接收端端口。接入 DataKit 时需显式指定为 `9529`。

- **`DD_ENV`**

    **INI**：`datadog.env`

    **默认值**：`null`

    设置程序环境，例如 `production`、`staging`。

- **`DD_SERVICE`**

    **INI**：`datadog.service`

    **默认值**：`null`

    设置应用服务名；生产环境建议显式设置。

- **`DD_SERVICE_MAPPING`**

    **INI**：`datadog.service_mapping`

    **默认值**：`null`

    重命名 APM 服务名，比如 `DD_SERVICE_MAPPING=pdo:payments-db,mysqli:orders-db`

- **`DD_TRACE_AGENT_CONNECT_TIMEOUT`**

    **INI**：`datadog.trace.agent_connect_timeout`

    **默认值**：`100`

    连接 trace 接收端的超时（毫秒）。远程部署、代理或跨集群网络时应结合实际延迟评估，不要盲目调大。

- **`DD_TAGS`**

    **INI**：`datadog.tags`

    **默认值**：`null`

    设置每个 span 默认追加的标签，例如 `key1:value1,key2:value2`。避免加入用户 ID、会话 ID、令牌和请求正文等高基数或敏感数据。

- **`DD_VERSION`**

    **INI**：`datadog.version`

    设置服务版本

- **`DD_TRACE_SAMPLE_RATE`**

    **INI**：`datadog.trace.sample_rate`

    **默认值**：`-1`

    设置 SDK 侧采样率，范围为 `0.0`（0%）到 `1.0`（100%）。它与 DataKit 接收端采样独立。
