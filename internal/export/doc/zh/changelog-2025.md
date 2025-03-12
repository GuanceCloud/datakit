# 更新日志

## 1.68.1(2025/02/28) {#cl-1.68.1}

本次发布属于 hotfix 修复，同时增加一些细节功能。内容如下：

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

