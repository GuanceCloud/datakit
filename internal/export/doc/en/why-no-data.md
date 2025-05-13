# How to Troubleshoot No Data Issues

---

After deploying DataKit, you would typically check the collected data directly from the monitoring cloud page. If everything is normal, the data will appear on the page quickly (the most direct is the host/process data in "Infrastructure"). However, for various reasons, there may be issues in the data collection, processing, or transmission process, leading to no data issues.

The following analyzes the possible causes of no data from several aspects:

- [Network Related](why-no-data.md#iss-network)
- [Host Related](why-no-data.md#iss-host)
- [Startup Issues](why-no-data.md#iss-start-fail)
- [Collector Configuration Related](why-no-data.md#iss-input-config)
- [Global Configuration Related](why-no-data.md#iss-global-settings)
- [Others](why-no-data.md#iss-others)

## Host Related {#iss-host}

Host-related issues are generally more subtle and often overlooked, making them difficult to troubleshoot (the lower-level, the harder to check), so here is a general list:

### Timestamp Anomaly {#iss-timestamp}

On Linux/Mac, you can view the current system time by entering `date`:

```shell
$ date
Wed Jul 21 16:22:32 CST 2021
```

In some cases, it may display like this:

```shell
$ date
Wed Jul 21 08:22:32 UTC 2021
```

This is because the former is China's Eastern Eight Zone time, and the latter is Greenwich Mean Time, which differs by 8 hours. However, in fact, the timestamps of these two times are the same.

If the current system time is significantly different from your mobile phone time, especially if it is ahead, then the monitoring cloud will not display this 'future' data.

Additionally, if the time is behind, the default viewer of the monitoring cloud will not display this data (the default viewer usually shows data from the last 15 minutes), but you can adjust the viewing time range in the viewer.

### Host Hardware/Software Not Supported {#iss-os-arch}

Some collectors are not supported on specific platforms, even if the configuration is turned on, there will be no data collection:

- There is no CPU collector on macOS
- Collectors such as Oracle/DB2/OceanBase/eBPF can only run on Linux
- Some Windows-specific collectors cannot run on non-Windows platforms
- The DataKit-Lite distribution only compiles a small number of collectors, and most collectors are not included in its binaries

### Network Related {#iss-network}

Network-related issues are relatively straightforward and common, and can be investigated using other methods (such as `ping/nc/curl` commands).

#### Cannot Connect to the Data Source {#iss-input-connect}

Since DataKit is installed on specific machines/Nodes, when it collects certain data, it may be unable to access the target due to network reasons (such as MySQL/Redis, etc.). In this case, you can find the problem through the collector debug:

```shell
$ datakit debug --input-conf conf.d/db/redis.conf
loading /usr/local/datakit/conf.d/db/redis.conf.sample with 1 inputs...
running input "redis"(0th)...
[E] get error from input = redis, source = redis: dial tcp 192.168.3.100:6379: connect: connection refused | Ctrl+c to exit.
```

#### Cannot Connect to Dataway {#iss-dw-connect}

If the host where DataKit is located cannot connect to Dataway, you can test the 404 page of Dataway:

```shell
$ curl -I http[s]://your-dataway-addr:port
HTTP/2 404
date: Mon, 27 Nov 2023 06:03:47 GMT
content-type: text/html
content-length: 792
access-control-allow-credentials: true
access-control-allow-headers: Content-Type, Content-Length
access-control-allow-methods: POST, OPTIONS, GET, PUT
access-control-allow-origin: *
```

If the status code 404 is displayed, it means the connection with Dataway is normal.

For SAAS, the Dataway address is `https://openway.<<<custom_key.brand_main_domain>>>`.

If you get the following result, it indicates there is a network problem:

```shell
curl: (6) Could not resolve host: openway.<<<custom_key.brand_main_domain>>>
```

If you find an error log like the following in the DataKit log (*/var/log/datakit/log*), it means there is a problem with the current environment's connection to Dataway, which may be restricted by the firewall:

```shell
request url https://openway.<<<custom_key.brand_main_domain>>>/v1/write/xxx/token=tkn_xxx  failed:  ... context deadline exceeded...
```

### Startup Issues {#iss-start-fail}

Since DataKit adapts to mainstream OS/Arch types, it is possible to encounter deployment issues on some OS distribution versions, and the service may be in an abnormal state after installation, preventing DataKit from starting.

### *datakit.conf* is incorrect {#iss-timestamp}

*datakit.conf* is the main configuration entry for DataKit. If it is incorrectly configured (TOML syntax error), it will prevent DataKit from starting, and the DataKit log will contain a similar log (different syntax error messages will vary):

```shell
# Manually start the datakit program
$ /usr/local/datakit
2023-11-27T14:17:15.578+0800    INFO    main    datakit/main.go:166     load config from /user/local/datakit/conf.d/datakit.conf...
2023-11-27T14:15:56.519+0800    ERROR   main    datakit/main.go:169     load config failed: bstoml.Decode: toml: line 19 (last key "default_enabled_inputs"): expected value but found "ulimit" instead
```

### Service Exception {#iss-service-fail}

Due to certain reasons (such as DataKit service startup timeout), the `datakit` service may be in an invalid state, which requires [some operations to reset the DataKit system service](datakit-service-how-to.md#when-service-failed).

### Keep Restarting or Not Starting {#iss-restart-notstart}

In Kubernetes, insufficient resources (memory) may cause DataKit OOM, leaving no time to perform specific data collection. You can check whether the memory resources allocated in *datakit.yaml* are appropriate:

```yaml
  containers:
  - name: datakit
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit:datakit:<VERSION>
    resources:
      requests:
        memory: "128Mi"
      limits:
        memory: "4Gi"
    command: ["stress"]
    args: ["--vm", "1", "--vm-bytes", "150M", "--vm-hang", "1"]
```

Here, the system requires a minimum of 128MB of memory ( `requests` ) to start DataKit. If the DataKit's own collection tasks are heavy, the default 4GB may not be enough, and you need to adjust the `limits` parameter.

### Port Occupied {#iss-port-in-use}

Some DataKit collectors need to open specific ports locally to receive data from the outside. If these ports are occupied, the corresponding collector will not be able to start, and the DataKit log will show information similar to the port being occupied.

The mainstream collectors affected by the port include:

- HTTP's 9529 port: some collectors (such as eBPF/Oracle/LogStream, etc.) push data to DataKit's HTTP interface
- StatsD 8125 port: used to receive StatsD's metric data (such as JVM-related metrics)
- OpenTelemetry 4317 port: used to receive OpenTelemetry metrics and Trace data

For more port occupancy, see [this list](datakit-port.md).

### Insufficient Disk Space {#iss-disk-space}

Insufficient disk space can lead to undefined behavior (DataKit's own logs cannot be written/diskcache cannot be written, etc.).

### Resource Occupation Exceeds Default Settings {#iss-ulimit}

After DataKit is installed, the number of files it opens is 64K by default (Linux). If a collector (such as the file log collector) opens too many files, subsequent files cannot be opened, affecting the collection.

In addition, opening too many files proves that the current collection has a serious congestion, which may consume too much memory resources, leading to OOM.

## Collector Configuration Related {#iss-input-config}

Collector configuration-related issues are generally straightforward, mainly including the following possibilities:

### No Data Generated by the Data Source {#iss-input-nodata}

Take MySQL as an example, some slow query or lock-related collections will only have data when corresponding problems occur, otherwise, there will be no data in the corresponding view.

In addition, some exposed Prometheus metric data collection objects may have the Prometheus metric collection turned off by default (or only localhost can access), which requires corresponding configuration on the data collection object for DataKit to collect this data. This kind of problem with  no data generation can be verified through the above collector debug function ( `datakit debug --input-conf ...` ).

For log collection, if the corresponding log file does not have new (relative to after DataKit starts) log data, even if there is already log data in the current log file, there will be no data collection.

For Profiling data collection, the service/application being collected also needs to turn on the corresponding function to have Profiling data pushed to DataKit.

### Access Permissions {#iss-permission-deny}

Many middleware collections require user authentication configuration, some of which need to be set on the data collection object. If the corresponding username/password configuration is incorrect, DataKit collection will report an error.

In addition, since DataKit uses toml configuration, some password strings require additional escaping (usually URL-Encode), for example, if the password contains the `@` character, it needs to be converted to `%40`.

> DataKit is gradually optimizing the existing password string (connection string) configuration method to reduce this kind of escaping.

### Version Issues {#iss-version-na}

The software version of some user environments may be too old or too new, not in the DataKit support list, which may cause collection problems.

It may also be possible to collect even if it is not in the DataKit support list, and we cannot test all version numbers. When there are incompatible/unsupported versions, feedback is needed.

### Collector Bug {#iss-datakit-bug}

You can directly go to [Bug Report](why-no-data.md#bug-report).

### Collector Configuration Not Enabled {#iss-no-input}

Since DataKit only recognizes the `.conf` configuration files in the *conf.d* directory, some collector configurations may be misplaced or have the wrong extension name, causing DataKit to skip its configuration. Correct the corresponding file location or file name.

### Collector Disabled {#iss-input-disabled}

In the main configuration of DataKit, some collectors can be disabled, and even if the collector is correctly configured in *conf.d*, DataKit will ignore this type of collector:

```toml
default_enabled_inputs = [
  "-disk",
  "-mme",
]
```

These collectors with a `-` in front are disabled. Remove the `-` in front or remove the item.

### Collector Configuration Error {#iss-invalid-conf}

DataKit collectors use TOML format configuration files. When the configuration file does not conform to TOML specifications or does not conform to the field definition types of the program (such as configuring integers as strings, etc.), there will be a configuration file loading failure, which will lead to the collector not being enabled.

DataKit has a built-in configuration check function, see [here](datakit-tools-how-to.md#check-conf).

### Configuration Method Error {#iss-config-mistaken}

There are two major types of collector configurations in DataKit:

- Host installation: Directly add the corresponding collector configuration in the *conf.d* directory
- Kubernetes installation:
    - You can directly mount the collector configuration through ConfigMap
    - You can also modify it through environment variables (if both ConfigMap and environment variables exist, the configuration in the environment variables will prevail)
    - You can also mark the collection configuration through Annotations (relative to environment variables and ConfigMap, its priority is the highest)
    - If the default collector list ( `ENV_DEFAULT_ENABLED_INPUTS` ) specifies a certain collector, and the same name and same configuration collector is added in the ConfigMap, its behavior is undefined, which may trigger the following single-instance collector issue

### Single-instance Collector {#iss-singleton}

A single-instance collector can only be enabled in one DataKit. If multiple collectors are enabled, DataKit loads the first one in file name order (if in the same `.conf` file, only the first one is loaded), and the others are not loaded. This may lead to the latter collectors not being enabled.

## Global Configuration Related {#iss-global-settings}

Some global configurations of DataKit segments also affect the collected data, in addition to the disabled collectors mentioned above, there are also the following aspects.

### Blacklist/Pipeline Impact {#iss-pipeline-filter}

Users may have configured a blacklist on the monitoring cloud page, which is used to discard data that meets certain characteristics and not upload it.

Pipeline itself also has the operation of discarding data ( `drop()` ).

Both of these discard behaviors can be seen in the output of `datakit monitor -V`.

In addition to discarding data, Pipeline may also modify data, which may affect the front-end query, such as cutting the time field and causing a large deviation in time.

### Disk Cache {#iss-diskcache}

DataKit has set up a disk cache mechanism for some complex data processing, which temporarily caches them to disk for peak shaving, and they will be reported later. By viewing the [disk cache-related metrics](datakit-metrics.md#metrics), you can learn about the corresponding data cache situation.

### Sinker Dataway {#iss-sinker-dataway}

If [Sinker Dataway](../deployment/dataway-sink.md) is enabled, some data will be discarded due to not matching the rules according to the existing Sinker rule configuration.

### IO Busy {#iss-io-busy}

Due to the bandwidth limit between DataKit and Dataway, the data upload is relatively slow, which affects the data collection (not in time for consumption). In this case, DataKit will discard the metrics data that are too late to process, and non-metric data will block the collection, resulting in no data being displayed on the monitoring cloud page.

### Dataway Cache {#iss-dataway-cache}

If there is a network failure between Dataway and the monitoring cloud center, Dataway will cache the data pushed by DataKit, and this part of the data may be delayed or ultimately discarded (the data exceeds the disk cache limit).

### Account Issues {#iss-workspace}

If the user's monitoring cloud account is overdue/exceeds the data usage, it will cause DataKit data reporting to have 4xx issues. This kind of problem can be directly seen in `datakit monitor`.

## Others {#iss-others}

`datakit monitor -V` will output a lot of status information, and due to resolution issues, some data will not be directly displayed, and you need to scroll in the corresponding table to see it.

However, some terminals do not support the current monitor's drag and drop operation, which is easily mistaken for no collection. You can view the monitor by specifying a specific module (each table header's red letter represents the module):

```shell
# Check the status of the HTTP API
$ datakit monitor -M H

# Check the collector configuration and collection status
$ datakit monitor -M I

# Check basic information and runtime resource usage
$ datakit monitor -M B,R
```

The above is some basic troubleshooting ideas for no data issues. The following introduces some of the functions and methods used during these troubleshooting processes in DataKit itself.

### Collect DataKit Runtime Information {#bug-report}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)

After various troubleshooting, it may still be impossible to find the problem. At this time, we need to collect various information of DataKit (such as logs, configuration files, profiles, and its own metric data). To simplify this process, DataKit provides a command that can obtain all related information at one time and package it into a file. The usage method is as follows:

```shell
$ datakit debug --bug-report
...
```

After successful execution, a zip file is generated in the current directory, with a naming format of `info-<timestamp in milliseconds>.zip`.

<!-- markdownlint-disable MD046 -->
???+ tip

    - Please make sure to collect bug report information during the DataKit operation, preferably when the problem occurs (such as high memory/CPU usage). With the help of DataKit's own metrics and profile data, we can locate some difficult problems more quickly.

    - By default, the command will collect profile data, which may have a certain performance impact on DataKit. You can disable the collection of profile by the following command ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0)):

    ```shell
    $ datakit debug --bug-report --disable-profile
    ```

    - If there is public network access, you can directly upload the file to OSS to avoid the hassle of file copying ([:octicons-tag-24: Version-1.27.0](changelog.md#cl-1.27.0)):

    ```shell hl_lines="7"
    # Here *must fill in* the correct OSS address/Bucket name and corresponding AS/SK
    $ datakit debug --bug-report --oss OSS_HOST:OSS_BUCKET:OSS_ACCESS_KEY:OSS_SECRET_KEY
    ...
    bug report saved to info-1711794736881.zip
    uploading info-1711794736881.zip...
    download URL(size: 1.394224 M):
        https://OSS_BUCKET.OSS_HOST/datakit-bugreport/2024-03-30/dkbr_co3v2375mqs8u82aa6sg.zip
    ```

    Paste the link address at the bottom to us (please make sure that the  file in OSS is publicly accessible, otherwise the link cannot be downloaded directly).

    - By default, the bug report will collect 3 times of DataKit's own metrics, and you can adjust the number of times here through `--nmetrics` ([:octicons-tag-24: Version-1.27.0](changelog.md#cl-1.27.0)):

    ```shell
    $ datakit debug --bug-report --nmetrics 10
    ```
<!-- markdownlint-enable -->

The list of files after unzipping is as follows:

```not-set
├── basic
│   └── info
├── config
│   ├── container
│   │   └── container.conf.copy
│   ├── datakit.conf.copy
│   ├── db
│   │   ├── kafka.conf.copy
│   │   ├── mysql.conf.copy
│   │   └── sqlserver.conf.copy
│   ├── host
│   │   ├── cpu.conf.copy
│   │   ├── disk.conf.copy
│   │   └── system.conf.copy
│   ├── network
│   │   └── dialtesting.conf.copy
│   ├── profile
│   │   └── profile.conf.copy
│   ├── pythond
│   │   └── pythond.conf.copy
│   └── rum
│       └── rum.conf.copy
├── data
│   └── pull
├── externals
│   └── ebpf
│       ├── datakit-ebpf.log
│       ├── datakit-ebpf.stderr
│       ├── datakit-ebpf.offset
│       └── profile
│           ├── allocs
│           ├── block
│           ├── goroutine
│           ├── heap
│           ├── mutex
│           └── profile
├── metrics
│   ├── metric-1680513455403
│   ├── metric-1680513460410
│   └── metric-1680513465416
├── pipeline
│   ├── local_scripts
│   │   ├── elasticsearch.p.copy
│   │   ├── logging
│   │   │   ├── aaa.p.copy
│   │   │   └── refer.p.copy
│   │   └── tomcat.p.copy
│   └── remote_scripts
│       ├── pull_config.json.copy
│       ├── relation.json.copy
│       └── scripts.tar.gz.copy
├── log
│   ├── gin.log
│   └── log
├── syslog
│   └── syslog-1680513475416
└── profile
    ├── allocs
    ├── heap
    ├── goroutine
    ├── mutex
    ├── block
    └── profile
```

File description

| File name       | Is it a directory | Description                                                                                                    |
| ---:            | ---:             | ---:                                                                                                        |
| `config`        | Yes              | Configuration files, including main configuration and enabled collector configurations                              |
| `basic`         | Yes              | Operating system and environment variable information of the running environment                                     |
| `data`          | Yes              | Blacklist files in the `data` directory, i.e., `.pull` files                                                            |
| `log`           | Yes              | The latest log files, including log and gin log, temporarily not supporting `stdout`                              |
| `profile`       | Yes              | When pprof is enabled ([:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) is enabled by default), profile data will be collected |
| `metrics`       | Yes              | Data returned by the `/metrics` interface, named in the format of `metric-<timestamp in milliseconds>`              |
| `syslog`        | Yes              | Only supports `linux`, based on `journalctl` to obtain related logs                                                     |
| `error.log`     | No               | Records error information that occurs during the command output process                                            |

### Sensitive Information Processing {#sensitive}

When collecting information, sensitive information (such as tokens, passwords, etc.) will be automatically filtered and replaced, with specific rules as follows:

- Environment variables

Only environment variables starting with `ENV_` are obtained, and environment variables with names containing `password`, `token`, `key`, `key_pw`, `secret` are desensitized, replaced with `******`

- Configuration files

The content of the configuration file is processed by regular replacement, such as:

- `https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` is replaced with `https://openway.<<<custom_key.brand_main_domain>>>?token=******`
- `pass = "1111111"` is replaced with `pass = "******"`
- `postgres://postgres:123456@localhost/test` is replaced with `postgres://postgres:******@localhost/test`

After the above processing, most sensitive information can be removed. Despite this, there may still be sensitive information in the exported files, which can be manually removed. Please make sure to confirm.

### Debug Collector Configuration {#check-input-conf}

[:octicons-tag-24: Version-1.9.0](changelog.md#cl-1.9.0)

We can debug whether the collector can normally collect data through the command line, such as debugging the disk collector:

``` shell
$ datakit debug --input-conf /usr/local/datakit/conf.d/host/disk.conf
loading /usr/local/datakit/conf.d/host/disk.conf with 1 inputs...
running input "disk"(0th)...
disk,device=/dev/disk3s1s1,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631702195i,inodes_total_mb=1631i,inodes_used=349475i,inodes_used_mb=0i,inodes_used_percent=0.02141781760611041,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064064000
disk,device=/dev/disk3s6,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631352732i,inodes_total_mb=1631i,inodes_used=12i,inodes_used_mb=0i,inodes_used_percent=0.0000007355858585707753,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064243000
disk,device=/dev/disk3s2,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631353840i,inodes_total_mb=1631i,inodes_used=1120i,inodes_used_mb=0i,inodes_used_percent=0.00006865463350366712,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064254000
disk,device=/dev/disk3s4,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631352837i,inodes_total_mb=1631i,inodes_used=117i,inodes_used_mb=0i,inodes_used_percent=0.000007171961659450622,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064260000
disk,device=/dev/disk1s2,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921841i,inodes_total_mb=4i,inodes_used=1i,inodes_used_mb=0i,inodes_used_percent=0.00002031760067015574,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064266000
disk,device=/dev/disk1s1,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921873i,inodes_total_mb=4i,inodes_used=33i,inodes_used_mb=0i,inodes_used_percent=0.000670476462923769,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064271000
disk,device=/dev/disk1s3,fstype=apfs free=503996416i,inodes_free=4921840i,inodes_free_mb=4i,inodes_total=4921892i,inodes_total_mb=4i,inodes_used=52i,inodes_used_mb=0i,inodes_used_percent=0.0010565042873756677,total=524288000i,used=20291584i,used_percent=3.8703125 1685509141064276000
disk,device=/dev/disk3s5,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1634318356i,inodes_total_mb=1634i,inodes_used=2965636i,inodes_used_mb=2i,inodes_used_percent=0.18146011694186712,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064280000
disk,device=/dev/disk2s1,fstype=apfs free=3697000448i,inodes_free=36103520i,inodes_free_mb=36i,inodes_total=36103578i,inodes_total_mb=36i,inodes_used=58i,inodes_used_mb=0i,inodes_used_percent=0.00016064889745830732,total=5368664064i,used=1671663616i,used_percent=31.137422570532436 1685509141064285000
disk,device=/dev/disk3s1,fstype=apfs free=167050518528i,inodes_free=1631352720i,inodes_free_mb=1631i,inodes_total=1631702197i,inodes_total_mb=1631i,inodes_used=349477i,inodes_used_mb=0i,inodes_used_percent=0.0214179401512444,total=494384795648i,used=327334277120i,used_percent=66.21042556354438 1685509141064289000
# 10 points("M"), 98 time series from disk, cost 1.544792ms | Ctrl+c to exit.
```

The command will start the collector and print the data collected by the collector in the terminal. The bottom will display:

- The number of points collected and their type (here "M" indicates time series data)
- The number of timelines (only for time series data)
- Collector name (here "disk")
- Collection time consumption

You can end the debug with Ctrl + c. To get the collected data as soon as possible, you can appropriately adjust the collection interval of the collector (if any).

<!-- markdownlint-disable MD046 -->
???+ tip

    - Some passively receiving data collectors (such as DDTrace/RUM) need to specify the HTTP service (`--http-listen=[IP:Port]`), and then use some HTTP client tools (such as `curl`) to send data to the corresponding address of DataKit. See `datakit help debug` for details

    - The collector configuration used for debugging can be of any extension, and does not necessarily [end with `.conf`](datakit-input-conf.md#intro). We can use file names such as *my-input.conf.test* specifically for debugging, while not affecting the normal operation of DataKit.
<!-- markdownlint-enable -->

### View Monitor Page {#monitor}

See [here](datakit-monitor.md)

### Check if Data is Generated through DQL {#dql}

This feature is supported on Windows/Linux/Mac, and it needs to be executed in Powershell on Windows.

> This feature is supported in DataKit [1.1.7-rc7](changelog.md#cl-1.1.7-rc7)

```shell
$ datakit dql
> Here you can enter DQL query statements...
```

For troubleshooting no data, it is recommended to check the corresponding metric set according to the collector documentation. Taking the MySQL collector as an example, the current documentation has the following metric sets:

- `mysql`
- `mysql_schema`
- `mysql_innodb`
- `mysql_table_schema`
- `mysql_user_status`

If the MySQL collector has no data, you can check if there is data in the `mysql` metric set:

``` python
#
# Check the latest mysql metric of the specified host (here is tan-air.local) from the mysql collector
#
M::mysql {host='tan-air.local'} order by time desc limit 1
```

Check if a certain host object has been reported, where `tan-air.local` is the expected host name:

```python
O::HOST {host='tan-air.local'}
```

Check the existing APM (tracing) data categories:

```python
show_tracing_service()
```

And so on, if the data is indeed reported, then it can be found through DQL, and as for the front end not displaying, it may be blocked by other filter conditions. Through DQL, whether it is data collected by DataKit or other means (such as Function), you can view the original data at zero distance, which is especially convenient for Debugging.

### Check if There are Exceptions Logs {#check-log}

Get the latest 10 ERROR, WARN level logs through Shell/Powershell

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
    $ cat /var/log/datakit/log | grep "WARN\|ERROR" | tail -n 10
    ...
    ```

=== "Windows"

    ```powershell
    PS > Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
    ...
    ```
<!-- markdownlint-enable -->

- If you find descriptions such as `Beyond...` in the logs, it is generally because the data volume exceeds the free quota
- If there are some `ERROR/WARN` words, it generally indicates that DataKit has encountered some problems

#### Check the Running Logs of a Single Collector {#check-input-log}

If no exceptions are found, you can directly view the running logs of a single collector:

```shell
# shell
tail -f /var/log/datakit/log | grep "<Collector Name>" | grep "WARN\|ERROR"

# powershell
Get-Content -Path "C:\Program Files\datakit\log" -Wait | Select-String "<Collector Name>" | Select-String "ERROR", "WARN"
```

You can also remove the `ERROR/WARN` filter to directly view the logs of the corresponding collector. If the logs are not enough, you can turn on the debug logs in `datakit.conf` to view more logs:

```toml
# DataKit >= 1.1.8-rc0
[logging]
    ...
    level = "debug" # Change the default of info to debug
    ...

# DataKit < 1.1.8-rc0
log_level = "debug"
```

#### Check gin.log {#check-gin-log}

For collectors that remotely send data to DataKit, you can check gin.log to see if there is any remote data being sent:

```shell
tail -f /var/log/datakit/gin.log
```

### Troubleshooting Guide {#how-to-trouble-shoot}

To facilitate everyone's problem troubleshooting, the following diagram lists some basic troubleshooting ideas, and you can follow its guidance to troubleshoot potential problems:

``` mermaid
graph TD
  %% node definitions
  no_data[No data];
  debug_fail{Debug failed};
  monitor[Check <a href='https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-monitor/'>monitor</a>];
  debug_input[<a href='https://docs.<<<custom_key.brand_main_domain>>>/datakit/why-no-data/#check-input-conf'>Debug input.conf</a>];
  read_faq[Read FAQ];
  dql[DQL query];
  beyond_usage[Data usage exceed];
  pay[Pay];
  filtered[Dropped by blacklist?];
  sinked[Data been Sinked?];
  check_time[Check machine time];
  check_token[Check workspace token];
  check_version[Check DataKit version];
  dk_service_ok[<a href='https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-service-how-to/'>DataKit service ok?</a>];
  check_changelog[<a href='https://docs.<<<custom_key.brand_main_domain>>>/datakit/changelog'>Check changelog if bug fixed</a>];
  is_input_ok[Collectors running ok?];
  is_input_enabled[Is collector enabled?];
  enable_input[Enable collector];
  dataway_upload_ok[Is upload ok?];
  iss[Create a new issue];

  no_data --> dk_service_ok --> check_time --> check_token --> check_version --> check_changelog;

  no_data --> monitor;
  no_data --> debug_input --> debug_fail;
  debug_input --> read_faq;
  no_data --> read_faq --> debug_fail;
  dql --> debug_fail;

  monitor --> beyond_usage -->|No| debug_fail;
  beyond_usage -->|Yes| pay;

  monitor --> is_input_enabled;

  is_input_enabled -->|Yes| is_input_ok;
  is_input_enabled -->|No| enable_input --> debug_input;

  monitor --> is_input_ok -->|No| debug_input;

  is_input_ok -->|Yes| dataway_upload_ok -->|Yes| dql;
  is_input_ok --> filtered --> sinked;

  trouble_shooting[<a href='https://docs.<<<custom_key.brand_main_domain>>>/datakit/why-no-data/#bug-report'>Bug report</a>];

  debug_fail --> trouble_shooting;
  trouble_shooting --> iss;
```
