---
title     : 'NFS'
summary   : 'NFS 指标采集'
tags:
  - '主机'
__int_icon      : 'icon/nfs'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor:
  - desc: '暂无'
    path: '-'
---

{{.AvailableArchs}}

---

NFS 指标采集器，采集以下数据：

- RPC 吞吐量指标
- NFS 挂载点指标（仅支持 NFSv3 和 v4）
- NFSd 吞吐量指标

## 配置 {#config}

### 前置条件 {#requirements}

- NFS 客户端环境正确配置
- NFS 客户端正确挂载至服务器的共享目录

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

### NFSd 开启 {#nfsd}

NFSd 是 NFS 服务的守护进程，是服务器端的一个关键组件，负责处理客户端发送的 NFS 请求。如果本地机器同时作为 NFS 服务器，则可开启该指标查看网络、磁盘 I/O、用户处理 NFS 请求的线程等统计信息。

如需开启，则需修改配置文件。

```toml
[[inputs.nfs]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ## 是否开启 NFSd 指标采集
  nfsd = true

...

```

### NFS 挂载点详细统计信息开启 {#nfs-mountstats}

默认开启的 nfs_mountstats 指标集仅展示挂载点磁盘用量以及 NFS 运行时间的统计信息，如需查看 NFS 挂载点的 R/W、Transport、Event、Operations 等信息则需要修改配置文件。

```toml
[[inputs.nfs]]
  
  ...

  ## NFS 挂载点指标配置
  [inputs.nfs.mountstats]
    ## 开启 R/W 统计信息
    # rw = true
    ## 开启传输统计信息 
    # transport = true
    ## 开启事件统计信息
    # event = true
    ## 开启操作统计信息
    # operations = true

...

```

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
