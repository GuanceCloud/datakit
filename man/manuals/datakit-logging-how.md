# DataKit 日志处理原理介绍

本篇用来介绍 DataKit 如如何处理日志的。在[另一篇文档](datakit-logging)中，我们介绍了 DataKit 是如何采集日志的，这两篇文档，可结合起来看，希望大家对整个日志处理有更全面的认识。

核心问题：

- 日志的采集为什么这么复杂
- 日志数据是如何处理的

## 为什么日志采集的配置这么复杂

我们从[这篇文档](datakit-logging)中可得知，因为日志来源多种多样，导致日志的配置方式多种多样，我们有必要在此做一番梳理，便于大家理解。

在日志的采集过程中，DataKit 有主动、被动两大类采集方式：

- 主动采集
	- 直接采集[磁盘文件日志](logging)
	- 采集[容器](container)产生的日志

- 被动采集
  - 通过 [HTTP](logstreaming)、[TCP/UDP](logging#7306c0d5) 以及 [Websocket](logfwd) 给 DataKit 注入日志

在这些不同形式的日志采集方式中，它们都要解决同一个核心问题：**DataKit 接下来如何处理这些日志？**

将这个核心问题再拆解一下，可细分为如下几个子问题：

- 确定 `source` 是什么：后续所有的日志处理，都依赖这个字段（额外还有一个 `service` 字段，但如果不指定，其值将设置成 source 值一样）
- Pipeline 如何配置：虽然不是必须配置，但应用很广
- 额外 Tag 配置：也不是必须配置，但某些时候，有其特殊功用
- 多行日志如何切割：需告诉 DataKit，目标日志是如何分割每条日志的（DataKit 默认以非空白字符开头的每一行为新日志）
- 是否有特殊的忽略策略：不是 DataKit 采集到的所有数据都需要处理，符合一定条件的话，可以选择不采集它们（虽然它们符合采集条件）
- 其它特色配置：如过滤颜色字符、文本编码处理等

目前，有如下几种方式来告诉 DataKit 如何处理拿到的日志：

- [日志采集器](logging) 中的 conf 配置

日志采集器中，[通过 conf 配置](logging#224e2ccd)要采集的文件列表（或者从哪个 TCP/UDP 端口读取日志流），在 conf 中可配置 source/Pipeline/多行切割/额外添加 tag 等多种设置。

如果是以 TCP/UDP 形式将数据发送给 DataKit，也只能通过 logging.conf 来配置后续的日志处理，因为 TCP/UDP 这种协议不便于附加额外的描述信息，它们只负责传送简单的日志流数据。

这种形式的日志采集，是最易于理解的一种方式。

- [容器采集器](container)中的 conf 配置

目前容器采集器 conf 针对日志，只能做最粗浅的配置（基于容器/Pod 镜像名），无法在此配置日志的后续处理（如 Pipeline/source 设置等），因为这个 conf 针对的是**当前主机上所有日志**的采集，而在容器环境下，一个主机上的日志多种多样，无法在此分门别类逐个配置。

- 通过在请求中告知 DataKit 如何配置日志处理

通过 HTTP 请求 DataKit 的 [logstreaming](logstreaming) 服务，在请求中带上各种请求参数，以告知 DataKit 如何处理收到的日志数据。 

- 在被采集对象（比如容器/Pod）上做特定的标注，告知 DataKit 如何处理它们产生的日志

前面提到，单纯在容器采集器 conf 中配置日志采集，因为细粒度太粗，不利于精细配置，但可以在[容器/Pod 上打上标注](container#f3cb35b8)，DataKit 会**主动去发现这些标注**，进而就知道每个容器/Pod 日志要如何处理了。

### 优先级说明

一般情况下，容器/Pod 上的标注优先级最高，它会覆盖 conf/Env 上的设置；其次 Env 优先级居中，它会覆盖 conf 中的配置；conf 中的配置，优先级最低，其中的配置随时可能被 Env 或标注中的设定覆盖。

> 目前尚没有直接跟日志采集/处理相关的 Env，后续可能会增加相关环境变量。

下面举个列子，在 container.conf 中，假定我们将名为 'my_test' 的镜像排除在日志采集之外：

```toml
container_exclude_log = ["image:my_test*"]
```

此时，DataKit 就不会采集所有通配该镜像名的容器或 Pod。但如果对应的 Pod 上做了对应 Annotation 标注：

```yaml
apiVersion: apps/v1
kind: Pod
metadata:
  name: test-app
  annotations:
    datakit/logs: |   # <----------
      [
        {
          "source": "my-testing-app",
          "pipeline": "test.p",
        }
      ]

spec:
   containers:
   - name : mytest
     image: my_test:1.2.3
```

即使在 container.conf 中我们排除了所有通配 `my_test.*` 的镜像，但因为该 Pod 带有特定的标注（`datakit/logs`），DataKit 仍然会采集该 Pod 的日志，并且可配置 Pipeline 等诸多设置。

## 日志数据是如何处理的

在 DataKit 中，日志目前要经过如下几个阶段的处理（按照处理顺序列举）：

- 采集阶段

从外部读取（接收）到日志后，采集阶段会进行基本的处理。这些处理包括日志分割（将大段文本分成多条独立的裸日志）、编解码（统一转成 UTF8 编码）、剔除一些干扰性的颜色字符等

- 单条日志切割 

如果对应的日志有配置 Pipeline 切割，那么每一条日志（含单条多行日志）都会通过 Pipeline 切割，Pipeline 主要又分为两个步骤：

  1. Grok/Json 切割：通过 Grok/Json，将单条 Raw 日志切割成结构化数据
	1. 对提取出来的字段，再精细处理：比如[补全 IP 信息](pipeline#9b1bba32)，[日志脱敏](pipeline#52a4c41c)等

- 黑名单（Filter）

[Filter 是一组过滤器](datakit-monitor)，它接收一组结构化数据，通过一定的逻辑判断，决定数据是否丢弃。Filter 是中心下发（DataKit 主动拉取）的一组逻辑运算规则，其形式大概如下：

```
{ source = 'datakit' AND bar IN [ 1, 2, 3] }
```

以日志为例，假定切割出来的 100 条日志中，有 10 条满足这里的条件（即 source 为 `datakit`，且字段 `bar` 字段的值出现在后面的列表中），那么这 10 条日志将不会上报到观测云，被默默丢弃。但在 [DataKit Monitor](datakit-monitor) 中能看到丢弃的统计情况。

- 上报观测云

经过上述这些步骤后，日志数据最终上报给观测云，在日志查看页面，即可看到日志数据。一般情况下，从日志产生，如果采集成功，到页面能看到数据，期间有 30s 左右的延迟，这期间，DataKit 本身数据也是最大 10s 才上报一次，中心也要经过一系列处理才最终入库。

## 延申阅读

- [DataKit 日志采集综述](datakit-logging)
- [如何调试 Pipeline](datakit-pl-how-to)
- [行协议黑名过滤器](datakit-filter)
