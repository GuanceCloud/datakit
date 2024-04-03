# 如何排查无数据问题

---

在部署完 Datakit 之后，一般会直接从观测云页面查看采集到的数据，如果一切正常，数据会很快在页面上展示（最直接是「基础设施」中的主机/进程等数据），但很多时候，因为各种原因，数据采集、处理或传输过程会出现一些问题，进而导致无数据问题。

下文从如下几个方面分析可能造成无数据的原因：

- [网络相关](why-no-data.md#iss-network)
- [主机相关](why-no-data.md#iss-host)
- [启动问题](why-no-data.md#iss-start-fail)
- [采集器配置相关](why-no-data.md#iss-input-config)
- [全局配置相关](why-no-data.md#iss-global-settings)
- [其它](why-no-data.md#iss-others)

## 网络相关 {#iss-network}

网络相关的问题比较直接，也比较常见，通过其它方式（`ping/nc/curl` 等命令）也能排查。

### 被采集对象无法连接 {#iss-input-connect}

由于 Datakit 安装在特定的机器/Node 上，当它采集某些数据时，可能网络原因，导致无法访问被采集的对象（比如 MySQL/Redis 等），此时通过采集器调试即可发现问题：

```shell
$ datakit debug --input-conf conf.d/db/redis.conf
loading /usr/local/datakit/conf.d/db/redis.conf.sample with 1 inputs...
running input "redis"(0th)...
[E] get error from input = redis, source = redis: dial tcp 192.168.3.100:6379: connect: connection refused | Ctrl+c to exit.
```

### 无法连接 Dataway {#iss-dw-connect}

如果 Datakit 所在的主机，无法连接 Dataway，可以测试一下 Dataway 的 404 页面：

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

如果显示状态码为 404，则表示与 Dataway 链接正常。

对 SAAS 而言，Dataway 地址为 `https://openway.guance.com`。

如果得到如下结果，则表示网络是有问题的：

```shell
curl: (6) Could not resolve host: openway.guance.com
```

如果在 Datakit 日志（*/var/log/datakit/log*）中发现类似如下这样的错误日志，则说明当前环境和 Dataway 的连接出现了一些问题，可能是防火墙做了限制：

```shell
request url https://openway.guance.com/v1/write/xxx/token=tkn_xxx failed:  ... context deadline exceeded...
```

## 主机相关 {#iss-host}

主机相关的问题，一般比较隐秘，往往不会关注这个点，所以较难排查（越低级越难查），所以这里大概列一下：

### 时间戳异常 {#iss-timestamp}

在 Linux/Mac 上，输入 `date` 即可查看当前系统时间：

```shell
$ date
Wed Jul 21 16:22:32 CST 2021
```

有些情况下，这里可能显示成这样：

```shell
$ date
Wed Jul 21 08:22:32 UTC 2021
```

这是因为，前者是中国东八区时间，后者是格林威治时间，两者相差 8 小时，但实际上，这两个时间的时间戳是一样的。

如果当前系统的时间跟你的手机时间相差甚远，特别是，它如果超前了，那么观测云上是看不到这些「将来」的数据的。

另外，如果时间滞后，观测云默认的查看器是看不到这些数据的（一般查看器默认显示最近 15min 的数据），可以在查看器上调一下查看的时间范围。

### 主机软硬件不支持 {#iss-os-arch}

某些采集器在一些特定的平台上是不支持的，即使开启了配置，也不会有数据采集：

- macOS 上没有 CPU 采集器
- Oracle/DB2/OceanBase/eBPF 等采集器都只能在 Linux 上运行
- 一些 Windows 特定的采集器，在非 Windows 平台也无法运行
- Datakit-Lite 发行版只编译了小部分采集器，大部分采集器都没有包含在其二进制中

## 启动问题 {#iss-start-fail}

由于 Datakit 适配了主流 OS/Arch 种类，不排除在某些 OS 的发行版上出现部署问题，服务安装完成后，可能服务处于异常状态，导致 Datakit 无法启动

### *datakit.conf* 有误 {#iss-timestamp}

*datakit.conf* 是 Datakit 主配置入口，如果它配置有误（toml 语法错误），会导致 Datakit 无法启动，且 Datakit 日志中有类似如下日志（不同语法错误信息不同）：

```shell
# 手动启动 datakit 程序
$ /usr/local/datakit
2023-11-27T14:17:15.578+0800    INFO    main    datakit/main.go:166     load config from /user/local/datakit/conf.d/datakit.conf...
2023-11-27T14:15:56.519+0800    ERROR   main    datakit/main.go:169     load config failed: bstoml.Decode: toml: line 19 (last key "default_enabled_inputs"): expected value but found "ulimit" instead
```

### 服务异常 {#iss-service-fail}

由于某些原因（如 Datakit 服务启动超时）会导致 `datakit` 服务处于无效状态，此时需要[一些操作来重置 Datakit 系统服务](datakit-service-how-to.md#when-service-failed)。

### 一直重启或不启动 {#iss-restart-notstart}

在 Kubernetes 中，可能资源不够（内存）会导致 Datakit OOM，无暇执行具体的数据采集。可以检查 *datakit.yaml* 中所分配的内存资源是否合适：

```yaml
  containers:
  - name: datakit
    image: pubrepo.guance.com/datakit:datakit:<VERSION>
    resources:
      requests:
        memory: "128Mi"
      limits:
        memory: "4Gi"
    command: ["stress"]
    args: ["--vm", "1", "--vm-bytes", "150M", "--vm-hang", "1"]
```

此处，要求系统上最低（`requests`）有 128MB 的内存才能启动 Datakit。如果 Datakit 自身采集任务繁重，可能默认的 4GB 是不够用的，需要调整 `limits` 参数。

### 端口被占用 {#iss-port-in-use}

部分 Datakit 采集器需要在本地开启特定的端口，以接收外部的数据，如果这些端口被占用，对应的采集器会无法启动，在 Datakit 日志中会出现类似端口被占用的信息。

受端口影响的主流采集器有：

- HTTP 的 9529 端口：部分采集器（如 eBPF/Oracle/LogStream 等采集器）是通过往 Datakit 的 HTTP 接口推送数据
- StatsD 8125 端口：用于接收 StatsD 的指标数据（如 JVM 相关指标）
- OpenTelemetry 4317 端口：用于接收 OpenTelemetry 指标和 Trace 数据

更多端口占用，参见[这个列表](datakit-port.md)。

### 磁盘空间不够 {#iss-disk-space}

磁盘空间不够会导致未定义行为(Datakit 自身日志无法写入/diskcache 无法写入等)。

### 占用资源超出默认设定 {#iss-ulimit}

Datakit 安装完后，其打开的文件数默认是 64K（Linux），如果某个采集（比如文件日志采集器）打开了太多文件，会导致后续文件无法打开，影响采集。

另外，打开的文件过多，证明当前采集出现了比较严重的拥堵，可能会消耗太多的内存资源，进而导致 OOM。

## 采集器配置相关 {#iss-input-config}

采集器配置相关的问题一般比较直接，主要有如下一些可能：

### 被采集对象无数据产生 {#iss-input-nodata}

以 MySQL 为例，一些慢查询或锁有关的采集，要确实出现对应问题，才会有数据，不然在对应的视图中是看不到数据的。

另外，一些暴露 Prometheus 指标的被采集对象，其 Prometheus 指标采集可能默认是关闭的（或者只能 localhost 才能访问），这些都需要在被采集对象上做对应的配置，Datakit 才能采集到这些数据。这种没有数据产生的问题，通过上面的采集器调试功能（`datakit debug --input-conf ...`）即可验证。

对日志采集而言，如果对应的日志文件没有新（相对 Datakit 启动以后）的日志产生，即使当前该日志文件中已经有日志数据，也不会有数据采集上来。

对 Profiling 类数据采集，也需要被采集的服务/应用开启对应的功能，才会有 Profiling 数据推送到 Datakit。

### 访问权限 {#iss-permission-deny}

很多中间件的采集，需要提供用户认证配置，这些配置有些需要在被采集对象上做设置。如果对应的用户名/密码配置有误，Datakit 采集会报错。

另外，由于 Datakit 采用 toml 配置，一些密码字符串需要一些额外的转义（一般是 URL-Encode），比如，如果密码中含有 `@` 字符，需要将它转成 `%40`。

> Datakit 在逐步优化现有的密码字符串（连接字符串）配置方式，以减少这类转义。

### 版本问题 {#iss-version-na}

某些用户环境的软件版本可能太老或太新，不在 Datakit 的支持列表中，可能会出现采集问题。

不在 Datakit 的支持列表重，可能也能采集，我们不可能测试所有的版本号。当有不兼容/不支持的版本，需要反馈。

### 采集器 Bug {#iss-datakit-bug}

这种直接走 [Bug Report](why-no-data.md#bug-report) 即可。

### 未开启采集器配置 {#iss-no-input}

由于 Datakit 只识别 *conf.d* 目录下 `.conf` 配置文件，一些采集器配置可能放错了位置，或者扩展名错误，导致 Datakit 略过了其配置。修正对应的文件位置或文件名即可。

### 采集器被禁用 {#iss-input-disabled}

在 Datakit 的主配置中，可以禁用某些采集器，即使在 *conf.d* 中正确配置了采集器，Datakit 也会忽略这类采集器：

```toml
default_enabled_inputs = [
  "-disk",
  "-mme",
]
```

这些前面带 `-` 的采集器，都是被禁用的。去掉前面的 `-` 或移除该项即可。

### 采集器配置有误 {#iss-invalid-conf}

Datakit 采集器使用 TOML 格式的配置文件，当配置文件不符合 TOML 规范，或者不符合程序字段定义的类型时（比如将整数配置成字符串等），会出现配置文件加载失败的问题，进而导致采集器不会开启。

Datakit 内置了配置检测功能，参见[这里](datakit-tools-how-to.md#check-conf)。

### 配置方式有误 {#iss-config-mistaken}

Datakit 中采集器配置有俩大类：

- 主机安装：直接在 *conf.d* 目录下增加对应的采集器配置即可
- Kubernetes 安装：
    - 可以通过 ConfigMap 直接挂载采集器配置
    - 也可以通过环境变量修改（如果同时存在 ConfigMap 和环境变量，最终以环境变量中的配置为准）
    - 也可以通过 Annotation 等方式标注采集配置（Annotation 标注相对环境变量和 ConfigMap，其优先级最高）
    - 如果同时在默认采集器列表（`ENV_DEFAULT_ENABLED_INPUTS`）中指定了某个采集器，在 ConfigMap 中又增加了同名、同配置采集器，其行为是未定义的，可能会触发下面提到的单例采集器问题

### 单例采集器 {#iss-singleton}

[单例采集器](datakit-input-conf.md#input-singleton)在一个 Datakit 中，只能开启一个，如果开启了多个采集器，Datakit 以文件名顺序加载第一个（如果在同一个 `.conf` 文件中，只加载第一个），其它的不再加载。这种可能导致排名靠后的采集器不会开启。

## 全局配置相关 {#iss-global-settings}

某些 Datakit 段的全局配置，也会影响被采集的数据，除了上面禁用的采集器之外，还有如下一些方面。

### 黑名单/Pipeline 影响 {#iss-pipeline-filter}

用户可能在观测云页面上配置了黑名单，其作用是丢弃符合某些特征的数据不予上传。

而 Pipeline 本身也有丢弃数据的操作（`drop()`）。

这俩类丢弃行为都可以在 `datakit monitor -V` 的输出中看到。

除了丢弃数据之外，Pipeline 还可能修改数据，这些修改可能导致前端查询被影响，比如切割了时间字段出现问题，导致时间出现较大的偏差。

### 磁盘缓存 {#iss-diskcache}

Datakit 对一些复杂数据的处理设置了磁盘缓存机制，这些数据由于处理消耗大，暂时将它们缓存到磁盘做削峰处理，它们会延后上报。通过查看[磁盘缓存相关的指标](datakit-metrics.md#metrics)，可获知对应的数据缓存情况。

### Sinker Dataway {#iss-sinker-dataway}

如果开启了 [Sinker Dataway](../deployment/dataway-sink.md)，根据已有的 Sinker 规则配置，一些数据因不匹配规则而丢弃。

### IO Busy {#iss-io-busy}

由于 Datakit 跟 Dataway 之间网络带宽限制，导致上报数据比较慢，进而影响了数据采集（来不及消费），在这种情况下，Datakit 会丢弃来不及处理的指标数据，而非指标数据，会阻塞采集，从而导致观测云页面看不到数据。

### Dataway 缓存 {#iss-dataway-cache}

Dataway 和观测云中心如果发生网络故障，Dataway 会缓存 Datakit 推送过来的数据，这部分数据可能延迟到达，或者最终丢弃（数据超过了磁盘缓存限额）

### 账号问题 {#iss-workspace}

如果用户观测云账号欠费/数据使用超量，会导致 Datakit 数据上报出现 4xx 问题。这种问题在 `datakit monitor` 能直接看到。

## 其它  {#iss-others}

`datakit monitor -V` 会输出非常多的状态信息，由于分辨率问题，一些数据不会直接显示，需要在对应的表格中滚动才能看到。

但是某些终端不支持当前 monitor 的拖拉操作，容易被误认为没有采集。可以通过指定特定的模块（每个表格头部的红色字母表示模块）来查看 monitor：

```shell
# 查看 HTTP API 的状态
$ datakit monitor -M H

# 查看采集器配置和采集情况
$ datakit monitor -M I

# 查看基本信息以及运行态资源占用情况
$ datakit monitor -M B,R
```

以上是无数据问题的一些基本的排查思路，下面介绍这些排查过程中用到的一些 Datakit 自身的功能以及方法。

## 收集 DataKit 运行信息 {#bug-report}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)

经过各种排查后，可能还是无法找到问题，这时候我们需要收集 Datakit 的各种信息（如日志、配置文件、profile 和自身指标数据等），为了简化这个过程，DataKit 提供了一个命令，可以一次性获取所有相关信息并将其打包到一个文件中。使用方式如下：

```shell
$ datakit debug --bug-report
...
```

<!-- markdownlint-disable MD046 -->
???+ tip


    - 默认情况下，该命令会收集 profile 数据，这可能会对 DataKit 产生一定的性能影响，可以通过下面命令来禁用采集 profile ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0))：
    
    ```shell
    $ datakit debug --bug-report --disable-profile
    ```
    
    执行成功后，在当前目录下生成一个 zip 文件，命名格式为 `info-<时间戳毫秒数>.zip`。
    
    - 如果有公网访问，可以直接将文件上传到 OSS，避免麻烦的文件拷贝（[:octicons-tag-24: Version-1.27.0](changelog.md#cl-1.27.0)）：
    
    
    ```shell hl_lines="7"
    # 此处*必须填上*正确的 OSS 地址/Bucket 名称以及对应的 AS/SK
    $ datakit debug --bug-report --oss OSS_HOST:OSS_BUCKET:OSS_ACCESS_KEY:OSS_SECRET_KEY
    ...
    bug report saved to info-1711794736881.zip
    uploading info-1711794736881.zip...
    download URL(size: 1.394224 M):
        https://OSS_BUCKET.OSS_HOST/datakit-bugreport/2024-03-30/dkbr_co3v2375mqs8u82aa6sg.zip
    ```
    
    将底部的链接地址贴给我们即可（请确保 OSS 中的文件是公网可访问的，否则该链接无法直接下载）。
    
    
    - 默认情况下，bug report 会收集 3 次 Datakit 自身指标，可以通过 `--nmetrics` 调整这里的次数（[:octicons-tag-24: Version-1.27.0](changelog.md#cl-1.27.0)）：
    
    ```shell
    $ datakit debug --bug-report --nmetrics 10
    ```
<!-- markdownlint-enable -->

解压后的文件列表参考如下：

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

| 文件名称       | 是否目录 | 说明                                                                                                    |
| ---:          | ---:     | ---:                                                                                                    |
| `config`      | 是       | 配置文件，包括主配置和已开启的采集器配置                                                                |
| `basic`       | 是       | 运行环境操作系统和环境变量信息                                                                                  |
| `data`        | 是       | `data` 目录下的黑名单文件，即 `.pull` 文件                                                                                |
| `log`         | 是       | 最新的日志文件，包括 log 和 gin log，暂不支持 `stdout`                                                  |
| `profile`     | 是       | pprof 开启时（[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2)已默认开启），会采集 profile 数据 |
| `metrics`     | 是       | `/metrics` 接口返回的数据，命名格式为 `metric-<时间戳毫秒数>`                                           |
| `syslog`      | 是       | 仅支持 `linux`, 基于 `journalctl` 来获取相关日志                                                        |
| `error.log`   | 否       | 记录命令输出过程中出现的错误信息                                                        |

### 敏感信息处理 {#sensitive}

信息收集时，敏感信息（如 token、密码等）会被自动过滤替换，具体规则如下：

- 环境变量

只获取以 `ENV_` 开头的环境变量，且对环境变量名称中包含 `password`, `token`, `key`, `key_pw`, `secret` 的环境变量进行脱敏处理，替换为 `******`

- 配置文件

配置文件内容进行正则替换处理，如：

- `https://openway.guance.com?token=tkn_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` 被替换成 `https://openway.guance.com?token=******`
- `pass = "1111111"` 被替换成 `pass = "******"`
- `postgres://postgres:123456@localhost/test` 被替换成 `postgres://postgres:******@localhost/test`

经过上述处理，能够去除绝大部分敏感信息。尽管如此，导出的文件中还是可能存在敏感信息，可以手动将敏感信息移除，请务必确认。

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
# 10 points("M"), 98 time series from disk, cost 1.544792ms | Ctrl+c to exit.
```

该命令会启动采集器，并将采集器采集到的数据在终端打印出来。底部会显示：

- 采集的点数以及其类型（此处 `M` 表示时序数据）
- 时间线数量（仅时序数据）
- 采集器名称（此处为 `disk`）
- 采集耗时

用 Ctrl + c 可以结束调试。为了尽快得到采集的数据，可以适当调整采集器的采集间隔（如果有）。

<!-- markdownlint-disable MD046 -->
???+ tip

    - 部分被动接收数据的采集器（比如 DDTrace/RUM）需要指定 HTTP 服务（`--hppt-listen=[IP:Port]`），然后通过一些 HTTP 客户端工具（比如 `curl`）将数据发送给 Datakit 对应地址。详见 `datakit help debug` 帮助

    - 调试用的采集器配置可以是任何形式的扩展名，不一定要[以 `.conf` 作为后缀](datakit-input-conf.md#intro)，我们可以用诸如 *my-input.conf.test* 这样的文件名专用于调试，同时又不影响 Datakit 的正常运行
<!-- markdownlint-enable -->

## 查看 Monitor 页面 {#monitor}

参见[这里](datakit-monitor.md)

## 通过 DQL 查看是否有数据产生 {#dql}

在 Windows/Linux/Mac 上，这一功能均支持，其中 Windows 需在 Powershell 中执行

> Datakit [1.1.7-rc7](changelog.md#cl-1.1.7-rc7) 才支持这一功能

```shell
$ datakit dql
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

- 如果日志中发现诸如 `Beyond...` 这样的描述，一般情况下，是因为数据量超过了免费额度
- 如果出现一些 `ERROR/WARN` 等字样，一般情况下，都表明 DataKit 遇到了一些问题

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
