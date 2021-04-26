{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

DataKit 之前版本会集成 Telegraf 插件，为了更好的使用和支持 Telegraf ，此前 DataKit 支持的 Telegraf 采集器全部废弃，用户如需使用 Telegraf 相关的采集器或者数据，需自行安装 Telegraf 并将数据打入 DataKit。

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

- macOS Homebrew: /usr/local/etc/telegraf.conf
- Linux debian and RPM packages: /etc/telegraf/telegraf.conf

修改配置文件如下：

```
[agent]
    interval = "10s"
    round_interval = true
    precision = "ns"
    collection_jitter = "0s"
    flush_interval = "10s"
    flush_jitter = "0s"
    metric_batch_size = 1000
    metric_buffer_limit = 100000
    logtarget = "file"
    logfile = "your_path.log"
    logfile_rotation_interval = ""

[[outputs.http]]
  ## URL is the address to send metrics to ,required
  url = "http://localhost:9529/v1/write/telegraf"
  method = "POST"
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


### 关于 Telegraf 数据 Tag 问题

- 可在 Telegraf 配置中加入 `golbal tag`,示例如下：

```shell
...
[global_tags]
  name = "zhangsan"
...

```

- DataKit 会默认加入一些 `golbal tag`，比如 `host`, Telegraf 数据如需覆盖 `host` ,需在 Telegraf 配置中加入 `host` Tag

### 关于 Telegraf 数据问题

如果 DataKit 采集器中有采集 `nginx` 指标集的数据，此时再开启 Telegraf 并写入 `nginx` 指标集的数据，数据可能会混乱甚至写入不成功，因此 DataKit 或 Telegraf 对于相同指标集的数据二选一