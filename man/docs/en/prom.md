
# Prometheus Exportor Data Collection
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:  · [:fontawesome-solid-flag-checkered:](index.md#legends "支持选举")

---

The Prom collector can obtain all kinds of metric data exposed by Prometheus Exporters, so long as the corresponding Exporter address is configured, the metric data can be accessed.

## Preconditions {#requirements}

Only metric data in Prometheus form can be accessed.

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/prom` directory under the DataKit installation directory, copy `prom.conf.sample` and name it `prom.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.prom]]
      # Exporter URLs
      # urls = ["http://127.0.0.1:9100/metrics", "http://127.0.0.1:9200/metrics"]
    
      # Ignoring request error to url
      ignore_req_err = false
    
      # Collector alias
      source = "prom"
    
      # Collection data output source
      # Configure this to write collected data to a local file instead of typing the data to the center
      # Then you can debug the locally saved metric set directly with the  datakit --prom-conf /path/to/this/conf
      # If url has been configured as the local file path, then --prom-conf takes precedence over debugging the data in the output path
      # output = "/abs/path/to/file"
    
      # Maximum size of data collected in bytes
      # When outputting data to a local file, you can set the upper limit of the size of the collected data
      # If the size of the collected data exceeds this limit, the collected data will be discarded
      # The maximum size of collected data is set to 32MB by default
      # max_file_size = 0
    
      # Filtering metrics type, optional values are counter, gauge, histogram, summary, untyped
      # Only counter and gauge metrics are collected by default
      # If empty, no filtering is performed
      metric_types = ["counter", "gauge"]
    
      # Metric name filter: Eligible metrics will be retained
      # Support regular, it can configure more than one, that is, satisfy one of them
      # If blank, no filtering is performed and all metrics are retained
      # metric_name_filter = ["cpu"]
    
      # Measurement name prefix
      # Configure this to prefix the measurement name
      measurement_prefix = ""
    
      # Measurement name
      # By default, the metric name will be cut with an underscore "_". The first field after cutting will be the measurement name, and the remaining fields will be the current metric name
      # If measurement_name is configured, the metric name is not cut
      # The final measurement name is prefixed with measurement_prefix
      # measurement_name = "prom"
    
      # TLS configuration
      tls_open = false
      # tls_ca = "/tmp/ca.crt"
      # tls_cert = "/tmp/peer.crt"
      # tls_key = "/tmp/peer.key"
    
      ## Set to true to turn on the election function
      election = true
    
      # Filter tags, configurable multiple tags
      # Matching tags will be ignored, but the corresponding data will still be reported
      # tags_ignore = ["xxxx"]
    
      # Custom authentication method, currently only Bearer Token is supported
      # token 和 token_file: You only need to configure one of them
      # [inputs.prom.auth]
      # type = "bearer_token"
      # token = "xxxxxxxx"
      # token_file = "/tmp/token"
      # Custom measurement name
      # You can group metrics that contain the prefix into one measurement
      # Custom measurement name configuration priority measurement_name configuration items
      #[[inputs.prom.measurements]]
      #  prefix = "cpu_"
      #  name = "cpu"
    
      # [[inputs.prom.measurements]]
      # prefix = "mem_"
      # name = "mem"
    
      # For data related to matching the following tag, discard these data and do not collect them
      [inputs.prom.ignore_tag_kv_match]
      # key1 = [ "val1.*", "val2.*"]
      # key2 = [ "val1.*", "val2.*"]
    
      # Add additional request headers to HTTP requests for data pull
      [inputs.prom.http_headers]
      # Root = "passwd"
      # Michael = "1234"
    
      # Rename tag key in prom data
      [inputs.prom.tags_rename]
        overwrite_exist_tags = false
        [inputs.prom.tags_rename.mapping]
        # tag1 = "new-name-1"
        # tag2 = "new-name-2"
        # tag3 = "new-name-3"
    
      # Call the collected metrics as logs to the center
      # When the service field is left blank, the service tag is set to the measurement name
      [inputs.prom.as_logging]
        enable = false
        service = "service_name"
    
      # Custom Tags
      [inputs.prom.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
    ```
    
    After configuration, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

???+ attention "Configuration of interval"

    Prometheus' metrics collection will cause some overhead (HTTP request) to the target service. To prevent unexpected configuration, the collection interval is currently 30s by default, and the configuration items are not obviously released in conf. If you must configure the collection interval, you can add this configuration in conf:

    ``` toml hl_lines="2"
    [[inputs.prom]]
        interval = "10s"
    ```

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

```
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

Because the data format of Prometheus is different from the line protocol format of Infuxdb. For Prometheus, the following is a piece of data exposed in a K8s cluster:

```
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

For Infuxdb, one way to organize the above data is

```
node_filesystem,tag-list available_bytes=1.21585664e+08,device_error=0,files=9.223372036854776e+18 time
```

Its organizational basis is:

- In Prometheus exposed metrics, if the name prefix is `node_filesystem`, then it is specified on the line protocol measurement `node_filesystem`.
- Place the original Prometheus metrics with their prefixes cut off into the metrics of the measurement `node_filesystem`.
- By default, all tags in Prometheus (that is, parts in `{}` remain in the row protocol of Infuxdb

To achieve this cutting purpose, you can configure `prom.conf` as follows

```
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
datakit tool --prom-conf prom.conf
```

Parameter description:

- `prom-conf`: Specifies the configuration file. By default, it looks for the `prom.conf` file in the current directory. If it is not found, it will look for the corresponding file in the `<datakit-install-dir>/conf.d/prom` directory.

Output sample:

```
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
