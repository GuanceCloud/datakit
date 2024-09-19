# Telegraf Data Access

---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple:

---

<!-- markdownlint-disable MD046 -->
???+ attention

    Before using Telegraf, it is recommended to determine whether DataKit can meet the expected data collection. If DataKit is already supported, Telegraf is not recommended for collection, which may lead to data conflicts and cause problems in use.

Telegraf is an open source data collection tool. DataKit can access the data set collected by Telegraf through simple configuration.
<!-- markdownlint-enable -->

## Telegraf Installation {#install}

Take Ubuntu as an example. For other systems, [see](https://docs.influxdata.com/telegraf/v1.18/introduction/installation/){:target="_blank"}：

Add installation source

```shell
curl -s https://repos.influxdata.com/influxdb.key | sudo apt-key add -
source /etc/lsb-release
echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
```

Install Telegraf

```shell
sudo apt-get update && sudo apt-get install telegraf
```

### Telegraf Configuration {#config}

Default profile path:

- Mac: `/usr/local/etc/telegraf.conf`
- Linux: `/etc/telegraf/telegraf.conf`
- Windows: The configuration file is in the Telegraf binary sibling directory (depending on the specific installation)

> Note: Under Mac, if you [install through `datakit install --telegraf`](datakit-tools-how-to#extras), the configuration directory is the same as Linux.

Modify the configuration file as follows:

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

# Here you can configure the global tag of the data collected by Telegraf
[global_tags]
  name = "zhangsan"

[[outputs.http]]
    ## URL is the address to send metrics to DataKit ,required
    url         = "http://localhost:9529/v1/write/metric?input=telegraf"
    method      = "POST"
    data_format = "influx" # You must select `influx` here, otherwise DataKit cannot parse the data

# More other configurations ...
```

If the [DataKit API location is adjusted](datakit-conf#config-http-server), you need to adjust the following configuration by setting the `url` to the real address of the DataKit API:

```toml
[[outputs.http]]
   ## URL is the address to send metrics to
   url = "http://127.0.0.1:9529/v1/write/metric?input=telegraf"
```

Telegraf's collection configuration is similar to DataKit, and it is also in [Toml format](https://toml.io/cn){:target="_blank"}Specifically, each collector basically takes `[[inputs.xxxx]]` as the entrance. Here, take `nvidia_smi` collection as an example:

```toml
[[inputs.nvidia_smi]]
  ## Optional: path to nvidia-smi binary, defaults to $PATH via exec.LookPath
  bin_path = "/usr/bin/nvidia-smi"

  ## Optional: timeout for GPU polling
  timeout = "5s"
```

## More Reading {#more-reading}

- [Detailed configuration of Telegraf collector](https://docs.influxdata.com/telegraf){:target="_blank"}
- [More Telegraf plug-ins](https://github.com/influxdata/telegraf#input-plugins){:target="_blank"}
