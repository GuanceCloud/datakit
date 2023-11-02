
# 如何排查无数据问题

---

大家在部署完数据采集之后（通过 DataKit 或 Function 采集），有时候在观测云的页面上看不到对应的数据更新，每次排查起来都心力憔悴，为了缓解这一状况，可按照如下的一些步骤，来逐步围歼「为啥没有数据」这一问题。

## 问题排查导图 {#how-to-trouble-shoot}

为便于大家排查问题，下图列举了一下基本的排查思路，大家可按照其指引来排查可能存在的问题：

``` mermaid
graph TD
  %% node definitions
  no_data[无数据];
  debug_fail{无果};
  monitor[查看 <a href='https://docs.guance.com/datakit/datakit-monitor/'>monitor</a> 情况];
  debug_input[<a href='https://docs.guance.com/datakit/why-no-data/#check-input-conf'>调试采集器配置</a>];
  read_faq[查看文档中的 FAQ];
  dql[DQL 查询];
  beyond_usage[数据是否超量];
  pay[成为付费用户];
  filtered[数据是否被黑名单丢弃];
  sinked[数据是否被 Sink];
  check_time[检查机器时间];
  check_token[检查工作空间空间 token];
  check_version[检查 Datakit 版本];
  dk_service_ok[<a href='https://docs.guance.com/datakit/datakit-service-how-to/'>Datakit 服务是否正常</a>];
  check_changelog[<a href='https://docs.guance.com/datakit/changelog'>检查 changelog 是否已修复</a>];
  is_input_ok[采集器是否运行正常];
  is_input_enabled[是否开启采集器];
  enable_input[开启采集器];
  dataway_upload_ok[上传是否正常];
  ligai[提交 <a href='https://ligai.cn/'>Ligai</a> 问题];

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

  trouble_shooting[<a href='https://docs.guance.com/datakit/why-no-data/#bug-report'>收集信息</a>];

  debug_fail --> trouble_shooting;
  trouble_shooting --> ligai;
```

## 调试采集器配置 {#check-input-conf}

[:octicons-tag-24: Version-1.9.0](changelog.md#cl-1.9.0)

我们可以通过命令行来调试采集器是否能正常采集到数据，如调试磁盘采集器：

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
# 10 points("M") from disk, cost 1.544792ms | Ctrl+c to exit.
```

该命令会启动采集器，并将采集器采集到的数据在终端打印出来。底部会显示：

- 采集的点数以及其类型（此处 `M` 表示时序数据）
- 采集器名称（此处为 `disk`）
- 采集耗时

用 Ctrl + c 可以结束调试。为了尽快得到采集的数据，可以适当调整采集器的采集间隔（如果有）。

<!-- markdownlint-disable MD046 -->
???+ tip

    - 部分被动接收数据的采集器（比如 DDTrace/RUM）需要指定 HTTP 服务（`--hppt-listen=[IP:Port]`），然后通过一些 HTTP 客户端工具（比如 `curl`）将数据发送给 Datakit 对应地址。详见 `datakit help debug` 帮助

    - 调试用的采集器配置可以是任何形式的扩展名，不一定要[以 `.conf` 作为后缀](datakit-input-conf.md#intro)，我们可以用诸如 *my-input.conf.test* 这样的文件名专用于调试，同时又不影响 Datakit 的正常运行
<!-- markdownlint-enable -->

## 检查 DataWay 连接是否正常 {#check-connection}

```shell
curl http[s]://your-dataway-addr:port
```

对 SAAS 而言，一般这样：

```shell
curl https://openway.guance.com
```

如果得到如下结果，则表示网络是有问题的：

```shell
curl: (6) Could not resolve host: openway.guance.com
```

如果发现如下这样的错误日志，则说明跟 DataWay 的连接出现了一些问题，可能是防火墙做了限制：

```shell
request url https://openway.guance.com/v1/write/xxx/token=tkn_xxx failed:  ... context deadline exceeded...
```

## 检查机器时间 {#check-local-time}

在 Linux/Mac 上，输入 `date` 即可查看当前系统时间：

```shell
date
Wed Jul 21 16:22:32 CST 2021
```

有些情况下，这里可能显示成这样：

```shell
Wed Jul 21 08:22:32 UTC 2021
```

这是因为，前者是中国东八区时间，后者是格林威治时间，两者相差 8 小时，但实际上，这两个时间的时间戳是一样的。

如果当前系统的时间跟你的手机时间相差甚远，特别是，它如果超前了，那么观测云上是看不到这些「将来」的数据的。

另外，如果时间滞后，你会看到一些老数据，不要以为发生了灵异事件，事实上，极有可能是 DataKit 所在机器的时间还停留在过去。

## 查看数据是否被黑名单过滤或 Pipeline 丢弃 {#filter-pl}

如果配置了[黑名单](datakit-filter.md)（如日志黑名单），新采集的数据可能会被黑名单过滤掉。

同理，如果 Pipeline 中对数据进行了一些[丢弃操作](../developers/pipeline/pipeline-built-in-function.md#fn-drop)，那么也可能导致中心看不到这些数据。

## 查看 Monitor 页面 {#monitor}

参见[这里](datakit-monitor.md)

## 通过 DQL 查看是否有数据产生 {#dql}

在 Windows/Linux/Mac 上，这一功能均支持，其中 Windows 需在 Powershell 中执行

> Datakit [1.1.7-rc7](changelog.md#cl-1.1.7-rc7) 才支持这一功能

```shell
datakit dql
> 这里即可输入 DQL 查询语句 ...
```

对于无数据排查，建议对照着采集器文档，看对应的指标集叫什么名字，以 MySQL 采集器为例，目前文档中有如下几个指标集：

- `mysql`
- `mysql_schema`
- `mysql_innodb`
- `mysql_table_schema`
- `mysql_user_status`

如果 MySQL 这个采集器没数据，可检查 `mysql` 这个指标集是否有数据：

``` python
#
# 查看 mysql 采集器上，指定主机上（这里是 tan-air.local）的最近一条 mysql 的指标
#
M::mysql {host='tan-air.local'} order by time desc limit 1
```

查看某个主机对象是不是上报了，这里的 `tan-air.local` 就是预期的主机名：

```python
O::HOST {host='tan-air.local'}
```

查看已有的 APM（tracing）数据分类：

```python
show_tracing_service()
```

以此类推，如果数据确实上报了，那么通过 DQL 总能找到，至于前端不显示，可能是其它过滤条件给挡掉了。通过 DQL，不管是 Datakit 采集的数据，还是其它手段（如 Function）采集的数据，都可以零距离查看原式数据，特别便于 Debug。

## 查看 Datakit 程序日志是否有异常 {#check-log}

通过 Shell/Powershell 给出最近 10 个 ERROR, WARN 级别的日志

```shell
# Shell
cat /var/log/datakit/log | grep "WARN\|ERROR" | tail -n 10

# Powershell
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
```

- 如果日志中发现诸如 `Beyond...` 这样的描述，一般情况下，是因为数据量超过了免费额度。
- 如果出现一些 `ERROR/WARN` 等字样，一般情况下，都表明 DataKit 遇到了一些问题。

### 查看单个采集器的运行日志 {#check-input-log}

如果没有发现什么异常，可直接查看单个采集器的运行日志：

```shell
# shell
tail -f /var/log/datakit/log | grep "<采集器名称>" | grep "WARN\|ERROR"

# powershell
Get-Content -Path "C:\Program Files\datakit\log" -Wait | Select-String "<采集器名称>" | Select-String "ERROR", "WARN"
```

也可以去掉 `ERROR/WARN` 等过滤，直接查看对应采集器日志。如果日志不够，可将 `datakit.conf` 中的调试日志打开，查看更多日志：

```toml
# DataKit >= 1.1.8-rc0
[logging]
    ...
    level = "debug" # 将默认的 info 改为 debug
    ...

# DataKit < 1.1.8-rc0
log_level = "debug"
```

### 查看 gin.log {#check-gin-log}

对于远程给 DataKit 打数据的采集，可查看 gin.log 来查看是否有远程数据发送过来：

```shell
tail -f /var/log/datakit/gin.log
```

## 上传 DataKit 运行日志 {#upload-log}

> Deprecated: 请使用 [Bug-Report 功能](why-no-data.md#bug-report)来代替。

排查 DataKit 问题时，通常需要检查 DataKit 运行日志，为了简化日志搜集过程，DataKit 支持一键上传日志文件：

```shell
datakit debug --upload-log
log info: path/to/tkn_xxxxx/your-hostname/datakit-log-2021-11-08-1636340937.zip # 将这个路径信息发送给我们工程师即可
```

运行命令后，会将日志目录下的所有日志文件进行打包压缩，然后上传至指定的存储。我们的工程师会根据上传日志的主机名以及 Token 传找到对应文件，进而排查 DataKit 问题。

## 收集 DataKit 运行信息 {#bug-report}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

在排查 DataKit 故障原因时，需要手动收集各种相关信息（如日志、配置文件、profile 和监控数据等），这通常比较繁琐。为了简化这个过程，DataKit 提供了一个命令，可以一次性获取所有相关信息并将其打包到一个文件中。使用方式如下：

```shell
datakit debug --bug-report
```

默认情况下，该命令会收集 profile 数据，这可能会对 DataKit 产生一定的性能影响，可以通过下面命令来禁用采集 profile ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0))：

```shell
datakit debug --bug-report --disable-profile
```

执行成功后，在当前目录下生成一个 zip 文件，命名格式为 `info-<时间戳毫秒数>.zip`。

解压后的文件列表参考如下：

```shell
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
│   └── .pull
├── metrics 
│   ├── metric-1680513455403 
│   ├── metric-1680513460410
│   └── metric-1680513465416 
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

文件说明

| 文件名称  | 是否目录 | 说明                                                                                                    |
| ---:      | ---:     | ---:                                                                                                    |
| `config`  | 是       | 配置文件，包括主配置和已开启的采集器配置                                                                |
| `basic`   | 是       | 运行环境操作系统和环境变量信息                                                                                  |
| `data`    | 是       | `data` 目录下的黑名单文件，即 `.pull` 文件                                                                                |
| `log`     | 是       | 最新的日志文件，包括 log 和 gin log，暂不支持 `stdout`                                                  |
| `profile` | 是       | pprof 开启时（[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2)已默认开启），会采集 profile 数据 |
| `metrics` | 是       | `/metrics` 接口返回的数据，命名格式为 `metric-<时间戳毫秒数>`                                           |
| `syslog`  | 是       | 仅支持 `linux`, 基于 `journalctl` 来获取相关日志                                                        |

### 敏感信息处理 {#sensitive}

信息收集时，敏感信息（如 token、密码等）会被自动过滤替换，具体规则如下：

- 环境变量

只获取以 `ENV_` 开头的环境变量，且对环境变量名称中包含 `password`, `token`, `key`, `key_pw`, `secret` 的环境变量进行脱敏处理，替换为 `******`

- 配置文件

配置文件内容进行正则替换处理，如：

``` not-set
https://openway.guance.com?token=tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` => `https://openway.guance.com?token=******
pass = "1111111"` => `pass = "******"
postgres://postgres:123456@localhost/test` => `postgres://postgres:******@localhost/test
```

经过上述处理，能够去除绝大部分敏感信息。尽管如此，如果导出的文件还存在敏感信息，可以手动将敏感信息移除，请务必确认。
