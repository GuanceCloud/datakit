---
title     : 'eBPF Tracing'
summary   : '关联 eBPF 采集的链路 span，生成链路'
tags:
  - '链路追踪'
  - 'EBPF'
__int_icon      : 'icon/ebpf'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

## 安装 {#install}

eBPF 链路功能分为 eBPF Span(以下简称 eSpan) 的采集器和汇集并链接 eSpan 生成 Trace 的链接器。

- eSpan 的采集功能由 DataKit 中的 `ebpf` 外部采集器实现
- 数据汇集和链接功能由 DataKit ELinker/DataKit 中的 `ebpftrace` 采集器实现。

单个 `ebpftrace` 采集器接收并链接来自多个 `ebpf` 采集器的 eSpan 数据，目前数量配比必须为 **`1:N`**。

### 安装 eSpan 的采集器 {#install-ebpf-agent}

在主机或者集群部署 [DataKit](../datakit/datakit-install.md)。

### 安装 eSpan 的链接器 {#install-linker}

有主机部署和 Kubernetes 部署安装方案：

- 主机部署 DataKit 的 ELinker 版本或者 DataKit，两种方式目前将互相覆盖：
    1. 安装 [*DataKit ELinker*](../datakit/datakit-install.md#elinker-install)。该版本不含 `ebpf` 采集器。
    1. 安装 [*DataKit*](../datakit/datakit-install.md#get-install) 。该方式后续可能会废弃。

- Kubernetes 部署 DataKit ELinker：

下载 [*datakit-elinker.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit/datakit-elinker.yaml)，执行命令 `kubectl apply -f datakit-elinker.yaml`，可通过指定命名空间 `datakit-elinker`，如 `kubectl -n datakit-elinker get all -owide` 查看相关资源

*为了降低误操作造成的数据污染可能性，推荐部署 DataKit ELinker 而非 DataKit。DataKit 的 ELinker 版本相较于 DataKit 的二进制和镜像大小分别减少约 50% 和 75%。*

## 配置 {#config}

### 前置条件 {#requirements}

如果数据量在 1e6 span/min，目前需要至少提供 4C 的 cpu 资源和 4G 的 mem 资源，推荐部署在使用 SSD 硬盘的主机上。

DataKit ELinker 或 DataKit 中的 `ebpftrace` 插件用于接收和链接 eBPF span , 最终实现链路 trace_id 的生成，并建立 span 间的父子关系。

请参考以下部署模型（如下图）： 需要使所有 `ebpf` 外部采集器的 [`ebpf-trace`](./ebpf.md#ebpf-trace) 插件生成的 eBPF span 数据**发送至同一个开启 `ebpftrace` 采集器的 DataKit ELinker 或 DataKit** 上

> 如果一个服务的三个应用 App 1 ～ 3 位于两个不同的节点，`ebpftrace` 目前根据 tcp seq 等来确认进程间的网络调用关系，需要对相关 eBPF span 进行链接以此生成 trace_id 和设置 parent_id。

![img0](./imgs/tracing.png)

### DataKit ELinker/DataKit 的 `ebpftrace` 插件配置 {#ebpftrace-config}

开启 DataKit ELinker 或 DataKit 中的 `ebpftrace` 插件。

配置项：

- `db_path`:
    - 说明： 存放数据库文件的目录。
    - 环境变量： `ENV_INPUT_EBPFTRACE_DB_PATH`
- `use_app_trace_id`:
    - 说明：是否继承来自网络路径上的 DataDog/OTEL 等 agent 传播的 trace id。
    - 环境变量： `ENV_INPUT_EBPFTRACE_USE_APP_TRACE_ID`
- `window`:
    - 说明：设置 eBPF Trace Span 的链接等待时间窗口，也可视为支持的 eBPF 链路的持续时间。
    - 环境变量： `ENV_INPUT_EBPFTRACE_WINDOW`
- `sampling_rate`:
    - 说明：设置链路采样率，范围 `0.0 - 1.0`，值为 `1.0` 不采样。**默认开启 `0.1(10%)` 采样**。
    - 环境变量：`ENV_INPUT_EBPFTRACE_SAMPLING_RATE`

配置方法：

- 主机部署方案的配置：
  配置文件位于 `/var/usr/local/datakit` 目录下的 `conf.d/ebpftrace` 目录，复制  `ebpftrace.conf.sample` 并命名为 `ebpftrace.conf`。

  ```toml
  [[inputs.ebpftrace]]
    db_path = "./ebpf_spandb"
    use_app_trace_id = true
    window = "20s"
    sampling_rate = 0.1
  ```

- Kubernetes 部署方案的配置：
  修改 `datakit-elinker.yaml` 中的环境变量，必须配置的为 `dataway` 地址 `ENV_DATAWAY`，按需配置采样率 `ENV_INPUT_EBPFTRACE_WINDOW`：

  ```yaml
  - name: ENV_DATAWAY
    value: https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-WORKSPACE-TOKEN> # Fill your real Dataway server and(or) workspace token
  - name: ENV_INPUT_EBPFTRACE_WINDOW
    value: 20s # ebpf trace span link window
  - name: ENV_INPUT_EBPFTRACE_SAMPLING_RATE
    value: '0.1' # 0.0 - 1.0 (1.0 means no sampling)
  - name: ENV_INPUT_EBPFTRACE_USE_APP_TRACE_ID
    value: 'true' # true means use app trace id (from otel, datadog ...) as ebpf trace id in ebpftrace
  - name: ENV_INPUT_EBPFTRACE_DB_PATH
    value: /usr/local/datakit/ebpf_spandb/
  ```

完成设置后将 [DataKit ELinker](../datakit/datakit-install.md#elinker-install)/DataKit 或相关 K8s Service 的 `<ip>:<port>` 提供给 eBPF 采集器用于 eBPF Span 的传输。

### DataKit 的 `ebpf` 插件配置 {#ebpf-config}

配置项细节见[eBPF 采集器环境变量和配置项](./ebpf.md#input-cfg-field-env)。

开启该采集器需要在配置文件中进行以下设置：

配置项 `trace_server` 的地址填写开启了 `ebpftrace` 插件的 DataKit ELinker/DataKit 的地址

```toml
[[inputs.ebpf]]
  enabled_plugins = [
    "ebpf-net",
    "ebpf-trace",
  ]

  l7net_enabled = [
    "httpflow",
  ]

  trace_server = "x.x.x.x:9529"

  trace_all_process = false
  
  trace_env_list = [
    "DK_BPFTRACE_SERVICE",
    "DD_SERVICE",
    "OTEL_SERVICE_NAME",
  ]
  trace_env_blacklist = []
  
  trace_name_list = []
  trace_name_blacklist = [
    ## The following two processes are hard-coded to never be traced,
    ## and do not need to be set:
    ##
    # "datakit",
    # "datakit-ebpf",
  ]
```

有以下几种方法选择是否对其他进程进行链路跟踪：

- 设置 `trace_all_process` 为 `true` 跟踪所有进程，可以配合 `trace_name_blacklist` 或者 `trace_env_blacklist` 排除部分不希望采集的进程
- 设置 `trace_env_list` 对包含任意一个指定**环境变量**的进程进行跟踪。
- 设置 `trace_name_list` 对包含任意一个指定**进程名**的进程进行跟踪。

可通过为被采集进程注入以下任意一个环境变，来设置 span 的 service name：

- `DK_BPFTRACE_SERVICE`
- `DD_SERVICE`
- `OTEL_SERVICE_NAME`
