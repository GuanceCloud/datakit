---
name: using-bug-report-as-debug-entry
description: Use when debugging with an external DataKit bug-report bundle such as info-<timestamp> or info-<timestamp>.zip, and you need to use the current bundle as the primary evidence while treating docs/bug-report as historical hints
---

# 使用 Bug Report 包建立排查入口

## Overview

这个技能关注的不是“如何修某个具体缺陷”，而是如何正确使用 bug report。

输入通常有两类：

1. 外部传入的现场包
   形如 `info-<timestamp>` 目录或 `info-<timestamp>.zip`
2. 仓库内的历史总结
   位于 `docs/bug-report/`

两者的职责不同：

- 外部 bug report 包是当前问题的一手现场
- `docs/bug-report/` 是历史总结，只能提供线索、证据模式和代码方向

不要把历史报告当结论库。先用当前 BR 包建立事实，再用历史报告找相似模式。

## When to Use

适用场景：

- 用户提供了 `info-<timestamp>` 目录
- 用户提供了 `info-<timestamp>.zip`
- 你要排查 DataKit 线上现场
- 你想结合 `docs/bug-report/` 中的历史经验缩小范围

不适用场景：

- 没有外部 BR 包，只有口述现象
- 当前问题和 DataKit 现场无关
- 你试图只靠 `docs/bug-report/` 代替当前排查

## Core Pattern

正确顺序：

1. 先读取当前 BR 包
2. 用当前 BR 包建立事实
3. 再查 `docs/bug-report/` 找相似证据模式
4. 回到当前 BR 包验证这些线索是否成立
5. 最后才判断是同一问题、同类问题，还是仅仅现象相似

## 使用步骤

### 1. 先确认输入是不是标准 BR 包

标准形式通常是：

- `info-<timestamp>`
- `info-<timestamp>.zip`

时间戳本身也有价值，它可以帮助你和 `metrics/`、`log/` 里的时间线做对齐。

如果用户给的是 zip，先解压。
如果用户给的是目录，先看目录结构。

### 2. 先读当前 BR 包，不要先翻历史报告

根据官方说明，优先关注这些目录：

- `basic`
- `config`
- `data`
- `log`
- `metrics`
- `pipeline`
- `profile`
- `externals`

不是每次都要把所有目录看完，但你至少要先回答四个问题：

1. 运行环境是什么
2. 实际启用了哪些输入和限制
3. 进程实际在持续做什么
4. 异常发生在哪一层

### 3. 推荐阅读顺序

建议按下面顺序进入：

1. `basic/info`
   用来判断系统、平台、架构、环境变量，也用于初步判断宿主机部署还是 k8s 内运行。

2. `config/`
   看主配置和各输入配置，确认到底启用了什么能力，尤其要注意高资源消耗输入和资源限制。

3. `data/pull` 或 `data/.pull`
   看中心下发配置、黑名单、远程 pipeline、DataWay 动态配置。

4. `log/log`、`log/gin.log`
   用于建立时间线，确认启动、接入、报错、重试、恢复等过程。

5. `metrics/`
   这是事实判断的核心。默认通常有多份采样，可以看趋势，不要只看单点。

6. `profile/`
   涉及 CPU、内存、OOM、热点模块时，必须进入。

7. `externals/`
   如果现场包含外部进程，这里往往是关键，不能后置。

### 4. 先识别运行形态，再决定排查重点

不要默认当前 BR 包一定来自 daemonset，也不要默认一定是宿主机 service。

建议交叉看这些信号：

- `basic/info` 里的环境变量
- `metrics/` 中 `datakit_uptime_seconds` 的 `docker="true|false"`
- `metrics/` 中 `resource_limit` 等标签
- `syslog/`、`log/` 中是否体现 `systemd`、container、pod 等线索

这一步决定后续重点：

- 宿主机场景更容易遇到系统时间、service 启动顺序、主机依赖类问题
- k8s/daemonset 场景要优先考虑 pod limit、容器重启、外部进程、节点环境

### 5. 指标优先级通常高于日志

日志负责回答“发生了什么”。
指标更适合回答“系统实际上持续在做什么”。

当用户说下面这些现象时，先看 `metrics/`：

- “本地有数据，服务端看不到”
- “数据断档”
- “pod 持续 OOM”
- “重启后恢复”
- “采集器开着，但效果异常”

优先用指标判断：

- 采集器是否在运行
- feed 是否发生
- 写 DataWay 是否成功
- 进程是否频繁重启
- 内存和 CPU 是否在持续拉高
- 外部进程是否一起占资源

不要被单条日志先带偏方向。

### 6. profile 负责回答“哪个模块在占资源”

当问题涉及 CPU、内存、OOM、卡顿、热点路径时，`profile/` 不是可选项。

使用原则：

- `profile/heap` 看当前仍在占用的 Go 堆内存
- `profile/allocs` 看累计分配量，适合识别高频分配模块
- `profile/profile` 看 CPU 热点
- 不要把 `allocs` 误判成“当前内存”
- 不要把单个 `heap` 误判成整个进程 RSS

### 7. 外部进程必须单独分析

如果 `externals/` 下有外部进程，例如：

- `externals/ebpf/profile/`

那就必须单独分析，因为：

- 主进程 `profile/` 看不到外部进程的堆
- pod 的真实内存可能是主进程和外部进程叠加导致
- `metrics/` 中外部进程 RSS 往往才接近真实 OOM 压力

分析资源问题时，至少要同时看三类信息：

1. `metrics/` 中主进程 RSS、Go heap、外部进程 RSS
2. 主进程 `profile/`
3. `externals/*/profile/`

### 8. 再用历史 bug report 找相似线索

`docs/bug-report/` 是历史总结，它的价值在于：

- 帮你识别相似现象
- 提供关键证据示例
- 告诉你哪些指标最值钱
- 提示相关代码路径
- 提示常见误判和排除项

正确用法：

1. 先完成当前 BR 包的现场判断
2. 再去历史报告中搜索相似现象
3. 提取历史报告中的关键证据模式和代码路径
4. 回到当前现场验证这些线索是否成立

### 9. 重点借用“历史证据模式”，不要直接借用“历史结论”

从历史报告中最值得借用的是：

- 触发条件
- 关键日志关键字
- 关键指标名
- 可疑代码路径
- 排除项

不要直接照搬的是：

- “这就是根因”
- “按上次那样改就行”
- “现象一样，所以一定是同一个问题”

### 10. 用当前 BR 包重建证据链

最终输出时，要能说清楚：

1. 用户现象是什么
2. 当前 BR 包里哪些文件证明这个现象成立
3. 问题落在采集、处理、写出、展示、资源还是时间窗口哪一层
4. 历史报告提供了哪些线索
5. 当前现场复现了哪些历史证据
6. 没有复现哪些历史证据

## 当前会话沉淀出的关键经验

基于 `~/Downloads/info-1774858681272` 和 `~/Downloads/info-1774866741044` 这两份现场，可以抽出几条稳定经验：

1. 先识别运行形态
   同样是“用户说 daemonset 有问题”，现场包未必真的来自 daemonset。
   `info-1774858681272` 实际是宿主机 service。
   `info-1774866741044` 才是 k8s/daemonset 场景。

2. “有数据但服务端看不到”要优先怀疑时间窗口问题
   如果 `metrics/` 里写出成功，DataKit 自身指标也在持续增长，问题未必在传输链路，也可能在时间戳落点。

3. OOM 不能只看主进程
   daemonset 场景下，pod OOM 可能来自主进程和外部进程叠加。
   只看 `profile/heap` 很容易低估真实内存压力。

4. `profile` 和 `metrics` 必须联动
   `profile` 解释模块归因。
   `metrics` 解释真实 RSS、uptime、重启趋势和外部进程占用。
   单看任意一边都容易误判。

5. 外部进程热点不能漏
   例如 `externals/ebpf/profile/` 能解释 `ebpf-net`、L7 flow、perf event、协议解析等开销。
   如果只看主进程 `profile/`，结论会缺一半。

## 常见错误

错误做法：

- 拿到 BR 包后先去翻历史报告，不先看当前现场
- 只根据 `docs/bug-report/` 下的旧结论做判断
- 把 `info-<timestamp>.zip` 当附件存档，不实际解压阅读
- 看到某条报错日志就直接下结论，不结合 `metrics/`
- 只看 `profile/heap`，不看 `metrics/` 中的 RSS
- 只看主进程 `profile/`，不看 `externals/*/profile/`
- 把 daemonset pod 的 OOM 误判成主进程单独泄漏

正确做法：

- 先把外部 BR 包读透
- 先建立当前现场的事实
- 再用历史报告缩小方向
- 用当前现场验证历史线索
- 资源问题时，把 `metrics/`、`profile/`、`externals/*/profile/` 联动分析

## 输出要求

当你基于 BR 包和历史报告完成一次排查后，输出里至少要说明：

1. 当前 BR 包中最关键的现场证据是什么
2. 当前现场属于宿主机、k8s，还是外部进程主导问题
3. 历史报告给了你哪些线索
4. 当前现场复现了哪些历史证据
5. 哪些历史证据没有复现
6. 这次是同一问题、同类问题，还是仅仅现象相似

## Related Docs

- `docs/bug-report/3006-ntp-stale-time-diff-after-clock-recovery.md`
- 官方说明：`https://docs.guance.com/datakit/bug-report-how-to/`
