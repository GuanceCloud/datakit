# DataKit 常见功能介绍

本课程系列主要介绍 DataKit 常用的一些功能，这些功能在 DataKit 日常使用和排错过程中出境的概率都比较高，有较高的学习价值。

预计这一系列课程分为如下几个主题：

1. DataKit Monitor：主要用在故障排查
1. DataKit DQL 查询：用于简便的数据查看和导出
1. DataKit 在容器环境下的采集配置
1. DataKit 第三方采集数据的接入

第一篇我们先从 DataKit Monitor 入手，这是除了采集之外，大家最常用的功能。

## DataKit Monitor 介绍

DataKit 可以简单理解为 DataKit 运行期间的一个观测面板，它可以呈现如下几个方面的数据：

- DataKit 基础信息：如 DataKit 版本、运行时长、选举状态等信息
- DataKit 资源消耗：运行期间 DataKit 自身的内存、CPU 使用情况
- 采集器运行情况：各个采集器的开启、运行情况，这也是大家最常查看的信息
- 其它各个功能模块的运行情况：包括黑名单、DataKit API、Pipeline 以及数据上传等信息展示

通过这些信息的展示，我们基本能排查 DataKit 中遇到的绝大部分问题。下面我们通过截图的方式，逐一说明如何阅读这些信息。

![](https://static.guance.com/images/teaching-school/dk-common-features/monitor.jpg)

## 基础信息面板

基础信息中，我们主要用来查看如下图几个信息：

![](https://static.guance.com/images/teaching-school/dk-common-features/basic.jpg)

每个信息的意义如下：

- `Hostname`：当前 DataKit 所在机器的主机名，如果是 K8s Daemonset 部署，那么这里是当前机器的 node-name。注意，这里的 node-name 跟主机名可能是不同的，DataKit 此处对 K8s 下的主机对象名做了优化。

- `Version`：版本号，这在排查问题时是一个最为关键的信息。有了版本号，支持工程师就能判定报告的问题是已知问题（可能已解决），还是新问题。这里也建议大家在报告问题时，无脑带上版本号信息。

- `Build`：版本打包时间，这个主要用来判断当前的版本是不是比较新的版本。

- `Branch`：当前版本的代码分支，有时候我们会发布一些测试版给用户试用，支持工程师能通过这个追到具体的代码分支，便于排查一些诸如代码奔溃（带代码文件和行号信息）类的问题。

- `Uptime`：当前 DataKit 的启动时长，有点类似于 Linux 中的 `uptime` 命令，用来判定当前 DataKit 启动时长。这个在判定 DataKit 资源是否泄露时可以作为一个参考。


- `Cgroup`：如果是 Linux 平台，主机安装时，一般都会默认开启该 Cgroup 限制，主要为了限制 DataKit 的资源消耗上限。截图中将 DataKit 的最大内存限制在 4GB，CPU 使用率控制在 5% ~ 20% 之间。

- `OS/Arch`：当前 DataKit 所在的操作系统和硬件平台。

- `Elected`：当前 DataKit 的选举状态。后面我们单独开一个章节来说明下选举的问题。

- `From`：当前 DataKit monitor 信息来源。DataKit 是可以跨机器查看其它 DataKit 的 monitor 信息的，一般情况下，命令 `datakit monitor` 是访问当前主机 9529 端口来获取 monitor 信息，也能通过 `datakit monitor --to <ip>:<port>` 来访问其它机器 DataKit 的信息。

### 选举信息

理解 DataKit 选举需要有一定的背景知识，选举开启与否、在哪选举、是否被选举上都影响具体数据的采集。此处我们详细说明一下。

#### 为什么要选举

在 DataKit 中，很大一部分的采集都是远程采集，比如采集 MySQL/Nginx/Redis 等指标，一般情况下，都不会将 DataKit 部署到这些服务所在的机器（因为它们可能是云服务）。故最常见的做法是在网络可达的前提下，在内部部署 DataKit 去采集目标服务（俗称远程采集）。

但在集群中，为了达到较为完整的可观测效果，我们的用户会在所有节点上都部署 DataKit，而且其采集配置和主配置都是一样的（如果每个 DataKit 配置都不同，难以统一管理），这就带来一个问题，如果集群内 10 个 DataKit 都配置了 MySQL 采集，那么岂不是 MySQL 指标数据要被采集十次？

为了解决重复采集的问题，观测云中心设置了一个选举机制，DataKit 启动时，如果其开启了选举，那么都会调用一下观测云中心选举接口。对中心选举接口而言，谁先调用成功，则将对应的采集权赋给谁，没有被选上的，都处于 stand-by 状态，听候调用。这就解决了大家争先恐后采集 MySQL 的问题。

除了解决重复采集的问题，选举还有一个重要的功能就是**避免采集中断**。对某个具体的 MySQL 而言，由于集群中有多个 DataKit 配置了它的采集，如果当前选上的 DataKit 因为某些原因下线，那么其它 stand-by 的 DataKit 会一拥而上，继续争夺下一轮的采集权。这样就保证了 MySQL 采集不会中断。

#### 选举信息解读

在 DataKit 主配置 *datakit.conf* 中，有这么一段来控制选举：

```toml
[election]
  enable = true
  enable_namespace_tag = false
  namespace = "vm-cluster"
  [election.tags]
```

注意到，这里有个 `namespace` 字段，这个是选举的命名空间。当用户有多个集群、且只有一个观测云工作空间时，有必要修改一下这个 `namespace` 字段。对一个观测云工作空间而言，其可以有无数个选举的命名空间，但每个命名空间，**同一时间只能有一个 DataKit 被选中**。

接下来我们看一下截图中的这段文本：

```
Elected vm-cluster::success|tan-vm
```

此处，`vm-cluster` 就是选举的 namespace，跟上面 *datakit.conf* 中所配置的一样，然后，`::` 后面表示「当前这个 DataKit 是否选举成功」。

如果我们在另一个 DataKit 上也做了如上一样的选举配置，其选举必定失败，如下截图所示（此处我们选择一个 Windows 上的 DataKit 为例）：

![](https://static.guance.com/images/teaching-school/dk-common-features/windows-election-failed.jpg)

此处的选举状态为 `defeat`，表示它当前选举失败，`|` 后面显示的是，当前这个选举 namespace 中被选中的 DataKit 的主机名（或 node-name）。

如果我们在某个 DataKit 上，发现 MySQL 一直没有采集到数据，除了检查采集配置外，一个重要的点就是查看该 DataKit 的选举状态，如果当前它选举失败，那么直接跳到选举成功（`|` 后那个主机名所在的 DataKit）的那个 DataKit 上查看即可。

如果我们将主机 `tan-vm` 上的 DataKit 停掉，稍等片刻后，这台 Windows 机器上的 DataKit 便会被选举上（前提是当前这个选举空间上只有这两个 DataKit）：

![](https://static.guance.com/images/teaching-school/dk-common-features/windows-election-ok.jpg)

## DataKit 运行资源面板

DataKit 是运行在用户环境，从用户角度而言，DataKit 所占资源越少越好。对支持工程师而言，从这个面板我们能判定当前运行的 DataKit 是否存在非预期的资源侵占（泄露）。

![](https://static.guance.com/images/teaching-school/dk-common-features/runtime.jpg)

此处用户主要关注如下三个字段：

- `CPU`：实际占用的 CPU（多核心）百分比
- `Mem`：实际占用的物理内存
- `OpenFiles`：DataKit 打开的文件数目（含磁盘文件、Socket 等）

其余的几个字段是 Golang 语言有关的指标，一般情况下，开发才需要关注这些信息。

## 采集器信息面板

采集器面板分为左右两个部分，左边是采集器开启情况，右边是采集器实际运行情况。

![](https://static.guance.com/images/teaching-school/dk-common-features/inputs.jpg)

在左边，第一列 `Inputs` 是采集器名称，也就是我们在采集器 *.conf* 配置中 `inputs` 后面的名字：

```
[inputs.cpu]
  ...
```

第二列 `Instances` 是采集器开启的个数，比如如下 MySQL 开启后，就会显示两个采集器：

```
[[inputs.mysql]]
   ...

[[inputs.mysql]]
   ...
```

第三列是采集器崩溃次数，如果此处显示为 6（如果有两个实例，则为 12，以此类推），则该采集器便处于罢工状态，需提供较为详细的日志给支持工程师排查。

### 采集器运行信息

采集器运行信息面板信息比较丰富，下面逐一说明。

![](https://static.guance.com/images/teaching-school/dk-common-features/inputs-runtime.jpg)

- 表头的 `Inputs Info(10 inputs)` 表示当前至少（可能有同名的采集器）有十个采集器有采集到数据。

- `Input` 跟左边列表类似，表示采集器名字。但某些特定的采集器，会对这个名字做一些自定义，比如，系统日志采集可能会显示成 *syslog/nginx*（此处以日志的 `source` 字段值来命名）。

- `Cat` 指数据的分类（category），最常见的如下这些类别： M（时序数据）、O（对象数据）、L（日志数据）、T（Tracing 数据）、R（RUM 数据）、N（网络等 eBPF 数据）、P（程序 Profiling 数据）、S（SCheck 数据）

- `Freq` 指数据的上报频率（分钟），一般开发才需要关注这个。

- `AvgFeed` 指每次采集/上报的数据点数，一般开发才需要关注这个。

- `Feeds` 总的上报次数，一般情况下跟采集次数正相关。

- ·TotalPts·：总的采集点数。

- `Filtered`：被黑名单过滤的点数。

- `1stFeed`：首次采集距今的时长。

- `LastFeed`：最后一次采集距今的时长，一般情况下，就靠这个来判断数据是否正在采集。如果跟预设的采集频率（比如 10s）相差甚远，则可能采集已经停止。

- `AvgCost`：每次采集的时间消耗，一般开发才需要关注这个。
- `MaxCost`：最大采集的时间消耗，一般开发才需要关注这个。
- `Error(date)`：采集过程中的报错信息（含该错误发生的时间），部分采集器如果出错（比如 MySQL 用户名密码填错或连接出错）会在该处有显示，该错误信息可以点击，点击完成后，在 Monitor 底部有红色字样展示详情。通过 ESC 或 Enter 可关闭该错误详情。

![](https://static.guance.com/images/teaching-school/dk-common-features/inputs-error.jpg)

有时候，如果采集器特别多，为便于查看，可以通过 `-I` 选项，选择只展示某些具体的采集器（以 `Input` 列中展示的名称来筛选），如：

```shell
# 只显示 logging/syslog 和 cpu 的采集情况，它们之间以英文逗号分割
datakit monitor -I logging/syslog,cpu
```

![](https://static.guance.com/images/teaching-school/dk-common-features/select-inputs.jpg)

此处表头会显示总共有 12 个采集器在采集数据，但只显示其中 2 个采集器的情况。

## 总结

本课程主要介绍了 DataKit Monitor 中主要信息面板，通过这些信息展示，我们能大致了解当前 DataKit 的运行状态，这对于排查问题有非常大的帮助。

后续我们将继续介绍 DataKit Monitor 中更多详情信息（`-V`）的解读，这些信息主要用于排查一些较为艰难的问题，它们需要一些 DataKit 的开发背景。但对于常见的问题排查，使用基本的 Monitor 输出即可。
