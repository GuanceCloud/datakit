# Container 采集器维护说明

<!-- markdownlint-disable MD013 -->

本文件适用于 `internal/plugins/inputs/container/` 及其子目录。根目录 `AGENTS.md` 和 `AGENTS_COMMON.md` 同时适用；若规则冲突，以本文件为准。本文描述当前实现，也列出不得误当成设计目标的已知边界。修改代码后应同步修正文档中的事实。

## 模块范围与事实来源

- `input.go` 定义 `Input`、默认值、配置保护范围和采集器生命周期。
- `impl.go` 选择普通 Docker/CRI、ECS Fargate、GCP Cloud API 及 Kubernetes 资源采集模式，并组装共享依赖。
- `container.go` 负责 runtime collector、指标/对象循环和日志发现调度；`container_gather.go` 构造容器指标与对象。
- `pod_watcher.go` 提供本节点 Pod informer、事件队列和日志扫描信号；`pod_metadata.go` 隔离 informer/API 两种 Pod 元数据来源并记录 cache miss。
- `container_log.go` 完成 runtime 容器扫描、Pod 元数据合并和有效日志配置选择；`log_config.go` 解析、补全和规范化配置。
- `log_coordinator.go` 拥有按 container ID 建立的日志任务、CRD 配置、tailer 创建/更新/关闭以及扫描信号广播。
- `crd_watcher.go` 监听 `ClusterLoggingConfig` 并触发全量日志 reconciliation。
- `kubernetes/` 是 Kubernetes 资源指标、对象、事件和旧版 `datakit/prom.instances` Pod Annotation 抓取子系统，不等同于 runtime 日志发现。
- runtime 列表、状态和日志路径的事实来源在 `internal/container/runtime/`；文件读取、offset、multiline 和 drain 语义的事实来源在 `internal/tailer/` 与 `internal/logtail/`。跨目录修改时必须一起审查和测试。

## 运行模式与所有权

- 普通主机/节点模式连接 Docker 或 CRI runtime，采集容器指标、对象和日志；在 Kubernetes 内还启动本节点 Pod watcher、日志 CRD watcher，以及独立的 Kubernetes 资源 collector。
- ECS Fargate 模式只采集 runtime 指标和对象，不启用本地文件日志 tailer。
- GCP Cloud API 模式使用 Cloud Monitoring/Logging；collector 受选举控制，其 Pod 元数据可能走直接 API provider，不应套用普通节点模式“完全不 GET”的假设。
- 同一次 `newContainerCollectors` 创建的 runtime collectors 共享一个 `containerLogCoordinator`。每个 collector 注册独立的容量为 1 的扫描信号 channel，`requestLoggingScan` 广播且合并尚未消费的信号。
- `containerTasks` 以 container ID 为 key；`taskMutex` 保护任务表，`containerLogTask.mu` 串行化单任务 reconciliation。不要在持有任务表锁时调用可能阻塞的 tailer 关闭逻辑。
- 包内存在 informer、metrics 注册表、`existingRuntimes` 和 goroutine group 等进程级状态。测试若替换这些对象必须恢复，未经审查不得 `t.Parallel()`。

## Kubernetes 元数据与 APIServer 约束

- 普通 Kubernetes 节点模式必须通过带 `spec.nodeName=<local-node>` field selector 的共享 Pod informer 获取元数据。不得恢复成每次容器指标、对象或日志扫描都对每个 Pod 发起 API `Get`；这会随容器数放大 APIServer QPS。
- informer 启动前的 RBAC 预检、初始 List 和持续 Watch 是预期请求。稳定运行时，runtime 扫描应读取本地 informer cache，而不是产生逐容器 APIServer 请求。
- `internal/kubernetes/client` 的 `LimitQPS=50`、`LimitBurst=50` 是每个 `Client` 各自的 token bucket，不是进程总额度。禁止在 scan、Pod 或 container 循环里新建 Client；这既绕过已有复用，也会按 Client 数放大全局 QPS。
- 必须区分三种状态：`podLabels == nil` 表示元数据未知；非 nil 空 map 表示 Pod 已获取但确实没有 label；非空 map 表示已知 label。CRD label selector 和 tag 补全依赖该区别。
- `not_synced`、`not_found` 和其他 cache error 都是暂时的元数据不可用，不等价于“Pod 没有 Annotation/CRD 配置”。已有日志任务遇到 `podMetadataUnavailable` 时必须保留上次成功的 `info`、有效配置和 tailer。
- 新容器第一次 cache miss 时仍允许利用 runtime 元数据和 `DATAKIT_LOGS_CONFIG`；无其他有效配置时可先建立默认 stdout 任务，待后续 cache hit 再补充 Pod 元数据和 reconcile。
- Pod 元数据缺失只影响 enrich/reconciliation，容器是否存活必须以成功的 runtime `ListContainers` 结果判断。runtime 列表失败时直接保留全部任务，绝不能用空 active set 清场。
- cache miss 日志应限速，metrics label 必须保持有限集合；不要把 namespace、Pod、container ID 等高基数字段加入内部 Prometheus metric label。

## 日志发现调度

默认 `logging_search_interval` 为 60 秒。`setup` 会调用 `config.ProtectedInterval`；默认开启 `ProtectMode` 时，指标、对象和日志周期被限制在 10 秒到 5 分钟，关闭保护模式时保留用户值。事件扫描与周期扫描都执行同一个全量 runtime reconciliation。

当前触发链如下：

1. `runLoggingDiscovery` 启动时立即执行一次 `gatherLogging("initial")`。
2. 周期 ticker 执行 `gatherLogging("ticker")`。
3. 本节点 Pod Add，以及 NodeName 改变或 phase 首次进入 Running 的 Update，分别安排 0、1、3 秒三次扫描请求。
4. Pod 进入 Terminating 或收到 Delete 时，延迟 1 秒请求扫描。UID 只用于队列标识和日志；事件不会按 UID 直接关闭任务。
5. CRD Add/Update/Delete 在处理后请求一次扫描。
6. 扫描信号会合并；0、1、3 秒的 retry queue item 按 attempt 而不是 Pod 区分，因此 Pod 突发也会在队列层合并。每个 runtime collector 的连续 Pod-event 扫描至少间隔 1 秒。不要通过扩大 worker 数、取消合并或恢复逐事件 API 查询来追求即时性。

Pod Annotation-only Update 会先进入 informer cache，但当前 `UpdateFunc` 不会因此立即请求日志扫描；新配置在下一次周期扫描或其他全量扫描时生效。用户文档所称热更新是“一个扫描周期内生效”，不是同步更新。

Docker 和 CRI 的 `ListContainers` 只返回 Running 容器。因此 0、1、3 秒重试只能提高发现概率，不能保证捕获运行窗口完全落在两次扫描之间的亚秒级容器。若从未有扫描看到该 container ID，就不会创建 tailer，后续删除扫描也没有可 drain 的任务；不得宣称亚秒 Pod 日志天然无损。

## 单次日志 reconciliation

`gatherLogging` 的顺序必须保持清晰：

1. 成功读取 runtime 当前 Running 容器；失败则返回且不清理任务。
2. 跳过 pause container，基于 runtime labels 建立 `containerLogInfo`。
3. 从 Pod metadata provider 合并 Annotation、Pod IP/labels/owner/image，并显式标记 metadata unavailable。
4. 计算 image/namespace 日志过滤结果，将容器及有效配置交给 coordinator。
5. 收集本次成功扫描看到的 active container IDs。
6. 全部容器处理完后，关闭 active set 中缺失的任务。

配置多个可用 runtime endpoint 时，当前会并行启动多个 collector，并共用同一个 coordinator。`cleanMissingContainerLog` 却只收到当前 collector 的 active set，因此一个 collector 可能误删另一个 runtime 的日志任务。这是当前可触发的已知缺陷，不是未来扩展才需要处理的问题；修复前不得把多 endpoint 视为日志采集安全的并行模式。

## 普通 Docker/CRI 日志配置优先级与规范化

以下规则仅适用于普通 Docker/CRI 本地日志路径。有效配置优先级从高到低为：

1. 容器环境变量 `DATAKIT_LOGS_CONFIG`。
2. 容器专属 Pod Annotation `datakit/<container-name>.logs`。
3. Pod 通用 Annotation `datakit/logs`。
4. 第一个匹配的 `ClusterLoggingConfig`；CRD 顺序必须保持稳定。
5. 通过 `container_include_log`/`container_exclude_log` 后的默认 stdout 配置。

显式 Env、Annotation 或匹配的 CRD 优先于默认日志过滤，并且整组覆盖低优先级来源，不做跨来源合并。需要同时采集 stdout 和容器内文件时，必须在同一个配置数组中声明。空字符串视为未配置。删除 Annotation 后通常会在下一次成功 reconciliation 回退到 CRD 或默认 stdout；若 fallback 又被过滤，当前存在下述旧 tailer 残留缺陷。

GCP Cloud Logging 不遵循上述优先级：它先执行 image/namespace include/exclude，只读取 Pod 通用和容器专属 Annotation，不读取 `DATAKIT_LOGS_CONFIG` 或 CRD；容器专属 Annotation 覆盖 Pod 通用 Annotation，`cloudStdoutLogConfig` 只选取数组中第一个 stdout-compatible 配置。

`newLogConfigs` 在正常路径上会把空 type 或 `stdout` 转为 runtime 类型和 runtime 日志路径，并补齐 source、容器/Pod/owner/global/label tags、多行选项及宿主机路径。多个配置解析为同一逻辑 path 时，后出现的配置获胜；修改该规则属于兼容性变更。

容器内文件路径的目标约束是：经 mount 映射或受约束的 merged rootfs fallback 转成宿主机路径，再加 `HOST_ROOT`/`/rootfs` 前缀，且不允许 `..` 逃逸 rootfs。当前实现尚未完整强制该约束，见下述已知边界。不要把容器内 path、宿主机 path 和 tailer recorder key 混为同一概念。

## 任务与 tailer reconciliation

| 条件 | 当前行为 |
| --- | --- |
| 新任务被过滤、配置 JSON 非法或全部 disable | 回滚尚未提交的任务；不留下活跃 tailer。 |
| 已有任务收到非法 JSON | 保留旧 config/tailer；不得先关闭再解析。 |
| 已有任务发生 Pod cache miss | 完整保留上次成功任务，不用空配置覆盖。 |
| metadata 从未知变为已知 | 刷新 `task.info`，重新计算 CRD、tags 和 config hash。 |
| enabled path 集合和 hash 均不变 | 不操作现有 tailer。 |
| enabled path 集合不变，但规范化 config hash 改变 | 对对应 path 调用异步 `Tailer.UpdateOptions`；不重开文件、不重置 offset。 |
| enabled path 集合变化，包括新增、删除或首次 disable 活跃 path | 当前实现关闭所有旧 tailer、清空切片，然后立即为所有 enabled path 重建。 |
| container ID 不再出现在一次成功的 runtime 扫描中 | 从任务表删除，并关闭其全部 tailer。 |

path 是任务内 tailer 的身份键。只改 source、service、pipeline、tags、编码、ANSI 或 multiline 等且 path 不变时应走 options 更新；`from_beginning` 对已经打开的文件不做回溯 seek，只影响以后打开的文件。

`Tailer.Close` 会先同步关闭 file watcher，再关闭 `shutdownChan`，但不等待 tailer 完全退出。正常 task removal 后，`Single` 在 context cancellation 路径中以 5 分钟上限执行 `readToEOF` 并 flush multiline 缓存；超时、读错误或全局 DataKit 退出可能提前结束。只有已经成功打开文件的 tailer 才有内容可 drain。

## 当前已知边界，不得固化为目标行为

- enabled path 集合只要变化，未变化的 path 也会被关闭并重建。旧 tailer 异步 drain，新 tailer 不等待其 recorder offset 落盘便可能打开相同文件，存在重复读取或边界不确定性。若修复，应按 path 做 retained/update、added/create、removed/close 的增量 diff，并设计明确交接。
- path 集合变化不是 failure-atomic：当前实现先关闭全部旧 tailer，再逐个创建新 tailer；任一新路径创建失败都可能留下部分配置甚至零 tailer，且无法自动恢复已关闭的旧任务。修复时应先验证/构建新状态，或提供明确 rollback/重试语义。
- 非 stdout 路径映射失败时，`newLogConfigs` 当前只记录 warning 并跳过该配置后续的 source、tags 和 multiline 补全，但不会从结果中移除该配置。`createTailer` 随后可能把原始 `cfg.Path` 加上 `HOST_ROOT`/`/rootfs` 后继续尝试，包含 `..` 时可能绕过预期 rootfs 边界。修复时应拒绝该配置，并保证旧任务不被破坏。
- 已有显式配置或 CRD 配置被删除后，如果容器又被 include/exclude 过滤，`addTask` 当前会在 reconciliation 前提前返回，旧 tailer 可能继续运行。过滤只应决定是否建立默认 stdout，不应让已失效的旧配置永久残留。
- `Tailer.UpdateOptions` 和 `Single.UpdateOptions` 使用有限容量的非阻塞 channel。channel 满时更新会被丢弃，但当前 coordinator 仍会推进 config hash，后续相同配置不会自动重试。
- `Single.applyOptions` 会重建 multiline parser，未先 flush 旧 parser 的待聚合内容；任意 options 热更新撞上未 flush 多行缓存时存在丢失风险。
- Annotation-only Update 不触发即时扫描；亚秒级 Running 窗口可能完全避开 runtime 扫描。
- 同一 Pod 内 container restart 时 Pod phase 可能一直是 Running，当前 Pod Update 条件不一定发出扫描请求，新 container ID 可能要等周期或其他事件扫描。
- Pod metadata 按 namespace/name 查询，当前未核对 cache 中 Pod UID 与 runtime label 的 Pod UID；同名 Pod 快速重建时，旧容器理论上可能短暂关联到新 Pod 元数据。
- 空数组 `[]` 和 `null` 当前等价于全部 disable；`[null]` 会在配置补全时触发 nil 解引用。修改解析逻辑时必须把这些输入作为回归用例，不能让用户配置导致采集器 panic。
- 上述风险必须在变更说明和测试结论中如实区分。不得用提高 APIServer QPS、无限提高扫描频率、绕过 channel 上限或取消资源边界的方式掩盖日志完整性问题。

## 并发、关闭与数据边界

- 新 goroutine 必须服从 `datakit.Exit` 或显式 context/channel，并由对应 group 等待；watcher 的 queue、informer、stop channel 和 runtime client 必须有单一所有者且可重复安全地收尾。
- `requestLoggingScan` 只表达“需要重新扫描”，不携带 Pod 快照。不要依赖事件对象仍存在，也不要把多个触发信号误认为必须执行相同次数的扫描。
- 任务创建失败必须可在下一次扫描重试，且不得留下不可重试或状态错位的占位任务。过滤、解析失败和全部 disable 的新任务应回滚；tailer 创建失败若保留空任务，后续扫描必须能够再次创建。对于已有任务，“失败不破坏最后一次可用状态”是修改目标；当前 path 全量重建尚不满足该目标。
- 采集数据继续经过现有 `io.Feeder`，保留 category、source、election、pipeline、global tags 和 point 所有权语义。
- measurement、字段/tag 类型、Annotation/Env/TOML key、默认值、过滤优先级和 metrics 名称/label 都是兼容性接口；变更时同步 sample、measurement、环境变量说明及中英文文档。
- 不得在日志中输出完整 Annotation、Env 日志配置、Authorization header、token、证书或原始日志载荷；诊断信息只记录必要的资源标识和有界错误原因。
- 修改 Kubernetes object/change 序列化时必须审查 point 的 `message`、`yaml`、`annotations` 和 diff。当前部分 object `message` 在删除临时字段前构造，可能包含 literal Env 或 Annotation 中的认证/header 配置；不得无意扩大敏感值暴露，若实施脱敏必须评估兼容性并补测试和文档。

## 可观测性

涉及日志发现或 Pod cache 的改动至少检查以下内部指标：

- `datakit_kubernetes_apiserver_requests_total{component,verb,resource,code}`：普通节点日志/metadata watcher 使用 `component="container_runtime"`；用 counter 增量计算 QPS，不看累计值猜测速率。
- `datakit_input_container_logging_discovery_cost_seconds{trigger}`：`trigger` 仅使用 `initial`、`ticker`、`pod-event` 等有限值。
- `datakit_input_container_logging_discovery_schedule_delay_seconds`：观察周期扫描排队/阻塞。
- `datakit_input_container_pod_cache_miss_total{usage,reason}`：`usage` 为 logging/metric/object，`reason` 为 not_synced/not_found/error。删除阶段的 not_found 可能是正常竞争，不能单凭此指标判定丢日志。
- `datakit_input_container_collect_cost_seconds{category}` 与 `datakit_input_container_collect_pts_total{category}`：确认指标/对象采集没有回归。
- `kubernetes/metrics.go` 中 legacy Pod Annotation Prometheus 的 active task、inflight scrape 和 result counter 必须维持 bounded labels，并在取消、超时和 panic 路径回收 gauge。

## 测试与验证

从相关包开始，使用仓库要求的 vendor、CGO 和无文件日志设置：

```sh
UT_EXCLUDE_INTEGRATION_TESTING=on GO111MODULE=on GOFLAGS=-mod=vendor \
  CGO_ENABLED=1 LOGGER_PATH=nul \
  go test -count=1 -timeout 1h ./internal/plugins/inputs/container/...
```

按改动边界补充运行：

- runtime 列表、状态、重连或日志路径：`./internal/container/runtime`。
- tailer options、offset、rotate、close 或 drain：`./internal/tailer` 及相关 `internal/logtail/...`。
- Kubernetes client/informer 或 APIServer metrics：`./internal/kubernetes/client`。
- mutex、channel、watcher、manager 或 shutdown 变更：对最小相关包再运行 `go test -race`；不要并行运行会替换全局 metrics/registry 的测试。

按改动影响从以下日志生命周期矩阵选择用例；触及对应行为时必须覆盖：新任务、重复扫描幂等、cache 未同步、cache not_found、已知空 labels、后续 metadata enrich、Pod UID 不匹配、Annotation 非法/删除/热更新、Env 优先级、CRD selector、filter、显式配置删除后 fallback 被过滤、`[]`/`null`/`[null]`、全部 disable、同 path options 更新、path 新增/删除、路径映射失败、`..` 逃逸、多 runtime endpoint、runtime List 失败、container 消失和并发 add/remove。修复已知缺陷时应断言目标安全行为，不要把缺陷快照固化成契约。涉及关闭时要断言 drain 后的精确日志集合，不只断言 tailer 数量。

需要 Kubernetes/Docker 运行验证时：

1. 复制部署 YAML 到临时文件，不覆盖用户原文件；镜像使用唯一测试 tag。
2. DataWay 指向本地可计数的 HTTP sink，禁止把验证数据发到真实中心。
3. 分别验证默认 stdout、Pod Annotation、容器专属 Annotation，以及创建/删除压力；每条日志携带 Pod/container/sequence 唯一键。
4. 等待 bounded drain 窗口后比较期望与实际唯一键集合，同时报告 missing 和 duplicate，不能只比较总条数。
5. 用上述 metrics 的测试前后增量报告 APIServer QPS、scan cost/delay、cache miss 和 collector/tailer 健康状态。
6. 清理 workload、临时 sink、测试镜像引用和 YAML；不要执行发布、推送或通知目标。

亚秒级 Pod 测试只能证明当前环境和样本下的结果。应分别报告“容器是否曾被 runtime 扫描发现”“tailer 是否成功打开”“关闭是否 drain 完成”，不能把一次零丢失结果推广为确定性保证。

## 完成检查

- 只修改预期源码、测试、文档和必要生成文件；保留用户已有工作区内容。
- Go 文件只格式化本次修改范围；测试恢复所有被替换的全局状态。
- `git diff --check` 通过，相关最小测试通过；明确列出未运行的 Docker、Kubernetes、CGO、race、跨平台和全仓库检查。
- 若行为变化影响本文件中的流程、优先级、不变量、已知边界或验证方式，同一个变更必须更新本文件。
