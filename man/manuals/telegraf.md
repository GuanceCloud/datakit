{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Telegraf 接入

> 注意：建议在使用 Telegraf 之前，先确 DataKit 是否能满足期望的数据采集。如果 DataKit 已经支持，不建议用 Telegraf 来采集，这可能会导致数据冲突，从而造成使用上的困扰。

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
- Windows：配置文件就在 Telegraf 二进制同级目录（视具体安装情况而定）

> 注意： Mac 下，如果通过 [`datakit --install telegraf` 安装](datakit-tools-how-to#df09fa95)，则配置目录和 Linux 一样。

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
    url         = "http://localhost:9529/v1/write/metric?input=telegraf"
    method      = "POST"
    data_format = "influx" # 此处必须选择 `influx`，不然 DataKit 无法解析数据

# 更多其它配置 ...
```

如果 [DataKit API 位置有调整](datakit-conf-how-to#f2d2b6db)，需调整如下配置，将 `url` 设置到 DataKit API 真实的地址即可：

```toml
[[outputs.http]]
   ## URL is the address to send metrics to
   url = "http://127.0.0.1:9529/v1/write/metric?input=telegraf"
```

Telegraf 的采集配置跟 DataKit 类似，也是 [Toml 格式](https://toml.io/cn)，具体每个采集器基本都是以 `[[inputs.xxxx]]` 作为入口，这里以开启 `nvidia_smi` 采集为例：

```toml
[[inputs.nvidia_smi]]
  ## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
  bin_path = "/usr/bin/nvidia-smi"

  ## Optional: timeout for GPU polling
  timeout = "5s"
```

Telegraf 采集器的详细配置，[参见这里](https://docs.influxdata.com/telegraf)

更多 Telegraf 插件，[参见这个列表](https://github.com/influxdata/telegraf#input-plugins)
