# 更新日志

## 1.92.0(2026/04/09) {#cl-1.92.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.92.0-new}

- 新增数据预聚合处理支持，覆盖聚合与尾采样链路（#2892）
- Pipeline 新增对 `llm` 类型数据的处理支持（#3001）
- 拨测采集器新增 SSL 证书有效期字段上报，支持输出证书过期时间和剩余天数（#3003）

### 问题修复 {#cl-1.92.0-fix}

- 修复 OpenTelemetry 在新版本校验下响应体过小导致的兼容性问题，完善 gRPC 响应内容（#3017）
- 修复 DDTrace 内存泄漏问题，优化大 trace 回收逻辑，避免 OOM（#3012）
- 修复日志采集中的 goroutine 泄漏问题，避免 Tailer 关闭路径跨实例等待导致资源无法回收（#3010）
- 修复 datakit sinker v2 中 `X-Global-Tags-V2` 未编码导致的 wrapped-url-error 问题（#3009）
- 修复 NTP 时间差在系统时间恢复后未自动清零的问题（#3006）

### 功能优化 {#cl-1.92.0-opt}

- 优化 SQLServer 和 Oracle 的 `database_instance` 优先级与 DBM 对象命名规则，避免跨节点数据混淆（#3011）
- 清理多个数据库采集器在 NewPoint 阶段注入 election tag 的逻辑，减少额外开销，并优化 Oracle 慢查询脱敏实现（#3004）
- 日志多行匹配新增 TiDB 慢日志默认规则（#3005）
- Journald 支持兼容更高版本的 systemd 库（#2996）
- 重构 GitLab 采集器 Prometheus 指标分类逻辑，补全指标字段并统一指标集命名（#2988）

### 兼容调整 {#cl-1.92.0-brk}

- 移除 upgrade 服务中的 HTTP Web 服务，升级管理统一转由 DCA 方式处理（#3007）

---

## 1.91.1(2026/04/07) {#cl-1.91.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.91.1-fix}

- 修复 DDTrace 内存泄漏问题：DDTraces.reset() 未正确释放内存导致 OOM，添加 shouldKeepInPool() 智能回收大 traces（#3012）
- 修复容器日志采集资源占用异常问题：修复 addTask 竞态条件，移除全局 goroutine group，优化 inotify 不可用时的扫描策略（#3010）
- 修复 NTP 时间差在系统时间恢复后仍被沿用的问题，现在时间差恢复时会自动清零（#3006）
- 修复 datakit sinker v2 中 X-Global-Tags-V2 header 值未编码导致的 wrapped-url-error 问题（#3009）

### 功能优化 {#cl-1.91.1-opt}

- SQLServer 和 Oracle 采集器优化 database_instance 优先级，优先使用配置中的标签，避免多节点数据混淆；max_queries 调整为 500，lookback_window 调整为 300s（#3011）
- 优化 SQLServer 和 Oracle 创建 point 时添加 election tag 的逻辑，减少每次创建的额外开销（#3004）

---

## 1.91.0(2026/03/26) {#cl-1.91.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.91.0-new}

- Kingbase 采集器新增 `server` 字段配置支持，可显式指定服务器标识，默认为 `host:port` 格式（#3002）
- bug report 新增外部采集器日志收集功能，自动收集 `[DataKit 安装目录]/externals` 目录下的 `.log` 文件（#2989）
- SQLServer 和 Oracle 采集器新增 `database_instance` 维度，通过查询数据库获取实例标识并作为 tag 写入（#2999）
- monitor 命令新增 `-Q (--quantile)` 选项，支持从 summary 指标中选择分位数（#2968）

### 问题修复 {#cl-1.91.0-fix}

- 修复 FireLens 日志流对嵌套 map/list 类型支持问题，现在将复合类型序列化为 JSON 字符串保存（#3000）
- 修复 Kingbase 采集器单例模式限制，现在支持多实例并发运行（#2995）
- 修复 logfwd 1.86.0 版本配置兼容性问题，支持 deprecated `LOGFWD_JSON_CONFIG` 环境变量自动转换为新格式（#2993）
- 修复 DataKit 缺少选举状态指标问题，确保未当选时也能正确上报选举状态指标（#2992）
- 修复 OpenTelemetry 采集器 parent_span_id 为零值时处理问题，将 `0000000000000000` 规范化为 `0`（#2987）
- 修复数据上传时 HTTP 请求格式错误导致 WAL 无限循环读取问题，现在会识别并丢弃脏数据（#2949）
- 修复 sinker header 值包含非法字符（如 `\n`）问题，现在对 header 值进行 URL 编码（#2947）

### 功能优化 {#cl-1.91.0-opt}

- 完善日志采集的多行匹配逻辑，移除已弃用的 `logging_auto_multiline_detection` 配置项，优化多行规则验证流程（#2990）
- 外部采集器支持交叉编译，提升多平台构建效率（#2994）
- Oracle 采集器升级指标集到 v2 版本，支持按指标类型分组配置采集间隔（tablespace/slow_query/process/system）（#2938）

---

## 1.90.0(2026/03/11) {#cl-1.90.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.90.0-new}

- APM 注入器新增 PHP 应用自动注入支持，包括 PHP 解释器检测、ddtrace 扩展安装和配置管理（#2986）
- Logstreaming 输入新增 AWS Firehose 数据源类型支持，接收并处理来自 AWS Firehose HTTP 端点的日志（#2979）
- Oracle 和 SQLServer 采集器新增 DBM（数据库监控）功能，包括查询指标、活动监控、会话聚合、连接指标、查询对象存储和执行计划存储（#2904）
- 主机安装器支持在安装时添加采集器配置，通过 `DK_INPUT_CONFIGS` 环境变量传递采集器配置（#2967）
- Journald 新增外部采集器实现（#2974）

### 问题修复 {#cl-1.90.0-fix}

- 修复 logfwd storage_index 配置优先级错误，环境变量 `LOGFWD_GLOBAL_STORAGE_INDEX` 现在优先于 CRD 配置（#2985）
- 修复 Helm chart DataWay token 明文暴露问题，支持自动创建 Kubernetes Secret 安全存储 token（#2981）
- 修复 OpenTelemetry 指标缺少 unit 和 description 字段问题，现在从 OTEL 指标中提取并传播这些字段（#2977）

### 功能优化 {#cl-1.90.0-opt}

- SNMP object 采集器暴露设备信息（device_type、device_vendor、device_hostname）并按接口名称合并接口条目（#2978）
- DataKit 安装器支持安装时配置采集器（#2967）
- 更新 APM 注入文档，包含 PHP 支持（#2986）
- 其他优化和问题修复

---

## 1.89.1(2026/02/12) {#cl-1.89.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.89.1-fix}

- 修复 DK 1.89.0 版本中，全局 `host` 标签设置 `host=__datakit_hostname` 时未正确使用 k8s 节点名称的问题（#2971）
- 修复采集器恢复失败阻塞选举心跳的问题，避免选举频繁切换（#2970）
- 修复意外采集 ECSFargate 容器日志时触发的错误（#2964）
- 修复选举模块状态管理，确保指标时间戳准确更新（#2970）

### 功能优化 {#cl-1.89.1-opt}

- flameshot 支持获取容器资源限制信息，优化容器环境下阈值计算准确性（#2966）
- DataKit 支持通过 datakit-operator 访问 k8s Pod 数据，为大规模集群提供 API Server 压力缓解方案（#2931）

---

## 1.89.0(2026/02/04) {#cl-1.89.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.89.0-new}

- 新增主机变更检测功能，支持用户、crontab、服务及文件变更监控（#2917）
- flameshot 支持持续采集模式，增加默认定时采集和阈值触发持续采集功能（#2953）
- 新增 DataKit 自身日志采集配置功能（#2950）

### 问题修复 {#cl-1.89.0-fix}

- 修复 Prometheus export 采集器 tags 优先级错误问题（#2960）
- 修复全局 `host` 标签设置 `host=__datakit_ip` 时无效的问题（#2956）
- 修复 eBPF 采集器导致 `istio-init` 容器不退出的问题（#2955）
- 修复容器日志采集使用默认 stdout 配置时存在无用操作的问题（#2962）
- 修复 WAL 锁文件使用 PID 导致退出后无法重用的问题（#2948）
- 修复 profile 采集器初始化时机问题，避免磁盘缓存未初始化导致的 panic（#2946）
- 修复 Statsd 指标采集，新增 event/service check 采集，这俩类数据目前以日志形式来采集（#2941）

### 功能优化 {#cl-1.89.0-opt}

- 为选举模块增加更多日志和指标，便于检测选举频繁切换和采集器暂停失败问题（#2957）
- 更新 DataKit HTTP 客户端指标，增加 URL 路径标签和请求体传输汇总指标（#2952）
- SQLServer 采集器新增 `sqlserver_host` 标签，并将 `instance` 标签改为 `counter_instance`（#2951）
- bug report 新增 Git 配置文件收集功能（#2939）
- Windows 进程采集器新增 status 字段支持（#2927）
- DDTrace 采集新增更多 `source_type` 支持（#2958）

---

## 1.88.1(2026/01/16) {#cl-1.88.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.88.1-fix}

- 在 1.87.2 版本中，OpenTelemetry 指标移除了全局主机 tag 追加，这一移除会造成比较大的影响，默认情况下还是追加这些 tag，如果需要移除，本版本新加一个 flag 来配置（#2942）
- Flameshot 中修复触发阈值判断问题（#2943）
- Pipeline 调试中增加 IPDB 配置功能（#2944）

---

## 1.88.0(2026/01/14) {#cl-1.88.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.88.0-new}

- 新增数据采集[可用性指标采集](../integrations/ingestion_canary.md)（#2900）
- DCA 新增 DataKit 存活检测（#2910）

### 问题修复 {#cl-1.88.0-fix}

- 修复 Pod 内存采集数值虚高问题（#2933）
- 修复 Pod 重启后 KubernetesPrometheus 未能重新采集的问题（#2936）
- 修复无法采集 DDTrace 中 NodeJS profile 的问题（#2937）[^2937]
- 修复多步拨测重试问题（#2915）
- 修复 AWS Lambda 扩展采集异常问题（#2918）

[^2937]: 要完整支持 DDTrace NodeJS profile 采集，底座仍需升级到最新版本。

### 功能优化 {#cl-1.88.0-opt}

- DataKit 日志输出中，给 `ERROR` 级别的日志单独一个文件（默认为 *error.log*），避免其被其它日志覆盖掉，同时 bug report 中也会带上这个错误日志（#2940）
- 优化磁盘缓存模块（WAL），新增更多指标和日志暴露，同时优化 *.pos* 文件对磁盘 io 的影响（#2935）
- SNMP 采集新增更多 yaml 配置，修复一些历史遗留问题（#2923）
- 容器日志采集和 logfwd 新增 `from_beginning_threshold_size` 配置项（#2934）
- 多个采集器采集的数据上增加了 `collector_source_ip` 字段，表示其数据来源（#2819）[^2819]
- 其它优化（#2928/#2932/#2930）

[^2819]: 这些采集器包括 `zipkin/logstreaming/beats_output` 等。

### 兼容调整 {#cl-1.88.0-brk}

- SNMP 采集的数据中移除了对象数据中的 `all` 冗余字段（#2923）
