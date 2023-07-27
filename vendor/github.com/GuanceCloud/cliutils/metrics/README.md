# Prometheuse 指标体系构建

任何一个服务需要有完整有效的指标体系来做自观测，本文说明如何使用 Prometheuse 方式来构建服务的指标体系。以下以 Datakit 为例来做说明。

## 如何给模块增加自身指标

对 Datakit 而言，自身的指标有几大类：

- 计数：某些操作的执行次数（如日志文件 rotate 次数）、某些物理量的计数（如字节数）
- 耗时：某些操作的耗时
- 报错信息：当某个模块报错时，也能通过 Prometheuse 方式来暴露，这种有点像日志。可以直接使用全局 metrics 模块提供的接口

## 计数（counter）

对于计数性质的指标，一般通过如下方式构建：

```golang
dwAPIVec = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Namespace : "datakit", // 所有 Datakit 指标，都统一用这个 Namespace
        Subsystem : "dataway", // 视具体模块而定，比如这里的 dataway 指 io 中的 dataway 模块
        Name      : "api_total", // 具体的指标名，比如这里指 HTTP API 的发送次数（当前是一个 counter）

        // 这样之后，就能看到一个完整的指标名：datakit_dataway_api_total

        Help: "The dataway API request count", // 该指标的说明文档
    },
    []string{"api", "status"}, // 可能的 label(tag) 名称
)

// 将该指标注册到 datakit 全局指标体系中
// NOTE: 注意，不要注册同样的指标名，不然这里会崩溃
metrics.MustRegister(dwAPIVec)
```

注意这里的 label 列表，对 dataway HTTP 请求而言，除了区分不同的 API 调用，每个 API 请求有不同的结果状态，故需要将状态这个维度添加进去，以区分不同结果状态下的指标。

注册完成后，在该模块代码中，就能无脑使用了：

```golang
// 请求 /v1/write/metric 成功
dwAPIVec.WithLabelValues("/v1/write/metric", "Status OK").Inc()
```

注意：

- `WithLabelValues` 的时候，label 值的顺序，必须跟注册该指标时指定的 label 列表顺序一致，不然会导致数据错乱
- `WithLabelValues` 中的 label 个数不能少于注册时的个数，否则会崩溃

在最终的 /metrics 接口返回中，能看到类似如下的返回：

```
# HELP datakit_dataway_api_total The dataway API request count
# TYPE datakit_dataway_api_total counter
datakit_dataway_api_total{api="/v1/write/metric",status="Status OK"} 169
```

## 计量（counter）

计量用于表示一些忽高忽低的指标，比如温度、CPU 使用率等，它不是单调递增的，

## 概要（summary）

概要用来表示一种自启动以来的总次数和总量之间的关系，比如网络请求总耗时和总次数，API 调用的总 Body 大小和次数，它自带两个字段：

- Sample Count：表示总次数
- Sample Sum：表示总量

比如，以 HTTP 请求为例，一个指标维度就是总的请求次数和总的请求耗时，即可组成一个 summary 指标：

```golang
httpLatencyVec = prometheuse.NewSummaryVec(
    prometheus.SummaryOpts{
        Namespace : "datakit",
        Subsystem : "http",

        Name : "api_cost_seconds", // 具体的指标名
        Help : "API request cost",
    },
    []string{"api", "status"}, // 这里的维度跟上面的基本一致
)
```

summary 可以这么用：

```golang
httpLatencyVec.WithLabelValues("/v1/write/metric", "Status OK").Observe(float64(time.Since(start))/float64(time.Second))
```

在最终的 /metrics 接口返回中，能看到类似如下的返回：

```
# HELP datakit_dataway_api_latency Dataway API request latency(ms)
# TYPE datakit_dataway_api_latency summary
datakit_http_api_cost_seconds_sum{api="/v1/write/metric",status="Status OK"}"} 3.1415926
datakit_http_api_cost_seconds_count{api="/v1/write/metric",status="Status OK"}"} 42
```

它表示「在 API `/v1/write/metric` 上总共有 42 次请求，总的请求耗时为 3.1415926 秒」，通过简单的除法，我们即可知道该 API 上的平均耗时。

### 概要的百分位

上面的方式只能计算平均值，但是我们可以在 summary 中设置一定的百分位，来获取最近一段时间的数据：

```golang
httpLatencyVec = prometheuse.NewSummaryVec(
    prometheus.SummaryOpts{
        Namespace : "datakit",
        Subsystem : "http",

        Name : "api_cost_seconds", // 具体的指标名
        Help : "API request cost",

        Objectives: map[float64][float64] {
            0.5:  0.05,
            0.75: 0.0075,
            0.95: 0.005,
        },
        MaxAge: 10 * time.Minute,
        AgeBuckets: 5,
    },
    []string{"api", "status"}, // 这里的维度跟上面的基本一致
)
```

这样，我们就能获取最近 10min 每个 API 上几个百分位（P50/P75/P95）的响应情况：

```
# HELP datakit_dataway_api_latency Dataway API request latency(ms)
# TYPE datakit_dataway_api_latency summary
datakit_http_api_cost_seconds{api="/v1/write/metric",status="Status OK",quantile="0.5"} 1.002858834
datakit_http_api_cost_seconds{api="/v1/write/metric",status="Status OK",quantile="0.75"} 1.002858834
datakit_http_api_cost_seconds{api="/v1/write/metric",status="Status OK",quantile="0.95"} 1.002858834
datakit_http_api_cost_seconds_sum{api="/v1/write/metric",status="Status OK"}"} 3.1415926
datakit_http_api_cost_seconds_count{api="/v1/write/metric",status="Status OK"}"} 42
```

## 指标命名规范

Prometheus 指标有自身的命名规范，参见[官方文档](https://prometheus.io/docs/practices/naming/)。

## 报错

> 报错信息的处理，还需进一步考虑一下，暂时不建议使用。

```golang
AddLastErr("my-module", "I got some error message")
```

在最终的 /metrics 接口返回中，能看到类似如下的返回：

```
# HELP datakit_lasterr Datakit internal errors(with error occurred unix timestamp)
# TYPE datakit_lasterr gauge
datakit_lasterr{message="I got some error message",source="my-module"} 1.678531768e+09
```

## 测试指标效果

如果如下代码，可以直接在自己的代码中获取 /metrics 接口返回的数据效果：

```golang

import (
    "github.com/prometheus/common/expfmt"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/GuanceCloud/cliutils/metrics"
)

apiLatencyVec = prometheuse.NewSummaryVec(
    prometheus.SummaryOpts{
        Namespace : "datakit",
        Subsystem : "dataway",

        Name : "api_cost_seconds",
        Help : "Dataway API request latency",
    },
    []string{"api", "status"},
)

// 构建一个 registry
reg := prometheus.NewRegistry()
reg.MustRegister(apiLatencyVec)

// 塞进去一个指标
apiLatencyVec.WithLabelValues("/v1/write/metric", "Status OK").Observe(float64(time.Since(start))/float64(time.Second))

// 获取 reg 上所有指标
mfs, err := reg.Gather()

// 此处即可看到 /metrics 接口返回的效果
fmt.Println(metrics.MetricFamily2Text(mfs))
```

## 使用用例

参见 [diskcache 的指标暴露实现](https://github.com/GuanceCloud/cliutils/blob/main/diskcache/metric.go)。
