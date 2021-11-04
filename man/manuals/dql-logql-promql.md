# DQL 与其它几种查询语言的对比

DQL 是观测云统一的查询语言，为便于大家学习这种语言，下面我们选取几种不同的查询语言来与之对比，以便大家能较为快速的理解和运用 DQL。

这里我们暂时选择 [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) 和 [LogQL](https://grafana.com/docs/loki/latest/logql/) 俩种语言。大家较为熟知的 SQL 语句因为其形式、功能等与 DQL 大相庭径，此处暂略。

PromQL 是 [Prometheuse](https://prometheus.io/) 中用于查询其时序数据的一种查询语言；LogQL 是用于 [Grafana Loki](https://grafana.com/oss/loki/) 的一种日志查询语言，它跟 DQL 一样，借鉴了 PromQL 的语法结构。总体上，这三种语言的结构类似，但细微处各有不同。下文将从如下几个方面加以阐述：

- 基本语法结构的差异
- 支持的常用预定义函数
- 常用查询写法对比

## 基本语法结构

| 查询语言  | 基本结构 |
| --------- | -------  |
| PromQL    | `指标 {条件过滤列表} [起始时间:结束时间]`
| LogQL     | `{stream-selector} log-pipeline` |
| DQL       | `namespace::指标集:(指标列表) [起始时间:结束时间:分组间隔] { 条件过滤列表 } GROUP-BY-clause ORDER-BY-clause` |

下面分别加以说明。

### PromQL

在 Prometheuse 中，相关指标是离散形式组织的。在其查询中，可直接查找对应的指标，如：

```
http_requests_total{environment="prometheus", method!="GET"}
```

此处即查找指标 `http_requests_total`，通过指定其 label 限制条件（`environment` 和 `method`）来过滤数据。

> 注：PromeQL 称这里的 label 限制条件为 Label Matchers。

### LogQL

顾名思义，LogQL 主要用于日志内容查询，如：

```
{container="query-frontend", namespace="loki-dev"} |= "metrics.go" | logfmt | duration > 10s and throughput_mb < 500
```

此处 `{...}` 里面的，LogQL 称之为 Stream Selector，其旨在于划定数据查询范围（类似于 SQL 中的 `FROM ...` 部分）；半部分则称之为 Log Pipeline，其主要处理日志信息的提取和过滤。

### DQL

DQL 覆盖面较为全面，相比于 PromQL 只能用于查找 Prometheuse 中的时序数据、LogQL 只能用于查找日志数据，DQL 作为全平台数据查询语言，其主要查询如下几种数据：

- 时序数据
- 日志数据
- 对象数据
- 应用性能追踪（APM）数据
- 用户行为检测（RUM）数据
- 关键事件数据
- 安全巡检数据
- ...

随着业务功能不断拓展，DQL 将封装更多不同的查询引擎（目前支持 InfluxDB 以及 ElasticSearch 两种）。其基本语法结构如下：

```python
namespace::measurement:(field-or-tag-list) { where-conditions } [time-range] BY-clause ORDER-BY-clause
```

如：

```
metric::cpu:(usage_system, usage_user) { usage_idle > 0.9 } [2d:1d:1h] BY hostname
```

此处，`metric` 指定了要查询时序数据（可简单理解成 MySQL 中的一个 DB），而 `cpu` 就是其中的一种指标集（类似于 MySQL 中的 Table），并且指定查找其中的两个字段 `usage_system` 和 `usage_user`；接着，`{...}` 中的表示过滤条件，最后 `[...]` 表示查询的时间范围：前天到昨天一段时间内，以 1h 为聚合间隔。

更多示例：

```
# 查询 K8s 中的 pod 对象（object）
object::kubelet_pod:(name, age) { cpu_usage > 30.0 } [10m] BY namespace

# 查找名为 my_service 应用的日志（message 字段）
logging::my_service:(message) [1d]

# 查看应用性能追踪（T 即 tracing）中，持续时间 > 1000us 的 span 数据，并且按照 operation 来分组
T::my_service { duration > 1000 } [10m] BY operation
```

## 横向对比

| 查询语言  | 主要领域                | 支持时序查询       | 支持日志查询 | 是否支持 HTTP API                                                  | 是否支持 Pipeline 切割              | 支持时间范围查找 | 支持 group by 聚合 |
| --------- | -------                 | ---                | -----        | ---------                                                          | ----                                | -----            | ---                |
| PromQL    | Prometheuse 指标查询    | 支持               | 不支持       | [支持](https://prometheus.io/docs/prometheus/latest/querying/api/) | 不支持                              | 支持             | [支持](https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators)               |
| LogQL     | 主要用于查询日志        | 支持从日志生成指标 | 支持         | [支持](https://grafana.com/docs/loki/latest/api/)                  | 支持                                | 支持             | [支持](https://grafana.com/docs/loki/latest/logql/#aggregation-operators)               |
| DQL       | DataFlux 全平台数据查询 | 支持               | 支持         | [支持](https://www.yuque.com/dataflux/datakit/apis#6c639732)       | 不支持（在 DataKit 端已预先切割好） | 支持             | 支持               |

### 数据处理函数支持情况

- [PromQL 支持的函数列表](https://prometheus.io/docs/prometheus/latest/querying/functions/#functions)
- [LogQL 支持的函数列表]((https://grafana.com/docs/loki/latest/logql/#metric-queries))
- [DQL 支持的函数列表](https://www.yuque.com/dataflux/doc/ziezwr)

## 常见查询语句写法对比

### 普通数据查询及过滤

```
# LogQL
{cluster="ops-tools1", namespace="dev", job="query-frontend"} |= "metrics.go" !="out of order" | logfmt | duration > 30s or status_code!="200"

# DQL
L::dev {cluster='ops-tools1', job='query=frontend', message != match("out of order"), (duraton > 30s OR stataus_code != 201)}

# PromQL（PromQL 不支持普通意义上的 OR 过滤）
http_requests_total{cluster='ops-tools1', job!='query=frontend', duration > 30s}
```

### 带聚合的查询以及过滤

```python
# LogQL
sum by (org_id) ({source="ops-tools",container="app-dev"} |= "metrics.go" | logfmt | unwrap bytes_processed [1m])

# PromQL
histogram_quantile(0.9, sum by (job, le) (rate(http_request_duration_seconds_bucket[10m])))

# DQL（注意，ops-tools 两边需加上 ``，不然被解析成减法表达式）
L::`ops-tools`:(bytes_processed) {filename = "metrics.go", container="app-dev"} [2m] BY sum(orig_id)
```
