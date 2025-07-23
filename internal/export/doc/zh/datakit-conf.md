
# DataKit 主配置
---

DataKit 主配置用来配置 DataKit 自己的运行行为。

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    其目录一般位于：

    - Linux/Mac: `/usr/local/datakit/conf.d/datakit.conf`
    - Windows: `C:\Program Files\datakit\conf.d\datakit.conf`

=== "Kubernetes"

    DaemonSet 安装时，虽然在对应目录下也存在这个文件，**但实际上 DataKit 并不加载这里的配置**。这些配是通过在 *datakit.yaml* 中[注入环境变量](datakit-daemonset-deploy.md#using-k8-env)来生成的。下面所有的配置，都能在 Kubernetes 部署文档中找到[对应的环境变量](datakit-daemonset-deploy.md#using-k8-env)配置。
<!-- markdownlint-enable -->

## DataKit 主配置示例 {#maincfg-example}

DataKit 主配置示例如下，我们可以根据该示例来开启各种功能（当前版本 {{ .Version }}）：

<!-- markdownlint-disable MD046 -->
??? info "*datakit.conf*"

    ```toml linenums="1"
    {{ CodeBlock .DatakitConfSample 4 }}
    ```
<!-- markdownlint-enable -->

## HTTP 服务的配置 {#config-http-server}

DataKit 会开启 HTTP 服务，用来接收外部数据，或者对外提供基础的数据服务。

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    ### 修改 HTTP 服务地址 {#update-http-server-host}

    默认的 HTTP 服务地址是 `localhost:9529`，如果 9529 端口被占用，或希望从外部访问 DataKit 的 HTTP 服务（比如希望接收 [RUM](../integrations/rum.md) 或 [Tracing](../integrations/datakit-tracing.md) 数据），可将其修改成：

    ```toml
    [http_api]
       listen = "0.0.0.0:<other-port>"
       # 或使用 IPV6 地址
       # listen = "[::]:<other-port>"
    ```

    注意，IPv6 支持需 [DataKit 升级到 1.5.7](changelog.md#cl-1.5.7-new)。

    #### 使用 Unix domain socket {#uds}

    DataKit 支持 UNIX domain sockets 访问。开启方式如下：`listen` 字段配置为<b>一个不存在文件的全路径</b>，这里以 `datakit.sock` 举例，可以为任意文件名。
    ```toml
    [http_api]
       listen = "/tmp/datakit.sock"
    ```
    配置完成后可以使用 `curl` 命令测试是否配置成功：`sudo curl --no-buffer -XGET --unix-socket /tmp/datakit.sock http:/localhost/v1/ping`。更多关于 `curl` 的测试命令的信息可以参阅[这里](https://superuser.com/a/925610){:target="_blank"}。

    ### HTTP 请求频率控制 {#set-http-api-limit}

    > [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 已经默认开启该功能。

    由于 DataKit 需要大量接收外部数据写入，为了避免给所在节点造成巨大开销，DataKit 默认给 API 设置了 20/s 的 QPS 限制：

    ```toml
    [http_api]
      request_rate_limit = 20.0 # 限制每个客户端（IP + API 路由）每秒发起请求的 QPS 限制

      # 如果确实有大量数据写入，可酌情调大限制，避免数据丢失（请求超限后客户端会收到 HTTP 429 错误码）
    ```

    ### 其它设置 {#http-other-settings}

    ```toml
    [http_api]
        close_idle_connection = true # 关闭闲置连接
        timeout = "30s"              # 设置服务端 HTTP 超时
    ```

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-http-api)
<!-- markdownlint-enable -->


### HTTP API 访问控制 {#public-apis}

[:octicons-tag-24: Version-1.64.0](changelog.md#cl-1.64.0)

出于安全考虑，DataKit 默认限制了一些自身 API 的访问（这些 API 只能通过 localhost 访问）。如果 DataKit 部署在公网环境，又需要通过其它机器或公网来请求这些 API，可以在 *datakit.conf* 中，修改如下 `public_apis` 字段配置：

```toml
[http_api]
  public_apis = [
    # 放行 DataKit 自身指标暴露接口 /metrics
    "/metrics",
    # ... # 其它接口
  ]
```

默认情况下，`public_apis` 为空。出于便捷和兼容性考虑，默认只开放了[部分接口](apis.md)，所有其它接口都是禁止外部访问的。而采集器对应的接口，比如 trace 类采集器，一旦开启采集器之后，其访问自动放开，默认就能外部访问。

Kubernetes 中增加 API 白名单参见[这里](datakit-daemonset-deploy.md#env-http-api)。

<!-- markdownlint-disable MD046 -->
???+ warning

    一旦 `public_apis` 不为空，则默认开启的那些 API 接口需要**再次手动添加**：

    ```toml
    [http_api]
      public_apis = [
        "/v1/write/metric",
        "/v1/write/logging",
        # ...
      ]
    ```
<!-- markdownlint-enable -->

## 全局标签（Tag）修改 {#set-global-tag}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

DataKit 允许给其采集的所有数据配置全局标签，全局标签分为两类：

- 主机类全局标签（GHT）：采集的数据跟当前主机绑定，比如 CPU/内存等指标数据
- 选举类全局标签（GET）：采集的数据来自某个公共（远程）实体，比如 MySQL/Redis 等，这些采集一般都参与选举，故这些数据上不会带上当前主机相关的标签

```toml
[global_host_tags] # 这里面的我们称之为「全局主机标签」
  ip   = "__datakit_ip"
  host = "__datakit_hostname"

[election]
  [election.tags] # 这里面的我们称之为「全局选举标签」
    project = "my-project"
    cluster = "my-cluster"
```

加全局标签时，有几个地方要注意：

1. 这些全局标签的值可以用 DataKit 目前已经支持的几个通配（双下划线（`__`）前缀和 `$` 都是可以的）：

    1. `__datakit_ip/$datakit_ip`：标签值会设置成 DataKit 获取到的第一个主网卡 IP
    1. `__datakit_hostname/$datakit_hostname`：标签值会设置成 DataKit 的主机名

1. 由于 [DataKit 数据传输协议限制](apis.md#lineproto-limitation)，不要在全局标签（Tag）中出现任何指标（Field）字段，否则会因为违反协议导致数据处理失败。具体参见具体采集器的字段列表。当然，也不要加太多标签，而且每个标签的 Key 以及 Value 长度都有限制。
1. 如果被采集上来的数据中，本来就带有同名的标签，那么 DataKit 不会再追加这里配置的全局标签
1. 即使 GHT 中没有任何配置，DataKit 仍然会在其中添加一个 `host=__datakit_hostname` 的标签。因为 hostname 是目前<<<custom_key.brand_name>>>平台数据关联的默认字段，故日志/CPU/内存等采集上，都会带上 `host` 这个 tag。
1. 这俩类全局标签（GHT/GET）是可以有交集的，比如都可以在其中设置一个 `project = "my-project"` 的标签
1. 当没有开启选举的情况下，GET 沿用 GHT（它至少有一个 `host` 的标签）中的所有标签
1. 选举类采集器默认追加 GET，非选举类采集器默认追加 GHT。

<!-- markdownlint-disable MD046 -->
???+ tip "如何区分选举和非选举采集器？"

    在采集器文档中，在顶部有类似如下标识，它们表示当前采集器的平台适配情况以及采集特性：

    :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:  · :fontawesome-solid-flag-checkered:

    若带有 :fontawesome-solid-flag-checkered: 则表示当前采集器是选举类采集器。
<!-- markdownlint-enable -->

### 全局 Tag 在远程采集时的设置 {#notice-global-tags}

因为 DataKit 会默认给采集到的所有数据追加标签 `host=<DataKit 所在主机名>`，但某些情况这个默认追加的 `host` 会带来困扰。

以 MySQL 为例，如果 MySQL 不在 DataKit 所在机器，但又希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名。

对这种情况，我们有两种方式可以绕过 DataKit 上的全局 tag：

- 在具体采集器中，一般都有一个如下配置，我们可以在这里面新增 Tag，比如，如果不希望 DataKit 默认添加 `host=xxx` 这个 Tag，可以在这里覆盖这个 Tag，以 MySQL 为例：

```toml
[[inputs.mysql.tags]]
  host = "real-mysql-host-name"
```

- 以 [HTTP API 方式往 DataKit 推送数据](apis.md#api-v1-write)时，可以通过 API 参数 `ignore_global_tags` 来屏蔽所有全局 Tag

<!-- markdownlint-disable MD046 -->
???+ info

    自 [1.4.20](changelog.md#cl-1.4.20) 之后，DataKit 默认会以被采集服务连接地址中的的 IP/Host 作为 `host` 的标签值。
<!-- markdownlint-enable -->

## DataKit 自身运行日志配置 {#logging-config}

DataKit 自身日志有两个，一个是自身运行日志（*/var/log/datakit/log*），一个是 HTTP Access 日志（*/var/log/datakit/gin.log*）。

DataKit 默认日志等级为 `info`。编辑 `datakit.conf`，可修改日志等级以及分片大小：

```toml
[logging]
  level = "debug" # 将 info 改成 debug
  rotate = 32     # 每个日志分片为 32MB
```

- `level`：置为 `debug` 后，即可看到更多日志（目前只支持 `debug/info` 两个级别）。
- `rotate`：DataKit 默认会对日志进行分片，默认分片大小为 32MB，总共 6 个分片（1 个当前写入分片加上 5 个切割分片，分片个数尚不支持配置）。如果嫌弃 DataKit 日志占用太多磁盘空间（最多 32 x 6 = 192MB），可减少 `rotate` 大小（比如改成 4，单位为 MB）。HTTP 访问日志也按照同样的方式自动切割。

## 高级配置 {#advance-config}

下面涉及的内容涉及一些高级配置，如果对配置不是很有把握，建议咨询我们的技术专家。

### 时间校准 {#ntp}

[:octicons-tag-24: Version-1.75.0](../datakit/changelog-2025.md#cl-1.75.0)

为避免本机时间偏差对数据采集的影响，DataKit 可通过调用 DataWay 接口（[:octicons-tag-24: Version-1.6.0](../deployment/dataway-changelog.md#cl-1.6.0)）来感知自身时间是否出现较大偏差。当感知到较大偏差后，DataKit 会校准当前时间（但不会修改系统时间）作为数据采集的时间。

在 *datakit.conf* 中，有如下配置项：

```toml
  # use dataway as NTP server
  [dataway.ntp]
    enable   = true  # default enabled
    interval = "5m"  # sync dataway time each 5min(minimal 1min)

    # if abs(datakit time - dataway time) >= diff, datakit will adjust data point
    # time with dataway time.
    diff     = "30s"  # minimal 5s
```

<!-- markdownlint-disable MD046 -->
???+ warning

    - 该行为默认开启，如果 DataWay 版本较低，最终效果仍旧是采用当前系统时间（即不做任何校准）
    - 目前 eBPP 相关的采集，由于其与 DataKit 是分离运行的，暂不支持时间矫正功能
<!-- markdownlint-enable -->

### IO 模块调参 {#io-tuning}

[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    某些情况下，DataKit 的单机数据采集量非常大，如果网络带宽有限，可能导致部分数据的采集中断或丢弃。可以通过配置 io 模块的一些参数来缓解这一问题：

    ```toml
    [io]
      feed_chan_size  = 1     # 数据处理队列长度
      max_cache_count = 1000  # 数据批量发送点数的阈值，缓存中超过该值即触发发送
      flush_interval  = "10s" # 数据发送的间隔阈值，每隔 10s 至少发送一次
      flush_workers   = 0     # 数据上传 worker 数（默认配额 CPU 核心 * 2）
    ```

    阻塞模式参见 [k8s 中的对应说明](datakit-daemonset-deploy.md#env-io)

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-io)
<!-- markdownlint-enable -->

### 资源限制  {#resource-limit}

由于 DataKit 上处理的数据量无法估计，如果不对 DataKit 消耗的资源做物理限制，将有可能消耗所在节点大量资源。这里我们可以借助 Linux 的 cgroup 和 Windows 的 job object 来限制，在 *datakit.conf* 中有如下配置：

```toml
[resource_limit]
  path = "/datakit" # Linux cgroup 限制目录，如 /sys/fs/cgroup/memory/datakit, /sys/fs/cgroup/cpu/datakit

  # 允许 CPU 核心数
  cpu_cores = 2.0

  cpu_max = 20.0 # 已弃用

  # 默认允许 4GB 内存(memory + swap)占用
  # 如果置为 0 或负数，则不启用内存限制
  mem_max_mb = 4096
```

如果 DataKit 超出内存限制后，会被操作系统强制杀掉，通过命令可以看到如下结果，此时需要[手动启动服务](datakit-service-how-to.md#when-service-failed)：

```shell
$ systemctl status datakit
● datakit.service - Collects data and upload it to DataFlux.
     Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: enabled)
     Active: activating (auto-restart) (Result: signal) since Fri 2022-02-30 16:39:25 CST; 1min 40s ago
    Process: 3474282 ExecStart=/usr/local/datakit/datakit (code=killed, signal=KILL)
   Main PID: 3474282 (code=killed, signal=KILL)
```

<!-- markdownlint-disable MD046 -->
???+ note

    - 资源限制只在[宿主机安装](datakit-install.md)的时候会默认开启
    - 只支持 CPU 使用率和内存使用量（mem+swap）控制，且只支持 Linux 和 windows ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0)) 操作系统。
    - CPU 使用率控制目前不支持这些 windows 操作系统： Windows 7, Windows Server 2008 R2, Windows Server 2008, Windows Vista, Windows Server 2003 和 Windows XP。
    - 非 root 用户改资源限制配置时，必须重装 service。
    - CPU 核心数限制会影响 DataKit 部分子模块的 worker 数配置（一般是 CPU 核心数的整数倍）。比如数据上传 worker 就是 CPU 核心数 * 2。而单个上传 worker 会占用默认 10MB 的内存用于数据发送，故 CPU 核心数如果开放较多，会影响 DataKit 整体内存的占用
    - [:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8) 开始支持 cgroup v2。如果不确定 cgroup 版本，可通过命令 `mount | grep cgroup` 来确认。
    - [:octicons-tag-24: Version-1.68.0](changelog-2025.md#cl-1.68.0) 支持在 *daktait.conf* 中配置 CPU 核心数限制，且弃用原来的百分比配置方式。百分比配置方式会因为不同主机的 CPU 核心数不同而出现 CPU 配额不同，在采集压力相同的情况下，可能会导致一些异常行为。老版本 DataKit 升级上来的时候，在升级命令中指定 `DK_LIMIT_CPUCORES` 环境变量即可。升级命令如果不指定，仍然沿用之前的百分比配置方式。如果重新安装 DataKit，则直接采用 CPU 核心数限额方式。
    - `cpu_max`: CPU 使用率是百分比制（最大值 100.0），以一个 8 核心的 CPU 为例，如果限额 `cpu_max` 为 20.0（即 20%），则 DataKit 最大的 CPU 消耗，在 top 命令上将显示为 160% 左右。

???+ failure "cgroup 设置失败"

    某些主机上，由于 DataKit 自动检测到 cgroup v1，在生成对应的 cgroup 规则时会报错：

    ``` txt
    cgroup setup err=...: open /sys/fs/cgroup/memory/datakit/memory.memsw.limit_in_bytes: permission denied
    ```

    该错误不是因为当前用户权限不够，而是因为内核中并未启用 Swap Accounting 所致。确认是否启用 Swap Accounting：

    ```shell
    # 看看是否有 swapaccount=1 或 cgroup.memory=swapaccount=1
    cat /proc/cmdline
    ```

    如果缺失，需要修改 */etc/default/grub* 中的 `GRUB_CMDLINE_LINUX` 或 `GRUB_CMDLINE_LINUX_DEFAULT`，在尾部添加 `swapaccount=1`，然后运行如下命令，并重启机器：

    ```shell
    sudo update-grub # Debian/Ubuntu
    # 或
    sudo grub2-mkconfig -o /boot/grub2/grub.cfg # CentOS/RHEL/Fedora。
    ```
<!-- markdownlint-enable -->

### 选举配置 {#election}

参见[这里](election.md#config)

### DataWay 参数配置 {#dataway-settings}

Dataway 部分有如下几个配置可以配置，其它部分不建议改动：

- `timeout`：上传<<<custom_key.brand_name>>>的超时时间，默认 30s
- `max_retry_count`：设置 Dataway 发送的重试次数（默认 1 次，最大 10 次）[:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)
- `retry_delay`：设置重试间隔基础步长，默认 1s。所谓基础步长，即第一次 1s，第二次 2s，第三次 4s，以此类推（以 2^n 递增）[:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)
- `max_raw_body_size`：控制单个上传包的最大大小（压缩前），单位字节 [:octicons-tag-24: Version-1.17.1](changelog.md#cl-1.17.1)
- `content_encoding`：可选择 v1 或 v2 [:octicons-tag-24: Version-1.17.1](changelog.md#cl-1.17.1)
    - v1 即行协议（默认 v1）
    - v2 即 Protobuf 协议，相比 v1，它各方面的性能都更优越。运行稳定后，后续将默认采用 v2

Kubernetes 下部署相关配置参见[这里](datakit-daemonset-deploy.md#env-dataway)。

#### WAL 队列配置 {#dataway-wal}

[:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0)

WAL 用于缓存 DataKit 来不及上传的数据，当突发有较大的数据采集时，如果来不及发送，DataKit 会将其写入磁盘队列，避免阻塞数据采集，影响数据的实时性。

WAL 磁盘队列有默认的磁盘大小限制，当缓存数据量超过该限制，新采集的数据就写不进去导致丢弃。如果不希望丢弃这些数据，可以将该数据类型（一般是日志 `L`）配置到 `no_drop_categories` 列表中。此时数据不会主动丢弃，但会阻塞数据采集。

在 `[dataway.wal]` 中，我们可以调整 WAL 队列的配置：

```toml
  [dataway.wal]
     max_capacity_gb = 2.0             # 2GB reserved disk space for each category(M/L/O/T/...)
     workers = 0                       # flush workers on WAL(default to CPU limited cores)
     mem_cap = 0                       # in-memory queue capacity(default to CPU limited cores)
     fail_cache_clean_interval = "30s" # duration for clean fail uploaded data
     #no_drop_categories = ["L"]       # category list that do not drop data if WAL disk full
```

磁盘文件位于 DataKit 安装目录的 *cache/dw-wal* 目录下：

```shell
/usr/local/datakit/cache/dw-wal/
├── custom_object
│   └── data
├── dialtesting
│   └── data
├── dynamic_dw
│   └── data
├── fc
│   └── data
├── keyevent
│   └── data
├── logging
│   ├── data
│   └── data.00000000000000000000000000000000
├── metric
│   └── data
├── network
│   └── data
├── object
│   └── data
├── profiling
│   └── data
├── rum
│   └── data
├── security
│   └── data
└── tracing
    └── data

13 directories, 14 files
```

此处，除了 *fc* 是失败重传队列，其它目录分别对应一种数据类型。当数据上传失败，这些数据会缓存到 *fc* 目录下，后续 DataKit 会间歇性将它们上传上去。

如果当前主机磁盘性能不足，可以尝试 [tmpfs 下使用 WAL](wal-tmpfs.md)。

### Sinker 配置 {#dataway-sink}

参见[这里](../deployment/dataway-sink.md)

### 使用 Git 管理 DataKit 配置 {#using-gitrepo}

参见[这里](git-config-how-to.md)

### 本地设置 Pipeline 默认脚本 {#pipeline-settings}

[:octicons-tag-24: Version-1.61.0](changelog.md#cl-1.61.0)

支持通过本地设置默认 Pipeline 脚本，如果与远程设置的默认脚本冲突，则倾向本地设置。

可通过两种方式配置：

- 主机方式部署，可在 DataKit 主配置文件中指定各类别的默认脚本，如下：

    ```toml
    # default pipeline
    [pipeline.default_pipeline]
        # logging = "<your_script.p>"
        # metric  = "<your_script.p>"
        # tracing = "<your_script.p>"
    ```

- 容器方式部署，可使用环境变量，`ENV_PIPELINE_DEFAULT_PIPELINE`，其值例如 `{"logging":"abc.p","metric":"xyz.p"}`

### 设置打开的文件描述符的最大值 {#enable-max-fd}

Linux 环境下，可以在 DataKit 主配置文件中配置 `ulimit` 项，以设置 DataKit 的最大可打开文件数，如下：

```toml
ulimit = 64000
```

ulimit 默认配置为 64000。在 Kubernetes 中，通过[设置 `ENV_ULIMIT`](datakit-daemonset-deploy.md#env-others) 即可。

### 采集器密码保护 {#secrets_management}

[:octicons-tag-24: Version-1.31.0](changelog.md#cl-1.31.0)

在配置文件中，我们可以对敏感的密码登信息加密。DataKit 在启动加载采集器配置文件时遇到 `ENC[]` 这种形式的字符串时，会主动解密密码，以得到正确的密码。

ENC 目前支持三种方式：

- 文件形式（`ENC[file:///path/to/enc4dk]`）：在对应的文件中填写正确的密码即可
- AES 加密方式（`ENC[aes://5w1UiRjWuVk53k96WfqEaGUYJ/Oje7zr8xmBeGa3ugI=]`）：需要在主配置文件 *datakit.conf*  中配置秘钥： `aes_key` 或者 `aes_key_file`, 秘钥长度是 16 位
- 环境变量方式：DataKit Kubernetes 安装时可通过[环境变量（`ENV_CRYPTO_*`）来设置](datakit-daemonset-deploy.md#env-others)

接下来以 MySQL 采集器为例，说明两种方式如何配置使用：

- 文件形式

    首先，将明文密码放到文件 `/usr/local/datakit/enc4mysql` 中，然后修改配置文件 mysql.conf:

    ```toml
    # 部分配置
    [[inputs.mysql]]
      host = "localhost"
      user = "datakit"
      pass = "ENC[file:///usr/local/datakit/enc4mysql]"
      port = 3306
      # sock = "<SOCK>"
      # charset = "utf8"
    ```

    DK 会从 `/usr/local/datakit/enc4mysql` 中读取密码并替换密码，替换后为 `pass = "Hello*******"`

- AES 加密方式

    首先在 `datakit.conf` 中配置秘钥：

    ```toml
    # crypto key or key filePath.
    [crypto]
      # 配置秘钥
      aes_key = "0123456789abcdef"
      # 或者，将秘钥放到文件中并在此配置文件位置。
      aes_Key_file = "/usr/local/datakit/mykey"
    ```

    `mysql.conf` 配置文件：

    ```toml
    pass = "ENC[aes://5w1UiRjWuVk53k96WfqEaGUYJ/Oje7zr8xmBeGa3ugI=]"
    ```

注意，通过 `AES` 加密得到的密文需要完整的填入。以下是代码示例：

<!-- markdownlint-disable MD046 -->
=== "Golang"

    ```go
    // AESEncrypt  加密。
    func AESEncrypt(key []byte, plaintext string) (string, error) {
        block, err := aes.NewCipher(key)
        if err != nil {
            return "", err
        }

        // PKCS7 padding
        padding := aes.BlockSize - len(plaintext)%aes.BlockSize
        padtext := bytes.Repeat([]byte{byte(padding)}, padding)
        plaintext += string(padtext)
        ciphertext := make([]byte, aes.BlockSize+len(plaintext))
        iv := ciphertext[:aes.BlockSize]
        if _, err := io.ReadFull(rand.Reader, iv); err != nil {
            return "", err
        }
        mode := cipher.NewCBCEncrypter(block, iv)
        mode.CryptBlocks(ciphertext[aes.BlockSize:], []byte(plaintext))

        return base64.StdEncoding.EncodeToString(ciphertext), nil
    }

    // AESDecrypt AES  解密。
    func AESDecrypt(key []byte, cryptoText string) (string, error) {
        ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
        if err != nil {
            return "", err
        }

        block, err := aes.NewCipher(key)
        if err != nil {
            return "", err
        }

        if len(ciphertext) < aes.BlockSize {
            return "", fmt.Errorf("ciphertext too short")
        }

        iv := ciphertext[:aes.BlockSize]
        ciphertext = ciphertext[aes.BlockSize:]

        mode := cipher.NewCBCDecrypter(block, iv)
        mode.CryptBlocks(ciphertext, ciphertext)

        // Remove PKCS7 padding
        padding := int(ciphertext[len(ciphertext)-1])
        if padding > aes.BlockSize {
            return "", fmt.Errorf("invalid padding")
        }
        ciphertext = ciphertext[:len(ciphertext)-padding]

        return string(ciphertext), nil
    }
    ```

=== "Java"

    ```java
    import javax.crypto.Cipher;
    import javax.crypto.spec.IvParameterSpec;
    import javax.crypto.spec.SecretKeySpec;
    import java.security.SecureRandom;
    import java.util.Base64;

    public class AESUtils {
        public static String AESEncrypt(byte[] key, String plaintext) throws Exception {
            javax.crypto.Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");

            SecureRandom random = new SecureRandom();
            byte[] iv = new byte[16];
            random.nextBytes(iv);
            IvParameterSpec ivParameterSpec = new IvParameterSpec(iv);
            cipher.init(Cipher.ENCRYPT_MODE, secretKeySpec, ivParameterSpec);
            byte[] encrypted = cipher.doFinal(plaintext.getBytes());
            byte[] ivAndEncrypted = new byte[iv.length + encrypted.length];
            System.arraycopy(iv, 0, ivAndEncrypted, 0, iv.length);
            System.arraycopy(encrypted, 0, ivAndEncrypted, iv.length, encrypted.length);

            return Base64.getEncoder().encodeToString(ivAndEncrypted);
        }

        public static String AESDecrypt(byte[] key, String cryptoText) throws Exception {
            byte[] ciphertext = Base64.getDecoder().decode(cryptoText);

            SecretKeySpec secretKeySpec = new SecretKeySpec(key, "AES");

            if (ciphertext.length < 16) {
                throw new Exception("ciphertext too short");
            }

            byte[] iv = new byte[16];
            System.arraycopy(ciphertext, 0, iv, 0, 16);
            byte[] encrypted = new byte[ciphertext.length - 16];
            System.arraycopy(ciphertext, 16, encrypted, 0, ciphertext.length - 16);

            Cipher cipher = Cipher.getInstance("AES/CBC/PKCS5Padding");
            IvParameterSpec ivParameterSpec = new IvParameterSpec(iv);
            cipher.init(Cipher.DECRYPT_MODE, secretKeySpec, ivParameterSpec);

            byte[] decrypted = cipher.doFinal(encrypted);

            return new String(decrypted);
        }
    }
    public static void main(String[] args) {
        try {
            String key = "0123456789abcdef"; // 16, 24, or 32 bytes AES key
            String plaintext = "HelloAES9*&.";
            byte[] keyBytes = key.getBytes("UTF-8");

            String encrypted = AESEncrypt(keyBytes, plaintext);
            System.out.println("Encrypted text: " + encrypted);

            String decrypt = AESDecrypt(keyBytes, encrypted);
            System.out.println("解码后的是："+decrypt);
        } catch (Exception e) {
            System.out.println(e);
            e.printStackTrace();
        }
    }
    ```
<!-- markdownlint-enable -->

### 远程任务 {#remote-job}

---

[:octicons-tag-24: Version-1.63.0](changelog.md#cl-1.63.0)

---

DataKit 接收中心下发任务并执行。目前支持 `JVM dump` 功能。

该功能是执行 `jmap` 命令，生成一个 jump 文件，并上传到 `OSS` `AWS S3 Bucket` 或者 `HuaWei Cloud OBS` 中。

安装 DK 之后会在安装目录下 `template/service-task` 生成两个文件：`jvm_dump_host_script.py` 和 `jvm_dump_k8s_script.py` 前者是宿主机模式下的脚本，后者是 k8s 环境下的。

DK 启动之后会定时执行脚本，如果修改脚本 那么 DK 重启之后会覆盖掉。

宿主机环境下，当前的环境需要有 `python3` 以及包。如果没有 需要安装 ：

```shell
# 有 python3 环境
pip install requests
# 或者
pip3 install requests

# 如果需要上传到华为云 OBS 需要安装库：
pip install esdk-obs-python --trusted-host pypi.org

# 如果需要上传到 AWS S3 需要安装 boto3:
pip install boto3
```

通过环境变量可以控制上传到多个存储捅类型，以下是配置说明， k8s 环境同理：

```toml
# upload to OSS
[remote_job]
  enable = true
  envs = [
      "REMOTE=oss",
      "OSS_BUCKET_HOST=host","OSS_ACCESS_KEY_ID=key","OSS_ACCESS_KEY_SECRET=secret","OSS_BUCKET_NAME=bucket",
    ]
  interval = "30s"

# or upload to AWS:
[remote_job]
  enable = true
  envs = [
      "REMOTE=aws",
      "AWS_BUCKET_NAME=bucket","AWS_ACCESS_KEY_ID=AK","AWS_SECRET_ACCESS_KEY=SK","AWS_DEFAULT_REGION=us-west-2",
    ]
  interval = "30s"

# or upload to OBS:
[remote_job]
  enable = true
  envs = [
      "REMOTE=obs",
      "OBS_BUCKET_NAME=bucket","OBS_ACCESS_KEY_ID=AK","OBS_SECRET_ACCESS_KEY=SK","OBS_SERVER=https://xxx.myhuaweicloud.com"
    ]
  interval = "30s"
```

K8S 环境下需要调用 Kubernetes API 所以需要 RBAC 基于角色的访问控制

配置相关：

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    其目录一般位于：

    - Linux/Mac: `/usr/local/datakit/conf.d/datakit.conf`
    - Windows: `C:\Program Files\datakit\conf.d\datakit.conf`

    修改配置，如果没有在最后添加：
    ```toml
    [remote_job]
      enable=true
      envs=["REMOTE=oss","OSS_BUCKET_HOST=<bucket_host>","OSS_ACCESS_KEY_ID=<key>","OSS_ACCESS_KEY_SECRET=<secret key>","OSS_BUCKET_NAME=<name>"]
      interval="100s"
      java_home=""
    ```

=== "Kubernetes"

    修改 DataKit yaml 文件，添加 RBAC 权限

    ```yaml

    ---

    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
    name: datakit
    rules:
    - apiGroups: ["rbac.authorization.k8s.io"]
      resources: ["clusterroles"]
      verbs: ["get", "list", "watch"]
    - apiGroups: [""]
      resources: ["nodes", "nodes/stats", "nodes/metrics", "namespaces", "pods", "pods/log", "events", "services", "endpoints", "persistentvolumes", "persistentvolumeclaims", "pods/exec"]
      verbs: ["get", "list", "watch", "create"]
    - apiGroups: ["apps"]
      resources: ["deployments", "daemonsets", "statefulsets", "replicasets"]
      verbs: ["get", "list", "watch"]
    - apiGroups: ["batch"]
      resources: ["jobs", "cronjobs"]
      verbs: [ "get", "list", "watch"]
    - apiGroups: ["<<<custom_key.brand_main_domain>>>"]
      resources: ["datakits"]
      verbs: ["get","list"]
    - apiGroups: ["monitoring.coreos.com"]
      resources: ["podmonitors", "servicemonitors"]
      verbs: ["get", "list"]
    - apiGroups: ["metrics.k8s.io"]
      resources: ["pods", "nodes"]
      verbs: ["get", "list"]
    - nonResourceURLs: ["/metrics"]
      verbs: ["get"]

    ---
    ```

    在上面的配置中，添加了 "pod/exec"，其他的保持和 yaml 一致即可。

    添加 remote_job 环境变量：

    ```yaml
    - name: ENV_REMOTE_JOB_ENABLE
      value: 'true'
    - name: ENV_REMOTE_JOB_ENVS
      value: >-
        REMOTE=oss,OSS_BUCKET_HOST=<bucket host>,OSS_ACCESS_KEY_ID=<key>,OSS_ACCESS_KEY_SECRET=<secret key>,OSS_BUCKET_NAME=<name>
    - name: ENV_REMOTE_JOB_JAVA_HOME
    - name: ENV_REMOTE_JOB_INTERVAL
      value: 100s

    ```

<!-- markdownlint-enable -->

配置说明：

1. `enable  ENV_REMOTE_JOB_ENABLE remote_job` 功能开关。
2. `envs  ENV_REMOTE_JOB_ENVS` 其中包括 `host` `access key` `secret key` `bucket` 信息，将获取到的 JVM dump 文件发送到 OSS 中，AWS 和 OBS 同理，更换环境变量即可。
3. `interval ENV_REMOTE_JOB_INTERVAL` DataKit 主动调用接口获取最新任务的时间间隔。
4. `java_home ENV_REMOTE_JOB_JAVA_HOME` 宿主机环境自动从环境变量（$JAVA_HOME）中获取，可以不用配置。

> 注意，使用的 Agent:`dd-java-agent.jar` 版本不应低于 `v1.4.0-guance`

### Point 缓存 {#point-pool}

[:octicons-tag-24: Version-1.28.0](changelog.md#cl-1.28.0)

> Point 缓存目前有额外的性能问题，不建议使用。

为了优化 DataKit 高负载情况下的内存占用，可以开启 Point Pool 来缓解：

```toml
# datakit.conf
[point_pool]
    enable = true
    reserved_capacity = 4096
```

同时，[DataKit 配置](datakit-conf.md#dataway-settings)中可以开启 `content_encoding = "v2"` 的传输编码（[:octicons-tag-24: Version-1.32.0](changelog.md#cl-1.32.0) 已默认启用 v2），相比 v1，它的内存和 CPU 开销都更低。

<!-- markdownlint-disable MD046 -->
???+ warning

    - 在低负载（DataKit 内存占用 100MB 左右）的情况下，开启 point pool 会增加 DataKit 自身的内存占用。所谓的高负载，一般指占用内存在 2GB+ 的场景。同时开启后也能改善 DataKit 自身的 CPU 消耗

<!-- markdownlint-enable -->

## 延伸阅读 {#more-reading}

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>宿主机安装</u>: 在服务器上安装 DataKit </font>](datakit-install.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Kubernetes 安装</u>: DaemonSet 安装 DataKit</font>](datakit-daemonset-deploy.md)
</div>
</font>
