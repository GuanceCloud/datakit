
# Datakit Sink
---

:warning: Datakit 侧的 Sinker 功能已经在[:octicons-tag-24: Version-1.13.0](changelog.md#cl-1.13.0)中废弃，请使用 [Dataway 端的 Sinker 方案](dataway-sink.md)。

---

如果希望将数据打到不同的工作空间，可以使用 Dataway Sinker 功能：

1. 在 DataKit 中，可以配置多个 Dataway Sinker 地址，这些 Dataway 一般只是 token 不同。除此之外，每个 Sinker 地址上可以附加一个或多个数据判定条件
1. 对满足条件的数据（一般通过判断 tag/field 上的 key-value 值），即将数据上传到对应的 Dataway
1. 如果数据不满足所有判定条件，数据继续会上传到默认的 Dataway 上

<!-- markdownlint-disable MD046 -->
???+ attention

    如果多个 Dataway Sinker 的判定条件之间存在交集，对同时满足多个判定条件的数据，它们会分别写入对应的工作空间，可能造成一定的数据重复。
<!-- markdownlint-enable -->

## Sinker 支持的数据类型 {#categories}

目前 Dataway Sinker 支持[所有种类类型](apis.md#category)。

## Dataway Sinker 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "*datakit.conf*"

    在 *datakit.conf* 的 `dataway` 下面增加以下片段（参考 Datakit [主配置模板](datakit-conf.md#maincfg-example)）:
    
    ```toml
    [[dataway.sinkers]]
      categories = ["M", "O"] # 此处还可以指定更多数据类型
      filters = [
        "{ cpu = 'cpu-total' }",
        "{ source = 'some-logging-source'}",
      ]
      url = "https://openway.guance.com?token=<YOUR-TOKEN>"
    
    [[dataway.sinkers]]
      another sinker...
      ...
    ```
    
    Dataway 的 Sinker 支持以下参数（`*` 为必填参数）:
    
    - `* categories`：表示 sinker 的数据类型，完整的数据类型列表参见[这里](apis.md#category)。
    - `* url`: 这里填写 dataway 的全地址(带 token)，如 `https://openway.guance.com?token=tkn_xxx`。
    - `filters`: 过滤规则。其配置规则跟 [行协议过滤器](datakit-filter.md) 一样。如果不配置任何规则，则表示无条件 sinker 过去。
    - `proxy`: HTTP 代理地址（IP:Port），形如 1.2.3.4:5678。

    配置完 Dataway Sinker 后，[重启 DataKit](datakit-service-how-to.md#manage-service)。

=== "Kubernetes"

    Kubernetes 中可以通过环境变量来配置 Sinker，参见[这里](datakit-daemonset-deploy.md#env-sinker)。

???+ attention

    虽然 Dataway 有[磁盘缓存](datakit-conf.md#io-disk-cache)功能，但 Dataway 上的 Sinker 暂时不具备这个功能，如果 Sinker 发送 Dataway 失败，那么数据就丢失了。
<!-- markdownlint-enable -->

## 延申阅读 {#more-readings}

- [Filter 写法](datakit-filter.md#howto)
- [主机安装时指定 Dataway Sinker](datakit-install.md#env-sink)
