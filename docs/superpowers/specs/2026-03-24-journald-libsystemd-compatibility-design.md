# Journald libsystemd 兼容性检测设计

## 目标

为 journald external collector 增加启动期检测能力，使其能够识别 DataKit 容器内运行时 `libsystemd` 无法读取目标主机 journald 文件的情况，输出明确告警，并避免进入“看起来正常启动，但实际上没有采集任何日志”的误导性状态。同时提供一个显式配置项，用于在启动 external collector 之前，从 `/rootfs` 下无条件复制指定的宿主机动态库到 DataKit 自己的运行目录，并让 journald collector 优先从该目录加载 `libsystemd`。

## 问题

当前 `internal/plugins/externals/journald/collect/` 下的采集器会先打开配置的 journal 路径，只要 `sdjournal` 接受该打开操作，就继续进入正常的读取循环。在已复现的 EKS 场景里，这会产生假阳性启动：

- DataKit DaemonSet Pod 运行时使用的是 `systemd/libsystemd 249`
- Amazon Linux 2023 EKS 节点写入 journal 时使用的是 `systemd 252`
- 主机上的 journal 文件本身是有效的，且在节点宿主机上可正常读取
- 如果 Pod 内部安装了 `journalctl`，对同一批 journal 文件执行 `journalctl` 可能会报 `unsupported feature`
- 采集器日志显示启动成功，但实际始终采集不到任何 entry

这本质上是 journal 文件格式与用户态运行时兼容性问题，不是内核版本问题，也不主要是路径解析问题。

## 非目标

- 自动下载或注入更新版本的 `libsystemd`
- 在运行时不兼容的情况下强行读取 journal
- 在 CI 中增加依赖真实 AL2023 journal 文件的环境相关集成测试
- 顺带重构与本问题无关的 journald 采集逻辑
- 在配置项关闭时自动猜测是否应该复制宿主机动态库

## 面向用户的结果

当 journald collector 启动时遇到不兼容的主机 journal 文件，运维人员应当看到一条高信号告警，说明：

- 当前 collector 运行时无法读取目标 journal 文件格式
- 触发诊断的配置路径以及采样到的具体目标
- 如果可获取，检测到的 reader 版本
- collector 将保持非激活状态
- 推荐的修复方向，例如使用更新版本的 bundled `libsystemd`

collector 不应崩溃退出，但也不能继续伪装成“已经正常开始采集”。

## 备选方案

### 方案 1：只做版本检测

检测当前运行时 `libsystemd` 或 `journalctl` 的版本，如果低于某个阈值则告警。

优点：

- 实现简单
- 启动开销低
- 日志解释直观

缺点：

- 版本差异只能作为启发式判断
- 真正的兼容性边界取决于 journal 文件特性，而不是纯版本号
- 可能出现误报或漏报

### 方案 2：只做探测

先解析出实际 journal 目标，然后使用与正常采集相同的 `sdjournal` 运行时栈，对目标执行一次最小化 open/read 探测。

优点：

- 直接覆盖真实失败路径
- 对当前 EKS 场景最准确
- 即使无法拿到版本信息，也能识别 unsupported-format

缺点：

- 需要额外的目标展开和错误分类逻辑
- 如果不额外补充，日志中缺少版本上下文

### 方案 3：版本信息加探测的混合方案

在可行时收集运行时版本信息，但以实际探测结果来决定 collector 行为。

优点：

- 行为判断准确，因为探测结果才是最终依据
- 日志包含版本上下文，排障效率更高
- 能直接解决“静默零采集”的核心问题

缺点：

- 代码量略高于前两种单一方案

## 推荐方案

采用混合方案。

collector 在启动时先根据配置决定是否执行一次显式的 host library prepare；如果开启，则无条件从 `/rootfs` 下复制配置指定的动态库文件列表到 DataKit 自己的 `external-libs` 目录，并让 journald external collector 优先从该目录加载 `libsystemd`。之后再尽可能获取版本上下文，并针对一个或多个具体 journal 目标执行轻量级兼容性探测。是否继续启动采集，以探测结果为准。如果识别出 unsupported format 或同类不兼容问题，则记录告警并保持 inactive。

## 设计概览

变更范围限定在 external journald collector，即 `internal/plugins/externals/journald/collect/`。设计中引入三个清晰职责：

### 0. Host library prepare 配置

在 `inputs.journald` 中新增显式配置项：

- `copy_node_libs`：`bool`，默认 `false`
- `copy_node_libs_files`：`[]string`，支持用户显式覆盖；如果未填写或为空，则程序必须回退到一组内置默认动态库文件名或 glob 列表

其语义必须明确如下：

- 这是一个显式 opt-in 行为
- 一旦 `copy_node_libs = true`，启动 external collector 前就从 `/rootfs` 的候选系统库目录中无条件复制 `copy_node_libs_files` 指定的动态库
- 如果 `copy_node_libs_files` 未填写或为空，则使用程序内置默认列表执行 copy
- copy 阶段本身不做兼容性判断，也不根据 probe 结果决定是否 copy
- copy 的目标目录固定为 DataKit 自己可管理的 `external-libs` 目录，并由 collector 启动时把 `LD_LIBRARY_PATH` 优先指向这里
- 最终是否可采集仍由后续 probe 决定，而不是由 copy 行为本身决定

`copy_node_libs_files` 的设计目标是给运维保留白名单调整能力，以便在不同发行版或 systemd 依赖组合下补充需要一起复制的动态库。对应的 sample 配置中应直接给出默认文件列表示例，避免用户在需要覆盖时不知道应该从哪些库开始填写。

首版内置默认列表应至少包含如下文件名或 glob：

- `libsystemd.so*`
- `liblz4.so*`
- `libzstd.so*`
- `liblzma.so*`
- `libcap.so*`
- `libgcrypt.so*`
- `libgpg-error.so*`
- `libselinux.so*`
- `libmount.so*`
- `libblkid.so*`
- `libacl.so*`
- `libpcre2-8.so*`

这组默认列表的定位是“首版起始白名单”，来源应以 `libsystemd.so` 在常见 Linux 发行版上的依赖树为基础，并允许用户通过 `copy_node_libs_files` 显式覆盖或补充。实现与文档都应避免把这组默认列表描述成对所有 node 发行版都完全充分的最终集合。

### 1. Journal 目标解析

把配置的 `paths` 展开为可探测目标。解析器需要支持：

- journal 根目录，例如 `/var/log/journal`、`/run/log/journal`、`/rootfs/var/log/journal`、`/rootfs/run/log/journal`
- 根目录下的 machine-id 子目录
- 直接配置的 machine-id 目录
- 直接配置的 `*.journal` 文件

解析结果应当是一个小而有序的具体目标集合，便于诊断与探测。

### 2. 兼容性探测

在进入正常采集循环之前，collector 应先对解析出的代表性目标执行一次轻量级 open/read 探测，并且必须走与正式采集相同的 `sdjournal` 运行时路径。

探测至少需要回答以下问题：

- 是否找到 journal 目标
- 是否能成功打开
- 是否能读取到至少一条 entry，或得到有意义的 journal 状态
- 底层库是否返回了不兼容或 unsupported format

探测结果应通过显式结构体表达，而不是在调用点临时解析字符串。

### 3. 启动诊断与降级

把探测结果映射为面向运维的有限原因集合：

- `unsupported-format`
- `permission-denied`
- `no-journal-files`
- `unexpected-open-error`
- `ok`

只要结果不是 `ok`，就输出一条显著告警，并且不要启动实际采集循环。这样 collector 会进入“有告警的降级状态”，而不是“看起来成功但没有任何结果”的误导性状态。

## 详细流程

collector 启动时：

1. 与现在一样解析配置。
2. 如果 `copy_node_libs = true`，则从 `/rootfs` 的候选系统库目录中无条件复制 `copy_node_libs_files` 指定的动态库到 `external-libs` 目录。
3. 启动 journald external collector 时，把 `LD_LIBRARY_PATH` 优先指向 `external-libs` 目录。
4. 将配置路径展开为候选 journal 目标。
5. 可选地收集运行时版本信息，用于诊断日志。
6. 对一个或多个候选目标执行兼容性探测。
7. 如果探测结果是 `ok`，继续进入正常的 journal 初始化和采集流程。
8. 如果探测结果是分类后的失败，输出一条醒目的告警，并在不进入读取循环的情况下返回。

在探测成功后，正常运行期采集行为保持不变。

## 多目标探测策略

collector 支持多个配置路径，因此必须明确多目标探测聚合策略。

策略如下：

- 按解析后路径列表的确定性顺序进行探测
- 探测目标必须与正式采集依赖的具体目标类型一致，不能用无关的虚拟代表目标替代
- 一旦任意一个目标探测成功，就停止后续探测
- 只要至少一个目标兼容，collector 就继续启动，不因其他目标失败而整体禁用
- 只有当全部目标都未成功时，collector 才保持 inactive

当没有任何目标成功时，collector 应从所有失败结果中选出“信息量最高”的失败原因。优先级顺序如下：

1. `unsupported-format`
2. `permission-denied`
3. `unexpected-open-error`
4. `no-journal-files`

如果存在多个相同严重级别的失败，则告警中的 `sampled target` 应选择该严重级别中、按探测顺序出现的第一个具体目标。这样既能保证实现和测试具备确定性，也避免在部分路径健康时出现误关停。

## 日志要求

告警日志应至少包含：

- 配置路径
- sampled target 路径
- 诊断原因
- 如果可获取，检测到的 reader 版本
- 如有价值，原始错误信息
- 明确说明 collector 将保持 inactive
- 修复建议

日志形态示例：

```text
journald compatibility check failed: reason=unsupported-format reader_version=249 target=/rootfs/var/log/journal/.../system.journal message="host journal requires newer libsystemd; collector will stay inactive"
```

应避免重复刷屏。一次启动过程中，只输出一次主诊断告警即可。

## 运行时版本检测

版本检测只是补充信息，不应替代真实探测。

可接受的信息来源，优先级如下：

- 如果存在稳定机制，优先读取运行时库元数据
- 如果环境中存在且调用成本低，可执行 `journalctl --version`
- 在 Kubernetes 容器环境中，不应假设 `journalctl` 一定安装；如果不存在，则版本信息可以为空，且不能影响 probe 和最终启动决策
- 如果以上都不可用，则不输出版本信息

即使版本检测失败，collector 也仍然要继续执行探测，并基于分类结果做出行为决策。

## 错误处理

### Unsupported journal format

- 记录 warning，而不是 fatal
- 不进入采集循环
- 日志中给出修复建议，例如挂载宿主机上兼容的 `systemd` 相关库到独立目录，并让 collector 优先从该目录加载 `libsystemd`
- 同时明确说明：宿主机上的 `libsystemd` 也只是兼容性候选，不保证一定兼容当前 DataKit 使用的 journald external binary
- 如果启用了 `copy_node_libs`，日志中还应明确说明 copy 行为只是前置准备，不代表复制后的库一定兼容

### Permission denied

- 记录包含目标路径的 warning
- 不进入采集循环
- 提示应聚焦在文件访问权限，而不是格式不兼容

### No journal files found

- 记录配置或路径相关 warning
- 不进入采集循环

### Unexpected open 或 probe error

- 记录包含原始错误的 warning
- 不进入采集循环

## 打包与运行时指引

产品文档中应明确说明：把 `LD_LIBRARY_PATH` 指向主机完整的 `/usr/lib64` 是不安全的，因为这样会把与容器不兼容的 glibc 组件一并带入 collector 进程。

推荐修复方向应优先包含：

- 从宿主机单独复制兼容的 `systemd` 相关库到一个独立目录，再挂入 DataKit 容器，并让 collector 优先从该目录加载 `libsystemd`
- 把宿主机 `systemd` 库作为兼容性候选方案，而不是绝对可用的固定方案；最终是否可用仍要以启动日志和 probe 结果为准
- 如果产品实现了 `copy_node_libs`，则产品文档还应说明：该配置只是让 DataKit 启动前无条件准备一组宿主机动态库，并不意味着这些库一定会让 journald probe 成功

本次检测能力本身不负责实现这些修复动作。

## 可能变更的文件

- `internal/plugins/externals/journald/collect/journald.go`
- 如果需要额外的探测结果类型或参数传递，可能改动 `internal/plugins/externals/journald/collect/common.go`
- `internal/plugins/externals/journald/collect/*_test.go`
- `internal/export/doc/en/inputs/journald.md`
- 如果该 collector 文档有中文对照，则同步修改 `internal/export/doc/zh/inputs/journald.md`

如果帮助逻辑超过很小的增量，应拆成单独文件，例如在同一 package 下新增 probe 或 diagnosis helper。

## 测试策略

采用确定性的单元测试，而不是依赖环境的集成测试。

覆盖范围应包括：

- journal 根目录、machine-id 目录和直接 journal 文件的路径展开
- `unsupported-format`、`permission-denied`、`no-journal-files`、通用 open error、成功路径等探测结果分类
- 在不兼容结果下，启动行为会阻止进入实际采集循环
- 告警日志包含 reason 和 sampled target

探测实现应暴露一个窄接口或 seam，以便测试注入合成结果，而不依赖真实 journal 文件和真实 `libsystemd` 版本。

## 风险

- 如果通过脆弱的字符串匹配来分类，容易误判
- 可能把“空但可读”的 journal 误判为不兼容
- 可能错误地把版本检测做成必需项，而不是辅助信息
- 如果 `copy_node_libs_files` 默认白名单不完整，可能导致复制后仍因缺少依赖库而启动失败
- 如果不同发行版的候选系统库目录识别不完整，可能导致启用了 `copy_node_libs` 但实际没有复制到预期动态库

设计应优先使用显式探测结果类型，并选择保守降级。
