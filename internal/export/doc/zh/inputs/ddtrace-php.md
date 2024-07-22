
# PHP

---

## 安装依赖 {#dependence}

PHP APM 插件安装，参见 [Datadog PHP 接入文档](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/php/#install-the-extension){:target="_blank"}。

## 配置 {#config}

根据 PHP 实际运行环境不同（Apache/NGINX），其配置有一些差异，详见 [Datadog PHP trace SDK 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}。

## 环境变量支持 {#envs}

下面是常用的 PPH APM 参数配置，完整的参数配置列表，参见 [Datadog 文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/php/){:target="_blank"}。

- **`DD_AGENT_HOST`**

    **INI**：`datadog.agent_host`

    **默认值**：`localhost`

    Datakit 监听的主机地址

- **`DD_TRACE_AGENT_PORT`**

    **INI**：`datadog.trace.agent_port`

    **默认值**：`8126`

    Datakit 监听端口号，此处需手动指定为 9529

- **`DD_ENV`**

    **INI**：`datadog.env`

    **默认值**：`null`

    设置程序环境信息，比如 `prod/pre-prod`

- **`DD_SERVICE`**

    **INI**：`datadog.service`

    **默认值**：`null`

    设置 APP 服务名

- **`DD_SERVICE_MAPPING`**

    **INI**：`datadog.service_mapping`

    **默认值**：`null`

    重命名 APM 服务名，比如 `DD_SERVICE_MAPPING=pdo:payments-db,mysqli:orders-db`

- **`DD_TRACE_AGENT_CONNECT_TIMEOUT`**

    **INI**：`datadog.trace.agent_connect_timeout`

    **默认值**：`100`

    Agent 连接 Datakit 超时配置 (单位 ms)，默认 100

- **`DD_TAGS`**

    **INI**：`datadog.tags`

    **默认值**：`null`

    设置每个 span 上都会默认追加的 tag 列表，例如：`key1:value1,key2:value2`

- **`DD_VERSION`**

    **INI**：`datadog.version`

    设置服务版本

- **`DD_TRACE_SAMPLE_RATE`**

    **INI**：`datadog.trace.smaple_rate`

    **默认值**：`-1`

    设置采样率从 0.0(0%) ~ 1.0(100%)。
