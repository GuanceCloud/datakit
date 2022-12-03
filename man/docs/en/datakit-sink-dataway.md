<!-- This file required to translate to EN. -->
# Dataway
---

如果希望将数据打到不同的工作空间，可以使用 Dataway Sink 功能：

1. 在 DataKit 中，可以配置多个 dataway sink 地址，这些 dataway 一般只是 token 不同。除此之外，每个 sink 地址上可以附加一个或多个数据判定条件
1. 对满足条件的数据（一般通过判断 tag/field 上的值），就打给对应的 dataway
1. 如果数据不满足所有判定条件，数据继续会打到默认的 dataway 上

目前 dataway sink 支持所有种类数据（M/O/L/T/...）。

???+ attention

    如果多个 dataway sink 的判定条件之间存在交集，对同时满足多个判定条件的数据，它们会分别写入对应的工作空间，可能造成一定的数据重复。

## Dataway Sink 配置 {#config}

- 第一步: 搭建后端存储

使用[观测云](https://console.guance.com/)的 Dataway, 或者自己搭建一个 Dataway 的 server 环境。

- 第二步: 增加配置

=== "datakit.conf"

    在 `datakit.conf` 中增加以下片段:
    
    ```toml
    ...
    [sinks]
       [[sinks.sink]]
         categories = ["M"]
         filters = ["{host='user-ubuntu'}", "{cpu='cpu-total'}"]
         target = "dataway"
         token = <YOUR-TOKEN1>
         url = "https://openway.guance.com"
 
       [[sinks.sink]]
         categories = ["M"]
         filters = ["{cpu='cpu-total'}"]
         target = "dataway"
         token = <YOUR-TOKEN2>
         url = "https://openway.guance.com"
    ...
    ```
    
    除了 Sink 必须配置[通用参数](datakit-sink-guide.md)外, Dataway 的 Sink 实例目前支持以下参数:
    
    - `url`(必须): 这里填写 dataway 的全地址(带 token)。
    - `token`(可选): 工作空间的 token。如果在 `url` 里面写了这里就可以不用填。
    - `filters`(可选): 过滤规则。类似于 io 的 `filters`, 但功能是截然相反的。sink 里面的 filters 匹配满足了才写数据; io 里面的 filters 匹配满足了则丢弃数据。前者是 `include` 后者是 `exclude`。
    - `proxy`(可选): 代理地址, 如 `127.0.0.1:1080`。

=== "Kubernetes"

    Kubernetes 中可以通过环境变量来配置 dataway sink，参见[这里](datakit-daemonset-deploy.md#env-sinker)。

- 第三步: [重启 DataKit](datakit-service-how-to.md#manage-service)

## 安装阶段设置 {#dw-setup}

Dataway Sink 支持安装时，通过环境变量的方式来设置：

```shell
DK_SINK_M="dataway://?url=https://openway.guance.com&token=<YOUR-TOKEN>&filters={host='user-ubuntu'}&filters={cpu='cpu-total'}" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

### Sink 多实例配置 {#multi-dw-sink}

针对单个数据类型，如果要指定多个 dataway sink 配置，可以以 `||` 来分割：

```shell
DK_SINK_M="dataway://?url=https://openway.guance.com&token=<TOKEN-1>&filters={host='user-ubuntu'}||dataway://?url=https://openway.guance.com&token=<TOKEN-2>&filters={host='user-centos'}" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

这里的意思是：

> 所有的时序数据（`M`），如果主机名（`host`）是 `user-ubuntu` 的，就打给 token 为 `TOKEN-1` 的空间，如果主机名是 `user-centos` 就打给 `TOKEN-2` 对应的工作空间。

以此类推，其它数据类型（如 L/O/T/...） 均可以做对应的设置。

## 延申阅读 {#more-readings}

- [Filter 写法](datakit-filter.md#howto)
