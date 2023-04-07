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

## 耗时

耗时性质的指标，有两个维度的计数，一个是次数，一个是耗时，故这里可以使用 summary：

```golang
dwAPILatencyVec = prometheuse.NewSummaryVec(
    prometheus.SummaryOpts{
        // 复用上面的 Namespace 和 Subsystem
        Namespace : "datakit",
        Subsystem : "dataway",

        Name : "api_latency", // 具体的指标名，比如这里指 HTTP API 的发送次数（当前是一个 counter）
        Help : "Dataway API request latency(ms)",
    },
    []string{"api", "status"}, // 这里的维度跟上面的基本一致
)

// 将该指标注册到 datakit 全局指标体系中
metrics.MustRegister(dwAPILatencyVec)
```

summary 可以这么用：

```golang
start := time.Now()

... // do request

// 请求 /v1/write/metric 成功
dwAPILatencyVec.WithLabelValues("/v1/write/metric", "Status OK").Observe(float64(time.Since(start)))
```

在最终的 /metrics 接口返回中，能看到类似如下的返回：

```
# HELP datakit_dataway_api_latency Dataway API request latency(ms)
# TYPE datakit_dataway_api_latency summary
datakit_dataway_api_latency_sum{api="/v1/write/metric",status="Status OK"}"} 178854
datakit_dataway_api_latency_count{api="/v1/write/metric",status="Status OK"}"} 1357
```

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

dwAPILatencyVec = prometheuse.NewSummaryVec(
    prometheus.SummaryOpts{
        Namespace : "datakit",
        Subsystem : "dataway",

        Name : "api_latency",
        Help : "Dataway API request latency(ms)",
    },
    []string{"api", "status"},
)

// 构建一个 registry
reg := prometheus.NewRegistry()
reg.MustRegister(dwAPILatencyVec)

// 塞进去一个指标
dwAPILatencyVec.WithLabelValues("/v1/write/metric", "Status OK").Observe(float64(time.Since(start)))

// 获取 reg 上所有指标
mfs, err := reg.Gather()

// 此处即可看到 /metrics 接口返回的效果
fmt.Println(metrics.MetricFamily2Text(mfs))
```

## 使用用例

参见 [diskcache 的指标暴露实现](https://github.com/GuanceCloud/cliutils/blob/main/diskcache/metric.go)。
