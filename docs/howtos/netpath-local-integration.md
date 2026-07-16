# NetPath 本地接入指南

本文验证完整本地链路：

```text
datakit-ebpf -> POST /v1/netpath/candidates -> inputs.netpath -> netpath point
```

建议先使用少量明确目标验证。动态发现只有在 `inputs.netpath.dynamic.enabled` 和 `inputs.ebpf.network_path_enabled` 都开启时才会生效。

## 1. 配置 NetPath

创建 `conf.d/netpath/netpath.conf`：

```toml
[[inputs.netpath]]
  protocol = "tcp"
  interval = "60s"
  timeout = "1s"
  max_ttl = 30
  traceroute_queries = 3
  e2e_queries = 10

  [inputs.netpath.dynamic]
    enabled = true
    token = "local-secret"
    protocol = "auto"
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
    env = "local"
```

建议保持 `monitor_ip_without_domain = false`。此时只有 IP、没有 hostname/domain 的候选目标会被丢弃；只有在明确需要探测纯 IP 路径时才设置为 `true`。

`dynamic.enabled` 默认是 `true`，设置为 `false` 可关闭 candidate API。DataKit HTTP 监听地址若暴露到主机外，必须使用非空 `token`。

## 2. 先验证 Candidate API

先手动上报一个候选目标。成功响应只表示候选目标进入 scheduler，实际探测会在后续 flush 中执行。

```bash
curl -sS -X POST "http://127.0.0.1:9529/v1/netpath/candidates" \
  -H "Content-Type: application/json" \
  -H "X-Datakit-Netpath-Token: local-secret" \
  -d '{
    "source": "manual",
    "host": "node-a",
    "tags": {
      "cluster": "local"
    },
    "tests": [
      {
        "hostname": "example.com",
        "port": 443,
        "protocol": "tcp",
        "origin": "manual",
        "namespace": "default",
        "source": {
          "hostname": "node-a",
          "ip": "127.0.0.1",
          "port": 50000,
          "process_name": "curl",
          "service_name": "manual-test"
        },
        "tags": {
          "direction": "outgoing"
        }
      }
    ]
  }'
```

期望响应：

```json
{
  "accepted": 1,
  "dropped": 0,
  "queue_size": 1,
  "drop_reasons": {}
}
```

`source.ip/source.port` 与 `dst_ip/dst_port` 是结构化原始四元组；`target_ip/port` 是候选流量观测到的目标。若同时存在 `hostname` 和 IP，DataKit 按 hostname 调度并主动探测，每个 traceroute run 独立解析域名，实际目标写入 `message.runs[].destination.ip_address`。自定义 `tags` 用于额外筛选，任意 `metadata` 不再接收或转发。

## 3. 开启 eBPF Candidate 上报

在已有 `inputs.ebpf` 配置中增加：

```toml
[inputs.ebpf]
  network_path_enabled = true
  network_path_api = "http://127.0.0.1:9529/v1/netpath/candidates"
  network_path_token = "local-secret"
  network_path_flush_interval = "10s"
  network_path_batch_size = 100
  network_path_http_timeout = "5s"
  network_path_queue_size = 1000
```

也可通过环境变量配置：

```bash
ENV_INPUT_EBPF_NETWORK_PATH_ENABLED=true
ENV_INPUT_EBPF_NETWORK_PATH_API=http://127.0.0.1:9529/v1/netpath/candidates
ENV_INPUT_EBPF_NETWORK_PATH_TOKEN=local-secret
```

`datakit-ebpf` 会提交出站 IPv4 TCP 连接候选目标。DataKit 负责后续过滤、去重、探测、traceroute 归一化和 point 上报。eBPF 只传递路径身份和源、目的关联信息，不传递连接统计到 netpath。

### Kubernetes 环境变量

Kubernetes 可以在 `ENV_DEFAULT_ENABLED_INPUTS` 的现有值末尾追加 `netpath`，自动创建 input；无需单独挂载 `netpath.conf`。动态候选默认开启，其余参数可通过 DaemonSet 环境变量覆盖：

```yaml
- name: ENV_INPUT_NETPATH_DYNAMIC_TTL
  value: "50m"
- name: ENV_INPUT_NETPATH_DYNAMIC_INTERVAL
  value: "20m"
- name: ENV_INPUT_NETPATH_DYNAMIC_FLUSH_INTERVAL
  value: "10s"
- name: ENV_INPUT_NETPATH_DYNAMIC_MAX_PER_MINUTE
  value: "150"
- name: ENV_INPUT_NETPATH_DYNAMIC_WORKERS
  value: "4"
- name: ENV_INPUT_NETPATH_DYNAMIC_E2E_QUERIES
  value: "10"
- name: ENV_INPUT_NETPATH_DYNAMIC_MONITOR_IP_WITHOUT_DOMAIN
  value: "true"
- name: ENV_INPUT_EBPF_NETWORK_PATH_ENABLED
  value: "true"
- name: ENV_INPUT_EBPF_NETWORK_PATH_API
  value: "http://127.0.0.1:9529/v1/netpath/candidates"
```

`ttl` 从同一路径最后一次被发现时重新计时；`interval` 是同一路径的实际 probe 周期；`flush_interval` 是调度扫描周期；`max_per_minute` 是所有动态路径共享的每分钟执行预算。

## 4. 检查结果和自监控

向一个有域名的目标产生出站 TCP 流量后，确认 `netpath` point 包含：

- `task_source=dynamic`、`origin=ebpf_netflow`、`source_service`、`src_ip`、`src_port`、`dst_ip`、`dst_port` 等路径上下文 tags；发生 DNAT 时另有 `dst_nat_ip`、`dst_nat_port`；
- Linux 上还会包含实际探测出口的 `probe_source_ip`、`probe_gateway_ip`、`probe_interface` 和 `probe_netns` tags；
- `traceroute_protocol`、`traceroute_status` 等路径 tags，`hop_count` field，以及包含完整逐跳路径 JSON 的 `message` field；
- 独立 E2E fields：`e2e_packets_sent`、`e2e_packets_received`、`e2e_probe_loss_percent`、`e2e_rtt_avg` 和 `e2e_rtt_variation_avg`；UDP 沉默时查看 `e2e_unknown`；
- 公网 hop 在服务端可进一步获得 `asn`、`as_name`、`as_prefix` 和 `cloud_provider`；私网、未响应或无法匹配的 hop 没有对应富化字段。
- 即使没有产生 hop，顶层 `source_host`、`src_ip`、`src_port`、`dst_ip`、`dst_port`、`src_cloud_provider` 和 `dst_cloud_provider` 仍可用于检索；未知端口使用 `"*"`，可选富化字段可能缺失。

`message` 成功时为 `runs[].destination` 和 `runs[].hops[]` 结构的 JSON，失败且没有路径时可能是错误文本。完整字段及前端解析规则见 [NetPath 前端数据对接说明](../specs/2026-07-10-netpath-frontend-data-contract.md)。

使用以下指标确认候选目标流转和调度压力：

```text
datakit_netpath_candidate_total{status,reason}
datakit_netpath_drop_total{stage,reason}
datakit_netpath_queue_size{queue}
datakit_netpath_store_size
datakit_netpath_task_total{source,protocol,status}
datakit_netpath_task_run_cost_seconds{source,protocol,status}
datakit_netpath_feed_error_total
```

`accepted` 不是完成信号。至少等待一个 `flush_interval`，再检查 `task_total`、结果 point 和 drop 指标。

## 5. 提升流量前先配置过滤

动态发现可能观察到大量无需探测的目标。先针对已知噪声 namespace 或目的地址配置过滤：

```toml
[[inputs.netpath.dynamic.filters]]
  name = "ignore-kube-system"
  namespaces = ["kube-system"]

[[inputs.netpath.dynamic.filters]]
  name = "ignore-private-databases"
  dest_cidrs = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
  ports = [5432, 6379]
```

匹配规则的候选目标仍返回 HTTP 200，但响应中会出现 `accepted = 0` 与 `filtered:ignore-kube-system` 之类的 `drop_reasons`。

## 6. 常见问题

| 现象 | 检查项 |
| --- | --- |
| HTTP 401 | `network_path_token` 必须与 `dynamic.token` 一致。 |
| HTTP 404 | 确认 `inputs.netpath.dynamic.enabled` 未设置为 `false`，并重载 DataKit 配置。 |
| `ip_without_domain` | 提供 `hostname`，或明确启用 `monitor_ip_without_domain`。 |
| `unsupported_protocol` | 使用 `tcp`、`udp`、`icmp` 或 `auto`。 |
| `input_chan_full` 或 `process_chan_full` | 降低候选流量，增加过滤，或调整队列、worker 和速率限制。 |
| `traceroute_fail_type=permission` | 为 DataKit 进程提供 TCP/UDP/ICMP traceroute 所需的 raw socket 权限。 |
| 本机请求返回 `403` | 确认 candidate URL 使用 `127.0.0.1`/`::1`；通过节点 IP 或本机反向代理访问时必须配置相同 token。 |
| 没有 eBPF 候选目标 | 确认存在出站 IPv4 TCP 流量、已开启 `network_path_enabled`、API 可达且 token 匹配。 |

reverse DNS 默认关闭。应在基础探测正常后再开启；它会对目标和可达 hop 发起 PTR 查询，并受 timeout 和 cache 限制。
