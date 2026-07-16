# NetPath 采集器路线图

生产接口以 `docs/specs/2026-07-09-netpath-collector-spec.md` 为准；本地验证步骤见 `docs/howtos/netpath-local-integration.md`。

## 当前交付

第一阶段交付独立的 `inputs.netpath` 采集器，用于网络路径探测。它与 HTTP、Browser、SSL、gRPC、WebSocket 和多步骤拨测保持独立。

已完成能力：

- 通过 `inputs.netpath` 配置静态 TCP、UDP 和 ICMP 目标。
- 通过 `POST /v1/netpath/candidates` 接收动态 TCP、UDP 和 ICMP 候选目标。
- 从 eBPF netflow 发现出站 IPv4 TCP 候选目标。
- 任务去重、TTL 续期、有界队列、放行限速、worker 和稳定 jitter。
- candidate filters、请求大小和批量限制、可选 token 鉴权及 IP-only 控制。
- 每次探测只产生一条 `netpath` 结果，traceroute hop 放在 JSON fields 中。
- 带有界缓存的可选 reverse DNS 富化。
- scheduler、队列、丢弃、探测耗时和上报错误 Prometheus 指标。

动态来源和 reverse DNS 默认关闭。生产灰度应从少量静态目标开始，再逐步开放带过滤规则的动态来源。

## 明确延期项

以下能力不在当前范围内：

- 中心下发 NetPath 任务和任务生命周期同步。
- on-demand 执行和任务优先级策略。
- 路径变化检测、拓扑差异分析和断言告警。
- 每 hop 独立 measurement。

## 后续工作

### 中心任务来源

- 定义带版本的任务协议，包括目标、协议、端口、间隔、超时、TTL、traceroute 参数、tags 和生命周期状态。
- 将 `server` 来源接入现有 scheduler，不改变动态 candidate 的现有语义。
- 定义静态、动态和中心任务间的优先级及容量分配。
- 实现任务更新、暂停、删除、重试和本地 fallback。

### 协议和结果增强

- 补齐 TCP SYN、IPv6 和 Paris traceroute 的权限模型、超时行为和结果结构。
- 基于归一化 traceroute 结果定义路径变化比较和告警语义。
- 在明确基数和保留策略后再评估每 hop measurement。

### 生产验证

- 验证各支持环境中的 traceroute 权限。
- 在代表性出站流量下评估候选速率、队列深度、task store 大小和结果量。
- 验证部署配置中的 eBPF 到 DataKit token 传递和 HTTP 监听暴露范围。
- 在大范围启用前验证 reverse DNS 延迟和缓存行为。
