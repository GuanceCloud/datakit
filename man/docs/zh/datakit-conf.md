{{.CSS}}
# DataKit 主配置
---

DataKit 主配置用来配置 DataKit 自己的运行行为。

=== "主机部署"

    其目录一般位于：
    
    - Linux/Mac: `/usr/local/datakit/conf.d/datakit.conf`
    - Windows: `C:\Program Files\datakit\conf.d\datakit.conf`

=== "Kubernates"

    DaemonSet 安装时，虽然在对应目录下也存在这个文件，**但实际上 DataKit 并不加载这里的配置**。这些配是通过在 datakit.yaml 中[注入环境变量](datakit-daemonset-deploy.md#using-k8-env)来生成的。下面所有的配置，都能在 Kubernates 部署文档中找到[对应的环境变量](datakit-daemonset-deploy.md#using-k8-env)配置。

## HTTP 服务的配置 {#config-http-server}

DataKit 会开启 HTTP 服务，用来接收外部数据，或者对外提供基础的数据服务。

=== "datakit.conf"

    ### 修改 HTTP 服务地址 {#update-http-server-host}
    
    默认的 HTTP 服务地址是 `localhost:9529`，如果 9529 端口被占用，或希望从外部访问 DataKit 的 HTTP 服务（比如希望接收 [RUM](rum.md) 或 [Tracing](datakit-tracing.md) 数据），可将其修改成：
    
    ```toml
    [http_api]
       listen = "0.0.0.0:<other-port>"
       # 或使用 IPV6 地址
       # listen = "[::]:<other-port>"
    ```

    注意，IPv6 支持需 [Datakit 升级到 1.5.7](changelog.md#cl-1.5.7-new)。

    #### 使用 Unix domain socket {#uds}

    Datakit 支持 UNIX domain sockets 访问。开启方式如下: `listen` 字段配置为<b>一个不存在文件的全路径</b>，这里以 `datakit.sock` 举例，可以为任意文件名。
    ```toml
    [http_api]
       listen = "/tmp/datakit.sock"
    ```
    配置完成后可以使用 `curl` 命令测试是否配置成功: `sudo curl --no-buffer -XGET --unix-socket /tmp/datakit.sock http:/localhost/v1/ping`。更多关于 `curl` 的测试命令的信息可以参阅[这里](https://superuser.com/a/925610)。
    
    ### HTTP 请求频率控制 {#set-http-api-limit}
    
    由于 DataKit 需要大量接收外部数据写入，为了避免给所在节点造成巨大开销，可修改如下 HTTP 配置（默认不开启）：
    
    ```toml
    [http_api]
      request_rate_limit = 1000.0 # 限制每个 HTTP API 每秒只接收 1000 次请求
    ```

    ### 其它设置 {#http-other-settings}

    ```toml
    [http_api]
        close_idle_connection = true # 关闭闲置连接
        timeout = "30s"              # 设置服务端 HTTP 超时
    ```

=== "Kubernates"

    参见[这里](datakit-daemonset-deploy.md#env-http-api)

## 全局标签（Tag）修改 {#set-global-tag}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

DataKit 允许给其采集的所有数据配置全局标签，全局标签分为两类：

- 主机类全局变量：采集的数据跟当前主机息息相关，比如 CPU/内存等指标数据
- 环境类全局变量：采集的数据来自某个公共实体，比如 MySQL/Redis 等，这些采集一般都参与选举，故这些数据上不会带上主机相关的全局 tag

```toml
[global_host_tags]
  ip         = "__datakit_ip"
  host       = "__datakit_hostname"

[global_election_tags]
  project = "my-project"
  cluster = "my-cluster"
```

加全局 Tag 时，有几个地方要注意：

- 这些全局 Tag 的值可以用 DataKit 目前已经支持的几个变量（双下划线（`__`）前缀和 `$` 都是可以的）：
  - `__datakit_ip/$datakit_ip`：标签值会设置成 DataKit 获取到的第一个主网卡 IP
  - `__datakit_hostname/$datakit_hostname`：标签值会设置成 DataKit 的主机名

- 由于 [DataKit 数据传输协议限制](apis.md#lineproto-limitation)，不要在全局标签（Tag）中出现任何指标（Field）字段，否则会因为违反协议导致数据处理失败。具体参见具体采集器的字段列表。当然，也不要加太多 Tag，而且每个 Tag 的 Key 以及 Value 长度都有限制。

- 如果被采集上来的数据中，本来就带有同名的 Tag，那么 DataKit 不会再追加这里配置的全局 Tag。
- 即使 `global_host_tags` 不配置任何全局 Tag，DataKit 仍然会在所有数据上尝试添加一个 `host=$HOSTNAME` 的全局 Tag。

### 全局 Tag 在远程采集时的设置 {#notice-global-tags}

因为 DataKit 会默认给采集到的所有数据追加标签 `host=<DataKit所在主机名>`，但某些情况这个默认追加的 `host` 会带来困扰。

以 MySQL 为例，如果 MySQL 不在 DataKit 所在机器，但又希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名。

对这种情况，我们有两种方式可以绕过 DataKit 上的全局 tag：

- 在具体采集器中，一般都有一个如下配置，我们可以在这里面新增 Tag，比如，如果不希望 DataKit 默认添加 `host=xxx` 这个 Tag，可以在这里覆盖这个 Tag，以 MySQL 为例：

```toml
[[inputs.mysql.tags]]
  host = "real-mysql-host-name" 
```

- 以 [HTTP API 方式往 DataKit 推送数据](apis.md#api-v1-write)时，可以通过 API 参数 `ignore_global_tags` 来屏蔽所有全局 Tag

???+ tip

    自 [1.4.20](changelog.md#cl-1.4.20) 之后，DataKit 默认会以被采集服务的 IP/Host 等字段为 `host` 字段，故这一问题升级之后将得到改善。建议大家升级到该版本来避免这一问题。

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

### IO 模块调参 {#io-tuning}

[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

=== "datakit.conf"

    某些情况下，DataKit 的单机数据采集量非常大，如果网络带宽有限，可能导致部分数据的采集中断或丢弃。可以通过配置 io 模块的一些参数来缓解这一问题：

    ```toml
    [io]
      feed_chan_size = 4096   # 数据处理队列（一个 job 一般都有多个 point）长度
      max_cache_count = 512   # 数据批量发送点数的阈值，缓存中超过该值即触发发送
      flush_interval = "10s"  # 数据发送的间隔阈值，每隔 10s 至少发送一次
    ```

    阻塞模式参见 [k8s 中的对应说明](datakit-daemonset-deploy.md#env-io)

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-io)


#### IO 磁盘缓存 {#io-disk-cache}

当 DataKit 发送数据失败后，为了不丢失关键数据，可以开启磁盘缓存。磁盘缓存的目的在于将发送失败的数据暂时存入磁盘，待条件允许时，再将数据发送出去。

=== "datakit.conf"

    ```toml
    [io]
      enable_cache = true   # 开启磁盘缓存
      cache_max_size_gb = 5 # 指定磁盘大小为 5GB
    ```

=== "Kubernetes"

    参见[这里](datakit-daemonset-deploy.md#env-io)

---

???+ attention

    目前不支持时序数据的缓存，除此之外的数据，都支持发送失败的磁盘缓存。另外，由于限制了磁盘大小，如果发送一直失败，导致磁盘超过上限，仍然会丢失数据（优先丢弃较老的数据）。

### cgroup 限制  {#enable-cgroup}

由于 DataKit 上处理的数据量无法估计，如果不对 DataKit 消耗的资源做物理限制，将有可能消耗所在节点大量资源。这里我们可以借助 cgroup 来限制，在 *datakit.conf* 中有如下配置：

```toml
[cgroup]
  path = "/datakit" # cgroup 限制目录，如 /sys/fs/cgroup/memory/datakit, /sys/fs/cgroup/cpu/datakit

  # 允许 CPU 最大使用率（百分制）
  cpu_max = 20.0

  # 允许 CPU 最小使用率（百分制）
  cpu_min = 5.0

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

???+ attention

    - cgroup 限制只在[宿主机安装](datakit-install.md)的时候会默认开启
    - cgourp 只支持 CPU 使用率和内存使用量（mem+swap）控制，且只支持 Linux 操作系统。

???+ tip

    Datakit 自 [1.5.8](changelog.md#cl-1.5.8) 开始支持 cgroup v2。如果不确定 cgroup 版本，可通过命令 `mount | grep cgroup` 来确认。

### 选举配置 {#election}

参见[这里](election.md#config)

### DataWay Sinker 配置 {#dataway-sink}

参见[这里](datakit-sink-dataway.md)

### 使用 Git 管理 DataKit 配置 {#using-gitrepo}

参见[这里](git-config-how-to.md)

### 设置打开的文件描述符的最大值 {#enable-max-fd}

Linux 环境下，可以在 Datakit 主配置文件中配置 `ulimit` 项，以设置 Datakit 的最大可打开文件数，如下：

```toml
ulimit = 64000
```

ulimit 默认配置为 64000。在 Kubernates 中，通过[设置 `ENV_ULIMIT`](datakit-daemonset-deploy.md#env-others) 即可。

## FAQ {#faq}

### cgroup 设置失败 {#cgoup-fail}

有时候启用 cgroup 会失败，在 [DataKit Monitor](datakit-monitor.md) 的 `Basic Info` 中会报告类似如下错误：

```
write /sys/fs/cgroup/memory/datakit/memory.limit_in_bytes: invalid argument
```

此时需手动删除已有 cgroup 规则库，然后再[重启 DataKit 服务](datakit-service-how-to.md#manage-service)。

```shell
sudo cgdelete memory:/datakit
```

> `cgdelete` 可能需额外安装工具包：
> 
> - Ubuntu: `apt-get install libcgroup-tools`
> - CentOS: `yum install libcgroup-tools`

### cgroup CPU 使用率说明 {#cgroup-how}

CPU 使用率是百分比制（==最大值 100.0==），以一个 8 核心的 CPU 为例，如果限额 `cpu_max` 为 20.0（即 20%），则 DataKit 最大的 CPU 消耗，==在 top 命令上将显示为 160% 左右==。`cpu_min` 同理。

## 延伸阅读 {#more-reading}

- [DataKit 宿主机安装](datakit-install.md)
- [DataKit DaemonSet 安装](datakit-daemonset-deploy.md)
- [DataKit 行协议过滤器](datakit-filter.md)
