# eBPF 采集器对 Kubernetes 工作负载影响排查记录

日期：2026-04-16

## 背景

线上出现了 Kubernetes Pod 异常停止相关事件，典型表现为：

- `FailedKillPod`
- `KillPodSandboxError`
- `context deadline exceeded`
- 就绪探针持续失败后，kubelet 停容器超时

目标不是只做分析，而是要把高风险路径收紧，先消除明显的生产风险。

## 本次排查范围

本次主要 review 了 external eBPF 采集器里四条路径：

1. `procwatch`
2. `l7flow`
3. `conntrack`
4. `l4log` / `bpf-netlog`

## 代码级结论

### 1. 最可疑路径：`procwatch` 对目标进程做动态 uprobe attach

`procwatch` 会跟踪进程生命周期，并在满足条件时对目标二进制的 `runtime.execute` 做 uprobe attach。

这条路径的特点是：

- 不是被动观测
- 会直接碰目标业务进程
- 目标如果是 Go 进程，命中面会很大

关键代码：

- `internal/procwatch/runtime.go`
- `internal/procwatch/catalog.go`
- `internal/cmd/run/run.go`

原来的问题在于：

- 只要启用了 `ebpf-trace`
- 就可能把目标进程 attach 面打开
- 而默认不是强白名单模式

这对 Kubernetes 上的业务容器、控制面容器、数据库类进程都不够安全。

### 2. 次高风险路径：`conntrack` kprobe 在内核热路径放大压力

`conntrack` 采集器在：

- `__nf_conntrack_hash_insert`
- `nf_ct_delete`

两个内核函数上挂了 kprobe。

这条路径不会改包，也不会改 conntrack 判定逻辑，但会在连接 churn 很高的节点上增加：

- conntrack 热路径 CPU
- BPF map update/delete 压力

它更像“节点放大器”，不一定是直接导致单个 Pod kill 失败的唯一原因，但会把已有的节点压力问题放大。

### 3. `l4log` / `bpf-netlog` 当前更像资源噪声源，不像首要元凶

这条路径当前已经做过多轮收敛：

- container netns fd 不再长期持有
- host-peer 共享 ring
- host/container NIC 枚举缓存
- listener watcher 稳定性优化

它仍然会带来：

- `setns`
- `/proc/<pid>/net/*` 扫描
- 抓包与重组的 CPU/内存消耗

但从代码结构看，它更像资源开销问题，而不是直接把目标业务进程 attach 进去。

## 本次落地改动

### 目标

先把最危险的“默认 attach 到目标业务进程”行为收紧。

### 具体改动

我把 `procwatch` 的目标进程 attach 策略改成了安全默认：

- 只有显式 `trace_all_proc = true`
- 或者显式配置了 `trace_name_list` / `trace_env_list`

才允许把目标进程标成 `traceable`，从而继续做 uprobe attach。

否则即使启用了 `ebpf-trace`：

- 仍然保留 catalog / 被动进程信息能力
- 但**不再默认 attach 到业务进程**

修改文件：

- `internal/procwatch/catalog.go`
- `internal/cmd/run/run.go`
- `internal/procwatch/catalog_test.go`

### 行为变化

调整前：

- 开了 `ebpf-trace`
- 即使没有明确目标范围
- 也可能对较大范围的目标进程做 attach

调整后：

- 开了 `ebpf-trace`
- 但没有 `trace_all_proc=true`
- 也没有白名单
- 则默认不 attach 目标进程

程序启动时会打印明确警告，提示当前处于安全默认模式。

## 第二轮硬化

在继续 review `procwatch` 和 namespace 相关路径后，又补了三处直接的代码硬化。

### 1. 没有 `trace_server` 时，不再允许目标进程 attach

之前的实现里：

- 只要启用了 `ebpf-trace`
- `procwatch` 的目标进程 attach 面就可能被打开

即使没有配置 `trace_server`，也可能进入“尝试 attach”的路径。

现在改成：

- 必须同时满足 `ebpf-trace` 开启
- 且 `trace_server` 非空
- 且满足显式 trace 范围

才允许真正打开目标进程 attach。

这进一步收缩了误触达业务进程的风险面。

### 2. 修正 `attachProcess()` 的状态一致性问题

review 时发现一个具体问题：

- 旧逻辑里，`attachProcess()` 会先准备 `inject`
- 再进入 attach 循环
- 即使实际没有任何 uprobe attach 成功，也可能把 `inject` 状态缓存下来

这样后续就会出现一种不一致状态：

- 代码以为这个二进制已经准备好注入信息
- 但实际上对应 hook 并没有成功挂上

这次已经修正为：

- 只有至少一个 hook 真正 attach 成功
- 才把 `inject` 写进 map
- 才把 `inject` 缓存在 `binaryRegistry`

同时，在二进制刷新路径上，旧的 `inject` 状态现在也会被显式清理，避免把过期状态带到下一轮。

### 3. namespace 切换恢复逻辑加强

`CallWithNetNS()` 现在改成 defer 方式统一恢复原 netns。

这样即使执行函数中途异常退出，线程把错误 netns 带出当前调用的风险也更低。

这不是本次 Kubernetes 问题的首要元凶，但属于必须补上的安全边界。

### 4. 进程退出清理 `proc_inject` 与 attach 回滚

本次又补了一层直接降低扰动风险的保护：

- `releaseProcess()` 现在会清理 `bpf-map bmap_procinject` 里对应 pid 的临时注入状态。
  这样可以避免旧 pid 映射长期残留，占用 BPF map 和触发误匹配。
- `attachProcess()` 在 `runtime.execute` uprobe 挂载后，如果 `proc_inject` map 拿不到或写入失败，会立即把当前已挂的 hook 解绑回滚。
  这样不会在半成功状态下把“已挂钩但未成功注入”这类脏状态留在 runtime 内。
- `procwatch.Stop()` 增加 runtime 空判断。当前 `NewProbeWatcher()` 会在未启用 trace 或无 trace 目标时返回 `Runtime=nil`，如果直接调用 `Shutdown()` 会触发 `nil`。这次改造避免了该分支下 stop 流程的 panic 风险。
- `procwatch.Start()` 同样补了 runtime 空判断。否则在“`trace_server` 已配置，但未显式配置 `trace_all_proc` / allowlist”这个安全模式里，`NewProbeWatcher()` 返回的 watcher 会带着 `Runtime=nil` 进入 `Start()`，随后在 `StartRuntime()` 处直接触发 panic。
- `run.parseFlags()` 现在会在每次解析配置前重置 feature 全局布尔值。旧实现依赖包级状态，但不会在二次解析时清空，容易在同进程的重复初始化、测试或未来热重载场景里把旧配置残留带到新配置，造成“实际上已关闭但逻辑仍开启”的假阳性行为。

## 为什么这样改

这次问题的约束不是“功能尽量全”，而是“线上容器不能继续被高风险默认行为影响”。

在现有架构下，最小风险、最容易快速止血的方案就是：

- 不先去猜具体是哪一个业务进程被打中
- 先把默认 attach 面收缩成显式 opt-in

这不会影响：

- 纯网络抓包
- `bpf-netlog`
- `host-peer` shared ring
- 被动网络统计

但会明显降低：

- 目标进程被动态 uprobe attach 的概率
- 对 Go 业务进程运行时的侵入性

## 已完成验证

已完成：

- `procwatch` 单测通过
- `cmd/run` 单测通过
- `l4log` 单测通过

执行命令：

```bash
go test ./internal/plugins/externals/ebpf/internal/procwatch ./internal/plugins/externals/ebpf/internal/cmd/run
go test ./internal/plugins/externals/ebpf/internal/procwatch ./internal/plugins/externals/ebpf/internal/cmd/run ./internal/plugins/externals/ebpf/internal/l4log
```

## 建议的线上验证步骤

1. 部署这版 collector。
2. 保持 `ebpf-net`、`bpf-netlog` 等被动路径可用。
3. 不配置 `trace_all_proc=true`。
4. 如果确实需要 trace，只给出明确白名单：
   - `trace_name_list`
   - 或 `trace_env_list`
5. 对比观察：
   - kubelet 事件里的 `FailedKillPod`
   - `KillPodSandboxError`
   - 停容器耗时
   - `datakit-ebpf` CPU / RSS

## 后续动作

如果这版上线后问题明显缓解：

- 说明高风险路径基本锁定在目标进程 attach 面
- 后续再考虑是否进一步把 `trace_all_proc` 变成更强提示或更严格约束

如果问题仍然存在：

- 下一优先级就是继续隔离 `conntrack`
- 再往后才是继续收 `l4log` / listener / netns 扫描路径
