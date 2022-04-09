{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 主配置

DataKit 主配置用来配置 DataKit 自己的运行行为，其目录一般位于：

- Linux/Mac: `/usr/local/datakit/conf.d/datakit.conf`
- Windows: `C:\Program Files\datakit\conf.d\datakit.conf`

> DaemonSet 安装时，虽然在对应目录下也存在这个文件，==但实际上 DataKit 并不加载这里的配置==。这些配是通过在 datakit.yaml 中[注入环境变量](datakit-daemonset-deploy#00c8a780)来生成的。

## HTTP 服务的配置

DataKit 会开启 HTTP 服务，用来接收外部数据，或者对外提供基础的数据服务。

### 修改 HTTP 服务地址

默认的 HTTP 服务地址是 `localhost:9529`，如果 9529 端口被占用，或希望从外部访问 DataKit 的 HTTP 服务（比如希望接收 [RUM](rum) 或 [Tracing](datakit-tracing) 数据），可将其修改成：

```toml
[http_api]
   listen = "0.0.0.0:<other-port>"
```

### HTTP 请求频率控制

由于 DataKit 需要大量接收外部数据写入，为了避免给所在节点造成巨大开销，可修改如下 HTTP 配置（默认不开启）：

```toml
[http_api]
	request_rate_limit = 1000.0 # 限制每个 HTTP API 每秒只接收 1000 次请求
```

## 全局标签（Tag）修改

DataKit 允许给其采集的所有数据配置全局标签，这些标签会默认添加到该 DataKit 采集的每一条数据上。这里是一个全局标签配置示例：

> 如果被采集上来的数据中，本来就带有同名的 Tag，那么 DataKit 不会再追加这里配置的全局 Tag。

```toml
[global_tags]
	ip         = "__datakit_ip"
	datakit_id = "$datakit_id"
	project    = "some_of_your_online_biz"
	other_tags = "..."                    # 可追加其它更多标签
```

加全局 Tag 时，有几个地方要注意：

- 这些全局 Tag 的值可以用 DataKit 目前已经支持的几个变量（双下划线（`__`）前缀和 `$` 都是可以的）：
  - `__datakit_ip/$datakit_ip`：标签值会设置成 DataKit 获取到的第一个主网卡 IP
  - `__datakit_id/$datakit_id`：标签值会设置成 DataKit 的 ID

- 由于 [DataKit 数据传输协议限制](apis#2fc2526a)，不要在全局标签（Tag）中出现任何指标（Field）字段，否则会因为违反协议导致数据处理失败。具体参见具体采集器的字段列表。当然，也不要加太多 Tag，而且每个 Tag 的 Key 以及 Value 长度都有限制。

### 全局 Tag 可能导致的问题

因为 DataKit 会默认给采集到的所有数据追加标签 `host=<DataKit所在主机名>`，但某些情况这个默认追加的 `host` 会带来困扰。

以 MySQL 为例，如果 MySQL 不在 DataKit 所在机器，肯定希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名。

对这种情况，我们有两种方式可以屏蔽掉这些全局 Tag：

- 在具体采集器中，一般都有一个如下配置，我们可以在这里面新增 Tag，比如，如果不希望 DataKit 默认添加 `host=xxx` 这个 Tag，可以在这里覆盖这个 Tag，以 MySQL 为例：

```toml
[[inputs.mysql.tags]]
  host = "real-mysql-host-name" 
```

- 以 [HTTP API 方式往 DataKit 推送数据](apis#f53903a9)时，可以通过 API 参数 `ignore_global_tags` 来屏蔽所有全局 Tag

## DataKit 自身运行日志配置

DataKit 自身日志有两个，一个是自身运行日志（*/var/log/datakit/log*），一个是 HTTP Access 日志（*/var/log/datakit/gin.log*）。 

DataKit 默认日志等级为 `info`。编辑 `datakit.conf`，可修改日志等级以及分片大小：

```toml
[logging]
	level = "debug" # 将 info 改成 debug
	rotate = 32     # 每个日志分片为 32MB
```

- `level`：置为 `debug` 后，即可看到更多日志（目前只支持 `debug/info` 两个级别）。
- `rotate`：DataKit 默认会对日志进行分片，默认分片大小为 32MB，总共 6 个分片（1 个当前写入分片加上 5 个切割分片，分片个数尚不支持配置）。如果嫌弃 DataKit 日志占用太多磁盘空间（最多 32 x 6 = 192MB），可减少 `rotate` 大小（比如改成 4，单位为 MB）。HTTP 访问日志也按照同样的方式自动切割。

## 高级配置

下面涉及的内容涉及一些高级配置，如果对配置不是很有把握，建议咨询我们的技术专家。

### cgroup 限制 

由于 DataKit 上处理的数据量无法估计，如果不对 DataKit 消耗的资源做物理限制，将有可能消耗所在节点大量资源。这里我们可以借助 cgroup 来限制，在 *datakit.conf* 中有如下配置：

```toml
[cgroup]
	path = "/datakit" # cgroup 限制目录，如 /sys/fs/cgroup/memory/datakit, /sys/fs/cgroup/cpu/datakit

  # 允许 CPU 最大使用率（百分制）
  cpu_max = 20.0

  # 允许 CPU 最使用率（百分制）
  cpu_min = 5.0

  # 默认允许 4GB 内存(memory + swap)占用
	# 如果置为 0 或负数，则不启用内存限制
	mem_max_mb = 4096 
```

如果 DataKit 超出内存限制后，会被操作系统强制杀掉，通过命令可以看到如下结果，此时需要[手动启动服务](datakit-service-how-to#147762ed)：

```shell
$ systemctl status datakit 
● datakit.service - Collects data and upload it to DataFlux.
     Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: enabled)
     Active: activating (auto-restart) (Result: signal) since Fri 2022-02-30 16:39:25 CST; 1min 40s ago
    Process: 3474282 ExecStart=/usr/local/datakit/datakit (code=killed, signal=KILL)
   Main PID: 3474282 (code=killed, signal=KILL)
```

> 注意：
> - 目前 cgroup 限制只在[宿主机安装](datakit-install)的时候会默认开启<!--，[DaemonSet 安装可使用专用方案]() -->
> - 目前 cgourp 只支持 CPU 使用率和内存使用量（mem+swap）控制，且支持 Linux 操作系统。

### 启用磁盘缓存（Alpha）

在 DataKit 日常运行中，有简单的缓存机制，如果发送 DataWay 失败，会缓存大概 1000 个数据点。一旦超出这个点数，就会丢弃掉。

可设置磁盘缓存来保证尽可能不丢弃：

- DataKit 重启
- 因大量数据发送失败而导致的数据清空

修改 *datakit.conf* 如下配置，即可开启磁盘缓存：

```toml
[io]
	enable_cache = true
```

开启磁盘缓存后，最大能缓存 1GB 的数据（目前不可配置），超过该大小的数据，将被丢弃。

### 使用 Git 管理 DataKit 配置

由于 DataKit 各种采集器的配置都是文本类型，如果逐个修改、生效，需要耗费大量的精力。这里我们可以使用 Git 来管理这些配置，其优点如下：

- 自动从远端 Git 仓库同步最新的配置，并自动生效
- Git 自带的版本管理，能有效的追踪各种配置的变更历史

在安装 DataKit 时（[DaemonSet 安装](datakit-daemonset-deploy)和[主机安装](datakit-install#f9858758)都支持），即可指定 Git 配置仓库。

#### 手动配置 Git 管理

Datakit 支持使用 git 来管理采集器配置、Pipeline 以及 Python 脚本。在 *datakit.conf* 中，找到 *git_repos* 位置，编辑如下内容：

```toml
[git_repos]
  pull_interval = "1m" # 同步配置间隔，即 1 分钟同步一次

  [[git_repos.repo]]
    enable = false   # 不启用该 repo

    ###########################################
		# Git 地址支持的三种协议：http/git/ssh
    ###########################################
    url = "http://username:password@github.com/path/to/repository.git"

		# 以下两种协议(git/ssh)，需配置 key-path 以及 key-password
    # url = "git@github.com:path/to/repository.git"
    # url = "ssh://git@github.com:9000/path/to/repository.git"
    # ssh_private_key_path = "/Users/username/.ssh/id_rsa"
    # ssh_private_key_password = "<YOUR-PASSSWORD>"

    branch = "master" # 指定 git branch
```

注意：开启 Git 同步后，原 `conf.d` 目录下的采集器配置将不再生效（*datakit.conf* 除外）。

#### 应用 Git 管理的 Pipeline 示例

我们可以在采集器配置中，增加 Pipeline 来对相关服务的日志进行切割。在开启 Git 同步的情况下，**DataKit 自带的 Pipeline 和 Git 同步下来的 Pipeline 均可使用**。在 [Nginx 采集器](nginx)的配置中，一个 pipeline 的配置示例：

```toml
[[inputs.nginx]]
    ...
    [inputs.nginx.log]
    ...
    pipeline = "my-nginx.p" # 具体加载哪里的 my-nginx.p，参见下面的 「约束」 说明
```

#### Git 管理的使用约束

在 Git 使用过程必须遵循以下约束:

- git repo 里面新建 `conf.d` 文件夹，下面放 DataKit 采集器配置
- git repo 里面新建 `pipeline` 文件夹，下面放置 Pipeline 文件
- git repo 里面新建 `python.d` 文件夹，下面放置 Python 脚本文件

下面以图例来说明：

```
datakit 根目录
├── conf.d
├── data
├── pipeline # 顶层 Pipeline 脚本
├── python.d # 顶层 python.d 脚本
├── externals
└── gitrepos
    ├── repo-1  # 仓库 1
    │   ├── conf.d    # 专门存放采集器配置
    │   ├── pipeline  # 专门存放 pipeline 切割脚本
    │   │   └── my-nginx.p # 合法的 pipeline 脚本
    │   │   └── 123     # 不合法的 Pipeline 子目录，其下文件也不会生效
    │   │       └── some-invalid.p
    │   └── python.d    存放 python.d 脚本
    │       └── core
    └── repo-2  # 仓库 2
        ├── ...
```

查找优先级定义如下:

1. 按 *datakit.conf* 中配置的 *git_repos* 次序（它是一个数组，可配置多个 Git 仓库），逐个查找指定文件名，若找到，返回第一个。比如查找 *my-nginx.p*，如果在第一个仓库目录的 *pipeline* 下找到，则以该找到的为准，**即使第二个仓库中也有同名的 *my-nginx.p*，也不会选择它**。

2. 在 *git_repos* 中找不到的情况下，则去 *<Datakit 安装目录>/pipeline* 目录查找 Pipeline 脚本，或者去 *<Datakit 安装目录>/python.d* 目录查找 Python 脚本。

### 设置打开的文件描述符的最大值

Linux 环境下，可以在 Datakit 主配置文件中配置 ulimit 项，以设置 Datakit 的最大可打开文件数，如下：

```toml
ulimit = 64000
```

ulimit 默认配置为 64000。

## FAQ

### cgroup 设置失败

有时候启用 cgroup 会失败，在 [DataKit Monitor](datakit-monitor) 的 `Basic Info` 中会报告类似如下错误：

```
write /sys/fs/cgroup/memory/datakit/memory.limit_in_bytes: invalid argument
```

此时需手动删除已有 cgroup 规则库，然后再[重启 DataKit 服务](datakit-service-how-to#147762ed)。

```shell
sudo cgdelete memory:/datakit
```

> `cgdelete` 可能需额外安装工具包：
> 
> - Ubuntu: `apt-get install libcgroup-tools`
> - CentOS: `yum install libcgroup-tools`

### cgroup CPU 使用率说明

CPU 使用率是百分比制（==最大值 100.0==），以一个 8 核心的 CPU 为例，如果限额 `cpu_max` 为 20.0（即 20%），则 DataKit 最大的 CPU 消耗，==在 top 命令上将显示为 160% 左右==。`cpu_min` 同理。

## 延伸阅读

- [DataKit 宿主机安装](datakit-install)
- [DataKit DaemonSet 安装](datakit-daemonset-install)
- [DataKit 行协议过滤器](datakit-filter)
