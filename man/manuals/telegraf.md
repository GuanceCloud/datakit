{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

Telegraf 是一个用 Go 编写的代理程序，是收集和报告指标和数据的代理。为了更好的使用和支持 Telegraf，DataKit 支持 Telegraf 数据接入

## Telegraf 使用并接入 DataKit 

### Telegraf 安装

以 Ubuntu 为例，其他系统 [参见](https://docs.influxdata.com/telegraf/v1.18/introduction/installation/)：

- 添加 `InfluxData repository`

```shell
curl -s https://repos.influxdata.com/influxdb.key | sudo apt-key add -
source /etc/lsb-release
echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
```

- 安装 Telegraf

```shell
sudo apt-get update && sudo apt-get install telegraf
```

### Telegraf 配置

默认配置文件路径：

- macOS Homebrew: `/usr/local/etc/telegraf.conf`
- Linux debian and RPM packages: `/etc/telegraf/telegraf.conf`

修改配置文件如下：

```
[agent]
    interval                  = "10s"
    round_interval            = true
    precision                 = "ns"
    collection_jitter         = "0s"
    flush_interval            = "10s"
    flush_jitter              = "0s"
    metric_batch_size         = 1000
    metric_buffer_limit       = 100000
    logtarget                 = "file"
    logfile                   = "your_path.log"
    logfile_rotation_interval = ""

[[outputs.http]]
    ## URL is the address to send metrics to DataKit ,required
    url         = "http://localhost:9529/v1/write/telegraf"
    method      = "POST"
    data_format = "influx"
    
```

DataKit 是通过 http 接入 Telegraf 数据，此处的 `outputs.http` 为必填字段，Telegraf Agent 详细配置，[参见](https://docs.influxdata.com/telegraf/v1.15/administration/configuration/)

加入 `Telegraf plugins`,以 `mem` 为例：

```shell
...
[[inputs.mem]]
# Read metrics about swap memory usage
#   # gather_memory_contexts = false
#   #   "/sys/fs/cgroup/memory",
#   #   "/sys/fs/cgroup/memory/child1",
#   #   "/sys/fs/cgroup/memory/child2/*",
#   # files = ["memory.*usage*", "memory.limit_in_bytes"]
#     "jvm.memory.pools.Metaspace.committed"
#   ## If true, collect telegraf memory stats.
#   # collect_memstats = true
#   ## Remember to change host address to fit your environment.
#   ## This collect all heap memory usage metrics.
#     name = "heap_memory_usage"
```

其他插件 [参见](https://docs.influxdata.com/telegraf/v1.15/plugins/)

- 开启 Telegraf

```shell
sudo service telegraf start
```


### 关于 Telegraf 数据全局标签问题

- 可在 Telegraf 配置中加入全局标签,示例如下：

```shell
...
[global_tags]
  name = "zhangsan"
...

```

- 在 Telegraf 加入的标签，DataKit 就不会覆盖它，因此可以通过这种方式屏蔽 DataKit 的全局标签

### 关于 Telegraf 和 DataKit 采集器冲突问题

- 如果 Datakit 已经存在的采集器，就不需要再使用 Telegraf 再采集了。在一般情况下，至少在 DataFlux 中，针对某一个采集器，DataKit 的采集会比 Telegraf 做的更好

