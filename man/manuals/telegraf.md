{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Telegraf 接入

Telegraf 是一个开源的数据采集工具。DataKit 通过简单的配置即可接入 Telegraf 采集的数据集。

## Telegraf 安装

以 Ubuntu 为例，其他系统 [参见](https://docs.influxdata.com/telegraf/v1.18/introduction/installation/)：

添加安装源

```shell
curl -s https://repos.influxdata.com/influxdb.key | sudo apt-key add -
source /etc/lsb-release
echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
```

安装 Telegraf

```shell
sudo apt-get update && sudo apt-get install telegraf
```

### Telegraf 配置

默认配置文件路径：

- Mac: `/usr/local/etc/telegraf.conf`
- Linux: `/etc/telegraf/telegraf.conf`

修改配置文件如下：

```toml
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

# 此处可配置 Telegraf 所采集数据的全局 tag
[global_tags]
  name = "zhangsan"

[[outputs.http]]
    ## URL is the address to send metrics to DataKit ,required
    url         = "http://localhost:9529/v1/write/telegraf"
    method      = "POST"
    data_format = "influx" # 此处必须选择 `influx`，不然 DataKit 无法解析数据

# 更多其它配置 ...
```

Telegraf 采集到的数据通过 DataKit 的 HTTP 接口接入，此处的 `[[outputs.http]]` 段必须指向 DataKit 的 /v1/write/metric 接口。Telegraf 采集器的详细配置，[参见这里](https://docs.influxdata.com/telegraf)

其他插件 [参见这个列表](https://github.com/influxdata/telegraf#input-plugins)

### 注意事项

- 在 Telegraf 加入的标签（可通过 `[global_tags]` 配置），DataKit 不会覆盖它，因此可以通过这种方式屏蔽 DataKit 的全局标签（如 `host` 等）
- 如果 DataKit 已经存在的采集器（如 CPU、内存、网络等），不建议使用 Telegraf 再采集了，这可能造成数据冲突。
