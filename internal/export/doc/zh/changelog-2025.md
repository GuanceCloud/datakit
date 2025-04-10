# 更新日志

## 1.71.0(2025/04/10) {#cl-1.71.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.71.0-new}

- Pyroscope 新增 Rust 语言支持[^2602]（#2602）

[^2602]: 该功能需较新版本底座配套支持。

### 问题修复 {#cl-1.71.0-fix}

- 修复 monitor 某些情况下可能导致崩溃的问题（#2610）
- 修复 APM 自动注入失败问题（#2607）
- 部分采集器自带的日志采集中补全日志长度自动分段，避免长日志丢失（#2613）
- 修复拨测采集器自定义 tag 无法上报的问题（#2616）

### 功能优化 {#cl-1.71.0-opt}

- Datakit 上传的数据包请求 HTTP 头中增加 X-Pkg-ID，用于数据包追踪（#2587）
- Kubernetes event 采集的数据中新增 `source_host/source_component` 字段（#2606）
- DDTrace 资源目录采集中将用户自定义注入的 tag 提到一级字段，便于做数据分流（#2609）
- 优化 DDTrace 采样策略（#2614）
- WAL 磁盘缓存允许部分数据分类在磁盘写满的情况下不丢弃数据（#2620）
- Profile 和 RUM session replay 数据增加 global tag 便于做分流（#2621）
- eBPF 采集增加更多 Kubernetes 标签，如 `cronjob/daemonset/statefulset` 等（#2571）
- 其它优化（#2615）

---

## 1.70.0(2025/03/26) {#cl-1.70.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.70.0-new}

- KubernetesPrometheus 新增 up 指标（#2577）

### 问题修复 {#cl-1.70.0-fix}

- 修复 Pod stdout 日志屏蔽采集，在某些情况下无效的问题（#2570）
- 优化超长（超过最大允许的 HTTP Post 长度）多行日志的处理（#2572）
- 修复 OpenTelemetry 在接收前端推送的 trace 数据时的跨域问题（#2592）
- 修复 APM 自动注入失效问题（#2594）
- 修复磁盘采集器获取用量失败的问题（#2597）
- 修复 Zipkin 采集器处理 `application/json; charset=utf-8` 问题（#2599）

### 功能优化 {#cl-1.70.0-opt}

- SQLServer 采集器增加 2008 版本支持（#2584）
- 数据库采集器（MySQL/Oracle/PostgreSQL/SQLServer）新增指标屏蔽功能（部分指标不予采集以缓解采集本身给数据库带来的压力），同时对采集过程中的 SQL 执行消耗做了指标记录（#2579）
- DDTrace 采集的资源目录中新增用户自定义的 tag（#2593）
- NFS 采集器增加读写延时指标（#2601）
- 修复 lsblk 采集器内存泄漏问题（!3458）

---

## 1.69.1(2025/03/18) {#cl-1.69.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.69.1-fix}

- 修复 Docker 容器 CPU 采集有误问题（#2589）
- 修复拨测采集器多步拨测的脚本执行导致的内存泄露问题（#2588）
- 优化多步拨测错误信息（#2567）
- 部分文档更新（#2590）

---

## 1.69.0(2025/03/12) {#cl-1.69.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.69.0-new}

- APM 自动注入增加注入 statsd 支持（#2573）
- Pipeline 新增 key event 类数据的处理（#2585）

### 问题修复 {#cl-1.69.0-fix}

- 修复主机重启后 `host_ip` 获取不到的问题（#2543）

### 功能优化 {#cl-1.69.0-opt}

- 优化进程采集器，增加若干跟进程有关的指标（#2366）
- DDTrace 优化 trace-id 字段的处理（#2569）
- OpenTelemetry 采集中增加 `base_service` 字段（#2575）
- 调整 WAL 默认设置，worker 数默认改成 CPU 限额核心数 * 8，同时安装/升级阶段支持指定 worker 数以及磁盘缓存大小（#2582）
- Datakit 容器环境下运行时，移除 pid 检测（#2586）

### 兼容调整 {#cl-1.69.0-brk}

- 优化磁盘采集器，默认屏蔽一些文件系统类型以及挂载点（#2566）

    调整磁盘指标采集，同时更新了主机对象中的磁盘列表采集，主要有如下差异：

    1. 新增了挂载点忽略选项：该调整主要是为了优化 Kubernetes 中 Datakit 获取磁盘列表时，过滤掉一些不必要的挂载点，比如 ConfigMap 配置挂载（`/usr/local/datakit/.*`）和 Pod 日志采集导致的挂载（`/run/containerd/.*`）；同时避免了新增的无效时间线（这些新增的时间线主要是挂载点不同导致的）。
    1. 新增文件系统忽略选项：对一些不太需要采集的文件系统（比如 `tmpfs/autofs/devpts/overlay/proc/squashfs` 等）默认做了忽略
    1. 主机对象采集中，也和 disk 指标采集做了同等的默认忽略策略。

    这样调整之后，时间线能大幅度减少，同时，我们在配置监控的时候，也更好理解，避免了挂载点繁多带来的困扰。


---

## 1.68.1(2025/02/28) {#cl-1.68.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.68.1-fix}

- 修复 OpenTelemetry 指标采集内存消耗问题（#2568）
- 修复 eBPF 解析 PostgreSQL 协议导致的崩溃问题（!3420）

---

## 1.68.0(2025/02/27) {#cl-1.68.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.68.0-new}

- 新增多步拨测功能（#2482）

### 问题修复 {#cl-1.68.0-fix}

- 修复日志采集多行缓存清理问题（!3419）
- 修复 xfsquota 默认配置问题（!3419）

### 功能优化 {#cl-1.68.0-opt}

- Zabbix Exporter 采集器增加低版本（v4.2+）兼容（#2555）
- Pipeline 处理日志时提供了 `setopt()` 函数来定制化日志等级的处理（#2545）
- OpenTelemetry 采集器在采集直方图（Histogram）类型的指标时，默认将其转换成 Prometheus 风格的直方图（#2556）
- 调整主机安装 Datakit 时的 CPU 限额方式，新装的 Datakit 默认使用基于 CPU 核心数的 limit 机制（#2557）
- Proxy 采集器增加来源 IP 白名单机制（#2558）
- Kubernetes 容器和 Pod 指标采集允许针对 namespace/image 等方式来进行定向采集（#2562）
- Kubernetes 容器和 Pod 的内存/CPU 补全基于 Limit 和 Request 的百分比采集（#2563）
- AWS 云同步新增 IPv6 支持（#2559）
- 其它问题修复（!3418/!3416）

### 兼容调整 {#cl-1.68.0-brk}

- OpenTelemetry 指标收集时，调整了指标集名字，原来的 `otel-service` 改成了 `otel_service`（!3412）

---

## 1.67.0(2025/02/12) {#cl-1.67.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.67.0-new}

- KubernetesPrometheus 支持采集时增加 HTTP header 设置，顺便支持通过字符串方式配置 bearer token（#2554）
- 增加 xfsquota 采集器（#2550）
- AWS 云同步增加 IMDSv2 支持（#2539）
- 新增 Pyroscope 采集器用于采集基于 Pyroscope 的 Java/Golang/Python Profiling 数据（#2496）

### 问题修复 {#cl-1.67.0-fix}
### 功能优化 {#cl-1.67.0-opt}

- 完善 DCA 配置有关的文档（#2553）
- OpenTelemetry 采集支持提取 event 字段为一级字段（#2551）
- 完善 DDTrace-Golang 文档，增加编译时插桩说明（#2549）

---

## 1.66.2(2025/01/17) {#cl-1.66.2}

本次发布属于 hotfix 修复，同时增加一些细节功能。内容如下：

### 问题修复 {#cl-1.66.2-fix}

- 修复 Pipeline 调试接口兼容性问题（!3392）
- 修复 UDS 监听问题（#2544）
- UOS 镜像增加 `linux/arm64` 支持（#2529）
- 修复 prom v2 采集器中 tag 优先级问题（#2546）以及 Bearer Token 问题（#2547）

---

## 1.66.1(2025/01/10) {#cl-1.66.1}

本次发布属于 hotfix 修复，同时增加一些细节功能。内容如下：

### 问题修复 {#cl-1.66.1-fix}

- 修复 prom v2 采集器时间戳精度问题（#2540）
- 修复 PostgreSQL index 这个 tag 跟 DQL 关键字冲突问题（#2537）
- 修复 SkyWalking 采集中 `service_instance` 字段缺失问题（#2542）
- 移除 OpenTelemetry 中无用配置字段，修复部分指标单位 tag（`unit`）缺失问题（#2541）

---

## 1.66.0(2025/01/08) {#cl-1.66.0}

本次发布为迭代发布，主要更新内容如下：

### 新功能 {#cl-1.66.0-new}

- 增加 KV 机制，支持通过拉取更新采集配置（#2449）
- 任务下发功能中，存储类型增加 AWS/华为云存储支持（#2475）
- 新增 [NFS 采集器](../integrations/nfs.md)（#2499）
- Pipeline 调试接口的测试数据支持更多 HTTP `Content-Type`（#2526）
- APM Automatic Instrumentation 新增 Docker 容器支持（#2480）

### 问题修复 {#cl-1.66.0-fix}

- 修复 OpenTelemetry 采集器无法接入 micrometer 数据的问题（#2495）

### 功能优化 {#cl-1.66.0-opt}

- 优化磁盘指标采集和对象中的磁盘采集（#2523）
- 优化 Redis slow log 采集，在 slow log 中新增客户端信息。同时，slow log 对低版本（<4.0）的 Redis（如 Codis）做了选择性支持（#2525）
- 调整 KubernetesPrometheus 采集器在采集指标过程中的错误重试机制，当目标服务短暂不在线时不再将其剔除采集（#2530）
- 优化 PostgreSQL 采集器默认配置（#2532）
- KubernetesPrometheus 采集的 Prometheus 指标，新增指标名裁剪配置入口（#2533）
- DDTrace/OpenTelemetry 采集器支持主动提取 `pod_namespace` 这个 tag（#2534）
- 完善日志采集 scan 机制，强制增加一个 1min 的 scan 机制，避免极端情况下的日志文件遗漏（#2536）

---

