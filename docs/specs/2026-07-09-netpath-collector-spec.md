# DataKit NetPath 采集器规范

`inputs.netpath` 从本机 DataKit 所在节点探测到 TCP 或 ICMP 目标的网络路径。目标可以来自静态配置，也可以由 `datakit-ebpf` 等本地组件动态上报。

实现目录：`internal/plugins/inputs/netpath`。

## 范围

该采集器负责路径探测、目标调度、traceroute 结果归一化和结果上报。不承载 HTTP、Browser、SSL、gRPC、WebSocket 或多步骤拨测。

当前实现支持：

- 通过 `[[inputs.netpath.targets]]` 配置静态目标。
- 通过 `POST /v1/netpath/candidates` 接收动态候选目标。
- TCP、UDP、ICMP 和 `auto` 协议。`auto` 在带端口时选择 TCP，否则选择 ICMP。Linux 支持三种协议；macOS 当前只支持 ICMP；Windows 不支持 NetPath。
- 目标去重、TTL、限速、worker 并发和稳定 jitter。
- 可选的目标及可达 hop 反向 DNS 富化。

动态发现和反向 DNS 默认关闭。未开启 `dynamic` 时，输入只执行静态目标。

## 配置

```toml
[[inputs.netpath]]
  protocol = "tcp"
  interval = "60s"
  timeout = "1s"
  max_ttl = 30
  traceroute_queries = 3
  e2e_queries = 10

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

  [inputs.netpath.dynamic]
    enabled = true
    token = ""
    protocol = "auto"
    contexts_limit = 5000
    ttl = "50m"
    interval = "20m"
    flush_interval = "10s"
    max_per_minute = 150
    workers = 4
    timeout = "1s"
    max_ttl = 30
    traceroute_queries = 3
    e2e_queries = 10
    input_queue = 1000
    process_queue = 1000
    max_tests_per_request = 1000
    max_body_bytes = 1048576
    monitor_ip_without_domain = false

  [inputs.netpath.reverse_dns]
    enabled = false
    timeout = "500ms"
    cache_ttl = "10m"
    cache_size = 4096

  [inputs.netpath.tags]
    env = "production"
```

### 静态目标

`target` 支持主机名或 IP；`ip` 可以作为仅 IP 目标的替代配置。TCP 目标必须设置 `port`，ICMP 目标无需端口。目标级的 `protocol`、`interval`、`timeout`、`max_ttl`、`traceroute_queries` 和 `e2e_queries` 会覆盖顶层同名设置。

### 动态目标

`dynamic` 仅影响 API 上报的候选目标：

| 配置项 | 含义 |
| --- | --- |
| `token` | 可选共享令牌，与请求头 `X-Datakit-Netpath-Token` 比较。 |
| `protocol` | `tcp`、`udp`、`icmp` 或 `auto`。显式协议会覆盖候选协议。 |
| `ttl` | 相同候选目标再次上报时续期的存活时间。 |
| `interval` | 去重后的动态任务执行间隔。 |
| `flush_interval` | 将可执行任务送入 worker 的频率，最小值为 `1s`。 |
| `max_per_minute` | 调度器每分钟的任务放行预算。 |
| `workers` | 最大探测并发数。 |
| `input_queue`、`process_queue` | 去重前和 worker 前的有界队列。 |
| `contexts_limit` | 活跃去重任务上下文的最大数量。 |
| `monitor_ip_without_domain` | 为 false 时，只有 IP 且无 hostname/domain 的候选目标会被丢弃。 |
| `filters` | 候选目标进入调度器前应用的排除规则。 |

生产上限：Linux UDP traceroute 的 `max_ttl <= 255`，TCP/ICMP traceroute 的有效 `max_ttl <= 60`；此外 `traceroute_queries <= 10`、`e2e_queries <= 100`、`workers <= 64`、`contexts_limit <= 100000`、队列长度 `<= 100000`、`max_per_minute <= 10000`、`max_tests_per_request <= 10000`、`max_body_bytes <= 16777216`。每个 candidate 的身份字段单项不超过 1024 bytes，计入请求级默认值后每个 test 合计不超过 4096 bytes。请求级 tags 与 test 级 tags 合计不超过 64 个，tag key 不超过 128 bytes，value 不超过 1024 bytes；请求级 tags 由同一请求产生的任务共享，不按 test 复制。

`traceroute_queries` 表示完整 traceroute 的执行次数，而不是同一个 TTL 的重试次数。每次执行从 TTL 1 开始，到达目标或达到 `max_ttl` 后结束，并独立写入 `message.runs[]`。目标为域名时，每个 run 独立执行 DNS 解析并使用返回的第一个 IPv4；DNS 返回顺序变化时，同一条记录的不同 runs 可能探测不同目标 IP，实际 IP 写入 `runs[].destination.ip_address`。不保证一次任务覆盖域名的全部 IP。UDP 在单次 run 内复用同一个源端口，以减少 ECMP 哈希变化导致的路径混合；不同 run 使用新的源端口，以观察可能的多路径分支。单次 run 内每个 probe 使用不同的 UDP datagram 长度，并校验 ICMP 引用报文中的长度，避免延迟回包被误归到后续 TTL。

`e2e_queries` 表示独立的端到端 probe 次数，默认 `10`。E2E 与 traceroute 并发执行且不探测中间 hop：TCP 重复建连并将连接成功或 RST/connection refused 都视为收到目的端响应；ICMP 使用 Echo Request/Reply；Linux UDP 使用正常端到端 IP TTL 的 UDP 并接收目的端 ICMP 响应。E2E 不使用 traceroute 的 `max_ttl`。目标为域名时，E2E 独立执行 DNS 解析，实际使用的 IPv4 写入 `e2e_dest_ip`，因此可能与任一 `runs[].destination.ip_address` 不同。开放 UDP 应用没有义务响应 datagram，因此 UDP 沉默记为 `unknown`，不计入 `e2e_probe_loss_percent` 的分母。

Traceroute 按任务协议发送真实的 IPv4 probe：TCP 使用带递增 TTL 的 TCP SYN 并接收中间 hop 的 ICMP Time Exceeded 或目的端的 SYN-ACK/RST-ACK，ICMP 使用 Echo，Linux UDP 使用 UDP datagram。三种 traceroute 都需要 DataKit 进程具备 raw socket 权限。

### 动态过滤规则

每个 `[[inputs.netpath.dynamic.filters]]` 都是排除规则。同一规则中的非空条件按 AND 匹配；同一条件中的多个值按 OR 匹配。任意规则匹配后即丢弃候选目标。支持源/目标 CIDR、源/目标主机、namespace、源服务、进程、容器、origin、协议和端口。

```toml
[[inputs.netpath.dynamic.filters]]
  name = "ignore-system"
  namespaces = ["kube-system"]

[[inputs.netpath.dynamic.filters]]
  name = "ignore-private-database"
  dest_cidrs = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
  ports = [5432, 6379]
```

## Candidate API

动态 API 默认通过 `POST /v1/netpath/candidates` 注册；设置 `dynamic.enabled = false` 可以关闭。只有请求来源和 HTTP server 实际接受连接的本地地址均为 loopback 时才允许省略 `token`；其他请求必须配置非空 `token`。因此 DataKit 使用节点 IP、反向代理或非 loopback 地址接收候选时，datakit-ebpf 必须发送相同 token。

```json
{
  "source": "ebpf_netflow",
  "host": "node-a",
  "tags": {
    "cluster": "production"
  },
  "tests": [
    {
      "hostname": "api.example.com",
      "target_ip": "203.0.113.10",
      "port": 443,
      "protocol": "tcp",
      "origin": "ebpf_netflow",
      "namespace": "4026532000",
      "source": {
        "hostname": "node-a",
        "ip": "192.0.2.20",
        "netns": "4026532000",
        "pid": 1234,
        "process_name": "checkout",
        "service_name": "checkout"
      },
      "tags": {
        "direction": "outgoing"
      }
    }
  ]
}
```

必须提供 `hostname` 或 `target_ip`/`ip`。TCP 候选目标还必须提供端口。若同时提供 `hostname` 和 IP，调度及主动探测优先使用 `hostname`，IP 保留为原始流量上下文；若同时提供 `target_ip` 与 `ip`，优先保留 `target_ip`。

成功或部分成功均返回 HTTP 200：

```json
{
  "accepted": 1,
  "dropped": 0,
  "queue_size": 1,
  "drop_reasons": {}
}
```

`accepted` 仅表示候选目标通过同步校验并进入输入队列，不代表已写入任务 store、已执行或已上报。store 满、任务过期和 `processCh` 满等异步结果通过指标观察。

同步丢弃原因包括 `unauthorized`、`bad_request`、`body_too_large`、`empty_tests`、`too_many_tests`、`invalid_identity`、`invalid_tags`、`missing_target`、`missing_port`、`unsupported_protocol`、`ip_without_domain`、`filtered` 和 `input_chan_full`。

## 调度

HTTP 路由注册或输入启动时会初始化 scheduler。worker 尚未启动时收到的候选目标保留在有界 `inputCh` 中，输入启动后继续处理。

动态任务经过三个阶段：

1. `inputCh` 接收完成校验的动态候选目标。
2. `taskStore` 按内部 schedule key 去重，维护下次执行时间并按 TTL 清理动态任务。
3. `processCh` 按 flush 预算向有限 worker 并发放行任务。

静态目标直接进入独立的本地调度 lane，使用独立的 flush、process queue 和 worker，不占用 dynamic 的 `contexts_limit`、`max_per_minute`、队列或 worker 配额。

动态 schedule key 由源主机、优先源标识（`service_name`、`process_name`、`container_id`、源 hostname）、目标 hostname（缺失时使用 IP）、端口和协议组成；仅发生目的地址转换时追加原始目的 IP 和端口，以区分不同的原始服务。同一非 NAT hostname 对应的不同观测 IP 共享一个调度任务；重复上报会续期并更新下次探测使用的结构化 candidate 属性，但不会提前已排定的执行时间。每次放行后按 key 添加最多 `min(interval/20, 30s)` 的稳定 jitter。

## 结果模型

每次执行产生一条网络（N）分类的 `netpath` point。

tags 表示探测身份和路径上下文，包括 `task_name`、`task_source`、`origin`、`run_type`、`protocol`、`src_ip`、`src_port`、`dst_ip`、`dst_port`、`dst_domain`、`namespace`、`source_host`、`source_process`、`source_service`、`netns`、`traceroute_protocol`、`traceroute_status` 和 `traceroute_fail_type`。四元组命名及 DNAT 字段 `dst_nat_ip`、`dst_nat_port` 与 NetFlow 对齐；`dst_*` 始终表示原始目的，`dst_nat_*` 仅在实际探测目标发生转换时存在。内部 schedule key 仅用于本地任务调度，不上传到 Dataway。Dataway 不上传 `branch_key`、`path_key`、`proto`、`target*`、`dest_*`、`source_ip`、`observed_dest_*` 和 `nat_dest_*` 等内部字段、旧字段或别名。

candidate 只接收结构化的端点、来源实体和自定义 tags，不再接收或转发任意 `metadata` fields。eBPF Candidate 不在 tags 中重复传递四元组、进程、服务或流量统计。

源和目的关联无需解析 hop JSON：`src_ip`、`src_port`、`dst_ip`、`dst_port`、`dst_domain`、`source_host`、`source_service`、`source_process`、`source_container_id` 和 `namespace` 都是 tags。即使没有产生 hop，端点四元组仍会保留；静态任务无法确定端口时使用 `"*"`，源 IP 未提供时回退到 `probe_source_ip`。`source_pid` 是 field。

`src_cloud_provider` 和 `dst_cloud_provider` 是端点级可选 tags。DataKit 会将已有的全局 `cloud_provider` 映射到 `src_cloud_provider`；Kodo 可以按顶层 `src_ip`、`dst_ip` 补齐端点云厂商，并在 `message.runs[].hops[]` 中为可识别的 hop 增加 `cloud_provider`。云厂商信息来自本地 IP 归属库，无法可靠匹配时保持缺失，不从 `as_name` 猜测。

Linux 会在每次 probe 完成后，以 traceroute 的单一实际目标在 DataKit 当前 probe network namespace 中查询路由，并补充 `probe_source_ip`、`probe_gateway_ip`、`probe_interface`、`probe_interface_mac` 和 `probe_netns` tags。这些 tags 描述实际 traceroute 的出口；同一记录的 runs 使用多个目标 IP 时没有单一目标，因此跳过本次顶层路由富化。查询失败不影响 probe 结果。

核心探测 fields 为 `test_run_id`、`scheduled_at`、`started_at`、`duration`、`source_pid`、`max_ttl`、`traceroute_queries`、`e2e_dest_ip`、E2E 质量 fields、`hop_count`、`traceroute_fail_reason` 和 `message`。TCP/ICMP 路径 runner 只执行 traceroute，不再上传 `success`、`path_success`、`path_destination_reached` 或旧的 `path_*` 质量别名。路径状态统一使用 `traceroute_status` tag；`e2e_status` 独立表示端到端 probe 状态。

traceroute 结果以标准 JSON 字符串写入 `message`，以便使用日志 `message` 的 JSON 加速：

```json
{
  "runs": [
    {
      "run_id": "1",
      "destination": {
        "ip_address": "203.0.113.10",
        "port": 443,
        "reverse_dns": ["api.example.com"]
      },
      "hops": [
        {
          "ttl": 1,
          "ip_address": "10.0.0.1",
          "reverse_dns": ["gateway.local"],
          "rtt": 0.1,
          "reachable": true
        },
        {
          "ttl": 2,
          "reachable": false
        }
      ]
    }
  ],
  "hop_count": {
    "avg": 2,
    "min": 2,
    "max": 2
  }
}
```

不可达 hop 使用 `reachable = false`，不会以 `"*"` 作为 IP 占位符。开启 reverse DNS 后，目标补充 `dst_reverse_dns`，可达 hop 补充 `reverse_dns`。

## 指标与失败类型

Prometheus 指标：

- `datakit_netpath_candidate_total{status,reason}`
- `datakit_netpath_drop_total{stage,reason}`
- `datakit_netpath_queue_size{queue}`
- `datakit_netpath_store_size`
- `datakit_netpath_task_total{source,protocol,status}`
- `datakit_netpath_task_run_cost_seconds{source,protocol,status}`
- `datakit_netpath_feed_error_total`

`datakit-ebpf` 的 candidate 客户端额外暴露：

- `dkebpf_netpath_candidate_requests_total{result}`
- `dkebpf_netpath_candidates_dropped_total{reason}`
- `dkebpf_netpath_pending_candidates`

candidate POST 在传输错误、HTTP 408/425/429 或 5xx 时保留原批次并指数退避重试，默认最多尝试 5 次；其他 4xx 视为永久错误。达到重试上限、队列已满或关闭时无法发送的 candidate 会计入 dropped 指标并记录 warning。

探测失败会归类为 `traceroute_fail_type`：`timeout`、`dns_error`、`permission`、`protocol_unsupported`、`invalid_target`、`connection_refused`、`target_unreachable`、`runner_error` 或 `unknown`。

## 暂不支持

中心下发任务、TCP SYN traceroute、IPv6 traceroute、路径变化检测、断言和每 hop 独立 measurement 不在当前实现范围内。
