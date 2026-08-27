---
title     : 'WinNetFlow'
summary   : '通过 Windows ETW 采集按进程的四层网络流量指标'
tags:
  - '网络'
__int_icon      : 'icon/windows'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

WinNetFlow 采集器通过 Windows 事件追踪（ETW）监听 `Microsoft-Windows-TCPIP` Provider 的 TCP/UDP 事件，按进程统计主机四层网络流量（字节、包数、重传、RTT、TCP 状态）。核心字段和端点角色与 Linux 版 `ebpf-net/netflow` 兼容，可复用基于 `src/dst`、`client/server`、`conn_side` 和流量字段的主机网络观测页面。

同时监听 `Microsoft-Windows-HttpService` Provider 采集 HTTP 请求指标（`httpflow`，默认开启），覆盖 IIS、HttpListener、ASP.NET Core 等基于 HTTP.sys 的 Web 服务，字段与 Linux 版 `ebpf-net/httpflow` 兼容。

## 配置 {#config}

### 前置条件 {#requirements}

- 操作系统：Windows 10 / Windows Server 2016 及以上（64 位）
- 采集器需要以管理员权限运行（创建实时 ETW 会话需要管理员权限）
- 与 eBPF 采集器互不依赖，可独立启用；两者在同一主机上只会启用其一

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable MD046 -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{$m.MarkdownTable}}

{{ end }}

## 已知限制 {#limitations}

- TCP 发送事件不携带包数字段，因此 TCP 的 `packets_written` 恒为 0；UDP 使用消息数近似包数；
- UDP 无连接语义，方向通过“非临时端口上的绑定套接字”启发式判断：命中监听端口标记为 `incoming`，否则 `outgoing`；
- 采集启动前已存在的 TCP 连接通过本地 TCP 监听端口快照判断方向：本地端点命中监听端口时标记为 `incoming`，否则为 `outgoing`；监听快照每 30 秒刷新一次；
- Windows ETW 不提供网络命名空间、Kubernetes 端点元数据、DNS 域名或 NAT 转换信息；`dst_nat_ip` 和 `dst_nat_port` 固定输出 `N/A`，需要的 Kubernetes 或命名空间标签可通过单独配置的全局标签补充；
- 高流量场景下 ETW 实时会话可能丢事件，采集器每 10 分钟输出一次会话统计（解码/丢弃/解析错误/会话丢失计数），异常时会告警日志。
- 每个周期跟踪的流数有上限（默认 65536，可配置 `max_flows`），连接风暴下超出的新流会被丢弃并计入 `flows_skipped` 统计。
- `httpflow` 仅覆盖走 HTTP.sys 的 HTTP 流量（IIS / HttpListener / ASP.NET Core 等）；自包含 socket 的 HTTP 服务（如部分 Go/Node 服务）不会被捕获。
- HTTP.sys 事件不携带 HTTP 版本与请求体大小，`http_version` 恒为空、`bytes_read` 恒为 0；`bytes_written` 仅在缓存命中响应（event 16）时可得。
- HTTP 进程归因为尽力而为：连接事件中的 PID 属于客户端，因此会被忽略；采集器在 HTTP.sys 上报响应时立即解析服务端 PID。如果服务端在 Windows 允许读取名称前已退出，`process_name` 可能为 `unknown`。
- HTTP 并发请求跟踪有上限（默认 65536，可配置 `max_http_requests`），超出的新请求会被丢弃并计入周期汇总日志；URL 查询参数不会被采集，以避免泄露敏感信息和产生无界指标基数；请求路径超过 `httpflow_path_limit`（默认 256）会被截断并置 `truncated`。
