# K8s 环境下采集器配置介绍

在 k8s 环境下，由于可能存在多种采集器的配置方式，大家在配置采集器的过程中，容易混淆不同配置方式之间的差异，本文简单介绍一下 K8s 环境下配置的最佳实践。

## K8s 环境下的配置方式

目前版本(>1.2.0)的 DataKit 支持如下几种方式的配置：

- 通过 [conf]() 配置
- 通过 ENV 配置
- 通过 Annotation 配置
- 通过 [Git](datakit-conf-how-to#5dd2079e) 配置
- 通过 [DCA](dca) 配置

如果进一步归纳，又可以分为两种类型：

- 基于 DataKit 的配置
	- conf
	- ENV
	- Git
	- DCA
- 基于**被采集实体**的配置
	- Annotation

由于存在这么多不同的配置方式，不同配置方式之间还存在优先级关系，下文以优先级从低到高的顺序，开始逐个分解。

### 通过 conf 配置

DataKit 运行在 K8s 环境中时，实际上跟运行在主机上并无太大差异，它仍然会去读取 _.conf_ 目录下的采集器配置。所以，通过 ConfigMap 等方式注入采集器配置是完全可行的，有的时候，甚至是唯一的方式，比如，在当前的 DataKit 版本中，MySQL 采集器的开启，只能通过注入 ConfigMap 方式。

### 通过 ENV 配置

在 K8s 中，我们启动 DataKit 时，是可以在其 yaml 中[注入很多环境变量的](datakit-daemonset-deploy#00c8a780)。除了 DataKit 的行为可以通过注入环境变量来干预，部分采集器也支持注入**专用的环境变量**，它们命名一般如下：

```shell
ENV_INPUT_XXX_YYY
```

此处 `XXX` 指采集器名字，`YYY` 即该采集器配置中的特定配置字段，比如 `ENV_INPUT_CPU_PERCPU` 用来调整 [CPU 采集器](cpu) _是否采集每个 CPU 核心的指标_（默认情况下，该选项是默认关闭的，即不采集每个核心的 CPU 指标）

需要注意的是，目前并不是所有的采集器都支持 ENV 注入。支持 ENV 注入的采集器，一般都是[默认开启的采集器](datakit-conf-how-to#764ffbc2)。通过 ConfigMap 开启的采集器，也支持 ENV 注入的（具体看该采集器是否支持），而且**默认以 ENV 注入的为准**。

> 环境变量注入的方式，一般只应用在 K8s 模式下，主机安装方式目前无法注入环境变量。

### 通过 Annotation 配置

目前 Annotation 配置的方式支持面较之 ENV 方式更为狭窄，它主要用来**标记被采集实体**，比如_是否需要开启/关闭某实体的采集（含日志采集、指标采集等）_

通过 Annotation 来干预采集器配置的场景比较特殊，比如在容器（Pod）日志采集器中，如果禁止采集所有日志（在容器采集器中 `container_exclude_log = [image:*]`），但只希望开启特定某些 Pod 的日志采集，那么就可以在特定的这些 Pod 上追加 Annotation 加以标记：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing-log-deployment
  labels:
    app: testing-log
spec:
  template:
    metadata:
      labels:
        app: testing-log
      annotations:
        datakit/logs: |    # <-------- 此处追加特定 Key 的 Annotation
          [
            {
              "source": "testing-source",   # 设置该 Pod 日志的 source
              "service": "testing-service", # 设置该 Pod 日志的 service
              "pipeline": "test.p"          # 设置该 Pod 日志的 Pipeline
            }
          ]
	...
```

> 注意：目前 Annotation 方式还不支持主流的采集器开启（目前只支持 [Prom](prom)）。后续会增加更多采集器。

到此为止，目前 DataKit 中，主流的几种 K8s 环境下的配置方式就这三种，它们优先级逐次提升，即 conf 方式优先级最低，ENV 次之，Annotation 方式优先级最高。

### Git 配置方式

Git 方式在主机模式和 K8s 模式下均支持，它本质上是一种 conf 配置，只是它的 conf 文件不是在默认的 _conf.d_ 目录下，而是在 DataKit 安装目录的 _gitrepo_ 目录下。如果开启了 Git 模式，那么默认的 _conf.d_ 目录下的**采集器配置将不再生效**（除了 _datakit.conf_ 这个主配置之外），但原来的 _pipeline_ 目录以及 _pythond_ 目录依然有效。从这一点可以看出，Git 主要用来管理 DataKit 上的各种文本配置，包括各种采集器配置、Pipeline 脚本以及 Python 脚本。

> 注意：DataKit 主配置（_datakit.conf_）不能通过 Git 来管理。

#### Git 模式下默认采集器的配置

在 Git 模式下，有一个非常重要的特征，即那些[默认开启的采集器](datakit-conf-how-to#764ffbc2) 的 **conf 文件是隐身的**，不管是 K8s 模式还是主机模式，故将这些默认开启的采集器配置文件用 Git 管理起来，需要做一些额外的工作，不然这会导致它们被**重复采集**。

在 Git 模式下，如果要调整默认采集器的配置（不想开启或要对其做对应的配置），有几种方式：

- 可将它们从 _datakit.conf_ 或者 _datakit.yaml_ 中移除掉。**此时它们就不是默认开启的采集器了**。
-	如果要修改特定采集器的配置，有如下几种方式：
	- 将它们的 conf 通过 Git 管理起来
	- 通过上文提及的 ENV 注入（具体要看该采集器是否支持 ENV 注入）
	- 如果该采集器支持 Annotation 标记，也可以通过该方式来调整

### DCA 配置方式

[DCA](dca) 配置方式实际上跟 Git 有点类似，它们都只能影响 DataKit 上的 conf/pipeline/pythond 文件配置。只是对 DCA 而言，它的功能没有 Git 强大，一般只用于小范围管理几台 DataKit 上的文件。

## 总结

至此，DataKit 上的几种配置方式都做了基本介绍，具体采集器是否支持特定的配置方式，还需要参考采集器文档。
