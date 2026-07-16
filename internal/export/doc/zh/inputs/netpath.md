---
title     : '网络路径'
summary   : '采集 TCP、UDP 和 ICMP 目标的网络路径与端到端质量'
tags:
  - '网络'
__int_icon      : 'icon/net'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

NetPath 采集器从 DataKit 所在节点主动探测目标的网络路径，支持 TCP、UDP 和 ICMP。目标可以通过静态配置指定，也可以由 `datakit-ebpf` 等本地流量源动态发现。NetPath 当前支持 Linux 和 macOS，暂不支持 Windows。

每次执行产生一条网络（N）分类的 `netpath` 数据，包括源和目的上下文、端到端时延、ICMP 丢包指标以及多次 traceroute 的逐跳结果。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `netpath.conf.sample` 并命名为 `netpath.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置完成后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)，也可以在 `ENV_DEFAULT_ENABLED_INPUTS` 中加入 `netpath` 后使用环境变量调整配置：

{{ CodeBlock .InputENVSampleZh 4 }}
<!-- markdownlint-enable -->

### 静态目标 {#static-targets}

通过 `[[inputs.netpath.targets]]` 配置静态目标：

```toml
[[inputs.netpath.targets]]
  name = "api-gateway"
  target = "api.example.com"
  port = 443
  protocol = "tcp"
  interval = "60s"
  timeout = "1s"
  max_ttl = 30
  traceroute_queries = 3
  e2e_queries = 10

  [inputs.netpath.targets.tags]
    service = "api"
```

协议为 `auto` 时，有端口的目标使用 TCP，没有端口的目标使用 ICMP。UDP 目前仅支持 Linux，并需要 DataKit 具备接收 raw ICMP 响应的权限。TCP/ICMP traceroute 的有效 `max_ttl` 上限为 `60`，Linux UDP traceroute 的上限为 `255`。

`traceroute_queries` 表示完整 traceroute 的执行次数。每次执行都会从 TTL 1 独立探测到目标或 `max_ttl`，并生成一个 `message.runs[]` 元素，而不是在同一个 TTL 上重试多次。目标为域名时，每个 run 独立执行 DNS 解析；实际目标 IP 写入 `runs[].destination.ip_address`，且不保证一次任务覆盖域名的全部 IP。

`e2e_queries` 表示独立端到端 probe 次数，默认 `10`。E2E 不检查中间 hop，并与 traceroute 并发执行：TCP 使用连接响应，ICMP 使用 Echo Reply，Linux UDP 使用目的端 ICMP 响应。E2E 发包使用正常的端到端 IP TTL，不使用 traceroute 的 `max_ttl`。域名会为 E2E 独立解析，选中的地址写入 `e2e_dest_ip`，可能与 `message.runs[].destination.ip_address` 不同。UDP 应用沉默记入 `e2e_unknown`，不会被直接当作丢包。

### 动态目标 {#dynamic-targets}

动态目标 API 默认为 `POST /v1/netpath/candidates`。`dynamic.enabled` 默认开启；如果 DataKit HTTP 监听地址会暴露到主机外，必须配置非空 `dynamic.token`，请求方通过 `X-Datakit-Netpath-Token` 请求头携带相同令牌。

动态候选有 hostname 时按 hostname 去重和探测，否则使用 IP，并在 `ttl` 生命周期内按 `interval` 周期执行。同一非 NAT hostname 的不同观测 IP 共享一个调度任务；发生目的地址转换时，原始目的 IP 和端口会额外参与任务身份。动态任务的 context、队列、速率和 worker 限制不会占用或丢弃已配置的静态目标。建议保持 `monitor_ip_without_domain = false`，并在扩大流量前使用 `dynamic.filters` 排除不需要探测的 namespace、网段或端口，避免产生高基数路径。

对于 hostname 候选，每次 DNS 解析后、traceroute 或 E2E 发包前，都会使用解析结果重新检查目的 host 和 CIDR 过滤规则。候选身份字段单项最长 1024 bytes；计入请求级默认值后，每个 test 的身份字段合计最长 4096 bytes。每个候选的请求级 tags 与 test 级 tags 合计最多 64 个；tag key 最长 128 bytes，value 最长 1024 bytes。固定协议字段（例如 `traceroute_status`、`traceroute_fail_type` 和端点四元组）不能被自定义 tags 覆盖。

`datakit-ebpf` 动态发现还需要在 eBPF 采集器中开启：

```toml
[inputs.ebpf]
  network_path_enabled = true
  network_path_api = "http://127.0.0.1:9529/v1/netpath/candidates"
  network_path_token = ""
```

### Reverse DNS {#reverse-dns}

Reverse DNS 默认关闭。开启后，DataKit 会对目的 IP 和有响应的 hop IP 执行 PTR 查询，并通过 TTL cache 限制重复查询：

```toml
[inputs.netpath.reverse_dns]
  enabled = true
  timeout = "500ms"
  cache_ttl = "10m"
  cache_size = 4096
```

## 数据结构 {#data}

一条结果的 tags 用于筛选任务、源端、目的端和路径，主要包括：

- 任务：`task_name`、`task_source`、`origin`、`run_type`、`protocol`；
- 四元组：`src_ip`、`src_port`、`dst_ip`、`dst_port`，发生 DNAT 时另有 `dst_nat_ip`、`dst_nat_port`；
- 端点上下文：`dst_domain`、`source_host`、`source_service`、`source_process`、`source_container_id`、`src_cloud_provider`、`dst_cloud_provider`；
- 实际探测出口：`probe_source_ip`、`probe_gateway_ip`、`probe_interface`、`probe_netns`；
- 状态：`traceroute_protocol`、`traceroute_status`、`traceroute_fail_type`、`e2e_status`。

源、目的 tags 不依赖逐跳结果。四元组命名与 NetFlow 对齐：`dst_*` 表示原始目的，`dst_nat_*` 表示实际探测使用的 DNAT 后目的；未知端口使用 `"*"`。即使探测没有产生 hop，顶层端点仍可用于检索。端点 `*_cloud_provider` 是可选云厂商信息。

NetPath 不上传 `branch_key` 或 `path_key`。历史记录应使用上述结构化的任务、源端和目的端 tags 检索；实际路由分支及其变化从 `message.runs[].hops[]` 计算。

路径完成状态使用 `traceroute_status`（`reached`、`partial` 或 `failed`）；它描述 traceroute 是否到达目的端，不表示端到端质量。

端到端质量统一使用独立的 `e2e_*` fields：

- `e2e_dest_ip`：E2E probe 独立解析并实际使用的 IPv4 地址；
- `e2e_packets_sent`、`e2e_packets_received`、`e2e_unknown`：probe 发送、响应和不确定结果数；
- `e2e_probe_loss_percent`：确定性 probe 的无响应比例，unknown 不进入分母；
- `e2e_rtt_avg`、`e2e_rtt_min`、`e2e_rtt_max`：端到端 RTT，单位微秒；
- `e2e_rtt_variation_avg`、`e2e_rtt_variation_max`：按发送序号相邻成功 probe 的 RTT 绝对差，单位微秒。

前端应使用 `e2e_rtt_avg` 展示路径整体时延，使用 `e2e_probe_loss_percent` 展示端到端 probe 无响应比例。`e2e_status` 与 `traceroute_status` 相互独立。TCP connection refused/RST 能证明目的端可达，因此计入 received；`e2e_probe_loss_percent` 不是某个中间设备的真实丢包率。

### 逐跳路径 {#message}

完整路径写入 `message` field，内容为标准 JSON：

```json
{
  "runs": [
    {
      "run_id": "1",
      "destination": {
        "ip_address": "8.8.8.8",
        "port": 443,
        "reverse_dns": ["dns.google"]
      },
      "hops": [
        {
          "ttl": 1,
          "ip_address": "10.0.0.1",
          "reverse_dns": ["gateway.local"],
          "rtt": 0.8315,
          "reachable": true
        },
        {
          "ttl": 2,
          "reachable": false
        },
        {
          "ttl": 3,
          "ip_address": "8.8.8.8",
          "rtt": 12.45,
          "reachable": true,
          "asn": 15169,
          "as_name": "GOOGLE",
          "as_prefix": "8.8.8.0/24",
          "cloud_provider": "gcp"
        }
      ]
    }
  ],
  "hop_count": {
    "avg": 3,
    "min": 3,
    "max": 3
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `runs[].run_id` | string | `message` 内一次 traceroute 的顺序 ID。 |
| `runs[].destination.ip_address` | string | 本次 run 实际探测的目标 IPv4。 |
| `runs[].destination.port` | uint16 | TCP/UDP 目标端口。 |
| `runs[].destination.reverse_dns` | string[] | 域名目标使用的 hostname。 |
| `runs[].hops[].ttl` | int | 跳序号。 |
| `runs[].hops[].reachable` | bool | 本次 TTL probe 是否收到响应。 |
| `runs[].hops[].ip_address` | string | hop IP；未响应时省略，不使用 `"*"` 占位。 |
| `runs[].hops[].reverse_dns` | string[] | 可选反向 DNS 名称。 |
| `runs[].hops[].rtt` | number | 本 TTL probe 的往返时间，单位毫秒；不是相邻 hop 之间的耗时。 |
| `runs[].hops[].asn` | uint64 | Kodo 使用本地离线数据库为公网 IP 富化的 ASN。 |
| `runs[].hops[].as_name` | string | 可选 ASN 组织名称。 |
| `runs[].hops[].as_prefix` | string | 可选 ASN 网络前缀。 |
| `runs[].hops[].cloud_provider` | string | Kodo 使用本地 IP 归属库富化的可选云厂商。 |
| `hop_count.avg/min/max` | number | 多次 traceroute 的 hop 数量统计。 |

ASN 和云厂商字段属于服务端可选富化。私网、CGNAT、未响应 hop、数据库缺失或无法匹配时不会出现对应字段。富化不会向第三方服务发送 hop IP，也不会根据 `as_name` 猜测云厂商。

如果探测在生成路径前失败，`message` 可能是错误文本而不是 JSON，应结合 `traceroute_status=failed`、`traceroute_fail_type` 和 `traceroute_fail_reason` 处理。

<!-- markdownlint-disable MD046 -->
???+ warning

    `reachable=false` 只表示本次 TTL probe 没有收到响应，可能由设备策略或 ICMP 限速导致，不能直接解释为该 hop 的真实网络丢包。逐 hop 不上传 `probe_count`、`response_count` 或 `timeout_count`。
<!-- markdownlint-enable -->

## 日志 {#logging}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## 安全和隐私 {#security}

- DataKit 会随 NetPath 结果上传 traceroute 中观测到的 hop IP，这是路径展示的基础数据。
- DataKit 不会将 hop IP 发送给外部 ASN 查询服务；ASN 在 Kodo 内使用离线数据库富化。
- Reverse DNS 在 DataKit 所在网络内发起 DNS 查询。对 DNS 暴露敏感时应保持关闭。
- 如果原始 hop IP 不允许离开客户网络，应在部署前评估是否启用 NetPath；仅将 ASN 富化迁移到 Kodo 不能隐藏上传给观测平台的原始 IP。

## 故障排查 {#troubleshooting}

| 现象 | 检查项 |
| --- | --- |
| Candidate API 返回 HTTP 401 | 请求令牌必须与 `dynamic.token` 一致。 |
| Candidate API 返回 HTTP 404 | 确认 `dynamic.enabled` 未关闭。 |
| `ip_without_domain` | 提供 hostname，或明确开启 `monitor_ip_without_domain`。 |
| `traceroute_fail_type=permission` | 提供 ICMP/UDP traceroute 所需的 raw socket 权限。 |
| 路径只有部分 hops | 检查 `traceroute_status=partial`；中间设备可能不响应 TTL probe。 |
| 没有 ASN | 确认 Kodo 已部署 ASN/ISP MMDB；私网和未响应 hop 不会有 ASN。 |
