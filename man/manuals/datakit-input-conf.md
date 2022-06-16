{{.CSS}}
# 采集器配置
---

- DataKit 版本：{{.Version}}
- 操作系统支持：全平台

DataKit 中采集器配置均使用 [Toml 格式](https://toml.io/cn)，所有采集器配置均位于 *conf.d* 目录下：

- Linux/Mac：`/usr/local/datakit/conf.d/`
- Windows：`C:\Program Files\datakit\conf.d\`

每个采集都分门别类，位于 *conf.d* 的下层子目录中。可参考具体的采集器配置说明，找到对应的子目录。

一个典型的配置采集器文件，其结构大概如下：

```toml
[[inputs.some_name]] # 这一行是必须的，它表明这个 toml 文件是哪一个采集器的配置
	key = value
	...

[[inputs.some_name.other_options]] # 这一行则可选，有些采集器配置有这一行，有些则没有
	key = value
	...
```

> 由于 DataKit 只会搜索 `conf.d/` 目录下以 `.conf` 为扩展的文件，故所有采集器配置==必须放在 *conf.d* 目录下（或其下层子目录下），且必须以 *.conf* 作为文件后缀==，不然 DataKit 会忽略该配置文件的处理。

## 如何修改采集器配置

目前部分采集器可以无需配置就能开启，有些则需要手动编辑配置。

### 同一个采集器开启多份采集

以 MySQL 为例，如果要配置多个不同 MySQL 采集，有两种方式：

1. 新加一个 conf 文件，比如 *mysql-2.conf*，可以将其跟已有的 *mysql.conf*  放在同一目录中。
1. 在已有的 mysql.conf 中，新增一段，如下所示：

> 不建议将多个不同的采集器配置到一个 conf 中，这可能导致一些奇怪的问题，也不便于管理。

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
  
    # 省略其它配置项...

##########################################
# 再来一个 MySQL 采集
##########################################
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项...

##########################################
# 下面继续再加一个
##########################################
[[inputs.mysql]]
	...
```

第二种方法管理起来可能更为简单，它将所有的同名采集器都用同一个 conf 管理起来了，第一种可能导致配置目录混乱。

总结一下，第二种多采集配置的结构如下：

```toml
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
```

这实际上是一个 Toml 的数组结构，==这种结构适用于所有采集器的多配置情况==。

### 关闭具体采集器

有时候，我们希望临时关闭某个采集器，也有两种方式：

1. 将对应的采集器 conf 重命名，比如 *mysql.conf* 改成 *mysql.conf.bak*，==只要保证文件后缀不是 conf 即可==
1. 在 conf 中，注释掉对应的采集配置，如

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
#    # 省略其它配置项...
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
  
    # 省略其它配置项...
``` 

相比而言，第一种方式更粗暴简单，第二种需小心修改，它可能会导致 Toml 配置错误。

### 采集器配置中的正则表达式 {#debug-regex}

在编辑采集器配置时，部分可能需要配置一些正则表达式。

由于 DataKit 绝大部分使用 Golang 开发，故涉及配置部分中所使用的正则通配，也是使用 Golang 自身的正则实现。由于不同语言的正则体系有一些差异，导致难以一次性正确的将配置写好。

这里推荐一个[在线工具来调试我们的正则通配](https://regex101.com/)。如下图所示：

![](imgs/debug-golang-regexp.png)

另外，由于 DataKit 中的配置均使用 Toml，故建议大家使用 `'''这里是一个具体的正则表达式'''` 的方式来填写正则（即正则俩边分别用三个英文单引号），这样可以避免一些复杂的转义。

## 默认开启的采集器 {#default-enabled-inputs}

DataKit 安装完成后，默认会开启一批采集器，无需手动开启。这些采集器一般跟主机相关，列表如下：

| 采集器名称                          | 说明                                                                           |
| ---------                           | ---                                                                            |
| [cpu](cpu.md)                       | 采集主机的 CPU 使用情况                                                        |
| [disk](disk.md)                     | 采集磁盘占用情况                                                               |
| [diskio](diskio.md)                 | 采集主机的磁盘 IO 情况                                                         |
| [mem](mem.md)                       | 采集主机的内存使用情况                                                         |
| [swap](swap.md)                     | 采集 Swap 内存使用情况                                                         |
| [system](system.md)                 | 采集主机操作系统负载                                                           |
| [net](net.md)                       | 采集主机网络流量情况                                                           |
| [host_processes](host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表                                      |
| [hostobject](hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等）                                 |
| [container](container.md)           | 采集主机上可能的容器或 Kubernetes 数据，假定主机上没有容器，则采集器会直接退出 |
