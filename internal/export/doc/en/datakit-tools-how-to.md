# Various Other Tool Usages

---

DataKit has many different small tools built-in for daily use. You can view the command-line help of DataKit through the following command:

``` shell
datakit help
```

> Note: Due to differences between different platforms, the specific help content may vary.

If you want to see how a specific command is used (such as `dql`), you can use the following command:

``` shell
$ datakit help dql
usage: datakit dql [options]

DQL used to query data. If no option specified, query interactively. Other available options:

      --auto-json      pretty output string if field/tag value is JSON
      --csv string     Specify the directory
  -F, --force          overwrite csv if file exists
  -H, --host string    specify datakit host to query
  -J, --json           output in json format
      --log string     log path (default "/dev/null")
  -R, --run string     run single DQL
  -T, --token string   run query for specific token(workspace)
  -V, --verbose        verbosity mode
```

<!--
## Viewing DataKit-related Events {#event}
During the operation of DataKit, some key events will be reported in the form of logs, such as the startup of DataKit and the running errors of collectors. In the command-line terminal, you can query through dql.

``` shell
datakit dql

dql > L::datakit limit 10;

-----------------[ r1.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vng'
   category 'input'
create_time 1639970679664
    date_ns 835000
       host 'demo'
    message 'elasticsearch Get "http://myweb:9200/_nodes/_local/name": dial tcp 150.158.54.252:9200: connect: connection refused'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:34 +0800 CST
-----------------[ r2.datakit.s1 ]-----------------
    __docid 'L_c6vvetpaahl15ivd7vn0'
   category 'input'
create_time 1639970679664
    date_ns 67000
       host 'demo'
    message 'postgresql pq: password authentication failed for user "postgres"'
     source 'datakit'
     status 'warning'
       time 2021-12-20 11:24:32 +0800 CST
-----------------[ r3.datakit.s1 ]-----------------
    __docid 'L_c6tish1aahlf03dqas00'
   category 'default'
create_time 1639657028706
    date_ns 246000
       host 'zhengs-MacBook-Pro.local'
    message 'datakit start ok, ready for collecting metrics.'
     source 'datakit'
     status 'info'
       time 2021-12-20 11:16:58 +0800 CST       
          
          ...       
```

Explanation of some fields:
- `category`: Category, the default is `default`, and it can also take the value `input`, indicating that it is related to the collector (`input`).
- `status`: Event level, which can take the values `info`, `warning`, `error`.
-->

## Debugging Commands {#debugging}

### Debugging the Blacklist {#debug-filter}

[:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)

To debug whether a piece of data will be filtered by the centrally configured blacklist, you can use the following command:

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ``` shell
    $ datakit debug --filter=/usr/local/datakit/data/.pull --data=/path/to/lineproto.data
    
    Dropped
    
        ddtrace,http_url=/webproxy/api/online_status,service=web_front f1=1i 1691755988000000000
    
    By 7th rule(cost 1.017708ms) from category "tracing":
    
        { service = 'web_front' and ( http_url in [ '/webproxy/api/online_status' ] )}
    ```

=== "Windows"

    ``` powershell
    PS > datakit.exe debug --filter 'C:\Program Files\datakit\data\.pull' --data '\path\to\lineproto.data'
    
    Dropped
    
        ddtrace,http_url=/webproxy/api/online_status,service=web_front f1=1i 1691755988000000000
    
    By 7th rule(cost 1.017708ms) from category "tracing":
    
        { service = 'web_front' and ( http_url in [ '/webproxy/api/online_status' ] )}
    ```
<!-- markdownlint-enable -->

The above output indicates that the data in the file *lineproto.data* is matched by the 7th rule (counting from 1) in the `tracing` category in the *.pull* file. Once matched, this piece of data will be discarded.

### Obtaining File Paths Using glob Rules {#glob-conf}

[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

In log collection, log paths can be configured using [glob rules](../integrations/logging.md#glob-rules).

You can debug the glob rules using DataKit. You need to provide a configuration file, and each line of the file is a glob statement.

An example of the configuration file is as follows:

``` shell
$ cat glob-config
/tmp/log-test/*.log
/tmp/log-test/**/*.log
```

A complete command example is as follows:

``` shell
$ datakit debug --glob-conf glob-config
============= glob paths ============
/tmp/log-test/*.log
/tmp/log-test/**/*.log

========== found the files ==========
/tmp/log-test/1.log
/tmp/log-test/logfwd.log
/tmp/log-test/123/1.log
/tmp/log-test/123/2.log
```

### Matching Text with Regular Expressions {#regex-conf}

[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

In log collection, [multiline log collection can be achieved by configuring regular expressions](../integrations/logging.md#multiline).

You can debug the regular expression rules using DataKit. You need to provide a configuration file, and the **first line of the file is the regular expression**, and the remaining content is the text to be matched (which can be multiple lines).

An example of the configuration file is as follows:

``` shell
$ cat regex-config
^\d{4}-\d{2}-\d{2}
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

A complete command example is as follows:

``` shell
$ datakit debug --regex-conf regex-config
============= regex rule ============
^\d{4}-\d{2}-\d{2}

========== matching results ==========
  Ok:  2020-10-23 06:41:56,688 INFO demo.py 1.0
  Ok:  2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Fail:  Traceback (most recent call last):
Fail:    File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
Fail:      response = self.full_dispatch_request()
Fail:  ZeroDivisionError: division by zero
  Ok:  2020-10-23 06:41:56,688 INFO demo.py 5.0
```

### Viewing the Running Status of DataKit {#using-monitor}

For the usage of monitor, please refer to [here](datakit-monitor.md).

<!-- markdownlint-disable MD013 -->
### Checking the Correctness of Collector Configuration {#check-conf}
<!-- markdownlint-enable -->

After editing the collector configuration file, there may be some configuration errors (such as incorrect configuration file format). You can check whether it is correct through the following command:

``` shell
datakit check --config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

### Viewing Workspace Information {#workspace-info}

To facilitate viewing workspace information on the server side, DataKit provides the following command to view it:

``` shell
datakit tool --workspace-info
{
  "token": {
    "ws_uuid": "wksp_2dc431d6693711eb8ff97aeee04b54af",
    "bill_state": "normal",
    "ver_type": "pay",
    "token": "tkn_2dc438b6693711eb8ff97aeee04b54af",
    "db_uuid": "ifdb_c0fss9qc8kg4gj9bjjag",
    "status": 0,
    "creator": "",
    "expire_at": -1,
    "create_at": 0,
    "update_at": 0,
    "delete_at": 0
  },
  "data_usage": {
    "data_metric": 96966,
    "data_logging": 3253,
    "data_tracing": 2868,
    "data_rum": 0,
    "is_over_usage": false
  }
}
```

### Debugging KV Files {#debug-kv}

When the collector configuration file is configured using the KV template, if you need to debug, you can use the following command for debugging.

``` shell
datakit tool --parse-kv-file conf.d/host/cpu.conf --kv-file data/.kv

[[inputs.cpu]]
  ## Collect interval, default is 10 seconds. (optional)
  interval = '10s'

  ## Collect CPU usage per core, default is false. (optional)
  percpu = false

  ## Setting disable_temperature_collect to false will collect cpu temperature stats for linux. (deprecated)
  # disable_temperature_collect = false

  ## Enable to collect core temperature data.
  enable_temperature = true

  ## Enable gets average load information every five seconds.
  enable_load5s = true

[inputs.cpu.tags]
  kv = "cpu_kv_value3"
```

### Viewing Cloud Attribute Data {#cloudinfo}

If the machine where DataKit is installed is a cloud server (currently supports `aliyun/tencent/aws/hwcloud/azure`), you can view some cloud attribute data through the following command. For example (marked as `-` means the field is invalid):

``` shell
datakit tool --show-cloud-info aws

           cloud_provider: aws
              description: -
     instance_charge_type: -
              instance_id: i-09b37dc1xxxxxxxxx
            instance_name: -
    instance_network_type: -
          instance_status: -
            instance_type: t2.nano
               private_ip: 172.31.22.123
                   region: cn-northwest-1
        security_group_id: launch-wizard-1
                  zone_id: cnnw1-az2
```

### Parsing Line Protocol Data {#parse-lp}

[:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6)

You can parse line protocol data through the following command:

``` shell
datakit tool --parse-lp /path/to/file
Parse 201 points OK, with 2 measurements and 201 time series
```

It can be output in JSON format:

``` shell
datakit tool --parse-lp /path/to/file --json
{
  "measurements": {  # List of metric sets
    "testing": {
      "points": 7,
      "time_series": 6
    },
    "testing_module": {
      "points": 195,
      "time_series": 195
    }
  },
  "point": 202,        # Total number of points
  "time_serial": 201   # Total number of timelines
}
```

### Data Recording and Replay {#record-and-replay}

[:octicons-tag-24: Version-1.19.0](changelog.md#cl-1.19.0)

Data import is mainly used to enter existing collected data. When demonstrating or testing, additional collection is not required.

#### Enabling Data Recording {#enable-recorder}

In *datakit.conf*, you can enable the data recording function. After enabling, DataKit will record the data to the specified directory for subsequent import:

``` toml
[recorder]
  enabled  = true
  path     = "/path/to/recorder"     # Absolute path, by default in the <DataKit installation directory>/recorder directory
  encoding = "v2"                    # Use protobuf-JSON format (xxx.pbjson), and you can also choose v1 (xxx.lp) in line protocol form (the former is more readable and supports more data types)
  duration = "10m"                   # Recording duration, starting from the startup of DataKit
  inputs   = ["cpu", "mem"]          # Record data of specified collectors (based on the names shown in the *Inputs Info* panel of monitor), and if empty, it means recording data of all collectors
  categories = ["logging", "metric"] # Recording types, and if empty, it means recording all data types
```

After the recording starts, the directory structure is roughly as follows (showing the `pbjson` format of time-series data here):

``` shell
[ 416] /usr/local/datakit/recorder/
├── [  64]  custom_object
├── [  64]  dynamic_dw
├── [  64]  keyevent
├── [  64]  logging
├── [  64]  network
├── [  64]  object
├── [  64]  profiling
├── [  64]  rum
├── [  64]  security
├── [  64]  tracing
└── [1.9K]  metric
    ├── [1.2K]  cpu.1698217783322857000.pbjson
    ├── [1.2K]  cpu.1698217793321744000.pbjson
    ├── [1.2K]  cpu.1698217803322683000.pbjson
    ├── [1.2K]  cpu.1698217813322834000.pbjson
    └── [1.2K]  cpu.1698218363360258000.pbjson

12 directories, 59 files
```

<!-- markdownlint-disable MD046 -->
???+ warning

    - After the data recording is completed, remember to turn off this function (`enable = false`). Otherwise, every time DataKit starts, recording will be launched, which may consume a large amount of disk space.
    - The collector name is not exactly the same as the name in the collector configuration (`[[inputs.some-name]]`), but the name shown in the first column of the *Inputs Info* panel of monitor. The name of some collectors may be like this: `logging/<some-pod-name>`. Here, the data directory it stores is */usr/local/datakit/recorder/logging/logging-some-pod-name.1705636073033197000.pbjson*, and the `/` in the collector name is replaced with `-` (to avoid an extra directory structure).
<!-- markdownlint-enable -->

#### Data Replay {#do-replay}

After DataKit records the data, you can save the data in this directory using Git or other methods (**make sure to keep the existing directory structure**). Then, you can import these data into <<<custom_key.brand_name>>> through the following command:

``` shell
$ datakit import -P /usr/local/datakit/recorder -D https://openway.<<<custom_key.brand_main_domain>>>?token=tkn_xxxxxxxxx

> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217783322857000.pbjson"(1 points) on metric...
+1h53m6.137855s ~ 2023-10-25 15:09:43.321559 +0800 CST
> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217793321744000.pbjson"(1 points) on metric...
+1h52m56.137881s ~ 2023-10-25 15:09:53.321533 +0800 CST
> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217803322683000.pbjson"(1 points) on metric...
+1h52m46.137991s ~ 2023-10-25 15:10:03.321423 +0800 CST
...
Total upload 75 kB bytes ok
```

Although the recorded data contains absolute timestamps (in nanoseconds), when playing back, DataKit will automatically shift these data to the current time (retaining the relative time intervals between data points), making it look like newly collected data.

You can obtain more help information about data import through the following command:

``` shell
$ datakit help import

usage: datakit import [options]

Import used to play recorded history data to <<<custom_key.brand_name>>>. Available options:

  -D, --dataway strings   dataway list
      --log string        log path (default "/dev/null")
  -P, --path string       point data path (default "/usr/local/datakit/recorder")
```

<!-- markdownlint-disable MD046 -->
???+ warning

    For RUM data, if there is no corresponding APP ID in the target workspace for playback, the data cannot be written. You can create a new application in the target workspace, change the APP ID to be consistent with that in the recorded data, or replace the APP ID in the existing recorded data with the APP ID of the corresponding RUM application in the target workspace.
<!-- markdownlint-enable -->

## Others {#others}

### Telegraf Integration {#telegraf}

> Note: Before using Telegraf, it is recommended to confirm whether DataKit can meet the expected data collection. If DataKit already supports it, it is not recommended to use Telegraf for collection, as it may cause data conflicts and usage troubles.

Install the Telegraf integration

``` shell
datakit install --telegraf
```

Start Telegraf

``` shell
cd /etc/telegraf
cp telegraf.conf.sample telegraf.conf
telegraf --config telegraf.conf
```

For usage matters of Telegraf, refer to [here](../integrations/telegraf.md).

### Security Checker Integration {#scheck}

Install the Security Checker

``` shell
datakit install --scheck
```

After a successful installation, it will run automatically. For the specific usage of the Security Checker, refer to [here](../scheck/scheck-install.md)

### eBPF Integration {#ebpf}

Install the DataKit eBPF collector. Currently, it only supports the `linux/amd64 | linux/arm64` platforms. For the usage instructions of the collector, see [DataKit eBPF Collector](../integrations/ebpf.md)

``` shell
datakit install --ebpf
```

If the prompt `open /usr/local/datakit/externals/datakit-ebpf: text file busy` appears, execute this command after stopping the DataKit service.

<!-- markdownlint-disable MD046 -->
???+ warning

    This command has been removed in [:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6-brk). The eBPF integration is built-in by default in the new version.
<!-- markdownlint-enable -->

### Update IP Database {#install-ipdb}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    - You can directly use the following command to install/update the IP geographic information database (here you can choose another IP address library `geolite2`, just replace `iploc` with `geolite2`):
    
    ``` shell
    datakit install --ipdb iploc
    ```
    
    - After updating the IP geographic information database, modify the *datakit.conf* configuration:
    
    ``` toml
    [pipeline]
      ipdb_type = "iploc"
    ```
    
    - Restart DataKit to take effect
    
    - Test whether the IP library takes effect
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province: 
       country: 
    ```

=== "Kubernetes(yaml)"

    - Modify *datakit.yaml* and uncomment the content between the 4 places marked with `---iploc-start` and `---iploc-end`.
    
    - Reinstall DataKit:
    
    ``` shell
    kubectl apply -f datakit.yaml
    
    # Ensure the DataKit container is started
    kubectl get pod -n datakit
    ```
    
    - Enter the container and test whether the IP library takes effect
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```

=== "Kubernetes(Helm)"

    - Add `--set iploc.enable` when deploying with Helm
    
    ``` shell
    helm install datakit datakit/datakit -n datakit \
        --set datakit.dataway_url="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-TOKEN>" \
        --set iploc.enable true \
        --create-namespace 
    ```
    
    For deployment matters of Helm, refer to [here](datakit-daemonset-deploy.md/#__tabbed_1_2).
    
    - Enter the container and test whether the IP library takes effect
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```
    
    If the installation fails, the output is as follows:
    
    ``` shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```
<!-- markdownlint-enable -->

### Automatic Command Completion {#completion}

> DataKit 1.2.12 supports this completion, and only two Linux distributions, Ubuntu and CentOS, have been tested. It is not supported on Windows and Mac.

During the use of the DataKit command line, due to the large number of command line parameters, the command prompt and completion function have been added here.

Most mainstream Linux systems support command completion. Taking Ubuntu and CentOS as examples, if you want to use the command completion function, you can additionally install the following software packages:

- Ubuntu: `apt install bash-completion`
- CentOS: `yum install bash-completion bash-completion-extras`

If these software packages are already installed before installing DataKit, the command completion function will be automatically included during the DataKit installation. If these software packages are updated after installing DataKit, you can execute the following operation to install the DataKit command completion function:

``` shell
datakit tool --setup-completer-script
```

Completion usage example:

``` shell
$ datakit <tab> # Enter \tab to get the following commands
dql       help      install   monitor   pipeline  run       service   tool

$ datakit dql <tab> # Enter \tab to get the following options
--auto-json   --csv         -F,--force    --host        -J,--json     --log         -R,--run      -T,--token    -V,--verbose
```

All the commands mentioned below can be operated in this way.

#### Obtaining the Automatic Completion Script {#get-completion}

If your Linux system is not Ubuntu or CentOS, you can obtain the completion script through the following command, and then add it one by one according to the shell completion method of the corresponding platform.

``` shell
# Export the completion script to the local datakit-completer.sh file
datakit tool --completer-script > datakit-completer.sh
```
