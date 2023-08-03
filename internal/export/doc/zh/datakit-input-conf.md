
# 采集器配置
---

## 简介 {#intro}

DataKit 中采集器配置均使用 [Toml 格式](https://toml.io/cn){:target="_blank"}。每个采集都分门别类，位于 `conf.d` 的下层子目录中：

- Linux/Mac：`/usr/local/datakit/conf.d/`
- Windows：`C:\Program Files\datakit\conf.d\`

一个典型的配置采集器文件，其结构大概如下：

```toml
[[inputs.some_name]] # 这一行是必须的，它表明这个 toml 文件是哪一个采集器的配置
    key = value
    ...

[[inputs.some_name.other_options]] # 这一行则可选，有些采集器配置有这一行，有些则没有
    key = value
    ...
```

<!-- markdownlint-disable MD046 -->
???+ tip

    由于 DataKit 只会搜索 `conf.d/` 目录下以 `.conf` 为扩展的文件，故所有采集器配置 **必须放在 `conf.d` 目录下（或其下层子目录下），且必须以 `.conf` 作为文件后缀**。否则，DataKit 会忽略处理该配置文件。
<!-- markdownlint-enable -->

## 默认开启的采集器 {#default-enabled-inputs}

DataKit 安装完成后，默认会开启一批采集器，无需手动开启。也可以按需[禁用所有默认采集器](datakit-install.md#common-envs)。

默认开启的采集器一般跟主机相关，列表如下：

| 采集器名称                                                        | 说明                                                                         |
| ---                                                               | ---                                                                          |
| [CPU（`cpu`）](../integrations/cpu.md)                            | 采集主机的 CPU 使用情况                                                        |
| [Disk（`disk`）](../integrations/disk.md)                         | 采集磁盘占用情况                                                             |
| [磁盘 IO（`diskio`）](../integrations/diskio.md)                   | 采集主机的磁盘 IO 情况                                                         |
| [内存（`mem`）](../integrations/mem.md)                           | 采集主机的内存使用情况                                                       |
| [Swap（`swap`）](../integrations/swap.md)                         | 采集 Swap 内存使用情况                                                         |
| [System（`system`）](../integrations/system.md)                   | 采集主机操作系统负载                                                         |
| [Net（`net`）](../integrations/net.md)                            | 采集主机网络流量情况                                                         |
| [主机进程（`host_processes`）](../integrations/host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表                                      |
| [主机对象（`hostobject`）](../integrations/hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等）                               |
| [容器（`container`）](../integrations/container.md)               | 采集主机上可能的容器或 Kubernetes 数据，假定主机上没有容器，则采集器会直接退出 |
| [Datakit（`dk`）](../integrations/dk.md)                          | 采集 Datakit 自身运行指标收集                                                |

### 移除默认采集器 {#disable-default-inputs}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    如果不希望开启某些采集器，可以通过在采集器名字前面加一个 `-` 的方式来屏蔽：
    
    ```toml
    # datakit.conf
    default_enabled_inputs = [
        "-cpu",
        "-disk",
        "-diskio",
        "-mem",
        "swap",
        "system",
        "hostobject",
        "net",
        "host_processes",
        "dk",
    ]
    ```
    
    同样，在安装阶段，也能通过[这种方式（`DK_DEF_INPUTS`）](datakit-install.md#common-envs)来禁用。

=== "Kubernetes"

    通过[环境变量（`ENV_ENABLE_INPUTS`）](datakit-daemonset-deploy.md#env-common)可设置要开启的采集器。
<!-- markdownlint-enable -->

## 修改采集器配置 {#modify-input-conf}

### 同一个采集器开启多份采集 {#input-multi-inst}

以 MySQL 为例，如果要配置多个不同 MySQL 采集，有两种方式：

- 方式一： 新加一个 `conf` 文件，比如 `mysql-2.conf`，可以将其跟已有的 `mysql.conf` 放在同一目录中。

- 方式二： 在已有的 `mysql.conf` 中，新增一段，如下所示：

```toml
# 第一个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...

#-----------------------------------------
# 再来一个 MySQL 采集
#-----------------------------------------
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...

#-----------------------------------------
# 下面继续再加一个
#-----------------------------------------
[[inputs.mysql]]
    ...
```

方式二多采集配置结构实际上是一个 Toml 的数组结构，**适用于所有采集器的多配置情况**，结构如下：

```toml
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 内容完全相同的两个采集器配置文件（文件名可以不同）为了防止误配置，只会应用其中一个。
    - 不建议将多个不同采集器（比如 MySQL 和 Nginx）配置到一个 `conf` 中，可能导致一些奇怪的问题，也不便于管理。
    - 部分采集器被限制为单实例运行，具体请查看 [只允许单实例运行的采集器](#input-singleton)。

???+ tip "两种方式对比"

    - 方式一，可能导致配置目录混乱。
    - 方式二，管理起来较为简单。它将所有的同名采集器，都用同一个 `conf` 管理起来。
<!-- markdownlint-enable -->

### 单实例采集器 {#input-singleton}

部分采集器只允许单实例运行，即使配置多份，也只会运行单个实例，这些单实例采集器列表如下：

| 采集器名称                                                        | 说明                                                                         |
| ---                                                               | -----------------------------------------------                              |
| [`cpu`](../integrations/cpu.md)                                   | 采集主机的 CPU 使用情况                                                        |
| [`disk`](../integrations/disk.md)                                 | 采集磁盘占用情况                                                             |
| [`diskio`](../integrations/diskio.md)                             | 采集主机的磁盘 IO 情况                                                         |
| [eBPF](../integrations/ebpf.md)                                   | 采集主机网络 TCP、UDP 连接信息，Bash 执行日志等                                 |
| [`mem`](../integrations/mem.md)                                   | 采集主机的内存使用情况                                                       |
| [`swap`](../integrations/swap.md)                                 | 采集 Swap 内存使用情况                                                         |
| [`system`](../integrations/system.md)                             | 采集主机操作系统负载                                                         |
| [`net`](../integrations/net.md)                                   | 采集主机网络流量情况                                                         |
| [`netstat`](../integrations/netstat.md)                           | 采集网络连接情况，包括 TCP/UDP 连接数、等待连接、等待处理请求等                |
| [主机进程（`host_processes`）](../integrations/host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表                                      |
| [主机对象（`hostobject`）](../integrations/hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等）                               |
| [容器（`container`）](../integrations/container.md)               | 采集主机上可能的容器或 Kubernetes 数据，假定主机上没有容器，则采集器会直接退出 |
| [Datakit（`dk`）](../integrations/dk.md)                          | 采集 Datakit 自己的运行情况，包括 CPU、Memory 等等                               |

### 关闭具体采集器 {#disable-inputs}

关闭某个具体的采集器，有两种方式：

<!-- markdownlint-disable MD046 -->
???+ tip "两种方式对比"

    - 方式一，更简单粗暴。
    - 方式二，需小心修改，可能会导致 Toml 配置错误。

- 方式一： 将对应的采集器 `conf` 重命名，比如 `mysql.conf` 改成 `mysql.conf.bak`，**只要保证文件后缀不是 `conf` 即可**。
- 方式二： 在 `conf` 中，注释掉对应的采集配置，如：

```toml

# 注释掉第一个 MySQL 采集
#[[inputs.mysql]]
#  host = "localhost"
#  user = "datakit"
#  pass = "<PASS>"
#  port = 3306
#  
#  interval = "10s"
#  
#  [inputs.mysql.log]
#    files = ["/var/log/mysql/*.log"]
#  
#  [inputs.mysql.tags]
#  
#    # 省略其它配置项 ...
#

# 保留这个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...
```
<!-- markdownlint-enable -->

### 采集器配置中的正则表达式 {#debug-regex}

在编辑采集器配置时，部分可能需要配置一些正则表达式。

由于 DataKit 绝大部分使用 Golang 开发，故涉及配置部分中所使用的正则通配，也是使用 Golang 自身的正则实现。由于不同语言的正则体系有一些差异，导致难以一次性正确的将配置写好。

这里推荐一个[在线工具来调试我们的正则通配](https://regex101.com/){:target="_blank"}。如下图所示：

<figure markdown>
  ![](https://static.guance.com/images/datakit/debug-golang-regexp.png){ width="800" }
</figure>

另外，由于 DataKit 中的配置均使用 Toml，故建议大家使用 `'''这里是一个具体的正则表达式'''` 的方式来填写正则（即正则俩边分别用三个英文单引号），这样可以避免一些复杂的转义。

## 更多阅读 {#more}

- [DataKit K8s 安装以及配置](datakit-daemonset-deploy.md)
- [通过 Git 管理采集器配置](git-config-how-to.md)
