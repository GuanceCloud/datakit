# 更新日志

## 1.94.0(2026/05/13) {#cl-1.94.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.94.0-new}

- 日志采集 socket logs 支持记录来源 IP，并使用 `collector_source_ip` 作为来源地址 tag（#3061）
- 拨测 HTTP 任务支持自定义协议版本，便于覆盖不同 HTTP 协议兼容性场景（#3041）
- PostgreSQL DBM 新增 SQL 执行量与 QPS 指标采集上报能力（#3046）
- MongoDB 支持 database object 采集上报（#3045）
- Doris 支持 object 采集上报（#3043）
- bug report 默认支持通过 Dataway 上传到观测云，并保留 OSS 直传方式（#3028）
- AWS Lambda 采集器完善函数调用链路，补充运行时事件与调用上下文关联能力（#2961）

### 问题修复 {#cl-1.94.0-fix}

- 修复 9529 HTTP API 未注册路由与未启用采集器路由的 403/404 返回策略，减少用户排障误导（#3054）
- 修复 DataKit HTTP 服务启动失败时主进程状态仍显示正常的问题，现在 HTTP 服务异常会使主进程退出（#3052）
- 修复 APM 自动注入在 arm64 动态库交叉编译与替换过程中的异常问题（#3050）
- 修复 Redis 采集器偶现 `concurrent map writes` 的问题，并优化 host tag 优先级处理（#3039）

### 功能优化 {#cl-1.94.0-opt}

- StatsD 默认关闭 DogStatsD event 与 service check 日志采集，仅在用户显式配置后开启，避免默认产生额外日志量（#3059）
- 优化 `DK_HTTP_LISTEN` 识别逻辑，支持直接填写 `ip:port`，并明确其与 `DK_HTTP_PORT` 的优先级（#3053）
- 优化 Pipeline 中 Grok 与 JSON 处理性能，提升日志处理吞吐（#3051）
- eBPF 采集器移除 CGO 依赖，并优化 netlog/netflow/L7/exporter 内存使用与运行时观测能力（#3049）
- 持续补充并优化日志自动多行规则，提升默认多行识别效果（#3048）
- 补齐 `cat`、`xfsquota`、`windowsremote`、`logfwdserver` 等采集器热加载能力（#3042）
- 优化 DataKit 启动过程，移除初始化阶段不必要的采集行为，降低启动耗时（#3038）
- OpenTelemetry 采集器复用 `cliutils/otlp` 共享解析器，收敛 metrics/logs/traces 解析主循环并保持 DataKit 本地语义（#3026）

---

## 1.93.0(2026/04/22) {#cl-1.93.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.93.0-new}

- MySQL 与 PostgreSQL DBM 采集器切换到新实现，完善数据库性能与对象采集链路（#2998）
- Pipeline 脚本支持对 `source` 做通配匹配，便于复用处理规则（#3036）
- Flameshot 新增 `jcmd` 与 `hprof` 采集支持，增强 Java 问题定位能力（#3031）

### 问题修复 {#cl-1.93.0-fix}

- 修复 aggregate 聚合链路中的发送与指标统计问题，完善兼容指标与并发发送行为（#3024）
- 修复 vSphere 事件采集时间与超时处理问题，并补充相关单元测试（#3037）
- 修复拨测模块若干稳定性问题，补全核心路径单元测试（#3018）

### 功能优化 {#cl-1.93.0-opt}

- 优化日志采集多行处理策略，统一手动/自动匹配行为并扩展默认规则集（#3029）
- 重构 eBPF 采集器链路，改为直接使用 cilium 相关能力并补强稳定性与兼容性（#3016）
- 更新 DCA 相关文档，补充控制台与使用说明（#3032）

---

## 1.92.1(2026/04/16) {#cl-1.92.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-1.92.1-fix}

- 修复 Pipeline 中 `add_key` 对 `list/map` 类型值未序列化的问题，现在会将复合类型规范化为字符串后再写入（#3035）
- 修复 SQLServer 和 Oracle `object` 采集未严格遵循采集间隔的问题，避免未到时间窗口时仍继续执行采集（#3034）
- 修复容器指标采集时 k8s `requests` 字段取值错误的问题，确保正确上报容器资源请求量（#3033）
- 修复 `datakit import` 回放场景下未继承 `datakit.conf` 中 DataWay 配置的问题，支持导入回放沿用主配置并保留命令行覆盖能力（#3030）

### 功能优化 {#cl-1.92.1-opt}

- 调整 DK 外部采集器与相关组件构建流程，完善多架构构建路径与编译兼容性（#3013）

---

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
