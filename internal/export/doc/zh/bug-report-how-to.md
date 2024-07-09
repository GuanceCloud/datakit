# 如何分析 Datakit Bug Report

---

## Bug Report 介绍 {#intro}

由于 Datakit 一般部署在用户环境，为了排查问题，需要获取各种现场数据，Bug Report（后面简称 BR） 就是用来收集这些信息，同时避免现场支持工程师或用户执行太多操作，降低沟通成本。

通过 BR，我们能获取各种 Datakit 在运行阶段的现场数据，按照 BR 下面的数据目录：

- *basic*：机器基本环境信息
- *config*：各种采集相关的配置
- *data*：中心的配置拉取情况
- *log*：Datakit 自身的程序日志
- *metrics*：Datakit 自身暴露的 Prometheus 指标
- *profile*：Datakit 自身 Profile 数据

下面就以上各个方面，分别说明如何通过这些已有的信息，来排查遇到的具体问题。

## 基本信息查看 {#basic}

BR 文件名形式一般为 `info-<timestamp-ms>.zip`，通过这个时间戳（单位毫秒），我们即可得知该 BR 的导出时间，这对后续的指标排查是有意义的。

在 *info* 文件中，会收集当前机器的操作系统信息，包括内核版本、发行版本、硬件架构等等。这些可以辅助我们排查问题。

除此之外，如果 Datakit 是容器安装的，还会收集一堆用户侧的环境变量配置情况，所有以 `ENV_` 开头的环境变量都是针对 Datakit 主配置或采集器配置。

## 查看采集配置 {#config}

在 *config* 目录下，收集了所有采集器的配置以及 Datakit 主配置，所有文件都以 `.conf.copy` 作为后缀。在排查数据问题时，这里的配置情况非常有帮助。

## 查看中心同步数据 {#pull}

在 *data* 目录下，有一个 *.pull* 的隐藏文件（较新的版本这个文件名为 *pull*，不再隐藏了），里面还有几类从中心拉取下来的配置信息：

``` shell
cat data/.pull | jq
```

结果是一个 JSON，如：

```json
{
  "dataways": null,
  "filters": {       # <--- 这里是黑名单列表
    "logging": [
      "{ ... }"
    ],
    "rum": [
      "{ ... }"
    ],
    "tracing": [
      "{ ... }",
    ]
  },
  "pull_interval": 10000000000,
  "remote_pipelines": null
}
```

有时候，用户会反馈数据缺失，很有可能是其配置的黑名单将数据丢弃了。这里的黑名单规则可以辅助我们排查这种数据丢失的情况。

## 日志分析 {#logging}

在 *log* 目录下，有两个文件：

- *log*：这是 Datakit 的程序运行日志。里面的信息可能不完整，因为 Datakit 会定期（默认 32MB）丢弃老的日志

在 *log* 文件中，我们可以搜索一下 `run ID`，自此以后，才是一个重新启动的运行日志。当然，可能搜索不到，这时候可以判断日志被 Rotate 掉了。

- *gin.log*：这是 Datakit 作为 HTTP 服务所记录的 access log

在接入了 DDTrace 等这类采集器的时候，分析 *gin.log* 有利于排查 DDTrace 数据采集的情况。

其它日志排查方式，参见[这里](why-no-data.md#check-log)。

## 指标分析 {#metrics}

指标分析是 BR 分析的重点，Datakit 自己暴露了非常多的[自身指标](datakit-metrics.md#metrics)，通过分析指标，我们能推断出 Datakit 各种行为。

以下各种指标都有各自不同的标签（label/tag），综合这些标签，能更好的定位问题。

### 数据采集指标 {#collector-metrics}

采集有关的指标有几个关键指标：

- `datakit_inputs_instance`：可知开启了哪些采集器
- `datakit_io_last_feed_timestamp_seconds`：各个采集器上次采集到数据的时间
- `datakit_inputs_crash_total`：采集器崩溃次数
- `datakit_io_feed_cost_seconds`：feed 阻塞时长，如果这个值较大，表明网络上传可能较慢，阻塞了采集器正常采集
- `datakit_io_feed_drop_point_total`：feed 时丢弃的数据点数（目前默认只有时序指标在阻塞时会丢弃）

综合分析上面这些指标，大概能还原各个采集器的运行情况。

### 黑名单/Pipeline 执行指标 {#filter-pl-metrics}

黑名单/Pipeline 是用户自定义的数据处理模块，这部分对数据的采集有重要影响：

- 黑名单主要用来丢弃数据，用户编写的规则可能会误杀一些数据，导致数据不完整
- Pipeline 除了处理数据之外，也可以丢弃数据（`drop()` 函数）。在处理数据的过程中，可能其 Pipeline 脚本消耗很大（比如复杂的正则表达式匹配），使得采集来不及，进而导致像日志跳档[^log-skip]这样的问题。

[^log-skip]: 所谓日志跳档，指采集的速度赶不上日志产生的速度。当用户日志设置了 rotate 机制的时候，第一个日志尚未采集完，第二个日志来不及采集，而被迎头赶上的第三个日志覆盖，此处第二个日志就跳档了，采集器根本就不会发现有第二个日志存在，进而跳过它，直接采集第三个日志文件

主要涉及的指标如下[^metric-naming]：

- `pipeline_drop_point_total`：Pipeline 丢弃的 point 数
- `pipeline_cost_seconds`：Pipeline 处理 point 的耗时，如果耗时较长（ms 级别），则可能导致采集阻塞
- `datakit_filter_point_dropped_total`：黑名单丢弃的 point 数

[^metric-naming]: 不同版本 Datakit，Pipeline 有关的指标命名可能不同。此处只列举它们共同的后缀名

### 数据上传指标 {#dataway-metrics}

数据上传指标主要指 Dataway 上报模块的一些 HTTP 有关的指标。

- `datakit_io_dataway_point_total`：上传总点数（不一定全部上传成功）
- `datakit_io_dataway_http_drop_point_total`：上传过程中，重传后仍失败，Datakit 会丢弃这些数据点
- `datakit_io_dataway_api_latency_seconds`：调用 Dataway API 的耗时。如果耗时较大，会阻塞采集器的运行
- `datakit_io_http_retry_total`：retry 数如果较多，表明网络质量不太好，也可能中心的压力很大

### 基础指标 {#basic-metrics}

基础指标主要指 Datakit 运行期间一些业务指标，它们包括：

- `datakit_cpu_usage`：CPU 消耗
- `datakit_heap_alloc_bytes/datakit_sys_alloc_bytes`：Golang 运行时 heap/sys 两种内存指标。如果出现 OOM，一般是后者的内存超过了内存限制
- `datakit_uptime_seconds`：Datakit 启动时长。启动时长是一个重要辅助指标
- `datakit_data_overuse`：如果工作空间欠费，Datakit 上报数据会失败，这个指标的值就是 1，否则为 0
- `datakit_goroutine_crashed_total`：崩溃的 Goroutine 计数。如果一些关键 Goroutine 崩溃，会影响 Datakit 的正常运行

### Monitor 查看 {#monitor-play}

Datakit 内置的 monitor 命令能播放 BR 中的一些关键指标，相当于一种可视化方式，相比查看苍白的数字，它显得更加友好一点：

```shell
$ datakit monitor -P info-1717645398232/metrics
...
```

由于默认 BR 会收集三份 metrics（每份数据相差 10s 左右），monitor 播放的时候，会有实时的数据更新。

### 指标无效问题 {#invalid-metrics}

BR 在分析问题时能提供非常多的帮助，但是很多时候，用户发现问题的时候，会重启 Datakit 进而丢失现场，导致 BR 收集到的数据无效。

此时我们可以通过 Datakit 内置的 [`dk` 采集器](../integrations/dk.md) 来采集其自身数据（建议将其添加到默认启动的采集器中，较新的 Datakit 版本[:octicons-tag-24: Version-1.11.0](changelog.md#cl-1.11.0)已经这么做了），上报给用户的空间，这相当于将 Datakit 自身的指标存档了。而在 `dk` 采集器中，可以更进一步开启所有自身指标采集（这会消耗更多时间线）

- Kubernetes 中安装时，通过 `ENV_INPUT_DK_ENABLE_ALL_METRICS` 来开启所有 Datakit 自身指标上报
- 主机安装，修改 `dk.conf`，在 `metric_name_filter` 中，打开第一个指标注释（`# ".*"`），相当于放行所有指标采集

这样会将 Datakit 暴露的所有指标都采集一份到用户的工作空间。在工作空间中，通过「内置视图」中搜索 `datakit`（选择「Datakit(New)」），即可看到这些指标的可视化效果。

## Profile 分析 {#profile}

Profile 分析主要面向开发者，通过 BR 中的 profile，我们能分析出在 BR 收集那一刻 Datakit 的内存/CPU 开销热点，透过这些 profile 分析，能指导我们更好的优化现有的代码，或者发现一些潜在的 bug。

在 *profile* 目录下，有如下一些文件：

- *allocs*：自 Datakit 启动一来的内存分配总量。通过这个文件我们能得知内存分配的重头在何处。有些地方可能没必要分配那么多内存
- *heap*：当前（收集 BR 那一刻）内存占用的分布。如果存在内存泄漏，这里大概率能看出来（内存泄漏一般发生在不需要那么多内存的模块，基本很容易看出来）
- *profile*：查看当前 Datakit 进程的 CPU 消耗。一些不必要的模块可能消耗了太多了 CPU（比如高频的 JSON 解析操作）

其它几个文件（*block/goroutine/mutex*）目前尚未用于问题排查。

通过如下命令，我们可以在浏览器中查看这些 profile 数据（建议用 Golang 1.20 以上的版本，它的可视化效果更好）：

```shell
go tool pprof -http=0.0.0.0:8080 profile/heap
```

我们可以在 shell 中做一个 alias，便于操作：

```shell
# /your/path/to/bashrc
__gtp() {
    port=$(shuf -i 40000-50000 -n 1) # 随机一个 40000 ~ 50000 之间的端口

    go tool pprof -http=0.0.0.0:${port} ${1}
}
alias gtp='__gtp'
```

```shell
source /your/path/to/bashrc
```

直接用如下命令即可：

```shell
gtp profile/heap
```

## 总结 {#conclude}

虽然 BR 不一定能解决所有问题，但能避免很多沟通上的信息差以及误导，还是建议大家在反馈问题的时候，提供对应的 BR。同时现有的 BR 也会不断改进，通过暴露更多的指标，收集更多其它方面的环境信息（比如 Tracing 有关的客户端信息等），进一步优化问题排查的体验。
