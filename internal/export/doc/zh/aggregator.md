---
title     : 'Aggregator'
summary   : 'DataKit 内部数据处理组件，负责聚合和尾采样规则的筛选、分包与转发'
tags:
  - '聚合'
  - '尾采样'
  - 'Dataway'
---

{{.AvailableArchs}}

---

`Aggregator` 是 DataKit 内部 IO 链路中的数据处理组件，而非独立的输入采集器。其主要功能包括：

- 根据 `aggr.toml` 配置规则筛选数据点并转发至 Dataway 聚合接口
- 根据 `tail-sampling.toml` 配置规则筛选 `tracing/logging/rum` 数据并转发至 Dataway 尾采样接口

DataKit 端仅负责配置加载、数据点筛选、打包和发送，实际的聚合计算与尾采样决策由 Dataway 端完成。

## 工作流程 {#workflow}

`Aggregator` 启动后首先加载配置，随后每分钟自动重载一次配置。

数据处理流程如下：

1. 执行聚合规则（`PickMetric`），命中规则的数据点打包发送至 `/v1/aggregate` 接口
2. 按数据类型执行尾采样规则（`PickTrace/PickLogging/PickRUM`），命中规则的数据发送至 `/v1/tail_sampling` 接口

不同数据类型支持的功能及处理行为：

| 数据类型 | 聚合规则 | 尾采样规则 | 原始数据点后续处理 |
| --- | --- | --- | --- |
| `tracing` | 支持 | 支持 | 命中尾采样规则后由 Aggregator 接管，不再进行普通写入 |
| `logging` | 支持 | 支持 | 未命中尾采样规则的数据点继续执行普通写入流程 |
| `rum` | 支持 | 支持 | 未命中尾采样规则的数据点继续执行普通写入流程 |
| `metric` 等其他类型 | 支持 | 不支持 | 原始数据点继续执行普通写入流程 |

## 主配置项 {#main-config}

`Aggregator` 的主要配置字段说明：

- `endpoints`：下游 Dataway 地址列表，需包含 `token` 查询参数，配置为空时使用 DataWay 地址。如果本地部署 DataWay 走级联模式可以使用该配置。
- `max_raw_body_size`：单次发送数据包的最大大小（压缩前，单位：字节）
- `use_local_config`：是否启用本地配置文件模式
- `local_config_dir`：本地配置文件目录路径
- `local_metric_config_file`：聚合规则配置文件名
- `local_tail_sampling_config_file`：尾采样规则配置文件名

其中：

- `[aggregator]` 是当前唯一有效的主配置入口
- 配置了 `endpoints` 时，DataKit 直接按这些地址拼接 `/v1/aggregate`、`/v1/tail_sampling`、`/v1/tail_sampling_config`
- `endpoints` 为空时，DataKit 复用 `[dataway]` 中已初始化的下游地址

配置示例：

```toml
[aggregator]
  endpoints = [
    "https://openway.example.com?token=<YOUR_WORKSPACE_TOKEN>",
  ]

  max_raw_body_size = 1048576

  use_local_config = true
  local_config_dir = "/usr/local/datakit/conf.d/aggr"
  local_metric_config_file = "aggr.toml"
  local_tail_sampling_config_file = "tail-sampling.toml"
```

本地配置文件默认路径：

- `/usr/local/datakit/conf.d/aggr/aggr.toml`
- `/usr/local/datakit/conf.d/aggr/tail-sampling.toml`

当 `use_local_config = false` 时，配置将通过 Dataway 动态拉取。

## 聚合配置 {#aggr-config}

### 顶层字段 {#aggr-top-level}

```toml
default_window = "15s"
action = ""
delete_rules_point = false
```

- `default_window`：默认聚合时间窗口
- `action`：默认处理动作，通常保持为空（可选值：`passthrough`、`drop`）

### 规则结构 {#aggr-rule-structure}

```toml
[[aggregate_rules]]
name = "otel_jvm_class_loaded_sum"
group_by = ["service_name"]

[aggregate_rules.select]
category = "metric"
measurements = ["otel_service"]
metric_name = ["jvm.classes.loaded"]
condition = ""

[aggregate_rules.algorithms."jvm_class_loaded_sum"]
method = "sum"
source_field = "jvm.classes.loaded"

[aggregate_rules.algorithms."jvm_class_loaded_sum".add_tags]
metric = "jvm_class_loaded_sum"
```

字段说明：

- `group_by`：数据聚合的分组维度
- `select.category`：数据筛选类别（支持 `metric`、`tracing`、`logging`、`rum` 等）
- `select.measurements`：measurement 白名单
- `select.metric_name`：指标字段白名单
- `select.condition`：数据过滤条件（为空时仅按 measurement/field 筛选）
- `algorithms.*.method`：聚合算法类型（字符串格式）
- `algorithms.*.source_field`：算法处理的源字段名，必须与数据点中的实际字段名一致

### method 可选值 {#aggr-methods}

当前支持的聚合算法：

- `sum`：求和
- `avg`：平均值
- `count`：计数
- `min`：最小值
- `max`：最大值
- `histogram`：直方图
- `stdev`：标准差
- `quantiles`：分位数
- `count_distinct`：去重计数
- `first`：第一个值
- `last`：最后一个值

注意事项：

- `expo_histogram` 算法当前暂不支持
- `method` 参数必须使用字符串格式

### 常见规则示例 {#aggr-examples}

#### 1. JVM 指标聚合 {#aggr-example-jvm}

```toml
[[aggregate_rules]]
name = "otel_jvm_threads_live_sum"
group_by = ["service_name"]

[aggregate_rules.select]
category = "metric"
measurements = ["otel_service"]
metric_name = ["jvm.threads.live"]
condition = ""

[aggregate_rules.algorithms."jvm_threads_live_sum"]
method = "sum"
source_field = "jvm.threads.live"
```

#### 2. Tracing 数据中统计 Root Span 数量 {#aggr-example-trace-root-span}

```toml
[[aggregate_rules]]
name = "trace_root_span_count"
group_by = ["service", "resource"]

[aggregate_rules.select]
category = "tracing"
measurements = ["opentelemetry"]
metric_name = ["span_id"]
condition = '{ parent_id = "0" }'

[aggregate_rules.algorithms."root_span.count"]
method = "count"
source_field = "span_id"
```

## 尾采样配置（tail-sampling.toml）{#tail-sampling-config}

### 顶层结构 {#tail-sampling-top}

```toml
version = 2

[trace]
data_ttl = "1m"
group_key = "trace_id"

[logging]
data_ttl = "1m"

[rum]
data_ttl = "1m"
```

配置说明：

- `version`：配置版本号，将随数据包一同发送至 Dataway
- `trace.group_key`：Trace 数据分组键，当前仅支持 `trace_id`
- `logging/rum`：通过 `group_dimensions` 配置按维度进行分组采样

### Trace 规则 {#trace-tail-sampling}

Trace 数据管道配置键名为 `sampling_pipeline`（注意：不是 `pipelines`）：

```toml
[trace]
data_ttl = "1m"
group_key = "trace_id"

[[trace.sampling_pipeline]]
name = "keep-resource"
type = "condition"
condition = '{ resource IN ["/resource"] }'
action = "keep"

[[trace.sampling_pipeline]]
name = "drop-304"
type = "condition"
condition = '{ http_status_code = "304" }'
action = "drop"
```

规则按配置顺序依次生效，因此规则顺序非常重要。推荐按照以下优先级编排：

1. 先写明确需要保留的 `keep` 规则
2. 再写明确需要丢弃的 `drop` 规则
3. 最后再使用概率采样规则作为**兜底** 如果没有概率作为兜底，没有匹配到规则的链路会被删掉。

推荐原因：

- `keep` 规则应优先保障关键链路、错误链路、慢链路等高价值数据不会被后续规则误丢弃
- `drop` 规则适合放在 `keep` 之后，用于清理已确认无需保留的数据
- 概率采样规则适合作为最后的兜底策略，用于在剩余数据上做比例保留

不推荐将兜底 `drop` 规则或大范围概率采样规则放在前面，否则后续更精确的 `keep` 规则可能无法达到预期效果。

推荐写法示例：

```toml
[trace]
data_ttl = "1m"
group_key = "trace_id"

[[trace.sampling_pipeline]]
name = "keep-error-trace"
type = "condition"
condition = '{ status = "error" }'
action = "keep"

[[trace.sampling_pipeline]]
name = "keep-slow-trace"
type = "condition"
condition = '{ duration > 1000000000 }'
action = "keep"

[[trace.sampling_pipeline]]
name = "drop-health-check"
type = "condition"
condition = '{ resource IN ["/health", "/ready"] }'
action = "drop"

[[trace.sampling_pipeline]]
name = "sample-rest"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.1
```

### Logging 规则 {#logging-tail-sampling}

```toml
[logging]
data_ttl = "1m"

[[logging.group_dimensions]]
group_key = "trace_id"

[[logging.group_dimensions.pipelines]]
name = "sample-by-trace-id"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.1
```

处理行为说明：

- 首先按照 `group_key` 对日志数据进行分组
- 命中组内采样规则的日志将发送至尾采样接口
- 缺少 `group_key` 的日志数据将透传，不参与该分组采样

### RUM 规则 {#rum-tail-sampling}

```toml
[rum]
data_ttl = "1m"

[[rum.group_dimensions]]
group_key = "session_id"

[[rum.group_dimensions.pipelines]]
name = "sample-by-session"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.2
```

### Pipeline 字段说明 {#pipeline-fields}

- `type = "condition"`：按条件匹配后执行相应动作，`action` 仅支持 `keep`（保留）或 `drop`（丢弃）
- `type = "probabilistic"`：按 `rate` 参数进行确定性概率采样（取值范围：`0 ~ 1`）
- `condition`：数据过滤条件，支持留空（表示匹配所有数据）
- `hash_keys`：预留字段，当前版本无需配置

当前版本限制：

- `derived_metrics` 功能暂未启用，配置该字段会导致错误

## 发送与性能行为 {#send-behavior}

`Aggregator` 在数据发送前会进行分包和并发优化：

- 根据 `max_raw_body_size` 配置自动拆分聚合批次和尾采样数据包
- 采用异步并发发送机制，工作线程数量上限为 `8`
- 使用 Protobuf 格式进行数据传输（`Content-Type: application/x-protobuf`）

`max_raw_body_size` 配置优先级：

1. `aggregator.max_raw_body_size`（最高优先级）
2. `dataway.max_raw_body_size`
3. DataKit 默认值

当尾采样数据发送返回 `412` 状态码时，DataKit 会自动重新发送尾采样配置至 `/v1/tail_sampling_config` 接口。

## 运行指标 {#metrics}

`Aggregator` 新增了以下 Prometheus 指标（完整前缀为 `datakit_io_`）：

| 指标名 | 类型 | 标签 | 说明 |
| --- | --- | --- | --- |
| `aggr_send_success_total` | Counter | `type`, `category` | Aggregator 发送成功次数 |
| `aggr_send_failed_total` | Counter | `type`, `category`, `reason` | Aggregator 发送失败次数 |
| `aggr_send_points_total` | Counter | `type`, `category` | Aggregator 成功发送的数据点数量 |
| `aggr_lost_points_total` | Counter | `type`, `category`, `reason` | Aggregator 发送失败导致的丢点数量 |
| `aggr_send_latency_seconds` | Summary | `type`, `category` | Aggregator 发送耗时（秒） |

标签说明：

- `type`：当前包含 `metric`、`tail_sampling`、`config`
- `category`：当前发送路径中主要为 `unknown`（配置下发为 `config`）
- `reason`：当前包含 `marshal`、`transport`、`network`、`server`、`other`

排查建议：

- `aggr_send_failed_total` 按 `reason` 上升时，优先定位编码、网络、下游状态码或传输层配置问题
- `aggr_lost_points_total > 0` 表示有数据在发送阶段丢失
- 结合 `aggr_send_points_total` 与 `aggr_send_latency_seconds` 可判断吞吐与时延是否异常

## 常见易错点 {#pitfalls}

1. `aggr.toml` 和 `tail-sampling.toml` 是两套独立的配置文件，不可混合编写
2. Trace 数据管道配置必须使用 `[[trace.sampling_pipeline]]` 语法
3. `method` 参数必须使用字符串格式，如 `sum`、`count`、`max` 等
4. `source_field` 必须与数据点中实际存在的字段名完全一致
5. `trace.group_key` 当前仅支持 `trace_id` 值
6. `logging/rum` 使用分组采样时，缺少 `group_key` 的数据点将透传处理

## 故障排查建议 {#troubleshooting}

若配置未按预期生效，请按以下顺序进行排查：

1. 确认已启用正确的配置来源（本地文件或远程拉取）
2. 检查本地配置文件路径是否正确及 TOML 语法是否合法
3. 验证 `category`、`measurements`、`metric_name`、`condition` 是否能匹配实际数据点
4. 确认 `source_field` 是否与数据点中的真实字段名一致
5. 检查尾采样 `group_key` 是否存在于数据点中
6. 查看 DataKit 日志中的相关信息，如 `loaded ... config`、`split ...`、`send ...` 及 HTTP 状态码等
