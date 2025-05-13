# How to Analyze a DataKit Bug Report

---

## Introduction to Bug Report {#intro}

As DataKit is typically deployed in user environments, various on-site data are required for troubleshooting. A Bug Report (hereinafter referred to as BR) is used to collect this information while minimizing the operations performed by on-site support engineers or users, thus reducing communication costs.

Through BR, we can obtain various on-site data of DataKit during its operation phase, according to the data directory below BR:

- *basic*: Basic machine environment information
- *config*: Various collection-related configurations
- *data*: Central configuration pull status
- *external*: eBPF related logging and profiles[:octicons-tag-24: Version-1.33.0](changelog.md#cl-1.33.0)
- *log*: DataKit's own program logs
- *metrics*: Prometheus metrics exposed by DataKit itself
- *profile*: Profile data of DataKit itself
- *pipeline*: Pipeline scripts[:octicons-tag-24: Version-1.33.0](changelog.md#cl-1.33.0)

Below, we will explain how to troubleshoot specific issues encountered through the information already available in these aspects.

## Viewing Basic Information {#basic}

The BR file name usually follows the format `info-<timestamp-ms>.zip`. With this timestamp (in milliseconds), we can determine the export time of the BR, which is meaningful for subsequent metric troubleshooting.

In the *info* file, the current machine's operating system information is collected, including kernel version, distribution version, hardware architecture, etc. These can assist us in troubleshooting issues.

In addition, if DataKit is installed in a container, it will also collect a bunch of user-side environmental variable configurations. All environment variables starting with `ENV_` are for DataKit's main configuration or collector configuration.

## Viewing Collector Configuration {#config}

Under the *config* directory, all collector configurations and DataKit's main configuration are collected, with all files suffixed with `.conf.copy`. When troubleshooting data issues, the configuration here is very helpful.

## Viewing Pulled Data {#pull}

Under the *data* directory, there is a hidden file named *.pull*(for newer version, the filename is *pull*), which contains several types of configuration information pulled from the server:

``` shell
cat data/.pull | jq
```

The result is a JSON, such as:

```json
{
  "dataways": null,
  "filters": {       # <--- This is the blacklist list
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

Sometimes, users report missing data, which is likely due to their configuration's blacklist discarding data. The blacklist rules here can help us troubleshoot this kind of data loss situation.

## Log Analysis {#logging}

Under the *log* directory, there are two files:

- *log*: This is the program running log of DataKit. The information inside may be incomplete because DataKit will periodically (default 32MB x 5) discard old logs.

In the *log* file, we can search for the `run ID`, and from then on, it is the log of a newly restarted run. Of course, it might not be found, in which case we can determine that the log has been Rotated.

- *gin.log*: This is the access log recorded by DataKit as an HTTP service.

When collectors like DDTrace are integrated, analyzing *gin.log* is beneficial for troubleshooting the data collection of DDTrace.

Other log troubleshooting methods can be found [here](why-no-data.md#check-log).

## Metric Analysis {#metrics}

Metric analysis is the focus of BR analysis. DataKit itself exposes a lot of [metrics](datakit-metrics.md#metrics). By analyzing these metrics, we can infer various behaviors of DataKit.

The following metrics have their own different labels (tags), and by synthesizing these labels, we can better locate problems.

### Data Collection Metrics {#collector-metrics}

There are several key metrics related to collection:

- `datakit_inputs_instance`: To know which collectors are enabled
- `datakit_io_last_feed_timestamp_seconds`: The last time each collector collected data
- `datakit_inputs_crash_total`: The number of times the collector crashed
- `datakit_io_feed_cost_seconds`: The duration of feed blocking. If this value is large, it indicates that the network upload(Dataway) may be slow, and blocking the collectors
- `datakit_io_feed_drop_point_total`: The number of data points discarded during feed (currently, only time series metrics are discarded when blocked)

By analyzing these metrics, we can roughly restore the running condition of each collector.

### Blacklist/Pipeline Execution Metrics {#filter-pl-metrics}

Blacklist/Pipeline is a user-defined data processing module, which has an important impact on data collection:

- The blacklist is mainly used to discard data. The rules written by the user may mistakenly kill some data, leading to incomplete data
- Pipeline, in addition to processing data, can also discard data (the `drop()` function). During the data processing process, the Pipeline script may consume a lot of time(such as complex regex match), and slow down the collector, thus leading to problems like log skipping[^log-skip].

[^log-skip]: The so-called log skipping refers to the collection speed not keeping up with the log generation speed. When the user's log is set with a rotate mechanism, the first log has not been collected, the second log is not collected in time, and is rotated by the third log that catches up, the second log is skipped here, the collector does not find the existence of the second log at all, and skips it directly to collect the third log file.

The main metrics involved are as follows[^metric-naming]:

- `pipeline_drop_point_total`: The number of points dropped by Pipeline
- `pipeline_cost_seconds`: The time taken for Pipeline to process points. If the time is long (reach to ms), it may lead to collector blocking
- `datakit_filter_point_dropped_total`: The number of points dropped by the blacklist

[^metric-naming]: Different versions of DataKit, the naming of Pipeline-related metrics may be different. Here only the common suffix names are listed.

### Data Upload Metrics {#dataway-metrics}

Data upload metrics mainly refer to some HTTP-related metrics of the Dataway reporting module.

- `datakit_io_dataway_point_total`: The total number of points uploaded (not necessarily all successfully uploaded)
- `datakit_io_dataway_http_drop_point_total`: During the upload process, if the data points still fail after retrying, DataKit will discard these data points
- `datakit_io_dataway_api_latency_seconds`: The time taken to call the Dataway API. If the time is long, it will block the operation of the collector
- `datakit_io_http_retry_total`: If the number of retries is high, it indicates that the network quality is not very good, and the center may be under a lot of pressure

### Basic Metrics {#basic-metrics}

Basic metrics mainly refer to some other metrics of DataKit, which include:

- `datakit_cpu_usage`: DataKit self CPU usage
- `datakit_heap_alloc_bytes/datakit_sys_alloc_bytes`: Golang runtime heap/sys memory metrics. If there is an OOM, it is generally the sys memory that exceeds the memory limit
- `datakit_uptime_seconds`: The duration that DataKit has been running. The startup duration is an important auxiliary metric
- `datakit_data_overuse`: If the workspace is overdue, DataKit's data reporting will fail, and the value of this metric is 1, otherwise it is 0
- `datakit_goroutine_crashed_total`: The count of crashed Goroutines. If some key Goroutines crashed, it will affect the behavior of DataKit

### Monitor Viewing {#monitor-play}

The built-in monitor command of DataKit can play some key metrics in BR. Compared with viewing pale numbers, it is more friendly:

```shell
$ datakit monitor -P info-1717645398232/metrics
...
```

Since the default BR will collect three sets of metrics (each set of data is about 10 seconds apart), when the monitor is playing, there will be real-time data updates.

### Invalid Metrics Issue {#invalid-metrics}

While BR can provide a lot of help when analyzing problems, many times when users find problems, they will restart DataKit and lose the scene, causing the data collected by BR to be invalid.

At this time, we can use the built-in [`dk` collector](../integrations/dk.md) of DataKit to collect its own data (it is recommended to add it to the collectors that start by default. The newer version of DataKit[:octicons-tag-24: Version-1.11.0](changelog.md#cl-1.11.0) has already done so), and report it to the user's space, which is equivalent to archiving DataKit's own metrics. And in the `dk` collector, you can further turn on all self-metric collection (this will consume more timelines)

- When installed in Kubernetes, turn on all DataKit self-metrics reporting through `ENV_INPUT_DK_ENABLE_ALL_METRICS`
- For host installation, modify `dk.conf`, and open the first metric comment in `metric_name_filter` (remove the comment `# ".*"`), which is equivalent to allowing all metrics to be collected

This will collect a copy of all the metrics exposed by DataKit to the user's workspace. In the workspace, search for `datakit` in the 'built-in views' (select 'DataKit(New)') to see the visual effect of these metrics.

## Pipeline {#pipeline}

[:octicons-tag-24: Version-1.33.0](changelog.md#cl-1.33.0)

If the user has configured Pipelines, we'll get a copy of these Pipeline scripts in the *pipeline* directory. By examining these Pipelines, we can identify issues with data field parsing; if certain Pipelines are found to be time-consuming, we can also offer optimization suggestions to reduce the resource consumption of the Pipeline scripts.

## External {#external}

[:octicons-tag-24: Version-1.33.0](changelog.md#cl-1.33.0)

In the *external* directory, logs and debug information from external collectors (currently primarily eBPF collector) are gathered to facilitate troubleshooting issues related to these external collectors.

## Profile Analysis {#profile}

Profile analysis is mainly aimed at developers. Through the profile in BR, we can analyze the hotspots of memory/CPU consumption of DataKit at the moment of BR. Through these profile analyses, we can guide us to better optimize the existing code or find some potential bugs.

Under the *profile* directory, there are the following files:

- *allocs*: The total amount of memory allocated since the start of DataKit. Through this file, we can know where the heavy memory allocation is. Some places may not need to allocate so much memory
- *heap*: The current (at the moment of collecting BR) distribution of memory usage. If there is a memory leak, it is very likely to be seen here (memory leaks generally occur in modules that do not need so much memory, which is basically easy to find out)
- *profile*: View the CPU consumption of the current DataKit process. Some unnecessary modules may consume too much CPU (such as high-frequency JSON parsing operations)

The other files (*block/goroutine/mutex*) are not currently used for troubleshooting.

Through the following command, we can view these profile data in the browser (it is recommended to use Golang above 1.20, its visualization effect is better):

```shell
go tool pprof -http=0.0.0.0:8080 profile/heap
```

We can do an alias in the shell for easy operation:

```shell
# /your/path/to/bashrc
__gtp() {
    port=$(shuf -i 40000-50000 -n 1) # Random a port between 40000 ~ 50000

    go tool pprof -http=0.0.0.0:${port} ${1}
}
alias gtp='__gtp'
```

```shell
source /your/path/to/bashrc
```

You can directly use the following command:

```shell
gtp profile/heap
```

## Summary {#conclude}

Although BR may not be able to solve all problems, it can avoid a lot of communication information differences and misguidance. It is still recommended that everyone provide the corresponding BR when reporting problems. At the same time, the existing BR will continue to improve, by exposing more metrics, collecting more other aspects of environmental information (such as Tracing-related client information, etc.), and further optimizing the experience of troubleshooting problems.

