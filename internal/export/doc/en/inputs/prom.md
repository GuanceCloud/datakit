---
title     : 'Prometheus Exporter'
summary   : 'Collect metrics exposed by Prometheus Exporter'
__int_icon      : 'icon/prometheus'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Prometheus Exporter Data Collection
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The Prom collector can obtain all kinds of metric data exposed by Prometheus Exporters, so long as the corresponding Exporter address is configured, the metric data can be accessed.

## Configuration {#config}

### Preconditions {#requirements}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
Only metric data in Prometheus form can be accessed.
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

???+ attention "Configuration of interval"

    Prometheus' metrics collection will cause some overhead (HTTP request) to the target service. To prevent unexpected configuration, the collection interval is currently 30s by default, and the configuration items are not obviously released in conf. If you must configure the collection interval, you can add this configuration in conf:

    ``` toml hl_lines="2"
    [[inputs.prom]]
        interval = "10s"
    ```
<!-- markdownlint-enable -->
### Configure Extra header {#extra-header}

The Prom collector supports configuring additional request headers in HTTP requests for data pull, as follows:

```toml
  [inputs.prom.http_headers]
  Root = "passwd"
  Michael = "1234"
```

### About Tag Renaming {#tag-rename}

> Note: For [DataKit global tag key](datakit-conf#update-global-tag), renaming them is not supported here.

`tags_rename` can replace the tag name of the collected Prometheus Exporter data, and `overwrite_exist_tags` is used to open the option of overwriting existing tags. For example, for existing Prometheus Exporter data:

```not-set
http_request_duration_seconds_bucket{le="0.003",status_code="404",tag_exists="yes", method="GET"} 1
```

Assume that the `tags_rename` configuration here is as follows:

```toml
[inputs.prom.tags_rename]
  overwrite_exist_tags = true
  [inputs.prom.tags_rename.mapping]
    status_code = "StatusCode",
    method      = "tag_exists", // 将 `method` 这个 tag 重命名为一个已存在的 tag
```

Then the final line protocol data will become (ignoring the timestamp):

```shell
# Note that tag_exists is affected here, and its value is the value of the original method
http,StatusCode=404,le=0.003,tag_exists=GET request_duration_seconds_bucket=1
```

If `overwrite_exist_tags` is disabled, the final data is:

```shell
# Neither tag_exists nor method has changed
http,StatusCode=404,le=0.003,method=GET,tag_exists=yes request_duration_seconds_bucket=1
```

Note that the tag name here is case-sensitive, and you can test the data with the following debugging tool to determine how to replace the tag name.

## Protocol Conversion Description {#proto-transfer}

Because the data format of Prometheus is different from the line protocol format of Influxdb. For Prometheus, the following is a piece of data exposed in a K8s cluster:

```not-set
node_filesystem_avail_bytes{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 1.21585664e+08
node_filesystem_avail_bytes{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 1.2623872e+08
node_filesystem_avail_bytes{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 3.7269504e+07
node_filesystem_avail_bytes{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 1.2808192e+07
node_filesystem_avail_bytes{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_avail_bytes{device="map auto_home",fstype="autofs",mountpoint="/home"} 0
# HELP node_filesystem_device_error Whether an error occurred while getting statistics for the given device.
# TYPE node_filesystem_device_error gauge
node_filesystem_device_error{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 0
node_filesystem_device_error{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 0
node_filesystem_device_error{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 0
node_filesystem_device_error{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 0
node_filesystem_device_error{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_device_error{device="map auto_home",fstype="autofs",mountpoint="/home"} 0
# HELP node_filesystem_files Filesystem total file nodes.
# TYPE node_filesystem_files gauge
node_filesystem_files{device="/dev/disk1s1",fstype="apfs",mountpoint="/"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk1s4",fstype="apfs",mountpoint="/private/var/vm"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk3s1",fstype="apfs",mountpoint="/Volumes/PostgreSQL 13.2-2"} 9.223372036854776e+18
node_filesystem_files{device="/dev/disk5s1",fstype="apfs",mountpoint="/Volumes/Git 2.15.0 Mavericks Intel Universal"} 9.223372036854776e+18
node_filesystem_files{device="map -hosts",fstype="autofs",mountpoint="/net"} 0
node_filesystem_files{device="map auto_home",fstype="autof
```

For Influxdb, one way to organize the above data is

```not-set
node_filesystem,tag-list available_bytes=1.21585664e+08,device_error=0,files=9.223372036854776e+18 time
```

Its organizational basis is:

- In Prometheus exposed metrics, if the name prefix is `node_filesystem`, then it is specified on the line protocol measurement `node_filesystem`.
- Place the original Prometheus metrics with their prefixes cut off into the metrics of the measurement `node_filesystem`.
- By default, all tags in Prometheus (that is, parts in `{}` )remain in the row protocol of Influxdb

To achieve this cutting purpose, you can configure `prom.conf` as follows

```not-set
  [[inputs.prom.measurements]]
    prefix = "node_filesystem_"
    name = "node_filesystem"
```

## Command Line Debug Measurement {#debug}

Because Prometheus exposes a lot of metrics, you don't necessarily need all of them, so DataKit provides a simple tool to debug `prom.conf` . If you constantly adjust the configuration of `prom.conf`, you can achieve the following purposes:

- Only Prometheus metrics that meet certain name rules are collected
- Collect only partial measurement data (`metric_types`), such as `gauge` type indicators and `counter` type metrics

DataKit supports debugging the configuration file of prom collector directly from the command line, copying a prom.conf template from conf.d/prom, filling in the corresponding Exporter address, and debugging this `prom.conf` through DataKit:

Debug `prom.conf` by executing the following command

```shell
datakit debug --prom-conf prom.conf
```

Parameter description:

- `prom-conf`: Specifies the configuration file. By default, it looks for the `prom.conf` file in the current directory. If it is not found, it will look for the corresponding file in the *<datakit-install-dir\>/conf.d/prom* directory.

Output sample:

```not-set
================= Line Protocol Points ==================

 prom_node,device=disk0 disk_written_sectors_total=146531.087890625 1623379432917573000
 prom_node,device=disk2 disk_written_sectors_total=0 1623379432917573000
 prom_node,device=disk4 disk_written_sectors_total=0 1623379432917573000
 prom_node memory_total_bytes=8589934592 1623379432917573000
 prom_node,device=XHC20 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=awdl0 network_transmit_bytes_total=1527808 1623379432917573000
 prom_node,device=bridge0 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=en0 network_transmit_bytes_total=2847181824 1623379432917573000
 prom_node,device=en1 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=en2 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=gif0 network_transmit_bytes_total=0 1623379432917573000
 prom_node,device=lo0 network_transmit_bytes_total=6818923520 1623379432917573000
 prom_node,device=p2p0 network_transmit_bytes_total=0 1623379432917573000
 ....
================= Summary ==================

Total time series: 58
Total line protocol points: 261
Total measurements: 3 (prom_node, prom_go, prom_promhttp)
```

Output description:

- Line Protocol Points: Generated line protocol points
- Summary: Summary results
    - Total time series: Number of timelines
    - Total line protocol points: Line protocol points
    - Total measurements: Number of measurements and their names.
