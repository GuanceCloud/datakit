# NetPath 前端数据对接说明

## 1. 文档目的

本文描述 DataKit NetPath 采集器上传并由 Kodo 富化后的数据结构，供前端进行路径列表、路径详情、拓扑和质量指标展示。

一条 NetPath 结果存储为一条网络（N 类）数据：

- 数据分类：`network`（简称 `N`）
- measurement/source：`netpath`
- Point 时间：本次探测的完成时间
- tags：任务、源端、目的端和路径上下文，适合筛选与分组
- fields：探测结果、时延、丢包和标识字段
- `message`：完整 traceroute 路径 JSON；探测无法生成路径时可能是错误文本

字段使用本文定义的规范名称。前端必须忽略不认识的字段，并将本文标记为“可选”的字段按缺失处理。

当前契约以 Dataway 实际收到的数据为准：顶层端点统一使用 NetFlow 风格的 `src_*`、`dst_*`，路径状态统一使用 `traceroute_*`。旧兼容字段不会继续上传，详见“旧字段迁移”。

## 2. 前端推荐处理流程

1. 使用 `traceroute_status` 判断逐跳探测状态：`reached`、`partial` 或 `failed`。
2. 使用 `traceroute_fail_type` 和 `traceroute_fail_reason` 展示失败原因。
3. 当 `message` 是 JSON 时，解析 `runs` 并展示逐跳路径；不能解析为 JSON 时，将其作为错误文本。
4. 使用 `e2e_*` 字段展示统一端到端质量；`e2e_status` 与 `traceroute_status` 相互独立。
5. 对公网 hop 展示 Kodo 富化的 ASN 和云厂商信息；富化字段缺失是正常情况。

## 3. Tags

存储层中 tags 和 fields 通常会以同级字段返回，但 tags 适合用于检索、筛选和分组。

### 3.1 任务和执行上下文

| 字段 | 类型 | 是否稳定存在 | 说明 |
| --- | --- | --- | --- |
| `path_key` | string | 是 | 稳定的逻辑探测路径标识，用于列表去重和关联历史记录。 |
| `task_name` | string | 是 | 探测任务名称。 |
| `task_source` | string | 是 | 任务来源。当前为 `local` 或 `dynamic`；`server` 为预留值。 |
| `origin` | string | 是 | 原始候选来源，例如 `local`、`ebpf_netflow`。 |
| `run_type` | string | 是 | 执行类型。当前为 `scheduled` 或 `dynamic`；`on_demand` 为预留值。 |
| `protocol` | string | 是 | 探测协议：`tcp`、`udp` 或 `icmp`。 |
| `traceroute_protocol` | string | 是 | 生成逐跳路径时实际使用的协议：`tcp`、`udp` 或 `icmp`。TCP 使用 TCP SYN traceroute。 |
| `traceroute_status` | string | 是 | `reached`、`partial` 或 `failed`。 |
| `traceroute_fail_type` | string | 失败时 | 标准化失败分类，详见“失败处理”。 |
| `e2e_status` | string | E2E 开启时 | `reached`、`partial`、`unknown` 或 `failed`，与 traceroute 状态相互独立。 |

### 3.2 流量四元组和目的端

| 字段 | 类型 | 是否稳定存在 | 说明 |
| --- | --- | --- | --- |
| `src_ip` | string | 通常 | 原始流量源 IP；静态任务未提供时回退到 `probe_source_ip`。源地址和路由都无法确定时缺失。 |
| `src_port` | string | 是 | 原始流量源端口，未知时为 `"*"`。 |
| `dst_ip` | string | 通常 | 原始流量目的 IP；与 NetFlow 语义一致。域名在失败前未解析出 IP 时可能缺失。 |
| `dst_port` | string | 是 | 原始流量目的端口，未知时为 `"*"`；与 NetFlow 语义一致。 |
| `dst_domain` | string | 可选 | 配置或动态发现的目的域名。 |
| `dst_nat_ip` | string | 发生 DNAT 时 | 流量源观察到的转换后目的 IP。每个 run 的实际探测 IP 以 `runs[].destination.ip_address` 为准。 |
| `dst_nat_port` | string | 发生 DNAT 时 | 流量源观察到的转换后目的端口。 |
| `dst_cloud_provider` | string | Kodo 可选富化 | `dst_ip` 对应的云厂商，适合顶层检索和分组。 |

`dst_*` 始终表示流量源看到的原始目的；只有候选流量发生目的转换时才提供 `dst_nat_*`。域名重新解析到不同 IP 不视为 DNAT。前端不要用 `dst_nat_*` 覆盖原始四元组。

DNAT 示例：原始连接为 `10.0.0.10:53000 -> 203.0.113.10:443`，实际探测目标为 `198.51.100.20:8443` 时，顶层 tags 为：

```json
{
  "src_ip": "10.0.0.10",
  "src_port": "53000",
  "dst_ip": "203.0.113.10",
  "dst_port": "443",
  "dst_nat_ip": "198.51.100.20",
  "dst_nat_port": "8443"
}
```

### 3.3 源端和网络上下文

| 字段 | 类型 | 是否稳定存在 | 说明 |
| --- | --- | --- | --- |
| `source_host` | string | 通常 | 发现该目标的源主机；动态候选未提供时回退到 DataKit 全局 `host`。 |
| `src_cloud_provider` | string | 可选 | 源端云厂商。DataKit 可从全局 `cloud_provider` 补齐，Kodo 也可按源 IP 富化。 |
| `source_service` | string | 动态候选可选 | 源服务名称。 |
| `source_process` | string | 动态候选可选 | 源进程名称。 |
| `source_container_id` | string | 动态候选可选 | 源容器 ID。 |
| `namespace` | string | 可选 | 候选来源的逻辑命名空间。 |
| `netns` | string | Linux 动态候选可选 | 候选流量所在的 Linux network namespace。 |
| `probe_source_ip` | string | Linux 路由查询成功时 | 主动探测实际使用的源 IP。 |
| `probe_gateway_ip` | string | Linux 路由查询成功时 | 主动探测实际使用的下一跳网关。 |
| `probe_interface` | string | Linux 路由查询成功时 | 主动探测使用的出口网卡。 |
| `probe_interface_mac` | string | Linux 路由查询成功时 | 出口网卡 MAC。 |
| `probe_netns` | string | Linux 路由查询成功时 | 主动探测实际执行所在的 network namespace。 |

`src_*`/`netns` 描述发现候选的原始流量；`probe_*` 描述本次主动探测实际选择的路由。前端不要将二者混为同一含义。

### 3.4 历史检索和自定义 Tags

`path_key` 用于列表去重和关联同一逻辑探测路径的历史记录。DataKit 将版本、`source_host` 和内部调度身份进行长度前缀编码，再对其计算 SHA-256，取前 128 bit，格式为 `np-v1-<32 位十六进制>`。静态任务的调度身份包含任务名、目的地址、端口和协议；动态任务还包含源实体、namespace、netns 和 DNAT 上下文。同一任务的执行时间、状态、延迟或 traceroute hop 变化不会改变 `path_key`。

`path_key` 不表示某次执行，也不表示实际 hop 分支。单次执行使用 `test_run_id`；实际路由分支及其变化从 `message.runs[].hops[]` 计算。需要按条件检索历史时，仍可结合 `task_source`、`origin`、`protocol`、`namespace`、`netns`、`source_host`、`source_service`、`source_container_id`、`dst_domain`、`dst_ip` 和 `dst_port`，发生 DNAT 时再结合 `dst_nat_ip`、`dst_nat_port`。

DataKit 全局 host tags、NetPath 配置 tags 和动态候选 tags 也会随记录上传。固定契约字段不能被自定义 tags 覆盖；`path_key` 由 DataKit 生成，自定义同名 tag 会被过滤；`branch_key` 不上传。全局 `cloud_provider` 会规范化为 `src_cloud_provider`，不再重复上传。其他扩展 tags 不属于固定协议，前端可以作为扩展筛选维度展示。

源、目的 tags 独立于 `message.runs[].hops[]`。即使 traceroute 没有产生 hop 或探测提前失败，前端仍应使用这些顶层 tags 检索、筛选和展示端点；不要从第一跳或最后一跳反推源、目的信息。

## 4. 顶层 Fields

### 4.1 唯一标识和时间

| 字段 | 类型 | 单位 | 说明 |
| --- | --- | --- | --- |
| `test_run_id` | string | - | 一次任务执行 ID。一次执行内保持不变。 |
| `scheduled_at` | int64 | Unix 微秒 | 计划执行时间。 |
| `started_at` | int64 | Unix 微秒 | 实际开始时间。 |
| `duration` | int64 | 微秒 | 整次探测耗时。 |

JavaScript 构造 `Date` 前应将 Unix 微秒除以 `1000`。当前 Unix 微秒仍在 JavaScript 安全整数范围内，但接口层优先保留为整数或字符串，避免后续精度问题。

### 4.2 状态和路径摘要

| 字段 | 类型 | 取值/单位 | 说明 |
| --- | --- | --- | --- |
| `hop_count` | int64 | 跳 | 当前路径的 TTL/hop 数摘要。完整多次结果以 `message.hop_count` 为准。 |
| `max_ttl` | int64 | 跳 | 应用协议上限后，本次 traceroute 实际使用的最大 TTL。 |
| `traceroute_queries` | int64 | 次 | 完整 traceroute 的执行次数，决定 `message.runs` 的数量。每个 run 从 TTL 1 独立探测到目标或 `max_ttl`，每个 TTL 发送一次 probe。 |
| `dst_reverse_dns` | string | - | 开启 reverse DNS 且解析成功时，顶层 `dst_nat_ip` 或 `dst_ip` 的反向域名。每个 run 的目标域名使用 `runs[].destination.reverse_dns`。 |
| `traceroute_fail_reason` | string | - | 原始 traceroute 失败原因。成功记录通常不存在。 |
| `message` | string/JSON | - | 成功或部分成功时为标准路径 JSON；无路径时可能为错误文本。 |

路径完成状态只看 `traceroute_status` tag，不再上传 `success`、`path_success` 或 `path_destination_reached` 等重复 fields。

### 4.3 独立 E2E 质量字段

E2E probe 与 traceroute 并发执行，不读取或推导中间 hop。TCP 使用连接响应，ICMP 使用 Echo Reply；Linux UDP 使用目的端 ICMP 响应。E2E 使用正常的端到端 IP TTL，不受 traceroute `max_ttl` 限制。TCP connection refused 证明目的端可达，因此计入 received；UDP 应用沉默无法区分开放端口和网络丢包，因此计入 unknown 而不是 loss。

| 字段 | 类型 | 单位 | 说明 |
| --- | --- | --- | --- |
| `e2e_dest_ip` | string | - | E2E probe 独立解析并实际使用的 IPv4；可能与任一 `message.runs[].destination.ip_address` 不同。 |
| `e2e_queries` | int64 | 次 | 配置的独立 E2E probe 数量，默认 `10`。 |
| `e2e_packets_sent` | int64 | 个 | 实际发出的 probe 数量。 |
| `e2e_packets_received` | int64 | 个 | 收到可识别目的端响应的 probe 数量。 |
| `e2e_unknown` | int64 | 个 | 结果不确定的 probe 数量，主要为开放 UDP 应用沉默。 |
| `e2e_probe_loss_percent` | float64 | `%` | 确定性 probe 中未收到响应的比例，范围 `0..100`；unknown 不进入分母。 |
| `e2e_rtt_avg` | float64 | 微秒 | 有效响应的平均 RTT。 |
| `e2e_rtt_min` | float64 | 微秒 | 有效响应的最小 RTT。 |
| `e2e_rtt_max` | float64 | 微秒 | 有效响应的最大 RTT。 |
| `e2e_rtt_variation_samples` | int64 | 对 | 用于 RTT variation 的连续成功 probe 对数。 |
| `e2e_rtt_variation_avg` | float64 | 微秒 | 按发送序号相邻的成功 probe 之间 RTT 绝对差的平均值。 |
| `e2e_rtt_variation_max` | float64 | 微秒 | 按发送序号相邻的成功 probe 之间 RTT 绝对差的最大值。 |
| `e2e_tcp_connection_refused` | int64 | 个 | TCP 收到 connection refused/RST 的次数；这些 probe 仍计入 received。 |
| `e2e_fail_reason` | string | - | E2E 初始化或执行错误，不改变 traceroute 主状态。 |

`e2e_probe_loss_percent` 是特定协议 probe 未收到可识别响应的比例，不应表述为链路中某个设备的真实丢包率。`e2e_rtt_variation_*` 是基于 RTT 的变化量，不是需要源、目的时钟同步的单向 IPDV。

### 4.4 前端质量字段使用规则

端到端质量只使用 `e2e_*` 字段，不上传 `path_latency`、`path_rtt_*`、`path_packet_loss_percent` 或 `path_packets_*` 等重复字段：

- 整体时延使用 `e2e_rtt_avg`，范围使用 `e2e_rtt_min` 和 `e2e_rtt_max`；
- 无响应比例使用 `e2e_probe_loss_percent`，同时展示 `e2e_packets_sent` 和 `e2e_packets_received` 作为样本量；
- RTT 变化使用 `e2e_rtt_variation_avg` 和 `e2e_rtt_variation_max`；
- traceroute 路径完成状态只使用 `traceroute_status`，不要从 E2E fields 反推。

前端展示毫秒时统一执行 `microseconds / 1000`。不要根据某一跳没有响应直接计算该路由器的真实丢包率；中间设备可能对 ICMP/UDP 超时报文限速或不响应。

### 4.5 动态候选扩展 Fields

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `source_pid` | int64 | 候选来源进程 PID。 |

动态候选不转发任意 `metadata` fields。固定结构之外只允许自定义 tags，前端仍应忽略不认识的扩展字段。

## 5. `message` 路径 JSON

### 5.1 完整结构

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
          "reverse_dns": ["gateway.example"],
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
          "reverse_dns": ["dns.google"],
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

### 5.2 Runs

| 字段 | 类型 | 是否必有 | 说明 |
| --- | --- | --- | --- |
| `runs` | array | 是 | 多次 traceroute 结果。可能为空。 |
| `runs[].run_id` | string | 是 | message 内部的 traceroute 次序 ID，例如 `1`、`2`。不要与顶层 `test_run_id` 混淆。 |
| `runs[].destination.ip_address` | string | 新数据必有 | 本次 run 实际探测的目标 IPv4。域名有多个 IP 时，不同 runs 可能不同。 |
| `runs[].destination.port` | uint16 | TCP/UDP 目标 | 本次 run 的目标端口。 |
| `runs[].destination.reverse_dns` | string[] | 域名目标 | 用于发起探测的目标域名。 |
| `runs[].hops` | array | 是 | 本次 traceroute 的逐 TTL 结果，按 `ttl` 升序展示。 |
| `hop_count.avg` | number | 是 | 各 run hop 数量平均值。 |
| `hop_count.min` | int | 是 | 各 run hop 数量最小值。 |
| `hop_count.max` | int | 是 | 各 run hop 数量最大值。 |

### 5.3 Hop

| 字段 | 类型 | 是否必有 | 来源 | 说明 |
| --- | --- | --- | --- | --- |
| `ttl` | int | 是 | DataKit | 跳序号，从 `1` 开始。 |
| `reachable` | bool | 是 | DataKit | 本次 TTL probe 是否收到响应。 |
| `ip_address` | string | 可选 | DataKit | 响应 hop 的 IP。`reachable=false` 时省略，不使用 `"*"` 占位。 |
| `reverse_dns` | string[] | 可选 | DataKit | hop 的反向 DNS 名称。可能为空或缺失。 |
| `rtt` | number | 可选 | DataKit | 本 TTL probe 从发出到收到响应的往返时间，单位毫秒；不是相邻 hop 之间的耗时。未响应时省略。 |
| `asn` | uint64 | 可选 | Kodo | 公网 IP 对应的自治系统编号。 |
| `as_name` | string | 可选 | Kodo | ASN 组织名称。 |
| `as_prefix` | string | 可选 | Kodo | ASN 数据库匹配到的网络前缀。 |
| `cloud_provider` | string | 可选 | Kodo | IP 归属库匹配到的云厂商；无法可靠匹配时缺失。 |

#### RTT 展示语义

`rtt` 是探测源到当前响应 hop 的往返时间，不是上一跳到当前 hop 的链路时延。前端节点可将其标记为“源到该 hop RTT”，不要使用“相邻 hop 延迟”或“链路延迟”。

不能使用 `当前 hop.rtt - 前一 hop.rtt` 计算单跳链路时延。匿名 hop、不同 run 的 ECMP 分支、排队抖动、ICMP 限速和往返路径不对称都会使差值失真，差值也可能为负数。当前数据无法可靠提供边级单跳延迟，因此拓扑连线不应展示该指标。路径整体质量使用顶层 `e2e_rtt_avg`；节点使用自身的 `rtt`。

ASN 和云厂商富化使用 Kodo 本地离线 IP 数据库，不会向第三方服务发送 hop IP。以下情况富化字段会缺失：

- 私网、CGNAT、回环、链路本地或组播地址；
- hop 未响应，没有 `ip_address`；
- ASN 数据库未部署；
- 数据库中没有匹配记录。

前端不得假设每一跳都有 ASN，也不要因为 ASN 缺失将 hop 判定为异常。

### 5.4 多 run 展示建议

- 目标为域名时按 `runs[].destination.ip_address` 区分本次 run 实际使用的 IP；不要假设同一条记录的所有 runs 都使用相同目标 IP，也不要假设一次记录覆盖 DNS 的全部 IP。
- 默认路径可以选择最新 run，或者选择到达目标且 hop 最完整的 run。
- 路径变化视图以 `ttl + ip_address` 对齐不同 runs。
- 同一 TTL 在不同 run 出现不同 IP 是正常的，可能由 ECMP 或路由变化造成。
- `reachable=false` 的匿名 hop 只能归属到 TTL，不能归属到某个 IP 或 ASN。
- 不上传逐 hop 的 `probe_count`、`response_count`、`timeout_count`；同一条记录的 TTL 响应统计可以从多个 run 推导，跨记录趋势应由服务端聚合。

### 5.5 TTL 响应覆盖率

在一条记录内，每个 run 对实际出现的每个 TTL 发送一次 probe，因此可以按 TTL 从 `runs[].hops[]` 推导：

- `probe_count(ttl)`：包含该 TTL 的 run 数量；分母使用实际数据，不直接使用配置的 `traceroute_queries`。
- `response_count(ttl)`：其中 `reachable=true` 的数量。
- `timeout_count(ttl)`：其中 `reachable=false` 的数量。
- `response_coverage(ttl) = response_count / probe_count`。
- `timeout_rate(ttl) = timeout_count / probe_count`。
- `avg_rtt(ttl)`：只对 `reachable=true` 且带 `rtt` 的样本求平均。

这些指标描述“TTL 位置的探测响应情况”，不等于某台路由设备的真实丢包率。超时样本没有 IP，不能归因给同 TTL 中曾响应的某个 hop；中间设备也可能只是不响应或限速 ICMP。前端原型中的“最高丢包率”应改为“最高 TTL 超时率”，并提示样本数。默认 3 个 run 时结果粒度只有 `0%`、`33.3%`、`66.7%`、`100%`，不适合用作稳定告警。

端到端无响应比例使用顶层 `e2e_probe_loss_percent`，并结合 `e2e_packets_sent`、`e2e_packets_received` 判断样本量。UDP 应用沉默会记入 `e2e_unknown`，不应当作确定丢包。

## 6. 成功、部分路径和失败

推荐状态映射：

| 条件 | 前端状态 | 说明 |
| --- | --- | --- |
| `traceroute_status=reached` | 成功 | traceroute 到达目的端。 |
| `traceroute_status=partial` | 部分路径 | 获得部分 hops，但未确认到达目的端。 |
| `traceroute_status=failed` | 失败 | 使用 `traceroute_fail_type` 和 `traceroute_fail_reason` 展示失败原因。 |
| `traceroute_status` 缺失 | 未知 | 兼容旧数据时按未知处理。 |

`traceroute_fail_type` 可能值：

| 值 | 含义 |
| --- | --- |
| `timeout` | 探测超时。 |
| `dns_error` | DNS 解析失败。 |
| `permission` | 缺少 raw socket、network namespace 等权限。 |
| `protocol_unsupported` | 当前平台或执行器不支持该协议。 |
| `invalid_target` | 目标或端口配置无效。 |
| `connection_refused` | 目标拒绝连接。 |
| `target_unreachable` | 无路由、网络不可达或主机不可达。 |
| `runner_error` | 其他执行器错误。 |
| `unknown` | 无法分类。 |

失败示例：

```json
{
  "traceroute_status": "failed",
  "traceroute_fail_type": "dns_error",
  "traceroute_fail_reason": "lookup api.example.invalid: no such host",
  "message": "lookup api.example.invalid: no such host"
}
```

此时前端解析 `message` JSON 失败是预期行为，应直接展示为错误详情。

## 7. 前端字段优先级

| 展示内容 | 首选字段 | 回退字段 |
| --- | --- | --- |
| 目的名称 | `dst_domain` | `dst_ip` |
| 目的 IP | `dst_ip` | `dst_nat_ip` 仅用于显示实际探测目标 |
| 源名称 | `source_service` | `source_host`、`src_ip` |
| 探测出口 | `probe_interface` + `probe_source_ip` | `src_ip` |
| 状态 | `traceroute_status` | 无 |
| 统一时延 | `e2e_rtt_avg` | 无 |
| 无响应比例 | `e2e_probe_loss_percent` | 无 |
| 跳数 | `message.hop_count` | `hop_count` |
| hop 名称 | `reverse_dns[0]` | `ip_address`、`TTL n` |
| ASN 名称 | `as_name` | `AS${asn}` |
| hop RTT | `message.runs[].hops[].rtt` | 无 |
| hop 设备类型 | 无 | `unknown` / 通用网络 hop |
| hop 云厂商 | `cloud_provider` | 不要根据 `as_name` 自动推断 |

### 7.1 旧字段迁移

以下字段不再发往 Dataway。前端、新查询和 Kodo 富化应只使用当前规范字段：

| 旧字段 | 当前字段/处理方式 |
| --- | --- |
| `source_ip` | `src_ip` |
| `dest_ip` | `dst_ip` |
| `dest_port` | `dst_port` |
| `dest_domain` | `dst_domain` |
| `observed_dest_ip`、`observed_dest_port` | `dst_ip`、`dst_port` |
| `nat_dest_ip`、`nat_dest_port` | `dst_nat_ip`、`dst_nat_port` |
| `dest_cloud_provider` | `dst_cloud_provider` |
| `dest_reverse_dns` | `dst_reverse_dns` |
| `status`、`path_status` | `traceroute_status` |
| `fail_type` | `traceroute_fail_type` |
| `fail_reason` | `traceroute_fail_reason` |
| `success`、`path_success`、`path_destination_reached` | 不再上传；使用 `traceroute_status` |
| `e2e_protocol` | 不再上传；E2E 使用任务 `protocol` |
| `e2e_success` | 不再上传；使用 `e2e_status` 和 `e2e_packets_received` |
| `task_id`、`result_id`、`finished_at` | 不再上传 |
| `branch_key` | 不再上传；路由分支从 `message.runs[].hops[]` 计算 |

迁移完成后不要同时查询新旧字段，否则会导致分组维度重复或新数据无法命中。

### 7.2 实体和图标降级

当前契约没有稳定的 `source_type`、`destination_type`、`device_type` 或实体 ID。`cloud_provider` 只表示 IP 归属到某个云厂商，不表示具体资源类型。图标只能按已有上下文做展示层降级，不能据此建立稳定的资产关联：

- Source：有 `source_container_id` 时可显示容器图标；否则有 `source_service` 时显示服务图标；否则有 `source_host` 时显示主机图标；均缺失时使用通用源端图标。
- Destination：有 `dst_domain` 时可显示域名/外部端点图标，否则使用通用目标图标。不能仅根据端口推断为服务器、负载均衡或数据库。
- Hop：有 `cloud_provider` 时可显示对应云厂商图标，否则使用通用网络 hop 图标。`as_name` 仅作为自治系统名称展示，不能代替 `cloud_provider`。

若产品必须按服务器、交换机、负载均衡、防火墙或云组件分组，需要服务端使用 hop IP 关联 CMDB、网络设备、Kubernetes 和云资源清单，再提供稳定的实体类型与 ID。仅靠 traceroute 和 ASN 数据无法完成该分类。

## 8. 完整记录示例

以下示例用扁平 JSON 表示存储查询返回的一条记录。实际查询接口可能另外返回时间、索引和工作空间元数据。

```json
{
  "source": "netpath",
  "path_key": "np-v1-74bd920d68104c2ec181d7de6cfa3f38",
  "task_name": "dns-google",
  "task_source": "dynamic",
  "origin": "ebpf_netflow",
  "run_type": "dynamic",
  "protocol": "tcp",
  "traceroute_protocol": "tcp",
  "traceroute_status": "reached",
  "dst_domain": "dns.google",
  "dst_ip": "8.8.8.8",
  "dst_cloud_provider": "gcp",
  "dst_port": "443",
  "source_host": "node-a",
  "src_ip": "10.0.0.10",
  "src_port": "53000",
  "src_cloud_provider": "aws",
  "source_service": "frontend",
  "probe_source_ip": "10.0.0.10",
  "probe_gateway_ip": "10.0.0.1",
  "probe_interface": "eth0",
  "test_run_id": "run-ab12cd34",
  "scheduled_at": 1783656000000000,
  "started_at": 1783656000001000,
  "duration": 14000,
  "hop_count": 3,
  "max_ttl": 30,
  "traceroute_queries": 3,
  "e2e_status": "reached",
  "e2e_dest_ip": "8.8.8.8",
  "e2e_queries": 10,
  "e2e_packets_sent": 10,
  "e2e_packets_received": 10,
  "e2e_unknown": 0,
  "e2e_probe_loss_percent": 0,
  "e2e_rtt_avg": 8200,
  "e2e_rtt_min": 7900,
  "e2e_rtt_max": 8700,
  "e2e_rtt_variation_samples": 9,
  "e2e_rtt_variation_avg": 180,
  "e2e_rtt_variation_max": 420,
  "message": "{\"runs\":[{\"run_id\":\"1\",\"destination\":{\"ip_address\":\"8.8.8.8\",\"port\":443,\"reverse_dns\":[\"dns.google\"]},\"hops\":[{\"ttl\":1,\"ip_address\":\"10.0.0.1\",\"rtt\":0.8315,\"reachable\":true},{\"ttl\":2,\"reachable\":false},{\"ttl\":3,\"ip_address\":\"8.8.8.8\",\"rtt\":8.2,\"reachable\":true,\"asn\":15169,\"as_name\":\"GOOGLE\",\"as_prefix\":\"8.8.8.0/24\",\"cloud_provider\":\"gcp\"}]}],\"hop_count\":{\"avg\":3,\"min\":3,\"max\":3}}"
}
```

## 9. 当前不在数据契约内的能力

- 不提供相邻 hop 的链路时延；`rtt` 是探测源到响应 hop 的 RTT，不能通过相邻 RTT 相减得到可靠链路延迟。
- 不提供 hop 级 `device_type`，不能直接按服务器、交换机、负载均衡、防火墙或云网络组件分组。
- `cloud_provider` 只提供云厂商归属，不提供具体云产品或资源类型；`as_name` 仍只表示 ASN 所属组织。
- 不提供稳定的源端、目的端或 hop 实体类型和实体 ID；图标只能按已有上下文降级展示。
- 不提供逐 hop 的真实丢包率；匿名 timeout 不能归因到具体路由器。
- 不提供逐 hop 的 `probe_count`、`response_count`、`timeout_count`。
- 不保证每个 hop 都能得到 IP、reverse DNS 或 ASN。
- 暂不提供路径变化事件；前端或服务端需要基于历史 runs 比较路径。
- 当前不支持 IPv6 traceroute 和 TCP SYN traceroute。
