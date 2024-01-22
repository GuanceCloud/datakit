# 各种其它工具使用
---

DataKit 内置很多不同的小工具，便于大家日常使用。可通过如下命令来查看 DataKit 的命令行帮助：

```shell
datakit help
```

> 注意：因不同平台的差异，具体帮助内容会有差别。

如果要查看具体某个命令如何使用（比如 `dql`），可以用如下命令：

```shell
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

## 数据录制和回放 {#record-and-replay}

[:octicons-tag-24: Version-1.19.0](changelog.md#cl-1.19.0)

数据导入主要用于录入已有的采集数据，再做演示或测试的时候，可以不用额外采集。

### 开启数据录制 {#enable-recorder}

在 *datakit.conf* 中，可开启数据录制功能。开启之后，Datakit 会将数据录制到指定的目录，以便于后续导入：

```toml
[recorder]
  enabled  = true
  path     = "/path/to/recorder"     # 绝对路径，默认在 <Datakit 安装目录>/recorder 目录下
  encoding = "v2"                    # 采用 protobuf-JSON 格式（xxx.pbjson），也可以选择 v1（xxx.lp），采用行协议形式（前者更便于阅读，且数据类型支持更全）
  duration = "10m"                   # 录制时长，从 Datakit 启动后开始计时
  inputs   = ["cpu", "mem"]          # 录制指定采集器的数据（以 monitor 中具体 feed 的名字为准），为空则表示录制所有采集器数据
  categories = ["logging", "metric"] # 录制类型，为空则表示录制所有数据类型
```

录制开始后，目录结构大致如下（此处展示的是时序数据的 `pbjson` 格式）：

```shell
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
???+ attention

    - 数据录制完成后，记得关闭该功能（`enable = false`），否则每次 Datakit 启动都会启动录制，可能会消耗大量磁盘
    - 采集器名字不完全等同于采集器配置中的名字（`[[inputs.some-name`]]`），而是 monitor *Inputs Info* 面板中第一列中显示的名字。  部分采集器的名字可能这样：`logging/some-pod-name`，此处其存放的数据目录为 */usr/local/datakit/recorder/logging/logging-some-pod-name.1705636073033197000.pbjson*，此处将采集器名字中的 `/` 替换成了 `-`
<!-- markdownlint-enable -->

### 数据回放 {#do-replay}

Datakit 录制完数据后，我们可以将该目录中的数据用 Git 或其它方式保存（**一定要确保好已有的目录结构**），然后，通过如下命令可以将这些数据导入到观测云：

```shell
$ datakit import -P /usr/local/datakit/recorder -D https://openway.guance.com?token=tkn_xxxxxxxxx

> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217783322857000.pbjson"(1 points) on metric...
+1h53m6.137855s ~ 2023-10-25 15:09:43.321559 +0800 CST
> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217793321744000.pbjson"(1 points) on metric...
+1h52m56.137881s ~ 2023-10-25 15:09:53.321533 +0800 CST
> Uploading "/usr/local/datakit/recorder/metric/cpu.1698217803322683000.pbjson"(1 points) on metric...
+1h52m46.137991s ~ 2023-10-25 15:10:03.321423 +0800 CST
...
Total upload 75 kB bytes ok
```

虽然录制的数据中带了绝对时间戳（纳秒），播放的时候，Datakit 会自动将这些数据平移到当前时间（保留各个数据点之间的相对时间间隔），让它看起来像是新采集的数据一样。

通过如下命令可以获取更多数据导入的帮助说明：

```shell
$ datakit help import

usage: datakit import [options]

Import used to play recorded history data to Guance Cloud. Available options:

  -D, --dataway strings   dataway list
      --log string        log path (default "/dev/null")
  -P, --path string       point data path (default "/usr/local/datakit/recorder")
```

<!-- markdownlint-disable MD046 -->
???+ attention

    对 RUM 数据而言，如果回放的目标工作空间没有对应的 APP ID，则数据无法写入，可以在目标工作空间新建一个应用，将 APP ID 改成和录制数据中的一致，或者替换已有的录制数据中 APP ID 为目标工作空间中对应 RUM 应用的 APP ID。
<!-- markdownlint-enable -->

## 查看 DataKit 运行情况 {#using-monitor}

monitor 用法[参见这里](datakit-monitor.md)

## 检查采集器配置是否正确 {#check-conf}

编辑完采集器的配置文件后，可能某些配置有误（如配置文件格式错误），通过如下命令可检查是否正确：

```shell
datakit check --config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

## 查看工作空间信息 {#workspace-info}

为便于大家在服务端查看工作空间信息，DataKit 提供如下命令查看：

```shell
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

## 查看 DataKit 相关事件 {#event}

DataKit 运行过程中，一些关键事件会以日志的形式进行上报，比如 DataKit 的启动、采集器的运行错误等。在命令行终端，可以通过 dql 进行查询。

```shell
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

部分字段说明

- `category`: 类别，默认为 `default`, 还可取值为 `input`， 表明是与采集器 (`input`) 相关
- `status`: 事件等级，可取值为 `info`, `warning`, `error`

## DataKit 更新 IP 数据库文件 {#install-ipdb}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    - 可直接使用如下命令安装/更新 IP 地理信息库（此处可选择另一个 IP 地址库 `geolite2`，只需把 `iploc` 换成 `geolite2` 即可）：
    
    ```shell
    datakit install --ipdb iploc
    ```
    
    - 更新完 IP 地理信息库后，修改 *datakit.conf* 配置：
    
    ``` toml
    [pipeline]
      ipdb_type = "iploc"
    ```
    
    - 重启 DataKit 生效

    - 测试 IP 库是否生效

    ```shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province: 
       country: 
    ```

=== "Kubernetes(yaml)"

    - 修改 *datakit.yaml*，打开 4 处带 `---iploc-start ` 和 `---iploc-end` 中间注释。
    
    - 重新安装 DataKit：
    
    ```shell
    kubectl apply -f datakit.yaml
    
    # 确保 DataKit 容器启动
    kubectl get pod -n datakit
    ```

    - 进入容器，测试 IP 库是否生效

    ```shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```

=== "Kubernetes(Helm)"

    - helm 部署添加 `--set iploc.enable`
    
    ```shell
    helm install datakit datakit/datakit -n datakit \
        --set datakit.dataway_url="https://openway.guance.com?token=<YOUR-TOKEN>" \
        --set iploc.enable true \
        --create-namespace 
    ```
    
    关于 helm 的部署事项，参见[这里](datakit-daemonset-deploy.md/#__tabbed_1_2)。
    
    - 进入容器，测试 IP 库是否生效

    ```shell
    datakit tool --ipinfo 1.2.3.4
            ip: 1.2.3.4
          city: Brisbane
      province: Queensland
       country: AU
           isp: unknown
    ```

    如果安装失败，其输出如下：
    
    ```shell
    datakit tool --ipinfo 1.2.3.4
           isp: unknown
            ip: 1.2.3.4
          city: 
      province:
       country:
    ```
<!-- markdownlint-enable -->

## DataKit 安装第三方软件 {#extras}

### Telegraf 集成 {#telegraf}

> 注意：建议在使用 Telegraf 之前，先确 DataKit 是否能满足期望的数据采集。如果 DataKit 已经支持，不建议用 Telegraf 来采集，这可能会导致数据冲突，从而造成使用上的困扰。

安装 Telegraf 集成

```shell
datakit install --telegraf
```

启动 Telegraf

```shell
cd /etc/telegraf
cp telegraf.conf.sample telegraf.conf
telegraf --config telegraf.conf
```

关于 Telegraf 的使用事项，参见[这里](../integrations/telegraf.md)。

### Security Checker 集成 {#scheck}

安装 Security Checker

```shell
datakit install --scheck
```

安装成功后会自动运行，Security Checker 具体使用，参见[这里](../scheck/scheck-install.md)

### DataKit eBPF 集成 {#ebpf}

安装 DataKit eBPF 采集器，当前只支持 `linux/amd64 | linux/arm64` 平台，采集器使用说明见 [DataKit eBPF 采集器](../integrations/ebpf.md)

```shell
datakit install --ebpf
```

如若提示 `open /usr/local/datakit/externals/datakit-ebpf: text file busy`，停止 DataKit 服务后再执行该命令

<!-- markdownlint-disable MD046 -->
???+ warning

    该命令在 [:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6-brk) 已经被移除。新版本默认就内置了 eBPF 集成。
<!-- markdownlint-enable -->

## 查看云属性数据 {#cloudinfo}

如果安装 DataKit 所在的机器是一台云服务器（目前支持 `aliyun/tencent/aws/hwcloud/azure` 这几种），可通过如下命令查看部分云属性数据，如（标记为 `-` 表示该字段无效）：

```shell
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

## 解析行协议数据 {#parse-lp}

[:octicons-tag-24: Version-1.5.6](changelog.md#cl-1.5.6)

通过如下命令可解析行协议数据：

```shell
datakit tool --parse-lp /path/to/file
Parse 201 points OK, with 2 measurements and 201 time series
```

可以以 JSON 形式输出：

```shell
datakit tool --parse-lp /path/to/file --json
{
  "measurements": {  # 指标集列表
    "testing": {
      "points": 7,
      "time_series": 6
    },
    "testing_module": {
      "points": 195,
      "time_series": 195
    }
  },
  "point": 202,        # 总点数
  "time_serial": 201   # 总时间线数
}
```

## DataKit 自动命令补全 {#completion}

> DataKit 1.2.12 才支持该补全，且只测试了 Ubuntu 和 CentOS 两个 Linux 发行版。其它 Windows 跟 Mac 均不支持。

在使用 DataKit 命令行的过程中，因为命令行参数很多，此处我们添加了命令提示和补全功能。

主流的 Linux 基本都有命令补全支持，以 Ubuntu 和 CentOS 为例，如果要使用命令补全功能，可额外安装如下软件包：

- Ubuntu：`apt install bash-completion`
- CentOS: `yum install bash-completion bash-completion-extras`

如果安装 DataKit 之前，这些软件已经安装好了，则 DataKit 安装时会自动带上命令补全功能。如果这些软件包是在 DataKit 安装之后才更新的，可执行如下操作来安装 DataKit 命令补全功能：

```shell
datakit tool --setup-completer-script
```

补全使用示例：

```shell
$ datakit <tab> # 输入 \tab 即可提示如下命令
dql       help      install   monitor   pipeline  run       service   tool

$ datakit dql <tab> # 输入 \tab 即可提示如下选项
--auto-json   --csv         -F,--force    --host        -J,--json     --log         -R,--run      -T,--token    -V,--verbose
```

以下提及的所有命令，均可使用这一方式来操作。

### 获取自动补全脚本 {#get-completion}

如果大家的 Linux 系统不是 Ubuntu 和 CentOS，可通过如下命令获取补全脚本，然后再按照对应平台的 shell 补全方式，一一添加即可。

```shell
# 导出补全脚本到本地 datakit-completer.sh 文件中
datakit tool --completer-script > datakit-completer.sh
```

## DataKit 调试命令 {#debugging}

### 调试黑名单 {#debug-filter}

[:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)

为了调试某条数据是否会被中心配置的黑名单过滤，我们可以用如下命令：


<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
    $ datakit debug --filter=/usr/local/datakit/data/.pull --data=/path/to/lineproto.data
    
    Dropped
    
        ddtrace,http_url=/webproxy/api/online_status,service=web_front f1=1i 1691755988000000000
    
    By 7th rule(cost 1.017708ms) from category "tracing":
    
        { service = 'web_front' and ( http_url in [ '/webproxy/api/online_status' ] )}
    ```

=== "Windows"

    ```powershell
    PS > datakit.exe debug --filter 'C:\Program Files\datakit\data\.pull' --data '\path\to\lineproto.data'
    
    Dropped
    
        ddtrace,http_url=/webproxy/api/online_status,service=web_front f1=1i 1691755988000000000
    
    By 7th rule(cost 1.017708ms) from category "tracing":
    
        { service = 'web_front' and ( http_url in [ '/webproxy/api/online_status' ] )}
    ```
<!-- markdownlint-enable -->

此处输出表明，文件 *lineproto.data* 中的这条数据，被位于 *.pull* 文件中 `tracing` 所在分类的第 7 条（从 1 开始计数）规则匹配。一旦匹配，则该条数据将被丢弃。

### 使用 glob 规则获取文件路径 {#glob-conf}
[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

在日志采集中，支持以 [glob 规则配置日志路径](../integrations/logging.md#glob-rules)。

通过使用 Datakit 调试 glob 规则。需要提供一个配置文件，该文件的每一行都是一个 glob 语句。

配置文件示例如下：

```shell
$ cat glob-config
/tmp/log-test/*.log
/tmp/log-test/**/*.log
```

完整命令示例如下：

```shell
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

### 正则表达式匹配文本 {#regex-conf}
[:octicons-tag-24: Version-1.8.0](changelog.md#cl-1.8.0)

在日志采集中，支持配置 [正则表达式实现多行日志采集](../integrations/logging.md#multiline)。

通过使用 Datakit 调试正则表达式规则。需要提供一个配置文件，该文件的**第一行是正则表达式**，剩余内容是被匹配的文本（可以是多行）。

配置文件示例如下：

```shell
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

完整命令示例如下：

```shell
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
