# 更新日志

## 2.7.0(2026/07/29) {#cl-2.7.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.7.0-new}

- 拨测采集器新增 NetPath 在线任务，支持 TCP、UDP 和 ICMP 网络路径探测及 debug 调试（#3160）
- DataWay gzip 数据上报支持可选的 `gzip-caesar-v1` 字节位移混淆，默认关闭；启用前需确保 Kodo 已兼容，并继续使用 HTTPS（#3163）
- Flameshot OSS hprof 上传支持 STS 临时凭证，同时保持静态 AK/SK 方式兼容（#3152）

### 问题修复 {#cl-2.7.0-fix}

- 修复 Oracle 采集器未持续上报 `collector.up` 的问题（#3162）

### 功能优化 {#cl-2.7.0-opt}

- MySQL 和 SQL Server Object 结构采集新增开关及库表范围过滤，降低大规模实例的采集开销（#3161, #3164）
- 浏览器拨测内置的 Lightpanda 升级至 0.3.6（#3167）

---

## 2.6.1(2026/07/23) {#cl-2.6.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.6.1-fix}

- 修复 Lightpanda 将浏览器拨测 URL 中的 `#` 编码为 `%23` 的问题（#3157）

### 功能优化 {#cl-2.6.1-opt}

- Promsd 采集器新增 `relabel_configs`，支持筛选及重标记 HTTP 和 File 服务发现目标，暂不支持 Consul（#3138）
- 优化通过 Kubernetes Pod 注解 `datakit/prom.instances` 配置的 Prometheus 采集调度与资源控制，提升配置刷新和任务清理稳定性（#3155）
- 完善主机部署 DataKit 接收跨网络 APM 上报时的 HTTP 监听与安全说明（!4099）
- 更新文档中过时的 DataKit CLI 命令与参数示例（!4101）

---

## 2.6.0(2026/07/16) {#cl-2.6.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.6.0-new}

- 拨测采集器支持通过 SSE 执行一次性任务，不影响周期调度；拨测结果新增 `trigger_type`，手动任务新增 `run_batch_id`（#3139）
- 拨测节点名称支持 `name_i18n` 多语言适配，并保留旧字段回退（#3126）
- SNMP Trap 新增 `source` 配置，支持自定义日志 Source，未配置时使用 `traps`（#3145）
- 新增 NetPath 网络路径采集器，支持 TCP、UDP、ICMP 目标探测及基于 eBPF 的 TCP 目标发现（#3143）

### 问题修复 {#cl-2.6.0-fix}

- 修复容器环境中 `ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK=false` 无法开启内网拨测的问题（#3147）
- 修复 OpenTelemetry 开启 `split_service_name` 后将 gRPC span 服务名替换为 `grpc` 的问题（#3148）

### 功能优化 {#cl-2.6.0-opt}

- 变更事件统一使用 unified diff，并优化 K8s annotation、env 和 probe 的变更展示（#3137）

### 兼容调整 {#cl-2.6.0-brk}

- Helm Chart 默认关闭 `componentHealthLivenessProbe`，避免旧版镜像因缺少 `/v1/health` 进入 CrashLoop；DataKit 2.5.0 及以上版本可手动开启

---

## 2.5.0(2026/07/08) {#cl-2.5.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.5.0-new}

- Oracle 采集器新增 ASM 磁盘组指标采集（#3136）
- Oracle 采集器新增数据库打开模式、资源上限及数据文件限额指标（#3136）
- Flameshot 新增 Go 语言 pprof 采集支持（#3132）
- Flameshot 新增主动 Heap Dump 与 hprof 上传至 OSS/S3（#3122）

### 问题修复 {#cl-2.5.0-fix}

- 修复 `gpu_smi` 掉卡告警逻辑死循环风险（#3144）
- 修复聚合链路 tail sampling 未按 sinker header 分组的问题（#3095）

### 功能优化 {#cl-2.5.0-opt}

- Redis 新增 `role_status` 指标字段，支持主从角色变化检测（#3141）
- 新增全局 `measurement_version` 配置，支持安装/升级时统一指定指标集版本，覆盖 MySQL、Oracle、PostgreSQL、SQLServer、Redis、JVM 等采集器（#3142）
- 新增 `/v1/health` 端点与 CRI 运行时故障自恢复，优化 K8s liveness 检测（#3140）

---

## 2.4.0(2026/07/02) {#cl-2.4.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.4.0-new}

- DataKit 连接 DataWay 时支持 IPv4/IPv6 双栈自适应，默认优先 IPv6 并自动回退到 IPv4，同时支持配置地址族策略（#3131）
- Pipeline 新增 `enable_grok_fast_path` 配置项，可显式控制 Grok fast path 优化，默认关闭（#3133）
- SNMP 新增用户 Profile 目录 `conf.d/snmp/extra_profiles`，支持覆盖或继承内置 Profile，避免升级时覆盖用户配置（#3128）

### 问题修复 {#cl-2.4.0-fix}

- 修复 `gpu_smi` 采集器对 NVIDIA SMI v12/v13 XML 的兼容问题，支持新版功耗字段及新增 GPU 指标（#3129）

### 功能优化 {#cl-2.4.0-opt}

- SNMP 内置 Profile 指标支持 `tags_ignore` 和 `tags_ignore_regexp` 过滤，减少高基数 tag 产生的无效时间线（#3130）
- 修正 MySQL、MySQL 用户状态、DBM 和 InnoDB 指标的类型、单位及描述，明确累计计数与当前状态值（#3135）

### 兼容调整 {#cl-2.4.0-brk}

- OpenTelemetry 采集器兼容 LoongSuite/Alibaba ARMS 的新版语义字段，补充 DB、HTTP/Network、MQ 和 RPC 属性映射（#3085）

---

## 2.3.0(2026/06/24) {#cl-2.3.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.3.0-new}

- 新增 GKE Autopilot 集成，通过 `datakit-gke-autopilot` Helm Chart/YAML 以 Deployment 方式部署 DataKit，并通过 GCP Cloud API 采集容器指标和 stdout/stderr 日志（#3107）

### 问题修复 {#cl-2.3.0-fix}

- 修复 PostgreSQL 表对象采集在不同版本间的兼容性问题，并修复分批采集时索引信息可能重复的问题（#3127）

### 功能优化 {#cl-2.3.0-opt}

- SNMP 采集调度优化：同一 IP 的同类采集任务会跳过重复执行，不同类型采集仍保持串行，降低重复采集和等待开销（#3121）
- 优化 eBPF 与 DataKit Operator 的通信效率，降低大规模集群下的查询和解析开销；该能力需要 DataKit Operator v1.8.9 及以上版本支持（#3125）
- Kafka 集成补充新版 Dashboard 和 Monitor 入口（#3040）
- DCA 升级至 0.1.7，优化管理端稳定性和交互体验，并补充前后端单元测试覆盖（#3065）
- 完善指标集元数据，`measurements-meta.json` 新增 `desc_i18n` 描述，并补充多个采集器的单位、标签和字段说明（!3956）

### 兼容调整 {#cl-2.3.0-brk}

- DataKit 上报的指标点新增 `__input_source` tag，取值形如 `dk.<input>`，用于标识指标来源采集器；HTTP API 和 Pipeline 生成的指标也会写入对应来源（#3124）

---

## 2.2.1(2026/06/18) {#cl-2.2.1}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.2.1-fix}

- 修复 PostgreSQL 采集器默认排除 `postgres` 数据库，导致默认库相关指标无法采集的问题；如需排除该数据库，可继续通过 `ignored_databases` 配置显式指定（#3123）

---

## 2.2.0(2026/06/17) {#cl-2.2.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.2.0-new}

- 新增 IBM AS/400 (IBM i) 外部采集器，通过 ODBC 连接采集系统、磁盘、作业、内存池、子系统、作业队列和消息队列等指标（#3082）
- vSphere 采集器新增虚拟机维度磁盘存储指标 `disk_used_latest`、`disk_provisioned_latest`、`disk_unshared_latest`（#3111）
- 拨测采集器新增 SSL/TLS 证书检测任务，支持证书过期时间、剩余天数、TLS 版本等检测（#3106）
- Pipeline 新增 `json_all` 和 `pt_kvs_set_map` 函数（#3109）
- SNMP 采集器新增 `oid_batch_size` 和 `bulk_max_repetitions` 配置，适配 iDRAC 等对 GetBulk 请求敏感的 SNMP Agent（#3093）
- PodMonitor/ServiceMonitor 改为 informer 架构，创建后修改 YAML 可动态生效，无需重启采集（#2832）

### 问题修复 {#cl-2.2.0-fix}

- 修复 APM 自动注入在特定场景下失败的问题（#3120）
- 修复 Windows 低延迟场景下 ICMP 拨测将 0ns 回包误判为丢包的问题（#3118）
- 修复 PostgreSQL 9.1+ 主从复制延迟指标因 numeric 类型转换失败导致无法上报的问题（#3114）
- 修复 prom_remote_write 采集器解析失败时缺少 `return` 导致继续处理无效数据的问题（#3105）
- 修复 compact Body 缓存 dump/load 后 PayloadType 字段不一致的 bug（#3102）
- 修复 diskio 采集器单测中读写速率偶发翻倍的稳定性问题（#3101）
- 修复磁盘使用率在特定条件下计算异常的问题（#3090）
- 修复 HTTP API reload 后请求限流和超时配置丢失的问题，热加载后配置与首次启动保持一致（#3079）

### 功能优化 {#cl-2.2.0-opt}

- 容器日志采集出现重复路径时，改为保留最后一条配置并告警，不再直接丢弃采集任务（#3119）
- 浏览器拨测结果中隐藏 Lightpanda 底层启动错误细节，避免暴露运行环境路径信息（#3103）
- 拨测节点名称变更后自动同步到 Dialtesting 任务的上报数据中（#3099）
- 补充 SNMP 自定义 YAML 模版字段格式说明文档（#3116）
- 更新数据库集成 Dashboard 路径指向新的 dashboard（#3113）
- CI 增加 Go module/golangci-lint 缓存复用，Docker buildx 支持 registry cache 加速镜像构建（#3096）

### 兼容调整 {#cl-2.2.0-brk}

- DK 自身指标采集改为全量/关闭模式，不再支持白名单过滤；Profile 采集改为手动控制（#3110）
- ddtrace 采集器 telemetry tag 兼容 `DD_TRACE_TAGS` 字段名，正确写入 JVM 指标 tags（#3108）

---

## 2.1.5(2026/06/16) {#cl-2.1.5}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.1.5-fix}

- 修复短生命周期 Pod 可能丢失日志的情况（#3117）

## 2.1.4(2026/06/12) {#cl-2.1.4}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.1.4-fix}

- 修复日志采集在 Pod 关闭时可能丢失数据的问题（#3115）

## 2.1.3(2026/06/12) {#cl-2.1.3}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.1.3-fix}

- 修复指标自动注入 `collector_source_ip` 导致时间线膨胀的问题，指标不再添加该标签；OpenTelemetry 指标新增关闭全局标签的配置项（#3112）

### 功能优化 {#cl-2.1.3-opt}

- 调整 `datakit debug --bug-report` 默认行为：默认仅生成并保留本地 zip，不再自动上传；需要上传时可显式使用 `--bug-report-dataway` 或继续使用 `--oss`（#3089）

---

## 2.1.2(2026/06/08) {#cl-2.1.2}

本次发布属于 hotfix 修复，内容如下：

### 问题修复 {#cl-2.1.2-fix}

- 修复 `process` 采集器在特殊环境下可能丢失 `container_id` 字段的问题，现在会遍历 `/proc/{pid}/cgroup` 中所有行以提取合法容器 ID（#3092）
- 修复日志采集在容器频繁创建/删除场景下可能产生过多路径扫描耗时指标，导致 Prometheus 指标内存占用过高甚至 OOM 风险的问题（#3097）
- 修复容器日志采集在容器删除竞态场景下可能导致 `Single` goroutine 无法退出的问题，避免 goroutine 泄露（#3100）

---

## 2.1.1(2026/06/04) {#cl-2.1.1}

本次发布属于 hotfix 修复，内容如下：

### 兼容调整 {#cl-2.1.1-brk}

- 浏览器拨测统一切换为 Lightpanda 引擎，并将 Lightpanda 内置到 DataKit 镜像中（#3098）

---

## 2.1.0(2026/06/03) {#cl-2.1.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-2.1.0-new}

- 新增浏览器拨测任务 `BROWSER`，可用于模拟真实浏览器访问页面、执行交互和上报拨测结果（#3072）
- DataKit CLI 新增 `datakit completion` 命令，支持生成 bash、zsh、fish、PowerShell 补全脚本（#3066）

### 功能优化 {#cl-2.1.0-opt}

- 优化 Pipeline 指数级性能回退问题，降低连续 `GREEDYDATA` Grok 解析场景下的处理耗时（#3094）

### 兼容调整 {#cl-2.1.0-breaking}

- DataKit CLI 迁移到 Cobra 命令框架，相关命令进行了调整，具体可查看[服务管理](datakit-service-how-to.md#manage-service)和[命令补全](datakit-tools-how-to.md#completion)文档（#3066）

---

## 2.0.0(2026/05/27) {#cl-2.0.0}

本次发布是 DataKit 首个 2.x 主线版本，正式启用独立的 datakit-v2 安装/升级源，主要有如下更新：

### 新加功能 {#cl-2.0.0-new}

- BPF 网络日志支持记录 HTTP header，便于在 L7 日志中保留 trace 相关上下文（#3081）
- HTTP API 新增 pull 接口支持，并支持单次数据上报时关闭 filter 处理（#3076）
- Prometheus Remote Write 采集器新增 `keep_exist_metric_name` 配置，支持保留原始指标名（#3075）
- Logfwd 支持通过环境变量配置 `from_beginning_threshold_size`（#3074）
- DDTrace 采集器实现 info 接口，完善采集器信息输出能力（#3073）
- `hostobject` 支持金山云 KEC 元数据适配（#3068）
- OceanBase 采集器兼容 4.x 版本系统视图（#3067）
- DataKit 支持自动 dump 自身 profile 并上报到中心，便于线上问题定位（#3064）
- Redis 采集器新增 database object 采集上报能力（#2984）

### 问题修复 {#cl-2.0.0-fix}

- 修复 HTTP/3 拨测 Timing 字段采集异常导致 Download 显示异常时长的问题（#3080）
- 修复 Kubernetes Prometheus 采集器遇到 ServiceAccount token 文件过期时的异常问题（#3071）
- 修复热加载后 6060 pprof 服务退出后不再重启的问题（#3063）
- 修复磁盘和 `hostobject` 采集过滤时机偏晚，可能触发特殊挂载点自动挂载并影响 object 上报的问题（#3044）
- 修复 Flameshot 中 `jcmd` 相关处理与 profile tag 重复问题（#3062）

### 功能优化 {#cl-2.0.0-opt}

- 调整 OSS/CDN 上 DataKit v2 相关文件路径，适配大版本发布流程（#3083）
- 优化拨测采集器 traceroute 默认配置（#3078）
- 调整 DataKit Sinker Header 处理，仅 point 写入类请求携带 global tag（#3070）
- 优化 profile 上报可观测性，补充 info 信息和相关指标（#3055）
- 支持新的 Docker API 版本，提升容器环境兼容性（#2991）

### 兼容调整 {#cl-2.0.0-brk}

- DataKit 升级到 v2 大版本，并保留 v1 永久暂存分支；升级逻辑会检测当前环境是否支持升级到 v2（#3008）
- DataKit v2 启用独立安装/升级源 `https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/`，与 1.x 源 `https://static.<<<custom_key.brand_main_domain>>>/datakit/` 相互隔离且不互通。
- 1.x 自动升级不会跨代到 2.x；从 1.x 升级到 2.x 需要手动将安装或升级 URL 中的 `datakit` 替换为 `datakit-v2`。
- DataKit v2 构建工具链升级到 Go 1.26.2；系统要求调整为 Linux kernel >= 3.2、Windows Server 2016 及以上、macOS 12+。

---

## 1.94.0(2026/05/13) {#cl-1.94.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.94.0-new}

- 日志采集 socket logs 支持记录来源 IP，并使用 `collector_source_ip` 作为来源地址 tag（#3061）
- 拨测 HTTP 任务支持自定义协议版本，便于覆盖不同 HTTP 协议兼容性场景（#3041）
- PostgreSQL DBM 新增 SQL 执行量与 QPS 指标采集上报能力（#3046）
- MongoDB 支持 database object 采集上报（#3045）
- Doris 支持 object 采集上报（#3043）
- bug report 默认支持通过 Dataway 上传，并保留 OSS 直传方式（#3028）
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

## 1.94.1(2026/07/08) {#cl-1.94.1}

本次发布属于 hotfix 修复，内容如下：

### 兼容调整 {#cl-1.94.1-brk}

- 新增全局 `measurement_version` 配置，支持安装/升级时统一指定指标集版本，覆盖 MySQL、Oracle、PostgreSQL、SQLServer、Redis、JVM 等采集器（#3142）

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

### Breaking Change {#cl-1.90.0-breaking}

- 删除旧的命令补全入口 `datakit tool --setup-completer-script`、`datakit tool --completer-script` 以及静态脚本 `datakit-completer.sh`，统一改为使用 `datakit completion`

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
