# Comparison of DQL and several other query languages

---

DQL is the unified query language of Observation Cloud. In order to make it easier for everyone to learn this language, below we select several different query languages to compare with them, so that everyone can understand and use DQL more quickly.

Here we temporarily choose two languages: [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/){:target="_blank"} and [LogQL](https://grafana.com/docs/loki/latest/logql/){:target="_blank"}. The well-known SQL statement is very different from DQL in its form and function, so it will not be mentioned here.

PromQL is a query language used in [Prometheus](https://prometheus.io/){:target="_blank"} to query its time series data; LogQL is a log query language used in [Grafana Loki](https://grafana.com/oss/loki/){:target="_blank"}. Like DQL, it draws on the syntax structure of PromQL. Overall, the three languages have similar structures, but they differ in subtle ways. The following will elaborate on the following aspects:

- Differences in basic grammatical structures
- Supported commonly used predefined functions
- Comparison of commonly used query writing methods

## Basic syntax structure {#syntax}

| query language | basic structure                                                                                                           |
| -------------- | ------------------------------------------------------------------------------------------------------------------------- |
| PromQL         | `metric-name {conditions} [start-time:end-time]`                                                                          |
| LogQL          | `{stream-selector} log-pipeline`                                                                                          |
| DQL            | `namespace::measurement:(metric-list) [start-time:end-time:time-interval] { conditions } GROUP-BY-clause ORDER-BY-clause` |

They are explained below.

### PromQL {#p}

In Prometheus, relevant metrics are organized in discrete form. In its query, you can directly find the corresponding indicators, such as:

``` not-set
http_requests_total{environment="prometheus", method!="GET"}
```

Here we look for the metric `http_requests_total` and filter the data by specifying its label constraints (`environment` and `method`).

> Note: PromQL calls the label constraints here Label Matchers.

### LogQL {#l}

As the name suggests, LogQL is mainly used for log content query, such as:

``` not-set
{container="query-frontend", namespace="loki-dev"} |= "metrics.go" | logfmt | duration > 10s and throughput_mb < 500
```

Here in `{...}`, LogQL calls it Stream Selector, which is designed to delineate the data query range (similar to the `FROM...` part in SQL); the half part is called Log Pipeline , which mainly deals with the extraction and filtering of log information.

### DQL {#d}

DQL has relatively comprehensive coverage. Compared with PromQL, which can only be used to find time series data in Prometheus, and LogQL, which can only be used to find log data, DQL, as a full-platform data query language, mainly queries the following types of data:

- Metric
- Logging
- Object
- Tracing
- RUM
- KeyEvent
- Security
- ...

As business functions continue to expand, DQL will encapsulate more different query engines (currently supporting InfluxDB and ElasticSearch). Its basic grammatical structure is as follows:

```python
namespace::measurement:(field-or-tag-list) { where-conditions } [time-range] BY-clause ORDER-BY-clause
```

如：

``` not-set
metric::cpu:(usage_system, usage_user) { usage_idle > 0.9 } [2d:1d:1h] BY hostname
```

Here, `metric` specifies that time series data is to be queried (can be simply understood as a DB in MySQL), and `cpu` is one of the metric sets (similar to Table in MySQL), and two of them are specified to be searched. The fields `usage_system` and `usage_user`; then, the ones in `{...}` represent the filter conditions, and finally `[...]` represents the time range of the query: from the day before yesterday to yesterday, with 1h as the aggregation interval.

More examples:

```not-set
# Query the pod object in K8s
object::kubelet_pod:(name, age) { cpu_usage > 30.0 } [10m] BY namespace

# Find the log of the application named my_service (message field)
logging::my_service:(message) [1d]

# View span data with duration > 1000us in application performance tracing (T stands for tracing), and group them by operation
T::my_service { duration > 1000 } [10m] BY operation
```

## Horizontal comparison {#compare}

| query language | Main areas                        | Support time series query             | Support log query | Whether to support HTTP API                                                          | Whether to support Pipeline               | Support time range search | Support group by aggregation                                                                                        |
| -------------- | --------------------------------- | ------------------------------------- | ----------------- | ------------------------------------------------------------------------------------ | ----------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| PromQL         | Prometheus metric query           | support                               | not support       | [support](https://prometheus.io/docs/prometheus/latest/querying/api/){:target="_blank"} | not support                               | support                   | [support](https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators){:target="_blank"} |
| LogQL          | Mainly used to query logs         | Supports generating metrics from logs | support           | [support](https://grafana.com/docs/loki/latest/api/){:target="_blank"}               | support                                   | support                   | [support](https://grafana.com/docs/loki/latest/logql/#aggregation-operators){:target="_blank"}                      |
| DQL            | DataFlux full platform data query | support                               | support           | [support](apis.md#api-raw-query){:target="_blank"}                                   | not support (Pre-cut on the DataKit side) | support                   | support                                                                                                             |

### Data processing function support {#funcs}

- [PromQL supported functions](https://prometheus.io/docs/prometheus/latest/querying/functions/#functions){:target="_blank"}
- [LogQL supported functions](https://grafana.com/docs/loki/latest/logql/#metric-queries){:target="_blank"}
- [DQL supported functions](../dql/funcs.md){:target="_blank"}

<!-- markdownlint-disable MD013 -->
## Comparison of common query statement writing methods {#basic-query}
<!-- markdownlint-enable -->
### General data query and filtering {#q-filter}

```not-set
# LogQL
{cluster="ops-tools1", namespace="dev", job="query-frontend"} |= "metrics.go" !="out of order" | logfmt | duration > 30s or status_code!="200"

# DQL
L::dev {cluster='ops-tools1', job='query=frontend', message != match("out of order"), (duraton > 30s OR stataus_code != 201)}

# PromQL（PromQL does not support OR filtering in the ordinary sense）
http_requests_total{cluster='ops-tools1', job!='query=frontend', duration > 30s}
```

### Querying and filtering with aggregation {#q-groupby}

```python
# LogQL
sum by (org_id) ({source="ops-tools",container="app-dev"} |= "metrics.go" | logfmt | unwrap bytes_processed [1m])

# PromQL
histogram_quantile(0.9, sum by (job, le) (rate(http_request_duration_seconds_bucket[10m])))

# DQL (note that ops-tools needs to be added with `` on both sides, otherwise it will be parsed into a subtraction expression)
L::`ops-tools`:(bytes_processed) {filename = "metrics.go", container="app-dev"} [2m] BY sum(orig_id)
```
