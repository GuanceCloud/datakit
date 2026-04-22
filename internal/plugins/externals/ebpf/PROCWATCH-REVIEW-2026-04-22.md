# Procwatch Review Notes - 2026-04-22

日期：2026-04-22

## 背景

本轮继续对 `internal/plugins/externals/ebpf/internal/procwatch` 做多轮 review，重点覆盖：

- `/proc` 读取与权限边界
- mount namespace / overlayfs / mountinfo 路径解析
- `l7flow` 与 `procwatch` 的集成一致性
- attach/detach 状态机
- PID reuse / stale attach 请求

## 本轮结论

### 1. `l7flow` 的 procwatch 接入已经恢复为按需解析

之前 `l7flow` 在 catalog cache miss 时只做 `ResolveLater()`，会导致新进程首批事件拿不到：

- `ProcessName`
- `ServiceName`
- `collectable` / blacklist 即时判定

现在改为同步 `LookupOrResolve()`：

- cache miss 时立即解析当前 pid
- 首批 L7 事件即可拿到服务信息
- blacklist 也能在首批事件上立即生效

相关文件：

- `internal/l7flow/net_tracer.go`
- `internal/l7flow/net_tracer_test.go`

### 2. `procwatch` 的 mount/path parser 之前对 namespace/overlayfs 不够稳

review 发现三处具体问题：

1. `parseMountInfoLine()` 没有还原 kernel `mountinfo` 的 octal escape
2. mountpoint 匹配使用纯 `strings.HasPrefix()`，`/foo` 会误匹配 `/foo-bar`
3. `/proc/<pid>/maps` 的 pathname 解析会丢掉路径中的空格

本轮已修正：

- mountinfo `root` / `mountPoint` 改为做 proc-style unescape
- mountpoint 匹配改为目录边界匹配
- `/proc/<pid>/maps` 改为保留第 6 列之后的整段 pathname

相关文件：

- `internal/procwatch/procfs_linux.go`
- `internal/procwatch/procfs_linux_test.go`

### 3. blacklist 语义已收紧为硬拒绝

`catalog.create()` 旧逻辑里，`trace_name_blacklist` / `trace_env_blacklist` 把当前进程判成 `collectable=false` 之后，
还会被后面的 parent inheritance 翻回去：

- 如果父进程是 collectable
- 子进程 blacklist 命中仍可能被放回 collectable

同时 `kernelFilter()` 在 parent inheritance 之前触发，用户态和内核态状态可能不一致。

本轮已修正：

- blacklist 命中作为 `hardDeny`
- parent inheritance 只用于非 hard deny 场景
- `kernelFilter()` 改为看最终 collectable 状态

相关文件：

- `internal/procwatch/catalog.go`
- `internal/procwatch/catalog_test.go`

### 4. attach 队列已有明显 PID reuse 风险

旧实现里 attach 队列只按 `pid` 排队：

- `sched_process_exec` 到 `attachLoop.drain()` 之间存在窗口
- 窗口内旧进程退出且 pid 被复用后
- 队列里的旧请求可能 attach 到新进程

本轮先做了一层低风险收口：

- attach request 改为携带 `pid + start_time`
- drain 前重新读取当前 `/proc/<pid>/stat`
- 如果当前 `start_time` 与排队时不一致，则直接跳过 attach

随后又继续收紧了一层：

- enqueue attach 时额外打开稳定的 `/proc/<pid>` 目录 fd
- drain 时优先通过这个 procfd 读取旧进程自己的 `stat`
- 如果旧进程已经退出，则旧 procfd 不会误读到复用 PID 的新进程
- 这进一步降低了 stale attach request 命中新进程的概率

本轮又把这条线继续推进到解析阶段：

- `Catalog` 新增基于 procfd 的 `ResolveWithProcFD()`
- attach 热路径在通过 procfd 身份校验后，不再退回按 pid 重新读 `/proc`
- 而是继续通过同一个 procfd 读取：
  - `stat`
  - `environ`
  - `exe`
  - `root`
  - `cmdline`

这样把“校验通过后又按 pid 重新读 `/proc`”这段剩余竞态也再压缩了一层。

这不是最终形态，但能直接缩小 stale request 命中新进程的概率。

相关文件：

- `internal/procwatch/runtime.go`
- `internal/procwatch/catalog.go`
- `internal/procwatch/procfs_linux.go`
- `internal/procwatch/runtime_test.go`

### 5. `resolveHostBinaryPath()` 已增加 `/proc/<pid>/root` 优先路径

旧实现只做 mount namespace 维度的路径翻译：

- 能覆盖很多 container mount 场景
- 但对同 mount namespace、不同 per-process root/chroot 的情况不够稳

本轮增加了一层保守 fallback：

- 先读 `/proc/<pid>/root`
- 尝试直接拼出目标进程视角下的二进制路径
- 只有命中常规文件时才直接采用
- 否则继续走原有 mountinfo resolver

这样不会破坏现有路径翻译，但能优先收敛：

- chroot
- per-process rootfs
- root 已可见、无需 mountinfo 反推的场景

相关文件：

- `internal/procwatch/helpers.go`
- `internal/procwatch/procfs_linux.go`
- `internal/procwatch/procfs_linux_test.go`

## 已执行验证

```bash
go test ./internal/plugins/externals/ebpf/internal/procwatch \
  ./internal/plugins/externals/ebpf/internal/l7flow \
  ./internal/plugins/externals/ebpf/internal/netflow
```

## 运行时验证补充

### container exit repro 第一轮结论

2026-04-22 在本地 VM `ebpf-stress-lab` 上执行：

```bash
cd internal/plugins/externals/ebpf
ITERATIONS=8 CONTAINER_TIMEOUT=30 KEEP_REPRO_ARTIFACTS=1 bash ./repro-procwatch-container-exit.sh
```

结果不是 hook 泄漏，而是 repro harness 自身漏了动态 uprobe 开关：

- `AddHooK count: 0`
- `DetachHook count: 0`
- eBPF 日志明确提示：
  - `ebpf-trace target-process uprobe attach is disabled by default; set enable_uprobe/--trace-uprobe to turn it on`

这说明第一轮失败不能用来判断 `procwatch` attach/detach 状态机是否正确，
只能说明旧 repro 脚本没有真正覆盖目标路径。

### 已跟进修正

- `repro-procwatch-container-exit.sh` 已显式加入 `--trace-uprobe`
- 脚本现在会在日志中检测 `target-process uprobe attach is disabled by default`
- 如果仍命中该提示，会直接失败并给出明确诊断

### container exit repro 第二轮结论

补上 `--trace-uprobe` 后再次执行，仍未进入动态 attach 路径。
日志继续给出更具体的原因：

- `ebpf-trace enabled without trace_server; target-process attach is disabled by default for safety`

这说明当前 runtime 安全策略要求：

- `--trace-uprobe`
- 非空 `--trace-server`
- 明确 allowlist 或 `trace_all_proc`

三者同时满足，目标进程 attach 才会真正开启。

### 已继续修正

- repro 脚本已进一步显式加入 `--trace-server`
- 失败诊断文案也改成同时检查：
  - `--trace-uprobe`
  - `--trace-server`
  - trace allowlist

### container exit repro 第三到第五轮结论

继续在 VM 上复跑后，最终定位到两个额外事实：

1. `procwatch` 用户态事件常量和内核侧 `REC_SCHED_*` 位定义不一致
2. repro 脚本过早核对 `DetachHook`，没有考虑 `procwatchDetachGracePeriod = 30s`

#### 事件位定义错位

C 侧：

- `REC_SCHED_FORK = 1`
- `REC_SCHED_EXEC = 2`
- `REC_SCHED_EXIT = 4`

Go 侧旧代码却把：

- `eventExec = 1`
- `eventExit = 2`

这会导致用户态把真正的 `exec` 事件按 `exit` 处理，
从而让新进程永远不会进入 attach 队列。

修正后第五轮复现已能看到：

- `procwatch exec event pid=...`
- `AddHooK: ... smrepro ...`

#### detach 宽限期导致的假阳性

第五轮日志显示：

- `AddHooK` 出现在 `03:47:37`
- 对应 `DetachHook` 出现在 `03:48:16`

这与 `procwatchDetachGracePeriod = 30s` 的设计一致，
说明脚本“容器跑完立即比对 attach/detach 计数”的策略会把正常的延迟 detach 误判成泄漏。

### 已继续修正

- `runtime.go` 中的 `eventFork/eventExec/eventExit` 位定义已对齐 C 侧
- 新增回归测试，确保 `exec` 事件会入 attach 队列、`exit` 事件不会误入
- repro 脚本改为在 `DETACH_SETTLE_TIMEOUT` 内轮询等待 hook 计数收敛后再判定失败

### container exit repro 第六轮结果

在补齐：

- 事件位定义修正
- `--trace-uprobe`
- `--trace-server`
- detach settle 等待

之后，VM 上执行：

```bash
cd internal/plugins/externals/ebpf
ITERATIONS=4 CONTAINER_TIMEOUT=30 DETACH_SETTLE_TIMEOUT=50 KEEP_REPRO_ARTIFACTS=1 bash ./repro-procwatch-container-exit.sh
```

结果通过：

- `AddHooK count: 1`
- `DetachHook count: 1`
- `procwatch repro passed`

这说明当前 `smrepro` 这条短命 Go 进程动态 uprobe 生命周期，至少在当前 VM 场景里已经能完成：

- 发现 exec
- 动态 attach
- 经宽限期后 detach 回收

但从日志细节看，这一轮命中的 `AddHooK` 来自本地 `/tmp/.../smrepro`，
容器内那几次超短命实例大多仍然在 attach drain 前就已经退出，
表现为大量：

- `skip stale procwatch attach for pid ...: no such process`

所以当前结论是：

- 本地短命 Go 进程的动态 uprobe 生命周期已跑通
- “容器内超短命实例也能稳定命中 attach” 这件事仍然没有被充分证明

这部分如果继续做，下一步更值得投入的是：

- 缩短 attach drain 延迟
- 或在 fork/exec 更早阶段排队
- 或把 repro workload 改成稍长寿命、但仍保持容器退出路径

### 后续继续优化与结果

为继续压缩容器短命进程的 miss 窗口，又做了三轮实改：

1. `sched_process_fork` 事件也进入 attach 队列，不再只依赖 `exec`
2. `attachLoop` 的 drain 周期从 `250ms` 压到 `50ms`
3. 对命中 `trace_name_list` 的 `exec` 事件走优先 attach drain

这几步之后，日志里已经能稳定看到容器内：

- `procwatch exec event pid=... name=smrepro`

但当时仍然失败在：

- `skip procwatch attach for pid ... (smrepro): traceable=false bin="" collectable=true`

继续下钻后又定位到两个容器路径问题：

1. `readProcessExePath*()` 过早要求 `/proc/<pid>/exe` 读出的路径必须能在宿主直接 `stat`
2. 当前 VM 上 `readlink /proc/<pid>/root` 只返回 `/`，但 `/proc/<pid>/root/<path>` 这个 proc portal 实际可访问容器内真实文件

对应修正：

- `normalizeProcessLinkTarget()` 改成只做 procfs 语义校验，不在 exe 读取阶段要求宿主 regular file
- `resolveHostBinaryPath*()` 增加 proc root portal 优先路径：
  - `/proc/<pid>/root/<path>`
  - `/proc/self/fd/<procfd>/root/<path>`

### 最新 VM 结果

继续在 `ebpf-stress-lab` 上执行：

```bash
cd internal/plugins/externals/ebpf
ITERATIONS=8 CONTAINER_TIMEOUT=30 DETACH_SETTLE_TIMEOUT=50 KEEP_REPRO_ARTIFACTS=1 bash ./repro-procwatch-container-exit.sh
```

结果：

- `AddHooK count: 9`
- `DetachHook count: 9`
- `procwatch repro passed`

这次 `AddHooK` 已不再只来自本地 `/tmp/.../smrepro`，
日志里可以看到多次容器 attach：

- `/proc/self/fd/<n>/root/usr/local/bin/smrepro`

说明当前 VM 场景下，容器内短命 Go 进程的动态 uprobe 生命周期也已经被打通。

### 还剩的风险

虽然容器 attach 已经跑通，但当前解析出来的二进制路径仍然偏向 proc portal 形式：

- `/proc/self/fd/<n>/root/usr/local/bin/smrepro`

这能工作，但不是最稳定的宿主 canonical path。
它的副作用是：

- 二进制 identity / dedupe 可能仍不够理想
- 同一镜像跨实例的 attach 复用机会可能偏少

## 继续收敛 Canonical Path

后续继续把容器路径从 proc portal 形式收敛到 overlayfs 宿主真实路径。

### 发现

在上一轮虽然容器 attach 已经跑通，但 `AddHooK` 里的路径仍然是：

- `/proc/self/fd/<n>/root/usr/local/bin/smrepro`

这会导致不同实例的同一二进制无法稳定复用 identity。

继续下钻后又定位到两处根因：

1. `readProcessExePath*()` 过早要求 `/proc/<pid>/exe` 目标必须能被宿主直接 `stat`
2. `mountinfo` resolver 之前只处理“设备号可映射到宿主 mount”场景，没有真正消费 overlayfs 的 `upperdir/lowerdir`

### 已修正

- `normalizeProcessLinkTarget()` 改成只做 procfs 语义校验，不在 exe 读取阶段要求宿主 regular file
- `mountinfo` 解析补充了：
  - `fsType`
  - `source`
  - `super options`
- `resolveMountPath()` 现在会在 overlay mount 上优先尝试：
  - `upperdir/<relpath>`
  - `lowerdir/<relpath>`
- `resolveHostBinaryPath*()` 先尝试 mountinfo/overlayfs canonical path，再回退到 proc root portal

### 最新结果

继续在 VM 上执行：

```bash
cd internal/plugins/externals/ebpf
ITERATIONS=8 CONTAINER_TIMEOUT=30 DETACH_SETTLE_TIMEOUT=50 KEEP_REPRO_ARTIFACTS=1 bash ./repro-procwatch-container-exit.sh
```

结果：

- `AddHooK count: 2`
- `DetachHook count: 2`
- `procwatch repro passed`

这里计数从之前的 `9/9` 收敛成 `2/2`，不是回退，而是 identity 变稳定了：

- 本地临时二进制收敛成 1 份
- 容器 overlayfs 二进制也收敛成 1 份

日志里容器 attach 路径已经变成稳定宿主层文件：

- `/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/214/fs/usr/local/bin/smrepro`

这说明当前 VM 场景下，容器内短命 Go 进程的动态 uprobe 生命周期和二进制 canonical path 都已经基本跑通。

### 下一步如果继续

最值得继续的是两件事：

1. 把 overlayfs `lowerdir+` / 更复杂转义场景补成专门测试
2. 进一步把 `snapshot id` 级路径收敛到镜像层/内容 hash 级 identity，减少跨实例重复 attach

## overlayfs 转义补强

最后又补了一轮 `overlayfs` parser review，重点看 `mountinfo` 里的：

- `lowerdir`
- `lowerdir+`
- 带 proc-style octal escape 的层路径

### 发现

之前虽然已经支持 overlayfs `upperdir/lowerdir`，但层路径拆分仍然有一个边角：

- 如果原始层路径里包含被转义的 `:`
- 例如 `\072`
- 直接按 `:` split 会把单个层路径错误拆成两段

这会导致：

- overlay lower layer 命中失败
- canonical path 回退到 proc portal 形式
- 跨实例 identity 稳定性再次变差

### 已修正

- 增加 `splitEscapedMountPathList()`
- 拆分 `lowerdir/lowerdir+` 时跳过 `\ddd` 形式的 octal escape
- 每段 layer path 再做 proc-style unescape
- 补了专门测试覆盖：
  - `\072` 转义分隔符保留
  - 带转义层路径的 overlay lowerdir 命中

### 快速回归

本地继续执行：

```bash
go test ./internal/plugins/externals/ebpf/internal/procwatch \
  ./internal/plugins/externals/ebpf/internal/l7flow \
  ./internal/plugins/externals/ebpf/internal/netflow
```

并在 VM 上做了一次快速复验：

```bash
cd internal/plugins/externals/ebpf
ITERATIONS=2 CONTAINER_TIMEOUT=30 DETACH_SETTLE_TIMEOUT=50 KEEP_REPRO_ARTIFACTS=1 bash ./repro-procwatch-container-exit.sh
```

结果仍通过：

- `AddHooK count: 2`
- `DetachHook count: 2`
- `procwatch repro passed`

这说明 overlayfs 转义补强之后，没有把前一轮已经收敛好的 canonical path / attach-detect 流程打回去。

## 外部参考

这些资料与本轮 review 直接相关：

- `/proc` 权限、`/proc/<pid>` 语义、PID reuse 与 procfd 行为
  - https://docs.kernel.org/filesystems/proc.html
- mount namespace 与 `/proc/<pid>/mountinfo`
  - https://man7.org/linux/man-pages/man7/mount_namespaces.7.html
- `mountinfo` 字段定义
  - https://man7.org/linux/man-pages/man5/proc_pid_mountinfo.5.html
- `/proc/<pid>/root` 提供的是目标进程的 filesystem view，而不只是普通 symlink
  - https://man7.org/linux/man-pages/man5/proc_pid_root.5.html
- `/proc/<pid>/exe` 的 ` (deleted)` 语义
  - https://man7.org/linux/man-pages/man5/proc_pid_exe.5.html
- `/proc/<pid>/maps` pathname 只有换行会做 octal escape，空格不会
  - https://man7.org/linux/man-pages/man5/proc_pid_maps.5.html
- overlayfs 在 `mountinfo` 中的 lowerdir/path escaping 行为
  - https://www.kernel.org/doc/html/latest/filesystems/overlayfs.html

## 下一阶段建议

### 第一优先级

- 把当前 `pid + start_time` 收口继续升级成更稳定的进程 identity
- 优先评估：
  - `pidfd_open()`
  - 打开 `/proc/<pid>` 获得稳定 procfd
  - 或在内核事件里补稳定 identity

### 第二优先级

- 把 `resolveHostBinaryPath()` 从“只按 mount namespace 解析”升级成“两级解析”
- 优先级顺序建议：
  1. 先尝试 `/proc/<pid>/root`
  2. 再 fallback 到当前 mountinfo resolver

这会更稳地覆盖：

- chroot
- container rootfs
- overlayfs
- bind mount 混合场景

### 第三优先级

- 增加针对复杂 namespace/rootfs 场景的 fixture 或 repro：
  - overlayfs lowerdir/lowerdir+ escape
  - bind mount
  - 同 mount namespace 不同 root
  - deleted exe
  - hidepid / ptrace denied
  - 短命 PID reuse
