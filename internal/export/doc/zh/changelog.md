# 更新日志
---

## 1.23.0(2024/01/11) {#cl-1.23.0}

### 新增功能 {#cl-1.23.0-new}

- Kubernetes 部署时支持通过环境变量（`ENV_DATAKIT_INPUTS`）配置任何采集器配置（#2068）
- 容器采集器支持更精细的配置，将 Kubernetes 对象 label 转换为采集数据的 tags（#2064）
    - `ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC`：支持将 label 转换成指标类数据的 tag
    - `ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2` 支持将 label 转换为非指标类（如对象/日志等）数据的 tag

### 问题修复 {#cl-1.23.0-fix}

- 修复容器采集器的 `deployment` 和 `daemonset` 字段偶发错误的问题（#2081）
- 修复容器日志采集在容器短暂运行并退出后，会丢失最后几行日志的问题（#2082）
- 修复 [Oracle](../integrations/oracle.md) 采集器慢查询 SQL 时间错误（#2079）
- 修复 Prom 采集器 `instance` 设置问题（#2084）

### 功能优化 {#cl-1.23.0-opt}

- 优化 Prometheus Remote Write 采集（#2069）
- eBPF 采集支持设置资源占用（#2075）
- 优化 Profiling 数据采集流程（#2083）
- [MongoDB](../integrations/mongodb.md) 采集器支持用户名和密码单独配置（#2073）
- [SQLServer](../integrations/sqlserver.md) 采集器支持配置实例名称（#2074）
- 优化 [ElasticSearch](../integrations/elasticsearch.md) 采集器视图和监控器（#2058）
- [KafkaMQ](../integrations/kafkamq.md) 采集器支持多线程模式（#2051）
- [SkyWalking](../integrations/skywalking.md) 采集器增加支持 Meter 数据类型（#2078）
- 更新一部分采集器文档以及其他 bug 修复（#2074/#2067）
- 优化 Proxy 代理安装时的升级命令（#2033）
- 优化非 root 用户安装时资源限制功能（#2011）

---

## 1.22.0(2023/12/28) {#cl-1.22.0}

### 新增功能 {#cl-1.22.0-new}

- 新增 [OceanBase](../integrations/oceanbase.md) 自定义 SQL 采集（#2046）
- 新增 [Prometheus Remote](../integrations/prom_remote_write.md) 黑名单/白名单（#2053）
- Kubernetes 资源数量采集添加 `node_name` tag（仅支持 Pod 资源）（#2057）
- Kubernetes Pod 指标新增 `cpu_limit_millicores/mem_limit/mem_used_percent_base_limit` 字段
- eBPF 采集器新增 `bpf-netlog` 插件 (#2017)

### 问题修复 {#cl-1.22.0-fix}

- 修复 [`external`](../integrations/external.md) 采集器僵尸进程问题（#2063）
- 修复容器日志 tags 冲突问题（#2066）
- 修复虚拟网卡信息获取失败问题 (#2050)
- 修复 Pipeline Refer table 和 IPDB 功能失效问题 (#2045)

### 优化 {#cl-1.22.0-opt}

- 优化 DDTrace 和 OTEL 字段提取白名单功能 （#2056）
- 优化 [SQLServer](../integrations/sqlserver.md) 采集器的 `sqlserver_lock_dead` 指标获取 SQL（#2049）
- 优化 [PostgreSQL](../integrations/postgresql.md) 采集器的连接库（#2044）
- 优化 [ElasticSearch](../integrations/elasticsearch.md) 采集器的配置文件，设置 `local` 默认为 `false`（#2048）
- K8s 安装时增加更多 ENV 配置项（#2025）
- 优化 Datakit 自身指标暴露
- 更新部分采集器集成文档

---

## 1.21.1(2023/12/21) {#cl-1.21.1}
本次发布属于 Hotfix 发布，修复如下问题：

- 修复 Prometheus Remote Write 不添加 Datakit 主机类 Tag 问题，主要兼容之前的老配置（#2055）
- 修复一批中间件默认的日志采集不加主机类 Tag 问题
- 日志采集修复中文字符颜色擦除乱码问题

---

## 1.21.0(2023/12/14)

本次发布属于迭代发布，主要有如下更新：

### 新增功能 {#cl-1.21.0-new}

- 添加 [ECS Fargate 采集模式](ecs-fargate.md)（#2018）
- 添加 [Prometheus Remote](../integrations/prom_remote_write.md) 采集器 tag 白名单（#2031）

### 问题修复 {#cl-1.21.0-fix}

- 修复 [PostgreSQL](../integrations/postgresql.md) 采集器版本检测问题（#2040）
- 修复 [ElasticSearch](../integrations/elasticsearch.md) 采集器帐号权限设置问题（#2036）
- 修复 [Host Dir](../integrations/hostdir.md) 采集器采集磁盘根目录崩溃问题（#2037）

### 优化 {#cl-1.21.0-opt}

- 优化 DDTrace 采集器：[去除 `message.Mate` 中重复的标签](../integrations/ddtrace.md#tags)（#2010）
- 优化容器内日志文件的路径搜寻策略（#2027）
- [拨测采集器](../integrations/dialtesting.md)增加 `datakit_version` 字段以及采集时间设置为任务开始执行的时间（#2029）
- 移除了 `datakit export` 命令优化二进制包大小（#2024）
- [调试采集器配置](why-no-data.md#check-input-conf) 中增加采集点的时间线数量（#2016）
- [Profile 采集](../integrations/profile.md)使用磁盘缓存实现异步化上报（#2041）
- 优化 Windows 下 Datakit 安装脚本（#2026）
- 更新一批采集器的内置视图和监控器

### Breaking Changes {#cl-1.21.0-brk}

- DDTrace 采集不再默认提取所有字段，这可能会导致某些页面自定义字段的数据缺失。可以通过编写 Pipeline 或者新的 JSON 查看语法（`message@json.meta.xxx`）来提取特定的字段

---

## 1.20.1(2023/12/07) {#cl-1.20.1}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.20.1-fix}

- 修复 DDTrace 一个采样 bug
- 修复 `error_message` 丢失信息的 bug
- 修复 Kubernetes Pod 对象数据没有正确采集 deployment 字段的 bug

## 1.20.0(2023/11/30) {#cl-1.20.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.20.0-new}

- [Redis](../integrations/redis.md) 采集器新增 hotkey 指标（#2019）
- monitor 命令支持播放 [bug report](why-no-data.md#bug-report) 中的指标数据（#2001）
- [Oracle](../integrations/oracle.md) 采集器增加自定义查询（#1929）
- [Container](../integrations/container.md) 容器内的日志文件支持通配采集（#2004）
- Kubernetes Pod 指标支持 `network` 和 `storage` 字段采集（#2022）
- [RUM](../integrations/rum.md) 新增配置支持对会话重放进行过滤（#1945）

### 问题修复 {#cl-1.20.0-fix}

- 修复 cgroup 在某些而环境下出现的 panic 错误（#2003）
- 修复 Windows 安装脚本在低版本 PowerShell 下执行失败（#1997）
- 修复磁盘缓存默认开启问题（#2023）
- 调整 Kubernetes Auto-Discovery 的 Prom 指标集命名风格（#2015）

### 功能优化 {#cl-1.20.0-opt}

- 优化内置采集器模板视图和监控器视图导出逻辑以及更新 MySQL/PostgreSQL/SQLServer 视图模板（#2008/#2007/#2013/#2024）
- 优化 Prom 采集器自身指标名称（#2014）
- 优化 Proxy 采集器，提供基本性能测试基准（#1988）
- 容器日志采集支持添加所属 Pod 的 Labels（#2006）
- 采集 Kubernetes 数据时默认使用 `NODE_LOCAL` 模式，需要添加额外的 [RBAC](../integrations/container.md#rbac-nodes-stats)（#2025）
- 优化链路处理流程（#1966）
- 重构 PinPoint 采集器，优化上下级关系 (#1947)
- APM 支持丢弃 `message` 字段以节约存储（#2021）

---

## 1.19.2(2023/11/20) {#cl-1.19.2}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.19.2-fix}

- 修复磁盘缓存 bug 导致 session replay 数据丢失问题
- 增加 Kubernetes 中资源采集耗时有关的 Prometheus 指标

---

## 1.19.1(2023/11/17) {#cl-1.19.1}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.19.1-fix}

- 修复磁盘缓存因 *.pos* 文件无法启动问题（[issue](https://github.com/GuanceCloud/cliutils/pull/59){:target="_blank"} ）

---

## 1.19.0(2023/11/16) {#cl-1.19.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.19.0-new}

- 支持 [OceanBase](../integrations/oceanbase.md) MySQL 模式采集（#1952）
- 新增[数据录制/播放](datakit-tools-how-to.md#record-and-replay)功能（#1738）

### 问题修复 {#cl-1.19.0-fix}

- 修复 Windows 低版本资源限制无效问题（#1987）
- 修复 ICMP 拨测问题（#1998）

### 功能优化 {#cl-1.19.0-opt}

- 优化 statsd 采集（#1995）
- 优化 Datakit 安装脚本（#1979）
- 优化 MySQL 内置视图（#1974）
- 完善 Datakit 自身指标暴露，增加完整 Golang 运行时等多项指标（#1971/#1969）
- 其它文档优化以及单元测试优化（#1952/#1993）
- 完善 Redis 指标采集，增加更多指标（#1940）
- TCP 拨测中允许增加报文（只支持 ASCII 文本）检测（#1934）
- 优化非 root 用户安装时的问题：
    - 可能因 ulimit 设置失败无法启动（#1991）
    - 完善文档，增加非 root 安装时的受限功能描述（#1989）
    - 调整非 root 安装时的前置操作，改为用户手动配置，避免不同操作系统可能存在的命令差异（#1990）
- MongoDB 采集器增加对老版本 2.8.0 的支持（#1985）
- RabbitMQ 采集器增加对低版本（3.6.X/3.7.X）的支持（#1944）
- 优化 Kubernetes 中 Pod 指标采集，以替换原始 Metric Server 方式（#1972）
- Kubernetes 下采集 Prometheus 指标时允许增加指标集名称配置（#1970）

### 兼容调整 {#cl-1.19.0-brk}

- 由于新增了数据录制/播放功能，故移除将数据写入文件的功能（#1738）

---

## 1.18.0(2023/11/02) {#cl-1.18.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.18.0-new}

- 新增 OceanBase 采集（#1924）

### 问题修复 {#cl-1.18.0-fix}

- 修复 Tracing 数据中较大 Tag 值兼容，现已调至 32MB（#1932）
- 修复 RUM session replay 脏数据问题（#1958）
- 修复指标信息导出问题（#1953）
- 修复 [v2 版本协议](datakit-conf.md#dataway-settings)构建错误问题

### 功能优化 {#cl-1.18.0-opt}

- 主机目录采集和磁盘采集中，新增挂载点等指标（#1941）
- KafkaMQ 支持 OpenTelemetry Tracing 数据处理（#1887）
- Bug Report 新增更多信息收集（#1908）
- 完善 Prom 采集过程中自身指标暴露（#1951）
- 更新默认 IP 库以支持 IPv6（#1957）
- 更新镜像名下载地址为 `pubrepo.guance.com`（#1949）
- 优化日志采集文件位置功能（#1961）
- Kubernetes
    - 支持 Node-Local Pod 信息采集，以缓解选举节点压力（#1960）
    - 容器日志采集支持更多粒度的过滤（#1959）
    - 增加 service 相关的指标采集（#1948）
    - 支持筛选 PodMonitor 和 ServiceMonitor 上的 Label 功能（#1963）
    - 支持将 Node Label 转换为 Node 对象的 Tag（#1962）

### 兼容调整 {#cl-1.18.0-brk}

- Kubernetes 不再采集 Job/CronJob 创建的 Pod 的 CPU/内存指标（#1964）

---

## 1.17.3(2023/10/31) {#cl-1.17.3}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.17.3-fix}

- 修复日志采集设置 Pipeline 无效问题（#1954）
- 修复 eBPF 在 arm64 平台无法运行的问题（#1955）

---

## 1.17.2(2023/10/27) {#cl-1.17.2}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.17.2-fix}

- 修复日志采集没有带 global host tag 的问题（#1942）
- 优化 Session Replay 数据的处理（#1943）
- 优化 Point 编码对非 UTF8 字符串的处理

---

## 1.17.1(2023/10/26) {#cl-1.17.1}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.17.1-fix}

- 修复拨测数据无法上传的问题

### 新加功能 {#cl-1.17.1-new}

- 新增通过 [eBPF 构建链路数据](../integrations/ebpftrace.md)，用来表示 Linux 进程/线程的调用关系（#1836）
- Pipeline 新增函数 [`pt_name`](../developers/pipeline/pipeline-built-in-function.md#fn-pt-name)（#1937）

### 功能优化 {#cl-1.17.1-opt}

- 优化 point 数据构建，提升内存使用效率（#1792）

---

## 1.17.0(2023/10/19) {#cl-1.17.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.17.0-new}

- Pod 添加 cpu_limit 指标 （#1913）
- `New Relic` 链路数据接入（#1834）

### 问题修复 {#cl-1.17.0-fix}

- 修复日志单行数据太长可能导致的内存问题（#1923）
- 修复 [disk](../integrations/disk.md) 采集器磁盘挂载点获取失败问题（#1919）
- 修复 helm 和 yaml 中的 Service 名称不一致问题（#1910）
- 修复 pinpoint span 中缺失 `agentid` 字段（#1897）
- 修复采集器中 `goroutine group` 错误处理问题（#1893）
- 修复[MongoDB](../integrations/mongodb.md) 采集器空数据上报问题（#1884）
- 修复 [rum](../integrations/rum.md) 采集器请求中出现大量 408 和 500 状态码（#1915）

### 功能优化 {#cl-1.17.0-opt}

- 优化 logfwd 的退出逻辑，避免因为配置错误导致程序退出影响到业务 Pod （#1922）
- 优化 [`ElasticSearch`](../integrations/elasticsearch.md) 采集器，增加索引指标集 `elasticsearch_indices_stats` 分片和副本等指标（#1921）
- 增加 [disk](../integrations/disk.md) 集成测试（#1920）
- DataKit monitor 支持 HTTPS（#1909）
- Oracle 采集器添加慢查询日志（#1906）
- 优化采集器 point 实现（#1900）
- [MongoDB](../integrations/mongodb.md) 采集器集成测试增加检测授权功能（#1885）
- 优化 Dataway 发送的重试功能，额外放出可配置参数

---

## 1.16.1(2023/10/09) {#cl-1.16.1}

本次发布属于 Hotfix 发布，修复如下问题：

### 问题修复 {#cl-1.16.1-fix}

- 修复 [K8s/容器采集器](../integrations/container.md) CPU 指标获取失败以及 containerd 下多行日志采集问题（#1895）
- 修复 [Prom 采集器](../integrations/prom.md)内存占用过大问题（#1905）

### Breaking Changes {#cl-1.16.1-bc}

- Tracing 数据采集的时候，所有 meta 信息中带 `-` 的字段名不会再被替换成 `_`。之所以这么修改，是为了避免 Tracing 数据和日志数据关联不上的问题（#1903）
- 所有 [Prom 采集器](../integrations/prom.md) 默认采用流式采集，以免未知的 Exporter 因数据量巨大造成 Datakit 大量的内存开销。

---

## 1.16.0(2023/09/21) {#cl-1.16.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.16.0-new}

- 新增 Neo4j 采集器（#1846）
- [RUM](../integrations/rum.md#upload-delete) 采集器新增 sourcemap 文件上传、删除和校验接口，并移除 DCA 服务中 sourcemap 上传和删除接口 (#1860)
- 新增 IBM Db2 采集器的监控视图和检测库（#1862）

### 问题修复 {#cl-1.16.0-fix}

- 修复环境变量 `ENV_GLOBAL_HOST_TAGS` 中使用 `__datakit_hostname` 无法获取主机 hostname 的问题 (#1874)
- 修复 [host_processes](../integrations/host_processes.md) 采集器指标数据缺少 `open_files` 字段 (#1875)
- 修复 Pinpoint 采集器 resource 大量为空的情况和 Pinpoint 占用内存过高问题 (#1857 #1849)

### 功能优化 {#cl-1.16.0-opt}

- 优化 Kubernetes 指标采集和对象采集的效率 (#1854)
- 优化日志采集的 metrics 输出 (#1881)
- Kubernetes Node 对象采集添加 unschedulable 和 node_ready 两个新字段 (#1886)
- [Oracle 采集器](../integrations/oracle.md)支持 Linux ARM64 架构（#1859）
- `logstreaming` 采集器增加集成测试（#1570）
- [Datakit 开发文档](development.md)中增加 IBM Db2 采集器内容（#1870）
- [Kafka](../integrations/kafka.md)、[MongoDB](../integrations/mongodb.md) 采集器文档完善（#1883）
- [MySQL](../integrations/mysql.md) 采集器监控帐号创建时，MySQL 8.0+ 默认采用 `caching_sha2_password` 加密方式 (#1882)
- 优化 [`bug report`](why-no-data.md#bug-report) 命令采集 syslog 文件过大问题（#1872）

### Breaking Changes {#cl-1.16.0-bc}

- 删除 DCA 服务中的 sourcemap 文件上传和删除接口，相关接口移至 [RUM](../integrations/rum.md#upload-delete) 采集器

---

## 1.15.1(2023/09/12) {#cl-1.15.1}

### 问题修复 {#cl-1.15.1-fix}

- 修复 logfwd 重复采集的问题

---

## 1.15.0(2023/09/07) {#cl-1.15.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.15.0-new}

- [Windows](datakit-install.md#resource-limit) 支持内存/CPU 限制（#1850）
- 新增 [IBM Db2 采集器](../integrations/db2.md)（#1818）

### 问题修复 {#cl-1.15.0-fix}

- 修复容器采集配置 include/exclude 的 double star 问题 (#1855)
- 修复一处 k8s Service 对象数据的字段错误

### 功能优化 {#cl-1.15.0-opt}

- [DataKit 精简版](datakit-install.md#lite-install)支持[日志](../integrations/logging.md)采集（#1861）
- [Bug Report](why-no-data.md#bug-report) 支持禁用 profile 数据采集（避免给当前 Datakit 造成压力）（#1868）
- Pipeline
    - 增加函数 `parse_int()` 和 `format_int()`（#1824）
    - 数据聚合函数 `agg_create()` 和 `agg_metric()` 支持输出任意类别的数据（#1865）
- 优化 Datakit 镜像大小（#1869）
- 文档
    - 增加[Datakit 指标性能测试报告](../integrations/datakit-metric-performance.md)（#1867）
    - 增加[external 采集器的使用文档](../integrations/external.md)（#1851）
    - 增加不同 Trace 传递说明的[文档](../integrations/tracing-propagator.md)（#1824）

---

## 1.14.2(2023/09/04) {#cl-1.14.2}

### 问题修复 {#cl-1.14.2-fix}

- 修复 Kubernetes 中 Pod 上 Prometheus Annotation 缺少 `instance` tag 的问题
- 修复 Pod 对象无法采集的问题

---

## 1.14.1(2023/08/30) {#cl-1.14.1}

### 问题修复 {#cl-1.14.1-fix}

- Kubernetes 中 Prometheus 指标采集优化（流式采集），避免可能的大量内存占用（#1853/#1845）
- 修复日志[颜色字符处理](../integrations/logging.md#ansi-decode)
    - Kubernetes 下环境变量为 `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES`

---

## 1.14.0(2023/08/24) {#cl-1.14.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.14.0-new}

- 新增采集器 [NetFlow](../integrations/netflow.md)（#1821）
- 新增[黑名单调试器](datakit-tools-how-to.md#debug-filter)（#1787）
- 新增 Kubernetes StatefulSet 指标和对象采集，新增 `replicas_desired` 对象字段（#1822）
- 新增 [DK_LITE](datakit-install.md#lite-install) 环境变量，用于安装 DataKit 精简版（#123）

### 问题修复 {#cl-1.14.0-fix}

- 修复 Container 和 Kubernetes 采集没有正确添加 HostTags 和 ElectionTags 的问题（#1833）
- 修复 [MySQL](../integrations/mysql.md#input-config) 自定义采集 Tags 为空时指标无法采集的问题（#1835）

### 功能优化 {#cl-1.14.0-opt}

- 增加 System 采集器中的 [process_count](../integrations/system.md#metric) 指标表示当前机器的进程数（#1838）
- 去掉 Process 采集器中的 [open_files_list](../integrations/host_processes.md#object) 字段（#1838）
- 增加[主机对象](../integrations/hostobject.md#faq)采集器文档中指标丢失的处理案例（#1838）
- 优化 Datakit 视图，完善 Datakit Prometheus 指标文档
- 优化 Pod/容器 日志采集的 [mount 方式](../integrations/container-log.md#logging-with-inside-config) (#1844)
- 增加 Process、System 采集器集成测试（#1841/#1842）
- 优化 etcd 集成测试（#1847）
- 升级 Golang 1.19.12（#1516）
- 增加通过 `ash` 命令[安装 DataKit](datakit-install.md#get-install) (#123)
- [RUM 采集](../integrations/rum.md)支持自定义指标集，默认的指标集新增 `telemetry`（#1843）

### 兼容调整 {#cl-1.14.0-brk}

- 移除 Datakit 端的 Sinker 功能，将其功能转移到 [Dataway 侧实现](../deployment/dataway-sink.md)（#1801）
- 移除 Kubernetes Deployment 指标数据的 `pasued` 和 `condition` 字段，新增对象数据 `paused` 字段

---

## 1.13.2(2023/08/15) {#cl-1.13.2}

### 问题修复 {#cl-1.13.2-fix}

- 修复 MySQL 自定义采集失败。(#1831)
- 修复 Prometheus Export 存在 Service 作用范围和执行的错误。（#1828）
- eBPF 采集器出现异常的 HTTP 响应码和延迟。(#1829)

### 功能优化 {#cl-1.13.2-opt}

- 完善容器采集的 image 字段取值。（#1830）
- MySQL 集成测试优化，提升测试速度。（#1826）

---

## 1.13.1(2023/08/11) {#cl-1.13.1}

- 修复容器日志 `source` 字段命名问题（#1827）

---

## 1.13.0(2023/08/10) {#cl-1.13.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.13.0-new}

- 主机对象采集器支持调试命令。（#1802）
- KafkaMQ 增加支持外部插件 handle 功能 (#1797)
- 容器采集支持 cri-o 运行时。 (#1763)
- Pipeline 增加用于指标生成的 `create_point` 函数 (#1803)
- 增加 PHP 语言的 Profiling 支持 （#1811）

### 问题修复 {#cl-1.13.0-fix}

- 修复 Cat 采集器 NPE 异常。
- 修复拨测采集器 http response_download 时间。（#1820）
- 修复 containerd 日志采集没有正常拼接 partial 日志的问题。（#1825）
- 修复 eBPF 采集器 `ebpf-conntrack` 插件探针失效问题。 (#1793)

### 功能优化 {#cl-1.13.0-opt}

- bug-report 命令优化（#1810）
- RabbitMQ 采集器支持多个同时运行。（#1756）
- 主机对象采集器调整。移除 state 字段。（#1802）
- 优化错误上报机制。解决 eBPF 采集器不能上报错误的问题。（#1802）
- Oracle 外置采集器增加在发生错误情况下将信息发送给中心。（#1802）
- 优化 Pythond 文档，增加 module not found 解决案例。（#1807）
- 部分采集器增加 global tag 的集成测试案例。（#1791）
- 优化 Oracle 集成测试。（#1802）
- OpenTelemetry 增加指标集和仪表板。
- 调整 k8s event 字段。 (#1766)
- 添加新的容器采集字段。 (#1819)
- eBPF 采集器增加流量字段至 `httpflow` 中。 (#1790)

---

## 1.12.3(2023/08/03) {#cl-1.12.3}

- 修复 Windows 下日志采集文件延迟释放问题（#1805）
- 修复新容器头部日志不采集的问题
- 修复几个正则表达式可能导致的崩溃问题（#1781）
- 修复安装包体积过大的问题（#1804）
- 修复日志采集器开启磁盘缓存可能失败的问题

---

## 1.12.2(2023/07/31) {#cl-1.12.2}

- 修复 OpenTelemetry Metric 和 Trace 路由配置问题

---

## 1.12.1(2023/07/28) {#cl-1.12.1}

- 修复老版本 DDTrace Python Profile 接入问题（#1800）

---

## 1.12.0(2023/07/27) {#cl-1.12.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.12.0-new}

- [HTTP API](apis.md##api-sourcemap) 增加 sourcemap 文件上传（#1782）
- 新增 .net Profiling 接入支持（#1772）
- 新增 Couchbase 采集器（#1717）

### 问题修复 {#cl-1.12.0-fix}

- 修复拨测采集器缺失 `owner` 字段问题（#1789）
- 修复 DDTrace 采集器缺失 `host` 问题，同时各类 Trace 的 tag 采集改为黑名单机制[^trace-black-list]（#1776）
- 修复 RUM API 跨域问题（#1785）

[^trace-black-list]: 各类 Trace 会在其数据上带上各种业务字段（称之为 Tag、Annotation 或 Attribute 等），Datakit 为了收集更多数据，默认这些字段都予以接收。

### 功能优化 {#cl-1.12.0-opt}

- 优化 SNMP 采集器加密算法识别方法；优化 SNMP 采集器文档，增加更多示例解释（#1795）
- 增加 Pythond 采集器 Kubernetes 部署示例，增加 Git 部署示例（#1732）
- 增加 InfluxDB、Solr、NSQ、Net 采集器集成测试（#1758/#1736/#1759/#1760）
- 增加 Flink 指标（#1777）
- 扩展 Memcached、MySQL 指标采集（#1773/#1742）
- 更新 Datakit 自身指标暴露（#1492）
- Pipeline 增加更多运算符支持（#1749）
- 拨测采集器
    - 增加拨测采集器内置仪表板（#1765）
    - 优化拨测任务启动，避免资源集中消耗（#1779）
- 文档更新（#1769/#1775/#1761/#1642）
- 其它优化（#1777/#1794/#1778/#1783/#1775/#1774/#1737）

---

## 1.11.0(2023/07/11) {#cl-1.11.0}

本次发布属于迭代发布，包含如下更新：

### 新加功能 {#cl-1.11.0-new}

- 新增 dk 采集器，移除 self 采集器(#1648)

### 问题修复 {#cl-1.11.0-fix}

- 修复 Redis 采集器时间线冗余的问题(#1743)，完善集成测试
- 修复 Oracle 采集器动态库安全问题(#1730)
- 修复 DCA 服务启动失败(#1731)
- 修复 MySQL/ElasticSearch 采集器集成测试(#1720)

### 功能优化 {#cl-1.11.0-opt}

- 优化 etcd 采集器(#1741)
- StatsD 采集器支持配置区分不同数据源(#1728)
- Tomcat 采集器支持 10 及以上版本，弃用 Jolokia(#1703)
- 容器日志采集支持配置容器内文件(#1723)
- SQLServer 采集器指标完善和集成测试功能重构(#1694)

### 兼容调整 {#cl-1.11.0-brk}

以下兼容性修改，可能会导致数据采集上的问题，如果您使用了以下的功能，请考虑是否升级，或者采用新的对应方案。

1. 移除容器日志的 `deployment` tag
1. 移除容器 stdout 日志的 `source` 字段以 `short_image_name` 来命名的逻辑。现在只使用容器名称或 Kubernetes 中的 label `io.kubernetes.container.name` 来命名[^cl-1.11.0-brk-why-1]。
1. 移除通过容器 label 采集其外挂文件路径的功能（`datakit/logs/inside`），改成通过[容器环境变量（`DATAKIT_LOGS_CONFIG`）](../integrations/container-log.md)的方式来实现[^cl-1.11.0-brk-why-2]。

[^cl-1.11.0-brk-why-1]: 在 Kubernetes 中，`io.kubernetes.container.name` 值是不变的，而主机容器中，容器名也不太变，故不再采用原始镜像名作为 `source` 字段的来源。
[^cl-1.11.0-brk-why-2]: 相比修改容器的 Label（一般情况下需要重新构建镜像），给容器追加环境变量更为方便（启动容器的时候，注入环境变量即可）。

---

## 1.10.2(2023/07/04) {#cl-1.10.2}

- 修复 Kubernetes 中 prom 采集器识别问题

## 1.10.1(2023/06/30) {#cl-1.10.1}

- 修复 OpenTelemetry HTTP 路由支持自定义
- 修复主机进程对象中启动时长（`started_duration`）字段缺失问题

---

## 1.10.0(2023/06/29) {#cl-1.10.0}

本次发布属于迭代发布，包含如下更新：

### 问题修复 {#cl-1.10.0-fix}

- 修复 Proxy 环境下 Profiling 数据上传问题（#1710）
- 修复升级过程中默认采集器开启问题（#1709）
- 修复 SQLServer 采集数据中日志被截断问题（#1689）
- 修复 Kubernetes 中 Metric Server 指标采集问题（#1719）

### 功能优化 {#cl-1.10.0-opt}

- KafkaMQ 支持 topic 级别的多行切割配置（#1661）
- Kubernetes DaemonSet 安装时支持通过 ENV 修改 Datakit 日志分片数和分片大小（#1711）
- Kubernetes Pod 指标和对象采集新增 `memory_capacity` 和 `memory_used_percent` 两个字段 (#1721)
- OpenTelemetry HTTP 路由支持自定义（#1718）
- Oracle 采集器优化 `oracle_system` 指标集丢失的问题，优化采集逻辑并增加部分指标（#1693）
- Pipeline 增加 `in` 运算符，增加 `value_type()` 和 `valid_json()` 函数，调整 `load_json()` 函数反序列化失败后的行为 (#1712)
- 主机进程对象中采集新增启动时长（`started_duration`）字段（#1722）
- 优化拨测数据发送逻辑（#1708）
- 更新更多集成测试（#1666/#1667/#1668/#1693/#1599/#1573/#1572/#1563/#1512/#1715）
- 模块重构以及优化（#1714/#1680/#1656）

### 兼容调整 {#cl-1.10.0-brk}

- Profile 数据的时间戳单位从纳秒改成微秒（#1679）

<!-- markdown-link-check-disable -->

---

## 1.9.2(2023/06/20) {#cl-1.9.2}

本次发布属于迭代中期发布，增加部分跟中心对接的功能以及一些 bug 修复和优化：

### 新加功能 {#cl-1.9.2-new}

- 新增 [Chrony 采集器](../integrations/chrony.md)（#1671）
- 新增 RUM Headless 支持（#1644）
- Pipeline
    - 新增 [offload 功能](../developers/pipeline/pipeline-offload.md)（#1634）
    - 重新调整了已有的文档结构（#1686）

### 问题修复 {#cl-1.9.2-fix}

- 修复一些可能导致崩溃的问题（!2249）
- HTTP 网络拨测增加 Host header 支持并修复随机的 error 报错（#1676）
- 修复 Kubernetes 中自动发现 Pod Monitor 和 Service Monitor 问题（#1695）
- 修复 Monitor 问题（#1702/!2258）
- 修复 Pipeline 数据误操作 bug（#1699）

### 功能优化 {#cl-1.9.2-opt}

- 在 Datakit HTTP API 返回中增加更多信息，便于错误排查（#1697/#1701）
- 其它重构（#1681/#1677）
- RUM 采集器增加更多 Prometheus 指标暴露（#1545）
- 默认开启 Datakit 的 pprof 功能，便于问题排查（#1698）

### 兼容调整 {#cl-1.9.2-brk}

- 移除 Kubernetes CRD `guance.com/datakits v1bate1` 对 logging 采集的支持（#1705）

---

## 1.9.1(2023/06/13) {#cl-1.9.1}

本次发布属于 bug 修复，主要修复如下问题：

- 修复 DQL 查询问题（#1688）
- 修复 HTTP 接口高频写入可能导致的崩溃问题（#1678）
- 修复 `datakit monitor` 命令参数覆盖问题（!2232）
- 修复 HTTP 上传数据时重试报错问题（#1687）

---

## 1.9.0(2023/06/08) {#cl-1.9.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.9.0-new}

- 新增 [NodeJS Profiling](../integrations/profile-nodejs.md) 接入支持（#1638）
- 新增点评 [Cat](../integrations/cat.md) 接入支持（#1593）
- 新增采集器配置[调试方法](why-no-data.md#check-input-conf)（#1649）

### 问题修复 {#cl-1.9.0-fix}

- 修复 K8s 中 Prometheus 指标采集导致的连接泄露问题（#1662）

### 功能优化 {#cl-1.9.0-opt}

- K8s DaemonSet 对象增加 `age` 字段（#1670）
- 优化 [PostgreSQL](../integrations/postgresql.md) 启动设置（#1658）
- SkyWalking 增加 [`/v3/log/`](../integrations/skywalking.md) 支持（#1654）
- 优化日志采集处理（#1652/#1651）
- 优化[升级文档](datakit-update.md#prepare)（#1653）
- 其它重构和优化（#1673/#1650/#1630）
- 新增若干集成测试（#1440/#1429）
    - PostgreSQL
    - 网络拨测

---

## 1.8.1(2023/06/01) {#cl-1.8.1}
本次发布属于 bug 修复，主要修复如下问题：

- 修复 KafkaMQ 多开情况下崩溃的问题（#1660）
- 修复 DaemonSet 模式下磁盘设备采集不全的问题（#1655）

---

## 1.8.0(2023/05/25) {#cl-1.8.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.8.0-new}

- Datakit 新增两个调试命令，便于用户在配置的过程中编写 glob 和正则表达式（#1635）
- 新增 DDTrace 和 OpenTelemetry 之间 Trace ID 双向透传功能（#1633）

### 问题修复 {#cl-1.8.0-fix}

- 修复拨测预检问题（#1629）
- 修复 SNMP 采集的两个字段问题（#1632）
- 修复升级服务跟其它服务的默认端口冲突问题（#1646）

### 功能优化 {#cl-1.8.0-opt}

- eBPF 采集 Kubernetes 网络数据时支持将 Cluster IP 转换为 Pod IP（需手动打开）（#1617）
- 新增一批集成测试（#1430/#1574/#1575）
- 优化容器网络相关的指标（#1397）
- Bug report 功能新增崩溃信息收集（#1625）
- PostgreSQL 采集器
    - 新增自定义 SQL 指标采集（#1626）
    - 新增 DB 级别的 tag （#1628）
- 优化 localhost 采集的 `host` 字段问题（#1637）
- 优化 Datakit 自身指标，新增 [Datakit 自身指标文档](datakit-metrics.md)（#1639/#1492）
- 优化 Pod 上的 Prometheus 指标采集，自动支持所有 Prometheus 指标类型（#1636）
- 新增 Trace 类采集的[性能测试文档](../integrations/datakit-trace-performance.md)（#1616）
- 新增 Kubernetes DaemonSet 对象采集（#1643）
- Pinpoint gRPC 服务支持 `x-b3-traceid` 透传 Trace ID（#1605）
- 优化集群选举策略（#1534）
- 其它优化（#1609#1624）

### 兼容调整 {#cl-1.8.0-brk}

- 容器采集器中，删除 `kube_cluster_role` 对象采集（#1643）

---

## 1.7.0(2023/05/11) {#cl-1.7.0}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.7.0-new}

- RUM Sourcemap 增加小程序支持（#1608）
- 增加新的采集选举策略，支持在 K8s 环境中进行 Cluster 级别的选举（#1534）

### 问题修复 {#cl-1.7.0-fix}

- Datakit 上传时，如果中心返回 5XX 状态码会导致四层连接数增加。本版本修复了该问题，同时在 [*datakit.conf*](datakit-conf.md#maincfg-example)（K8s 中可通过[环境变量配置](datakit-daemonset-deploy.md#env-dataway)）  中暴露更多连接有关的配置参数（DK001-15）

### 功能优化 {#cl-1.7.0-opt}

- 优化进程对象采集，默认关闭部分可能导致高消耗的字段（比如打开文件数/端口数）采集，这些字段通过采集器配置或者环境变量均可手动开启。这些字段可能很重要，但默认情况下我们还是认为，不应该因此导致主机上意外的性能开销（#1543）
- Datakit 自身指标优化：
    - 拨测采集器增加 Prometheus 指标暴露，便于排查拨测采集器自身一些潜在问题（#1591）
    - 增加 Datakit 上报时 HTTP 层面指标暴露（#1597）
    - 增加 KafkaMQ 采集时的指标暴露
- 优化 PostgreSQL 指标采集，增加了更多相关指标（#1596）
- 优化 JVM 有关的指标采集，主要是文档更新（#1600）
- Pinpoint
    - 增加更多开发者文档（#1601）
    - Pinpoint 修复 gRPC Service 支持（#1605）
- 优化磁盘指标采集在不同平台上的差异（#1607）
- 其它工程优化（#1621/#1611/#1610）
- 增加若干集成测试（#1438/#1561/#1585/#1435/#1513）

---

## 1.6.1(2023/04/27) {#cl-1.6.1}

本次发布属于 Hotfix 发布，修复如下问题：

- 老版本升级上来可能导致黑名单不生效(#1603)
- [Prom](../integrations/prom.md) 采集 `info` 类数据问题(#1544)
- 修复 Dataway Sinker 模块可能导致的数据丢失问题(#1606)

---

## 1.6.0(2023/04/20) {#cl-1.6.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.6.0-new}

- 新增 [Pinpoint](../integrations/pinpoint.md) API 接入(#973)

### 功能优化 {#cl-1.6.0-opt}

- 优化 Windows 安装脚本和升级脚本输出方式，便于在终端直接黏贴复制(#1557)
- 优化 Datakit 自身文档构建流程(#1578)
- 优化 OpenTelemetry 字段处理(#1514)
- [Prom](prom.md) 采集器支持采集 `info` 类型的 label 并将其追加到所有关联指标上（默认开启）(#1544)
- 在 [system 采集器](system.md)中，新增 CPU 和内存占用百分比指标(#1565)
- Datakit 在发送的数据中，增加数据点数标记（`X-Points`），便于中心相关指标构建(#1410)
    - 另外优化了 Datakit HTTP 的 `User-Agent` 标记，改为 `datakit-<os>-<arch>/<version>` 这种形态
- [KafkaMQ](kafkamq.md)：
    - 支持处理 Jaeger 数据(#1526)
    - 优化 SkyWalking 数据的处理流程(#1530)
    - 新增第三方 RUM 接入功能(#1581)
- [SkyWalking](skywalking.md) 新增 HTTP 接入功能(#1533)
- 增加如下集成测试：
    - [Apache](apache.md)(#1553)
    - [JVM](jvm.md)(#1559)
    - [Memcached](memcached.md)(#1552)
    - [MongoDB](mongodb.md)(#1525)
    - [RabbitMQ](rabbitmq.md)(#1560)
    - [Statsd](statsd.md)(#1562)
    - [Tomcat](tomcat.md)(#1566)
    - [etcd](etcd.md)(#1434)

### 问题修复 {#cl-1.6.0-fix}

- 修复 [JSON 格式](apis.md#api-json-example)写入数据时无法识别时间精度的问题(#1567)
- 修复拨测采集器不工作的问题(#1582)
- 修复 eBPF 在欧拉系统上验证器问题(#1568)
- 修复 RUM sourcemap 段错误问题(#1458)
<!-- - 修复进程对象采集器可能导致高 CPU 问题，默认情况下关闭了部分高消耗字段（listen 端口）的采集(#1543) -->

### 兼容调整 {#cl-1.6.0-brk}

- 移除老的命令行风格，比如，原来的 `datakit --version` 将不再生效，须以 `datakit version` 代之。详见[各种命令的用法](datakit-tools-how-to.md)

---

<!-- markdownlint-disable -->

## 1.5.10(2023/04/13) {#cl-1.5.10}

本次发布属于紧急发布，主要有如下更新：

### 新加功能 {#cl-1.5.10-new}

- 支持自动发现并采集 [Pod 上的 Prometheus 指标](kubernetes-prom.md#auto-discovery-metrics-with-prometheus)(#1564)
- Pipeline 新增聚合类函数(#1554)
    - [agg_create()](../developers/pipeline/pipeline-built-in-function.md#fn-agg-create)
    - [agg_metric()](../developers/pipeline/pipeline-built-in-function.md#fn-agg-metric)

### 功能优化 {#cl-1.5.10-opt}

- 优化了 Pipeline 执行性能，大约有 30% 左右性能提升
- 优化日志采集中历史位置记录操作(#1550)

---

## 1.5.9(2023/04/06) {#cl-1.5.9}
本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.9-new}

- 新增[伺服服务](datakit-update.md#remote)，用来管理 Datakit 升级(#1441)
- 新增[故障排查功能](why-no-data.md#bug-report)(#1377)

### 问题修复 {#cl-1.5.9-fix}

- 修复 Datakit 自身 CPU 指标获取，保持 monitor 和 `top` 命令获取到的 CPU 同步(#1547)
- 修复 RUM 采集器 panic 错误(#1548)

### 功能优化 {#cl-1.5.9-opt}

- 优化升级功能，避免 *datakit.conf* 文件被破坏(#1449)
- 优化 [cgroup 配置](datakit-conf.md#resource-limit)，移除 CPU 最小值限制(#1538)
- 优化 *self* 采集器，我们能选择是否开启该采集器，同时对其采集性能做了一些优化(#1386)
- 由于有了新的故障排查手段，简化了现有 monitor 展示(#1505)
- [Prom 采集器](prom.md)允许增加 *instance tag*，以保持跟原生 Prometheus 体系一致(#1517)
- [DCA](dca.md) 增加 Kubernetes 部署方式(#1522)
- 优化日志采集的磁盘缓存性能(#1487)
- 优化 Datakit 自身指标体系，暴露更多 [Prometheus 指标](apis.md#api-metrics)(#1492)
- 优化 [/v1/write](apis.md#api-v1-write)(#1523)
- 优化安装过程中 token 出错提示(#1541)
- monitor 支持自动从 *datakit.conf* 中获取连接地址(#1547)
- 取消 eBPF 对内核版本的强制检查，尽量支持更多的内核版本(#1542)
- [Kafka 订阅采集](kafkamq.md)支持多行 JSON 功能(#1549)
- 新增一大批集成测试(#1479/#1460/#1436/#1428/#1407)
- 优化 IO 模块的配置，新增上传 worker 数配置字段(#1536)
    - [Kubernetes](datakit-daemonset-deploy.md#env-io)
    - [*datakit.conf*](datakit-conf.md#io-tuning)

### 兼容调整 {#cl-1.5.9-brk}

- 本次移除了大部分 Sinker 功能，只保留了 [Dataway 上的 Sinker 功能](datakit-sink-dataway.md)(#1444)。同时 sinker 的[主机安装配置](datakit-install.md#env-sink)以及 [Kubernetes 安装配置](datakit-daemonset-deploy.md#env-sinker)都做了调整，其中的配置方式也跟之前不同，请大家升级的时候，注意调整
- 老版本的[发送失败磁盘缓存](datakit-conf.md#io-disk-cache)由于性能问题，我们替换了实现方式。新的实现方式，其缓存的二进制格式不再兼容，如果升级的话，老的数据将不被识别。建议先**手动删除老的缓存数据**（老数据可能会影响新版本磁盘缓存），然后再升级新版本的 Datakit。尽管如此，新版本的磁盘缓存，仍然是一个实验性功能，请谨慎使用
- Datakit 自身指标体系做了更新，原有 DCA 获取到的指标将有一定的缺失，但不影响 DCA 本身功能的运行

---

## 1.5.8(2023/03/24) {#cl-1.5.8}
本次发布属于迭代发布，主要是一些问题修复和功能完善。

### 问题修复 {#cl-1.5.8-fix}

- 修复容器日志采集可能丢失的问题(#1520)
- Datakit 启动后自动创建 Pythond 目录(#1484)
- 移除 [主机目录](hostdir.md) 采集器单例限制(#1498)
- 修复一个 eBPF 数值构造的问题(#1509)
- 修复 Datakit monitor 参数识别问题(#1506)

### 功能优化 {#cl-1.5.8-opt}

- 补全 Jenkins 采集器内存有关的指标(#1489)
- 完善 [cgroup v2](datakit-conf.md#resource-limit) 支持(#1494)
- Kubernetes 安装时增加环境变量（`ENV_CLUSTER_K8S_NAME`）来配置 cluster 名称(#1504)
- Pipeline
    - [`kv_split()`](../developers/pipeline/pipeline-built-in-function.md#fn-kv_split) 函数增加强制保护措施，避免数据膨胀(#1510)
    - 关于 JSON 的处理，优化了 [`json()`](../developers/pipeline/pipeline-built-in-function.md#fn-json) 和 [`delete()`](../developers/pipeline/pipeline-built-in-function.md#fn-delete) 删除 key 的功能。
- 其它工程上的优化(#1500)

### 文档调整 {#cl-1.5.8-doc}

- 增加 Kubernetes 全离线安装[文档](datakit-offline-install.md#k8s-offline)(#1480)
- 完善 StatsD 以及 DDTrace-Java 有关的文档(#1481/#1507)
- 补充 TDEngine 有关的文档(#1486)
- 移除 disk 采集器文档中的过时字段描述(#1488)
- 完善 Oracle 采集器文档(#1519)

## 1.5.7(2023/03/09) {#cl-1.5.7}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.7-new}

- Pipeline
    - `json` 函数增加 [key 删除](../developers/pipeline/pipeline-built-in-function.md#fn-json) 功能(#1465)
    - 增加函数 [`kv_split()`](../developers/pipeline/pipeline-built-in-function.md#fn-kv_split)(#1414)
    - 增加[时间函数](../developers/pipeline/pipeline-built-in-function.md#fn-datetime)(#1411)
- 增加 [IPv6 支持](datakit-conf.md#config-http-server)(#1454)
- 磁盘 IO 采集支持 [io wait 扩展指标](diskio.md#extend)(#1472)
- 容器采集支持 [Docker 和 containerd 共存](container.md#requrements)(#1401)
- 整合 [Datakit Operator 配置文档](datakit-operator.md)(#1482)

### 问题修复 {#cl-1.5.7-fix}

- 修复 Pipeline Bugs(#1476/#1469/#1471/#1466)
- 修复 *datakit.yaml* 缺少 `request` 导致的容器 Pending(#1470)
- 修复云同步过程中反复探测问题(#1443)
- 修复日志磁盘缓存的编码错误(#1474)

### 功能优化 {#cl-1.5.7-opt}

- 优化 Point Checker(#1478)
- 优化 Pipeline [`replace()`](../developers/pipeline/pipeline-built-in-function.md#fn-replace.md) 性能(#1477)
- 优化 Windows 下 Datakit 安装流程(#1404)
- 优化 [配置中心](confd.md) 的配置处理流程(#1402)
- 添加 [Filebeat](beats_output.md) 集成测试能力(#1459)
- 添加 [Nginx](nginx.md) 集成测试能力(#1399)
- 重构 [OpenTelemetry Agent](opentelemetry.md)(#1409)
- 重构 [Datakit Monitor 信息](datakit-monitor.md#specify-module)(#1261)

---

## 1.5.6(2023/02/23) {#cl-1.5.6}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.6-new}

- 命令行增加[解析行协议功能](datakit-tools-how-to.md#parse-lp)(#1412)
- Datakit yaml 和 helm 支持资源 limit 配置(#1416)
- Datakit yaml 和 helm 支持 CRD 部署(#1415)
- 添加 SQLServer 集成测试(#1406)
- RUM 支持 [resource CDN 标注](rum.md#cdn-resolve)(#1384)

### 问题修复 {#cl-1.5.6-fix}

- 修复 RUM 请求返回 5xx 问题(#1412)
- 修复日志采集路径错误问题(#1447)
- 修复 K8s Pod(`restarts`) 字段问题(#1446)
- 修复 DataKit filter 模块崩溃问题(#1422)
- 修复 Point 构建中 tag key 命名问题(#1413#1408)
- 修复 Datakit Monitor 字符集问题(#1405)
- 修复 OTEL tag 覆盖问题(#1396)
- 修复 public API 白名单问题(#1467)

### 功能优化 {#cl-1.5.6-opt}

- 优化拨测中无效任务的处理(#1421)
- 优化 Windows 下安装提示(#1404)
- 优化 Windows 中 Powershell 安装脚本模板(#1403)
- 优化 K8s 中 Pod/ReplicaSet/Deployment 的关联方法(#1368)
- 重构 point 数据结构及功能(#1400)
- Datakit 自带 [eBPF](ebpf.md) 采集器二进制安装(#1448)
- 安装程序地址改成 CDN 地址，优化下载问题(#1457)

### 兼容调整 {#cl-1.5.6-brk}

- 由于内置了 eBPF 采集器，移除多余命令 `datakit install --datakit-ebpf`(#1400)

---

## 1.5.5(2023/02/09) {#cl-1.5.5}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.5-new}

- Datakit 主机安装可自定义默认采集器开启(#1392)
- 提供 OTEL 的错误追踪(#1309)
- 提供 RUM Session 回放能力(#1283)

### 问题修复 {#cl-1.5.5-fix}

- 修复日志堆积问题(#1394)
- 修复 conf.d 重复启动采集器问题(#1385)
- 修复 OTEL 数据关联问题(#1364)
- 修复 OTEL 采集数据字段覆盖问题(#1383)
- 修复 Nginx Host 识别错误(#1379)
- 修复拨测超时(#1378)
- 修复云厂商实例识别(#1382)

### 功能优化 {#cl-1.5.5-opt}

- Datakit Pyroscope Profiling 多程序语言识别(#1374)
- 优化 CPU,Disk,eBPF,Net 等中英文文档(#1375)
- 优化 ElasticSearch, PostgreSQL, DialTesting 等英文文档(#1373)
- 优化 DCA,Profiling 文档(#1371#1372)
- 优化日志采集流程(#1366)
- [IP 库安装文档更新](datakit-tools-how-to.md) 配置方法文档支持(#1370)

---

## 1.5.4(2023/01/13) {#cl-1.5.4}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.4-new}

- [Confd 增加 Nacos 后端](confd.md)(#1315#1327)
- 日志采集器添加 LocalCache 特性(#1326)
- 支持 [C/C++ Profiling](profile-cpp.md)(#1320)
- RUM Session Replay 文件上报(#1283)
- WEB DCA 支持远程更新 config(#1284)

### 问题修复 {#cl-1.5.4-fix}

- 修复 K8S 采集失败数据丢失(#1362)
- 修复 K8S Host 字段错误(#1351)
- 修复 K8S Metrics Server 超时(#1353)
- 修复 Containerd 环境下 annotation 配置问题(#1352)
- 修复 Datakit 重新加载过程中采集器崩溃(#1359)
- 修复 Golang Profiler 函数执行时间计算错误(#1335)
- 修复 Datakit Monitor 字符集问题(#1321)
- 修复 async-profiler 服务现实问题(#1290)
- 修复 Redis 采集器 `slowlog` 问题(#1360)

### 功能优化 {#cl-1.5.4-opt}

- 优化 SQL 数据资源占用较高问题(#1358)
- 优化 Datakit Monitor(#1222)

---

## 1.5.3(2022/12/29) {#cl-1.5.3}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.3-new}

- Prometheus 采集器支持通过 Unix Socket 采集数据(#1317)
- 允许[非 root 用户运行 DataKit](datakit-install.md#common-envs)(#1153)

### 问题修复 {#cl-1.5.3-fix}

- 修复 netstat 采集器链接数问题(#1276/#1336)
- 修复 Go profiler 差值问题(#1328)
- 修复 Datakit 重启超时问题(#1297)
- 修复 Kafka 订阅消息被截断问题(#1338)
- 修复 Pipeline `drop()` 函数无效问题(#1343)

### 功能优化 {#cl-1.5.3-opt}

- 优化 eBPF 中 `httpflow` 协议判定(#1318)
- 优化 Windows 下 Datakit 安装升级命令(#1316)
- 优化 Pythond 使用封装(#1304)
- Pipeline 提供更详细的操作报错信息(#1262)
- Pipeline Ref-Table 提供基于 SQLite 的本地化存储实现(#1158)
- 优化 SQLServer 时间线问题(#1345)

---

## 1.5.2(2022/12/15) {#cl-1.5.2}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.2-new}

- 新增 [Golang Profiling](profile-go.md) 接入(#1265)

### 问题修复 {#cl-1.5.2-fix}

- logfwd 修复不采集问题(#1288)
- 修复 cgroup 不生效问题(#1293)
- 修复 DataKit 服务操作超时问题(#1297)
- 修复 SQLServer 采集可能卡死的问题(#1289)

### 功能优化 {#cl-1.5.2-opt}

- logfwd 支持通过 `LOGFWD_TARGET_CONTAINER_IMAGE` 来支持 image 字段注入(#1299)
- trace 采集器：
    - 优化 error-stack/error-message 格式问题(#1307)
    - SkyWalking 兼容性调整，支持 8.X 全序列(#1296)
- eBPF `httpflow` 增加 `pid/process_name` 字段(#1218/#1124)，优化内核版本支持(#1277)
- *datakit.yaml* 有调整，建议更新新的 yaml(#1253)
- GPU 显卡采集支持远程模式(#1312)
- 其它细节优化(#1311/#1260/#1301/#1291/#1298/#1305)

### 兼容调整 {#cl-1.5.2-brk}

- 移除 `datakit --man` 命令

---

## 1.5.1(2022/12/01) {#cl-1.5.1}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.1-new}

- 新增 Python Profiling 接入(#1146)
- Pythond 新增自定义事件上报功能(#1174)
- netstat 支持特定端口的指标采集(#1276)

### 问题修复 {#cl-1.5.1-fix}

- 修复 API write 接口 JSON 写入的时间戳精度问题(#1264)
- 修复 Windows GPU 数据采集问题(#1268)
- 其它问题修复(#1273/#1278/#1279/#1285/#1281/#1282)

### 功能优化 {#cl-1.5.1-opt}

- 优化 Redis 采集器 CPU 使用率采集，增加了新的指标字段(#1263)
- 优化 logfwd 采集器配置(#1280)
- 补全主机对象的字段采集，增加网络、磁盘等相关字段(#1252)

---

## 1.5.0(2022/11/17) {#cl-1.5.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.5.0-new}

- 新增 [SNMP 采集器](snmp.md)(#1068)
- 新增 [IPMI 采集器](ipmi.md)(#1085)

### 问题修复 {#cl-1.5.0-fix}

- 修复 Git 管理配置文件模式下启动未预期采集器问题(#1250)
- 修复 Jaeger 链路采集 TraceID 问题(#1251)
- 修复极端情况下容器采集器存在的内存泄露问题(#1256)
- 修复 Windows 下代理安装问题(#1244)
- 其它修复(#1259)

### 功能优化 {#cl-1.5.0-opt}

- 新增批量注入 [DDTrace-Java 工具](../developers/ddtrace-attach.md)(#786)
- [最新 DDTrace-Java SDK](../developers/ddtrace-guance.md) 增强了 SQL 脱敏功能(#789)
- 远程 Pipeline 优化（以下两个功能，要求 Studio 升级到 2022/11/17 以后的版本）：
    - Pipeline 支持来源映射关系配置，便于实现 Pipeline 和数据源之间的批量配置(#1211)
    - Pipeline 提供了函数分类信息，便于远程 Pipeline 编写(#1150)
- 优化 [Kafka 消息订阅](kafkamq.md)，不再局限于获取 SkyWalking 相关的数据，同时支持限速、多版本覆盖、采样以及负载均衡等设定(#1212)
- 通过提供额外配置参数（`ENV_INPUT_CONTAINER_LOGGING_SEARCH_INTERVAL`），缓解短生命周期 Pod 日志采集问题(#1255)
- 纯容器环境下，支持[通过 label 方式](container-log.md#logging-with-annotation-or-label)配置容器内日志采集(#1187)
- [SQLServer 采集器](sqlserver.md)增加更多指标集采集(#1216)
- 新增 Pipeline 函数(#1220/#1224)
    - [sample()](../developers/pipeline/pipeline-built-in-function.md#fn-sample)：采样函数
    - [b64enc()](../developers/pipeline/pipeline-built-in-function.md#fn-b64enc)：Base64 编码函数
    - [b64dec()](../developers/pipeline/pipeline-built-in-function.md#fn-b64dec)：Base64 解码函数
    - [append()](../developers/pipeline/pipeline-built-in-function.md#fn-append)：列表追加函数
    - [url_parse()](../developers/pipeline/pipeline-built-in-function.md#fn-url-parse)：HTTP URL 解析函数

- 各种文档完善(#1242/#1238/#1247)

## 1.4.20(2022/11/03) {#cl-1.4.20}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.4.20-new}

- 完善 Prometheus 生态兼容，增加 [ServiceMonitor 和 PodMonitor 采集识别](kubernetes-prometheus-operator-crd.md)(#1130)
- 增加基于 async-profiler 的 [Java Profiling 接入](profile-java-async-profiler.md)(#1240)

### 问题修复 {#cl-1.4.20-fix}

- 修复 Prom 采集器日志错乱问题(#1226)
- 修复 DDTrace trace-id 转换溢出问题，该问题可能导致 trace/span 丢失(#1235)
- 修复 ElasticSearch 采集器采集中断问题(#1236)
- 修复 Git 模式下采集器仍然会多开的问题(#1241)

### 功能优化 {#cl-1.4.20-opt}

- eBPF 采集增加 [interval 参数](ebpf.md#config)，便于调节采集的数据量(#1106)
- 所有远程采集器默认以其采集地址作为 `host` 字段的取值，避免远程采集时可能误解 `host` 字段的取值(#1120)
- DDTrace 采集到的 APM 数据，能自动提取 error 相关的字段，便于中心做更好的 APM 错误追踪(#1161)
- MySQL 采集器增加额外字段（`Com_commit/Com_rollback`）采集(#1206)
- 优化 GPU 采集器以适配更多显卡厂商(#1232)
- 其它完善(#1204/#1231/#1233)

---

## 1.4.19(2022/10/20) {#cl-1.4.19}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.4.19-new}

- DataKit 采集器配置和 Pipeline 支持[通过 etcd/Consul 等配置中心](confd.md)来同步(#1090)

### 问题修复 {#cl-1.4.19-fix}

- 修复 Windows Event 采集不到数据的问题(#1200)
- 修复 prom 调试器不工作的问题(#1192)

### 功能优化 {#cl-1.4.19-opt}

- Prometheus Remote Write 优化
    - 采集支持通过正则过滤 tag(#1197)
    - 支持通过正则过滤指标集名称(#1196)

- Pipeline 优化(#1188)
    - 优化 [grok()](../developers/pipeline/pipeline-built-in-function.md#fn-grok) 等函数，使得其可以用在 `if/else` 语句中，以判定操作是否生效
    - 增加 [match()](../developers/pipeline/pipeline-built-in-function.md#fn-match) 函数
    - 增加 [cidr()](../developers/pipeline/pipeline-built-in-function.md#fn-cidr) 函数(#733)
    <!-- - Pipeline 函数增加分类支持，便于用户在 Studio 页面上更快速定位操作函数(#1150) -->

- 进程采集器增加打开的文件列表详情字段(#1173)
- 完善外部接入类数据（T/R/L）的磁盘缓存和队列处理(#971)
- Monitor 上增加用量超支提示：在 monitor 底部，如果当前空间用量超支，会有红色文字 `Beyond Usage` 提示(#1025)
- 优化日志采集 position 功能，在容器环境下会将该文件外挂到宿主机，避免 DataKit 重启后丢失原有 position 记录(#1203)
- 优化稀疏日志场景下采集延迟问题(#1202)


### 兼容调整 {#cl-1.4.19-brk}

由于更换了日志采集的 position 存储（存储位置和存储格式都换了），更新本版本后，原有 position 将失效。新升级本版本后，升级间隙产生的日志将不被采集，请慎重。

---

## 1.4.18(2022/10/13) {#cl-1.4.18}

本次发布属于 Hotfix 发布，主要有如下更新：

- 修复 Docker 日志 16k 截断问题(#1185)
- 修复自动多行情况下日志被吞噬、导致容器重启后无日志采集的问题(#1162)
- 优化 eBPF DNS 数据采集，自动追加 Kubernetes 相关的 tag，同时预聚合部分数据，减少采集的数据量(#1186)
- 支持从 Kafka 中订阅基于 SkyWalking 的日志数据(#1155)
- 优化主机对象采集字段(#1171)
- 其它一些细节优化(#1159/#1177/#1160)

---

## 1.4.17(2022/10/8) {#cl-1.4.17}

本次发布属于迭代发布，主要有如下更新。

### 新功能 {#cl-1.4.17-new-features}

- 新增 [Promtail 采集器](promtail.md)(#644)
- 新增 [NVIDIA GPU 指标采集器](nvidia_smi.md)(#1005)
- 支持发现（需手动开启） Kubernetes 集群中带有 Prometheus Service 的服务，并对之实施 Prometheus 指标采集(#1123)
- 支持从 Kafka 中订阅基于 SkyWalking 的指标、日志、Trace 类数据，并将其分别以对应的数据类型上传到观测云(#1127)

### 问题修复 {#cl-1.4.17-fix}

- 修复 logging socket 采集器奔溃问题(#1129)
- 修复 Redis 采集问题(#1134)
- 修复 MySQL 采集器采集 PolarDB 时因报错中断采集的问题(#1147)
- 修复 git 模式下，部分默认开启的采集器不工作的问题(#1154)
- 修复 Kafka 指标集缺失问题(#1170)
- 修复拨测采集器数据上传错误问题(#1175)
- 修复 statsd 采集器日志问题(#1164)

### 优化 {#cl-1.4.17-opt}

- 替换部分可能导致漏洞的三方库(#1100)
- DataKit API 返回中新增特殊的 HTTP header 避免 CORB 问题(#1136)

- 网络拨测
    - 针对 TCP/HTTP 增加 IP 字段(#1108)
    - 调整部分字段的单位(#1113)

- 调整远程 Pipeline 调试 API(#1128)
- 增加采集器[单例运行控制](datakit-input-conf.md#input-singleton)(#1109)
- IO 模块中，日志类数据（除指标外均为日志类数据）上报均改成阻塞模式(#1121)
- 优化安装/升级过程中的终端提示(#1145)
- 其它文档以及性能优化(#1152/#1149/#1148)

### Breaking Changes {#cl-1.4.17-bc}

- Redis 采集器中，原来 latency 时序数据改为日志数据(#1144)
- 移除环境变量 `ENV_K8S_CLUSTER_NAME`，建议用全局 tag 方式来设置 Kubernetes 集群名称(#1152)

---

## 1.4.16(2022/09/15) {#cl-1.4.16}

本次发布属于迭代发布，主要有如下更新。

### 新功能 {#cl-1.4.16-new-features}

- 增加自动云同步功能，不再需要手动指定云厂商(#1074)
- 支持将 k8s label 作为 tag 同步到 pod 的指标和日志中(#1101)
- 支持将 k8s 中各类 yaml 信息采集到对应的[对象数据](container.md#objects)上(#1102)
- Trace 采集支持自动提取一些关键 meta 信息(#1092)
- 支持安装过程中指定安装源地址，以简化[离线安装](datakit-offline-install.md)流程(#1065)
- [Pipeline](../developers/pipeline/index.md) 新增功能：
    - 新增 for 循环/字典/数组支持(#1037/#1093)
    - 新增算数表达式支持(#798)
    - Pipeline 出错信息将在采集的数据上展示(#784/#1091)
    - 如果时间字段切割出错，支持自动修正时间字段(`time`)，以避免控制台页面上时间无法展示(#1091)
    - 新增 [len()](../developers/pipeline/pipeline-built-in-function.md#fn-len) 函数

### 问题修复 {#cl-1.4.16-fix}

- 修复 OOM 后 DataKit 服务不会自动启动的问题(#691)
- 修复 prom 采集器过滤指标问题(#1084)
- 修复 MySQL 采集器的指标单位、文档等问题(#1122)
- 修复 MongoDB 采集器问题(#1096/#1098)
- 修复 Trace 数据中采集了一些无效字段问题(#1083)

---

## 1.4.15(2022/09/13) {#cl-1.4.15}

本次发布属于 Hotfix 发布，大幅度提高日志类数据的采集和发送效率。

---

## 1.4.14(2022/09/09) {#cl-1.4.14}

本次发布属于 Hotfix 发布，主要有如下更新：

- 修正[磁盘采集器](disk.md)指标采集，自动忽略一些非物理磁盘；主机对象上的磁盘也做了对应的处理(#1106)
- 修正磁盘采集器在 Windows 上采集不到指标的问题(#1114)
- 修复 Git 管理配置的情况下，部分资源泄露导致的数据重复采集问题(#1107)
- 修复 [SQLServer 采集器](sqlserver.md) 复杂密码导致无法连接的问题(#1119)
- 修复 [DQL API 请求](apis.md#api-raw-query)丢失 `application/json` Content-Type 问题(#1119)
- 调整 Pipeline 有关的文档，将其移到「自定义开发」目录下：

<figure markdown>
  ![](https://static.guance.com/images/datakit/cl-1.4.14-dk-docs.gif){ width="300"}
</figure>

---

## 1.4.13(2022/09/01) {#cl-1.4.13}

### 采集器功能调整 {#cl-1.4.13-features}

- 优化 IO 模块的数据处理，提升数据吞吐效率(#1078)
- 在各类 Trace 上加上的磁盘缓存功能(#1023)
- DataKit 自身指标集增加 goroutine 使用有关的指标集（`datakit_goroutine`）(#1039)
- MySQL 采集器增加 `mysql_dbm_activity` 指标集(#1047)
- 增加 [netstat 采集器](netstat.md)(#1051)
- TDengine 增加日志采集(#1057/#1076)
- 优化磁盘采集器中的 `fstype` 过滤，默认只采集常见的文件系统（#1063/#1066）
- 日志采集器中，针对每条日志，增加字段 `message_length` 表示当前日志长度，便于通过长度来过滤日志(#1086)
- CRD 支持通过 DaemonSet 来定位 Pod 范围(#1064)
- eBPF 移除 `go-bindata` 依赖（#1062）
- 容器采集器中默认会打开 [k8s 和容器相关的指标](container.md#metrics)，这在一定程度上会消耗额外的时间线（#1095）
- Datakit 自带 DDTrace-Java SDK 已更新最新版本（*[Datakit 安装目录]/data/dd-java-agent-0.107.1.jar*）

### Bug 修复 {#cl-1.4.13-bugfix}

- 修复 DataKit 自身 CPU 使用率计算错误(#983)
- 修复 SkyWalking 中间件识别问题(#1027)
- 修复 Oracle 采集器退出问题(#1042/#1048)
- 修复 Sink DataWay 失效问题(#1056)
- 修复 HTTP /v1/write/:category 接口 JSON 写入问题(#1059)

### Breaking changes {#cl-1.4.13-br}

- GitLab 以及 Jenkins 采集器中，CI/CD 数据有关的时间字段做了调整，以统一前端页面的数据展示效果(#1089)

### 文档调整 {#cl-1.4.13-docs}

- 几乎每个章节都增加了跳转标签，便于其它文档永久性引用
- Pythond 文档已转移到开发者目录
- 采集器文档从原来「集成」移到 「DataKit」文档库(#1060)

<figure markdown>
  ![](https://static.guance.com/images/datakit/cl-1.4.13-dk-docs.gif){ width="300"}
</figure>

- DataKit 文档目录结构调整，减少了目录层级

<figure markdown>
  ![](https://static.guance.com/images/datakit/cl-1.4.13-dk-doc-dirs.gif){ width="300"}
</figure>

- 几乎每个采集器都增加了 k8s 配置入口

<figure markdown>
  ![](https://static.guance.com/images/datakit/cl-1.4.13-install-selector.gif){ width="800" }
</figure>

- 调整文档头部显示，除了操作系统标识外，对支持选举的采集器，增加选举标识

<figure markdown>
  ![](https://static.guance.com/images/datakit/cl-1.4.13-doc-header.gif){ width="800" }
</figure>

---

## 1.4.12(2022/08/26) {#cl-1.4.12}

本次发布属于 Hotfix 发布，主要有如下更新：

- 调整 Windows 下 CPU 采集的取值，以跟 Windows 进程监视器上的数值保持一致(#1002)
- 调整发送 Dataway 时的加锁行为，该行为可能导致数据发送变慢的问题
- 日志采集：
    - io 行为默认改成阻塞形式
    - 默认开启多行识别
    - 调整文件 rotate 尾部采集策略，避免可能出现的巨大数据包(#1072)
    - 调整环境变量相关的文档说明(#1071)
    - 日志的行协议中增加 `log_read_time` 字段，用来记录采集时的 UNIX 时间戳(#1077)

### Breaking Changes

- 移除 io 模块的全局阻塞（`blocking_mode`）以及按数据分类（`blocking_categories`）来设置阻塞的功能（默认都不开启）。**这个选项如果被人为打开，在新版本中，将不再生效**。

---

## 1.4.11(2022/08/17) {#cl-1.4.11}

### 新功能 {#cl-1.4.11-newfeature}

- Pipeline 中新增 [Ref-Table 功能](../developers/pipeline/pipeline-refer-table/)(#967)
- DataKit 9529 HTTP [支持绑定到 domain socket](datakit-conf.md#uds)(#925)
    - 对应的 [eBPF 采集](ebpf.md) 和 [Oracle 采集](oracle.md)，其配置方式也需做对应变更。
- RUM sourcemap 增加 Android R8 支持(#1040)
- CRD 增加日志配置支持(#1000)
    - [完整示例](kubernetes-crd.md#example)

### 优化 {#cl-1.4.11-optimize}

- 优化[容器采集器](container.md)文档
- 新增 [常见 Tag](common-tags.md) 文档(#839)
- 优化[选举的配置](election.md#config)和一些相关的命名(#1026)
- 选举类采集器在 DataKit 开启选举的情况下，仍然支持在特定的采集器上关闭选举功能(#927)
- 支持指定数据类型的 [io block 配置](datakit-daemonset-deploy.md#env-io)(#1021)
- DDTrace 采集器的采样增加 meta 信息识别(#927)
- DataKit 自身指标集增加 9529 [HTTP 请求相关指标](self.md#datakit_http)(#944)
- 优化 [Zipkin 采集](zipkin.md)的内存使用(#1013)
- DDTrace 采集器在[开启磁盘缓存](ddtrace.md#disk-cache)后，默认变成阻塞式 IO feed(#1038)
- [eBPF](ebpf.md#measurements) 增加进程名（`process_name`）字段(#1045)
- [DCA](dca.md) 新版本发布
- 日志类 HTTP 数据写入（Log Streaming/Jaeger/OpenTelemetry/Zipkin）均增加队列支持(#971)
- 日志采集增加自动多行支持(#1024)

### Bug 修复 {#cl-1.4.11-bugs}

- 修复 [MySQL 采集器](mysql.md) 连接泄露问题(#1041)
- 修复 Pipeline JSON 取值问题(#1036)
- 修复 macOS 上 ulimit 设置无效问题(#1032)
- 修复 sinker-Dataway 在 Kubernetes 中无效问题(#1031)
- 修复 [HTTP 数据写入类接口](apis.md#api-v1-write)数据校验问题(#1046)
- 修复 eBPF 采集器因内核变更后结构体偏移计算失败问题(#1049)
- 修复 DDTrace close-resource 问题(#1035)

---

## 1.4.10(2022/08/05) {#cl-1.4.10}
本次发布属于迭代发布，主要有如下更新：

- 部分数据类型发送失败后，支持缓存到磁盘，延后再发送(#945)
- 支持通过不同的 Dataway 地址，将满足条件的数据发送到不同的工作空间(#896)
- Sourcemap 增加 Android 和 iOS 支持(#886)

- 容器采集器相关更新：
    - 修复 Kubernetes 中 Node 主机操作系统信息采集错误(#950)
    - Kubernetes 中 Prom 采集不再自动追加 pod 相关信息，避免时间线暴增(#965)
    - Pod 对象中追加对应 yaml 信息(#969)

- Pipeline 相关更新：
    - 优化 Pipeline 执行步骤(#1007)
    - [`grok()`](../developers/pipeline/pipeline-built-in-function.md#fn-grok) 和 [`json()`](../developers/pipeline/pipeline-built-in-function.md#fn-json) 函数默认执行 trim-space 操作(#1001)

- DDTrace 相关更新：
    - 修复潜在的 goroutine 泄露问题(#1008)
    - 支持配置磁盘缓存来缓解内存占用问题(#1014)

- 其它 Bug 修复：
    - 优化行协议构造(#1016)
    - 日志采集中，移除定期清理尾部数据功能，以缓解可能导致的日志截断问题(#1012)

### Breaking Changes {cl-1.4.10-break-changes}

由于 RUM 中新增了 Sourcemap 支持，有了更多的配置选项，故 RUM 采集器不再默认开启，需[手动开启](rum.md#config)。

---

## 1.4.9(2022/07/26) {#cl-1.4.9}

本次发布属于 Hotfix 发布，主要有如下更新：

- eBPF `httpflow` 增加 Linux 4.5 及以上内核版本支持(#985)
- 修复 external 类采集器选举模式下的问题(#976/#946)
- 修复容器采集器导致的奔溃问题(#956/#979/#980)
- 修复 Redis `slowlog` 采集问题(#986)

---

## 1.4.8(2022/07/21) {#cl-1.4.8}

本次发布属于迭代发布，主要有如下更新：

- prom 采集器的内置超时改为 3 秒(#958)

- 日志相关问题修复：
    - 添加日志采集的 `log_read_offset` 字段(#905)
    - 修复日志文件在 rotate 后没有正确读取尾部遗留内容的问题(#936)

- 容器采集相关问题修复：
    - 修复对环境变量 `NODE_NAME` 的不兼容问题(#957)
    - k8s 自动发现的 prom 采集器改为串行式的分散采集，每个 k8s node 只采集自己机器上的 prom 指标(#811/#957)
    - 添加日志 source 和多行的的[映射配置](container.md#env-config)(#937)
    - 修复容器日志替换 source 后还使用之前的 multiline 和 Pipeline 的 bug(#934/#923)
    - 修正容器日志，设置文件活跃时长是 12 小时(#930)
    - 优化 docker 容器日志的 image 字段(#929)
    - 优化 k8s pod 对象的 host 字段(#924)
    - 修复容器指标和对象采集没有添加 host tag 的问题(#962)

- eBPF 相关：
    - 修复 Uprobe event name 命名冲突问题
    - 增加更多[环境变量配置](ebpf.md#config)，便于 k8s 环境的部署

- 优化 APM 数据接收接口的数据处理，缓解卡死客户端以及内存占用问题(#902)

- SQLServer 采集器修复：
    - 恢复 TLS1.0 支持(#909)
    - 支持通过 instance 采集过滤，以减少时间线消耗(#931)

- Pipeline 函数 `adjust_timezone()` 有所调整(#917)
- [IO 模块优化](datakit-conf.md#io-tuning)，提高整体数据处理能力，保持内存消耗的相对可控(#912)
- Monitor 更新：
    - 修复繁忙时 Monitor 可能导致的长时间卡顿(#933)
    - 优化 Monitor 展示，增加 IO 模块的信息展示，便于用于调整 IO 模块参数
- 修复 Redis 奔溃问题(#935)
- 去掉部分繁杂的冗余日志(#939)
- 修复选举类采集器在非选举模式下不追加主机 tag 的问题(#968)

---

## 1.4.7(2022/07/11) {#cl-1.4.7}

本次发布属于 Hotfix 发布，主要修复如下问题

- 选举有关
    - 修复 `election_namespace` 设置错误的问题(#915)
    - `enable_election_namespace` 这个 tag 的设置默认关闭，需[手动开启](election.md#config)
    - *datakit.conf* 中 `namespace` 字段将被弃用（仍然可用），改名为 `election_namespace`

- 修复采集器堵塞问题(#916)
    - DataKit 移除调用中心的心跳接口
    - DataKit 移除调用中心的 Dataway 列表接口

- [容器采集器](container.md)支持通过额外的配置（`ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP`）来修改 sidecar 容器的日志来源（`source`） 字段(#903)
- 修复黑名单在 Monitor 上的展示问题(#904)

---

## 1.4.6(2022/07/07) {#cl-1.4.6}

- 调整[全局 tag](datakit-conf.md#set-global-tag) 的行为，避免选举类采集的 tag 分裂(#870)
- [SQLServer 采集器](sqlserver.md)增加选举支持(#882)
- [行协议过滤器](datakit-filter.md)支持所有数据类型(#855)
- 9529 HTTP 服务增加[超时机制](datakit-conf.md#http-other-settings)(#900)
- MySQL
    - [dbm 指标集名字](mysql.md#logging)调整(#898)
    - `service` 字段冲突问题(#895)
- [容器对象](container.md#docker_containers)增加字段 `container_runtime_name` 以区分不同层次的容器名(#891)
- Redis 调整 [`slowlog` 采集](redis.md#redis_slowlog)，将其数据改为日志存储(#885)
- 优化 [TDEngine 采集](tdengine.md)(#877)
- 完善 Containerd 日志采集，支持默认格式的日志自动解析(#869)
- [Pipeline](../developers/pipeline/index.md) 增加 [Profiling 类数据](profile.md)支持(#866)
- 容器/Pod 日志采集支持在 Label/Annotation 上[额外追加 tag](container-log.md#logging-with-annotation-or-label)(#861)
- 修复 [Jenkins CI](jenkins.md#jenkins_pipeline) 数据采集的时间精度问题(#860)
- 修复 Tracing resource-type 值不统一的问题(#856)
- eBPF 增加 [HTTPS 支持](ebpf.md#https)(#782)
- 修复日志采集器可能的奔溃问题(#893)
- 修复 prom 采集器泄露问题(#880)
- 支持通过[环境变量配置 io 磁盘缓存](datakit-conf.md#using-cache)(#906)
- 增加 [Kubernetes CRD](kubernetes-crd.md) 支持(#726)
- 其它 bug 修复(#901/#899)

---

## 1.4.5(2022/06/29) {#cl-1.4.5}

本次发布属于 Hotfix 发布，主要修复日志采集器因同名文件快速删除并新建而导致采集中断问题(#883)

如果大家有计划任务在定期打包日志，可能会触发这个 Bug，**建议升级**。

---

## 1.4.4(2022/06/27) {#cl-1.4.4}

本次发布属于 Hotfix 发布，主要更新如下内容：

- 修复日志采集因 pos 处理不当导的不采集问题，该问题自 1.4.2 引入，**建议升级** (#873)
- 修复 TDEngine 可能导致的 crash 问题
- 优化 eBPF 数据发送流程，避免积攒过多消耗太多内存导致 OOM(#871)
- 修复采集器文档错误

---

## 1.4.3(2022/06/22) {#cl-1.4.3}

本次发布属于迭代发布，主要更新如下内容：

- Git 同步配置支持无密码模式(#845)
- Prom 采集器
    - 支持日志模式采集(#844)
    - 支持配置 HTTP 请求头(#832)
- 支持超 16KB 长度的容器日志采集(#836)
- 支持 TDEngine 采集器(810)
- Pipeline
    - 支持 XML 解析(#804)
    - 远程调试支持多类数据类型(#833)
    - 支持 Pipeline 通过 `use()` 函数调用外部 Pipeline 脚本(#824)
- 新增 IP 库（MaxMindIP）支持(#799)
- 新增 DDTrace Profiling 集成(#656)
- containerd 日志采集支持通过 image 和 K8s Annotation 配置过滤规则(#849)
- 文档库整体切换到 MkDocs(#745)
- 其它杂项(#822)

### Bug 修复 {#cl-1.4.3-bugfix}

- 修复 socket 采集器奔溃问题(#858)
- 修复部分采集器配置中空 tags 配置导致的奔溃问题(#852)
- 修复 IPDB 更新命令问题(#854)
- Kubernetes Pod 日志和对象上增加 `pod_ip` 字段(848)
- DDTrace 采集器恢复识别 trace SDK 上的采样设定(#834)
- 修复 DaemonSet 模式下，外部采集器（eBPF/Oracle）上的 `host` 字段可能跟 DataKit 自身不一致的问题(#843)
- 修复 stdout 多行日志采集问题(#859)
---

## 1.4.2(2022/06/16) {#cl-1.4.2}

本次发布属于迭代发布，主要更新如下内容：

- 日志采集支持记录采集位置，避免因为 DataKit 重启等情况导致的数据漏采(#812)
- 调整 Pipeline 在处理不同类数据时的设定(#806)
- 支持接收 SkyWalking 指标数据(#780)
- 优化日志黑名单调试功能：
    - 在 Monitor 中会展示被过滤掉的点数(#827)
    - 在 *datakit/data* 目录下会增加一个 *.pull* 文件，用来记录拉取到的过滤器
- Monitor 中增加 DataKit 打开文件数显示(#828)
- DataKit 编译器升级到 Golang 1.18.3(#674)

### Bug 修复 {#1.4.2-bugfix}

- 修复 `ENV_K8S_NODE_NAME` 未全局生效的问题(#840)
- 修复日志采集器中文件描述符泄露问题，**强烈推荐升级**(#838)
- 修复 Pipeline `group_in` 问题(#826)
- 修复 ElasticSearch 采集器配置的 `http_timeout` 解析问题(#821)
- 修复 DCA API 问题(#747)
- 修复 `dev_null` DataWay 设置无效问题(#842)

----

## 1.4.1(2022/06/07) {#cl-1.4.1}

本次发布属于迭代发布，主要更新如下内容：

- 修复 toml 配置文件兼容性问题(#195)
- 增加 [TCP/UDP 端口检测](socket)采集器(#743)
- DataKit 跟 DataWay 之间增加 DNS 检测，支持 DataWay DNS 动态切换(#758)
- [eBPF](ebpf) L4/L7 流量数据增加 k8s deployment name 字段(#793)
- 优化 [OpenTelemetry](opentelemetry) 指标数据(#794)
- [ElasticSearch](elasticsearch) 增加 AWS OpenSearch 支持(#797)
- [行协议限制](apis#2fc2526a)中，字符串长度限制放宽到 32MB(#801)
- [prom](prom) 采集器增加额外配置，支持忽略指定的 tag=value 的匹配，以减少不必要的时序时间线(#808)
- Sink 增加 Jaeger 支持(#813)
- Kubernetes 相关的[指标](container#7e687515)采集，默认全部关闭，以避免时间线暴增问题(#807)
- [DataKit Monitor](monitor)增加动态发现（比如 prom）的采集器列表刷新(#711)

### Bug 修复 {#cl-1.4.1-bugfix}
- 修复默认 Pipeline 加载问题(#796)
- 修复 Pipeline 中关于日志 status 的处理(#800)
- 修复 [Filebeat](beats_output) 奔溃问题(#805)
- 修复 [Log Streaming](logstreaming) 导致的脏数据问题(#802)

----

## 1.4.0(2022/05/26) {#cl-1.4.0}

本次发布属于迭代发布， 次版本号进入 1.4 序列。主要更新如下内容：

- Pipeline 做了很大调整(#761)
    - 所有数据类型，均可通过配置 Pipeline 来额外处理数据(#761/#739)
    - [grok()](pipeline#965ead3c) 支持直接将字段提取为指定类型，无需再额外通过 `cast()` 函数进行类型转换(#760)
    - Pipeline 增加[多行字符串支持](pipeline#3ab24547)，对于很长的字符串（比如 grok 中的正则切割），可以通过将它们写成多行，提升了可读性(#744)
    - 每个 Pipeline 的运行情况，通过 `datakit monitor -V` 可直接查看(#701)
- 增加 Kubernetes [Pod 对象](container#23ae0855-1) CPU/内存指标(#770)
- Helm 增加更多 Kubernetes 版本安装适配(#783)
- 优化 [OpenTelemetry](opentelemetry)，HTTP 协议增加 JSON 支持(#781)
- DataKit 在自动纠错行协议时，对纠错行为增加了日志记录，便于调试数据问题(#777)
- 移除时序类数据中的所有字符串指标(#773)
- 在 DaemonSet 安装中，如果配置了[选举](election)的命名空间，对参与选举的采集器，其数据上均会新增特定的 tag（`election_namespace`）(#743)
- CI 可观测，增加 [Jenkins](jenkins) 支持(#729)

### Bug 修复 {#cl-1.4.0-bugfix}

- 修复 monitor 中 DataWay 统计错误(#785)
- 修复日志采集器相关 bug(#783)
    - 有一定概率，日志采集会导致脏数据串流的情况
    - 在文件日志采集的场景下（磁盘文件/容器日志/logfwd），修复被采集日志因为 truncate/rename/remove 等因素导致的采集不稳定问题（丢失数据）
- 其它 Bug 修复(#790)

----

## 1.2.20(2022/05/22) {#cl-1.2.20}

本次发布属于 hotfix 发布，主要修复如下问题：

- 日志采集功能优化(#775)
    - 去掉 32KB 限制（保留 32MB 最大限制）(#776)
    - 修复可能丢失头部日志的问题
    - 对于新创建的日志，默认从头开始采集（主要是容器类日志，磁盘文件类日志目前无法判定是否是新创建的日志）
    - 优化 Docker 日志处理，不再依赖 Docker 日志 API

- 修复 Pipeline 中的 [decode](pipeline#837c4e09) 函数问题(#769)
- OpenTelemetry gRPC 方式支持 gzip(#774)
- 修复 [Filebeat](beats_output) 采集器不能设置 service 的问题(#767)

### Breaking changes {#cl.1.2.20-bc}

对于 Docker 类容器日志的采集，需要将宿主机（Node）的 */varl/lib* 路径挂载到 DataKit 里面（因为 Docker 日志默认落在宿主机的 */var/lib/* 下面），在 *datakit.yaml* 中，`volumeMounts` 和 `volumes` 中新增如下配置：

```yaml
volumeMounts:
- mountPath: /var/lib
  name: lib

# 省略其它部分 ...

volumes:
- hostPath:
    path: /var/lib
  name: lib
```

----

## 1.2.19(2022/05/12) {#cl-1.2.19}

本次发布属于迭代发布，主要更新如下内容：

- eBPF 增加 arm64 支持(#662)
- 行协议构造支持自动纠错(#710)
- DataKit 主配置增加示例配置(#715)
- [Prometheus Remote Write](prom_remote_write) 支持 tag 重命名(#731)
- 修复 DCA 客户端获取工作空间不全的问题(#747)
- 合并社区版 DataKit 已有的功能，主要包含 Sinker 功能以及 [Filebeat](beats_output) 采集器(#754)
- 调整容器日志采集，DataKit 直接支持 containerd 下容器 stdout/stderr 日志采集(#756)
- 修复 ElasticSearch 采集器超时问题(#762)
- 修复安装程序检查过于严格的问题(#763)
- 调整 DaemonSet 模式下主机名获取策略(#648)
- Trace 采集器支持通过服务名（`service`）通配来过滤资源（`resource`）(#759)
- 其它一些细节问题修复

----

## 1.2.18(2022/05/06) {#cl-1.2.18}

本次发布属于 hotfix 发布，主要修复如下问题：

- [进程采集器](host_processes.md)的过滤功能仅作用于指标采集，对象采集不受影响(#740)
- 缓解 DataKit 发送 DataWay 超时问题(#741)
- [GitLab 采集器](gitlab.md) 稍作调整(#742)
- 修复日志采集截断的问题(#749)
- 修复各种 trace 采集器 reload 后部分配置不生效的问题(#750)

----

## 1.2.17(2022/04/27) {#cl-1.2.17}

本次发布属于迭代发布，主要涉及如下几个方面：

- [容器采集器](container#7e687515)增加更多指标（`kube_` 开头）采集(#668)
- DDTrace 和 OpenTelemetry 采集器支持通过 HTTP Status Code（`omit_err_status`）来过滤部分错误的 trace
- 修复几个 Trace 采集器（DDtrace/OpenTelemetry/Zipkin/SkyWalking/Jaeger）在 git 模式下配置 reload 不生效的问题(#725)
- 修复 GitLab 采集器不能 tag 导致的奔溃问题(#730)
- 修复 Kubernetes 下 eBPF 采集器对 Pod 标签（tag）不更新的问题(#736)
- [prom 采集器](prom.md) 支持 [Tag 重命名](prom#e42139cb)(#719)
- 完善部分文档描述

----

## 1.2.16(2022/04/24) {#cl-1.2.16}

本次发布属于 hotfix 修复，主要涉及如下几个方面(#728)：

- 修复安装程序可能的报错导致无法继续安装/升级，目前选择容忍部分情况的服务操作错误
- 修复 Windows 安装脚本的拼写错误，该错误导致 32 位安装程序下载失败
- 调整 Monitor 关于选举情况的展示
- 开启选举的情况下，修复 MongoDB 死循环导致无法采集的问题

----

## 1.2.15(2022/04/21) {#cl-1.2.15}

本次发布属于迭代发布，含大量问题修复：

- Pipeline 模块修复 Grok 中[动态多行 pattern](datakit-pl-how-to#88b72768) 问题(#720)
- 移除掉一些不必要的 DataKit 事件日志上报(#704)
- 修复升级程序可能导致的升级失败问题(#699)
- DaemonSet 增加[开启 pprof 环境变量](datakit-daemonset-deploy#cc08ec8c)配置(#697)
- DaemonSet 中所有[默认开启采集器](datakit-input-conf#764ffbc2)各个配置均支持通过环境变量配置(#693)
- Tracing 采集器初步支持 Pipeline 数据处理(#675)
    - [DDtrace 配置示例](ddtrace#69995abe)
- 拨测采集器增加失败任务退出机制(#54)
- 优化 [Helm 安装](datakit-daemonset-deploy#e4d3facf)(#695)
- 日志新增 `unknown` 等级（status），对于未指定等级的日志均为 `unknown`(#685)
- 容器采集器大量修复
    - 修复 cluster 字段命名问题(#542)
    - 对象 `kubernetes_clusters` 这个指标集改名为 `kubernetes_cluster_roles`
    - 原 `kubernetes.cluster` 这个 count 改名为 `kubernetes.cluster_role`
    - 修复 namespace 字段命名问题(#724)
    - 容器日志采集中，如果 Pod Annotation 不指定日志 `source`，那么 DataKit 将按照[此优先级来推导日志来源](container#6de978c3)(#708/#723)
    - 对象上报不再受 32KB 字长限制（因 Annotation 内容超 32KB）(#709)
    - 所有 Kubernetes 对象均删除 `annotation` 这一 field
    - 修复 prom 采集器不会随 Pod 退出而停止的问题(#716)
- 其它问题修复(#721)

---

## 1.2.14(2022/04/12) {#cl-1.2.14}

本次发布属于 hotfix 发布，同时包含部分小的修改和调整：

- 修复日志采集器的 monitor 展示问题以及部分出错日志等级调整(#706)
- 修复拨测采集器内存泄露问题(#702)
- 修复主机进程采集器奔溃问题(#700)
- 日志采集器采集选项 `ignore_dead_log = '10m'` 默认开启(#698)
- 优化 Git 管理的配置同步逻辑(#696)
- eBPF 修复 `netflow` 中错误的 IP 协议字段(#694)
- 丰富 GitLab 采集器字段

---

## 1.2.13(2022/04/08) {#cl-1.2.13}

本次发布属于迭代发布，更新内容如下：

- 增加宿主机运行时的[内存限制](datakit-conf#4e7ff8f3)(#641)
    - 安装阶段即支持[内存限制配置](datakit-install#03be369a)
- CPU 采集器增加 [load5s 指标](cpu#13e60209)(#606)
- 完善 *datakit.yaml* 示例(#678)
- 支持主机安装时通过 [cgroup 限制内存](datakit-conf#4e7ff8f3)使用(#641)
- 完善日志黑名单功能，新增 `contain/notcontain` 判定规则(#665)
    - 支持在 *datakit.conf* 中[配置日志/对象/Tracing/时序指标这几类黑名单](datakit-filter#045b45e3)
    - 注意：升级该版本，要求 DataWay 升级到 1.2.1+
- 进一步完善 [containerd 下的容器采集](container)(#402)
- 调整 monitor 布局，增加黑名单过滤情况展示(#634)
- DaemonSet 安装增加 [Helm 支持](datakit-daemonset-deploy)(#653)
    - 新增 [DaemonSet 安装最佳实践](datakit-daemonset-bp)(#673)
- 完善 [GitLab 采集器](gitlab)(#661)
- 增加 [ulimit 配置项](datakit-conf#8f9f4364)用于配置文件打开数限制(#667)
- Pipeline [脱敏函数](pipeline#52a4c41c)有更新，新增 [SQL 脱敏函数](pipeline#711d6fe4)(#670)
- 进程对象和时序指标[新增 `cpu_usage_top` 字段](host_processes#a30fc2c1-1)，以跟 `top` 命令的结果对应(#621)
- eBPF 增加 [HTTP 协议采集](ebpf#905896c5)(#563)
- 主机安装时，eBPF 采集器默认不再会安装（减少二进制分发体积），如需安装[需用特定的安装指令](ebpf#852abae7)(#605)
    - DaemonSet 安装不受影响
- 其它 Bug 修复（#688/#681/#679/#680）

---

## 1.2.12(2022/03/24) {#cl-1.2.12}

本次发布属于迭代发布，更新内容如下：

1. 增加 [DataKit 命令行补全](datakit-tools-how-to#9e4e5d5f)功能(#76)
1. 允许 DataKit [升级到非稳定版](datakit-update#42d8b0e4)(#639)
1. 调整 Remote Pipeline 的在 DataKit 本地的存储，避免不同文件系统差异导致的文件名大小写问题(#649)
1. (Alpha)初步支持 [Kubernetes/Containerd 架构的数据采集](container#b3edf30c)(#402)
1. 修复 Redis 采集器的不合理报错(#671)
1. OpenTelemetry 采集器字段微调(#672)
1. 修复 [DataKit 自身采集器](self) CPU 计算错误(#664)
1. 修复 RUM 采集器因 IPDB 缺失导致的 IP 关联字段缺失问题(#652)
1. Pipeline 支持调试数据上传至 OSS(#650)
1. DataKit HTTP API 上均会[带上 DataKit 版本号信息](apis#be896a47)
1. [网络拨测](dialtesting)增加 TCP/UDP/ICMP/Websocket 几种协议支持(#519)
1. 修复[主机对象采集器](hostobject)字段超长问题(#669)
1. Pipeline
    - 新增 [decode()](pipeline#837c4e09) 函数(#559)，这样可以避免在日志采集器中去配置编码，可以在 Pipeline 中实现编码转换
    - 修复 Pipeline 导入 pattern 文件可能失败的问题(#666)
    - [add_pattern()](pipeline#89bd3d4e) 增加作用域管理

---

## 1.2.11(2022/03/17) {#cl-1.2.11}

本次发布属于 hotfix 发布，同时包含部分小的修改和调整：

- 修复 Tracing 采集器资源过滤（`close_resource`）的算法问题，将过滤机制下放到 Entry Span 级别，而非之前的 Root Span
- 修复[日志采集器](logging)文件句柄泄露问题(#658)，同时新增配置（`ignore_dead_log`），以忽略不再更新（删除）的文件
- 新增[Datakit 自身指标文档](self)
- DaemonSet 安装时
    - [支持安装 IPDB](datakit-tools-how-to#11f01544)(#659)
    - 支持[设定 HTTP 限流（ENV_REQUEST_RATE_LIMIT）](datakit-daemonset-deploy#00c8a780)(#654)

---

## 1.2.10(2022/03/11) {#cl-1.2.10}

修复 Tracing 相关采集器可能的奔溃问题

---

## 1.2.9(2022/03/10) {#cl-1.2.9}

本次发布属于迭代发布，更新内容如下：

- DataKit 9529 HTTP 服务添加 [API 限流措施](datakit-conf#39e48d64)(#637)
- 统一各种 Tracing 数据的[采样率设置](datakit-tracing#64df2902)(#631)
- 发布 [DataKit 日志采集综述](datakit-logging)
- 支持 [OpenTelemetry 数据接入](opentelemetry)(#609)
- 支持[禁用 Pod 内部部分镜像的日志](container#2a6149d7)(#586)
- 进程对象采集[增加监听端口列表](host_processes#a30fc2c1-1)(#562)
- eBPF 采集器[支持 Kubernetes 字段关联](ebpf#35c97cc9)(#511)

### Breaking Changes {#cl-1.2.9-bc}

- 本次对 Tracing 数据采集做了较大的调整，涉及几个方面的不兼容：

    - [DDtrace](ddtrace) 原有 conf 中配置的 `ignore_resources` 字段需改成 `close_resource`，且字段类型由原来的数组（`[...]`）形式改成了字典数组（`map[string][...]`）形式（可参照 [conf.sample](ddtrace#69995abe) 来配置）
    - DDTrace 原数据中采集的 [tag `type` 字段改成 `source_type`](ddtrace#01b88adb)

---

## 1.2.8(2022/03/04) {#cl-1.2.8}

本次发布属于 hotfix 修复，内容如下：

- DaemonSet 模式部署时，*datakit.yaml* 添加[污点容忍度配置](datakit-daemonset-deploy#e29e678e)(#635)
- 修复 Remote Pipeline 拉取更新时的 bug(#630)
- 修复 DataKit IO 模块卡死导致的内存泄露(#646)
- 在 Pipeline 中允许修改 `service` 字段(#645)
- 修复 `pod_namespace` 拼写错误
- 修复 logfwd 的一些问题(#640)
- 修复日志采集器在容器环境下采集时多行粘滞问题(#633)

---

## 1.2.7(2022/02/22) {#cl-1.2.7}

本次发布属于迭代发布，内容如下：

- Pipeline
    - Grok 中增加[动态多行 pattern](datakit-pl-how-to#88b72768)，便于处理动态多行切割(#615)
    - 支持中心下发 Pipeline(#524)，这样一来，Pipeline 将有[三种存放路径](pipeline#6ee232b2)
    - DataKit HTTP API 增加 Pipeline 调试接口 [`/v1/pipeline/debug`](apis#539fb60e)

<!--
- APM 功能调整(#610)
    - 重构现有常见的 Tracing 数据接入
    - 增加 APM 指标计算
    - 新增 [OTEL(OpenTelemetry)数据接入]()

!!! Delay
-->

- 为减少默认安装包体积，默认安装不再带 IP 地理信息库。RUM 等采集器中，可额外[安装对应的 IP 库](datakit-tools-how-to#ab5cd5ad)
    - 如需安装时就带上 IP 地理信息库，可通过[额外支持的命令行环境变量](datakit-install#f9858758)来实现
- 容器采集器增加 [logfwd 日志接入](logfwd)(#600)
- 为进一步规范数据上传，行协议增加了更多严格的[限制](apis#2fc2526a)(#592)
- [日志采集器](logging)中，放开日志长度限制（`maximum_length`）(#623)
- 优化日志采集过程中的 Monitor 显示(#587)
- 优化安装程序的命令行参数检查(#573)
- 重新调整 DataKit 命令行参数，大部分主要的命令已经支持。另外，**老的命令行参数在一定时间内依然生效**(#499)
    - 可通过 `datakit help` 查看新的命令行参数风格
- 重新实现 [ DataKit Monitor](datakit-monitor)

### 其它 Bug 修复

- 修复 Windows 下安装脚本问题(#617)
- 调整 *datakit.yaml* 中的 ConfigMap 设定(#603)
- 修复 Git 模式下 Reload 导致部分 HTTP 服务异常的问题(#596)
- 修复安装包 isp 文件丢失问题(#584/#585/#560)
- 修复 Pod annotation 中日志多行匹配不生效的问题(#620)
- 修复 TCP/UDP 日志采集器 _service_ tag 不生效的问题(#610)
- 修复 Oracle 采集器采集不到数据的问题(#625)

### Breaking Changes

- 老版本的 DataKit 如果开启了 RUM 功能，升级上来后，需[重新安装 IP 库](datakit-tools-how-to#ab5cd5ad)，老版本的 IP 库将无法使用。

---

## 1.2.6(2022/01/20)

本次发布属于迭代发布，内容如下：

- 增强 [DataKit API 安全访问控制](rum#b896ec48)，老版本的 DataKit 如果部署了 RUM 功能，建议升级(#578)
- 增加更多 DataKit 内部事件日志上报(#527)
- 查看 [DataKit 运行状态](datakit-tools-how-to#44462aae)不再会超时(#555)

- [容器采集器](container)一些细节问题修复

    - 修复在 Kubernetes 环境主机部署时崩溃问题(#576)
    - 提升 Annotation 采集配置优先级(#553)
    - 容器日志支持多行处理(#552)
    - Kubernetes Node 对象增加 _role_ 字段(#549)
    - [通过 Annotation 标注](kubernetes-prom)的 [Prom 采集器](prom) 会自动增加相关属性（_pod_name/node_name/namespace_）(#522/#443)
    - 其它 Bug 修复

- Pipeline 问题修复

    - 修复日志处理中可能导致的时间乱序问题(#547)
    - 支持 _if/else_ 语句[复杂逻辑关系判断支持](pipeline#1ea7e5aa)

- 修复日志采集器 Windows 中路径问题(#423)
- 完善 DataKit 服务管理，优化交互提示(#535)
- 优化现有 DataKit 文档导出的指标单位(#531)
- 提升工程质量(#515/#528)

---

## 1.2.5(2022/01/19)

- 修复[Log Stream 采集器](logstreaming) Pipeline 配置问题(#569)
- 修复[容器采集器](container)日志错乱的问题(#571)
- 修复 Pipeline 模块更新逻辑的 bug(#572)

---

## 1.2.4(2022/01/12)

- 修复日志 API 接口指标丢失问题(#551)
- 修复 [eBPF](ebpf) 网络流量统计部分丢失问题(#556)
- 修复采集器配置文件中 `$` 字符通配问题(#550)
- Pipeline _if_ 语句支持空值比较，便于 Grok 切割判断(#538)

---

## 1.2.3(2022/01/10)

- 修复 *datakit.yaml* 格式错误问题(#544)
- 修复 [MySQL 采集器](mysql)选举问题(#543)
- 修复因 Pipeline 不配置导致日志不采集的问题(#546)

---

## 1.2.2(2022/01/07)

- [容器采集器](container)更新：
    - 修复日志处理效率问题(#540)
    - 优化配置文件黑白名单配置(#536)
- Pipeline 模块增加 `datakit -M` 指标暴露(#541)
- [ClickHouse](clickhousev1) 采集器 config-sample 问题修复(#539)
- [Kafka](kafka) 指标采集优化(#534)

---

## 1.2.1(2022/01/05)

- 修复采集器 Pipeline 使用问题(#529)
- 完善[容器采集器](container)数据问题(#532/#530)
    - 修复 short-image 采集问题
    - 完善 k8s 环境下 Deployment/Replica-Set 关联

---

## 1.2.0(2021/12/30) {#cl-1.2.0}

### 采集器更新

- 重构 Kubernetes 云原生采集器，将其整合进[容器采集器](container.md)。原有 Kubernetes 采集器不再生效(#492)
- [Redis 采集器](redis.md)
    - 支持配置 [Redis 用户名](redis.md)(#260)
    - 增加 Latency 以及 Cluster 指标集(#396)
- [Kafka 采集器](kafka.md)增强，支持 topic/broker/consumer/connection 等维度的指标(#397)
- 新增 [ClickHouse](clickhousev1.md) 以及 [Flink](flinkv1.md) 采集器(#458/#459)
- [主机对象采集器](hostobject)
    - 支持从 [`ENV_CLOUD_PROVIDER`](hostobject#224e2ccd) 读取云同步配置(#501)
    - 优化磁盘采集，默认不会再采集无效磁盘（比如总大小为 0 的一些磁盘）(#505)
- [日志采集器](logging) 支持接收 TCP/UDP 日志流(#503)
- [Prom 采集器](prom) 支持多 URL 采集(#506)
- 新增 [eBPF](ebpf) 采集器，它集成了 L4-network/DNS/Bash 等 eBPF 数据采集(507)
- [ElasticSearch 采集器](elasticsearch) 增加 [Open Distro](https://opendistro.github.io/for-elasticsearch/){:target="_blank"} 分支的 ElasticSearch 支持(#510)

### Bug 修复

- 修复 [Statsd](statsd)/[RabbitMQ](rabbitmq) 指标问题(#497)
- 修复 [Windows Event](windows_event) 采集数据问题(#521)

### 其它

- [Pipeline](pipeline)
    - 增强 Pipeline 并行处理能力
    - 增加 [`set_tag()`](pipeline#6e8c5285) 函数(#444)
    - 增加 [`drop()`](pipeline#fb024a10) 函数(#498)
- Git 模式
    - 在 DaemonSet 模式下的 Git，支持识别 `ENV_DEFAULT_ENABLED_INPUTS` 并将其生效，非 DaemonSet 模式下，会自动开启 *datakit.conf* 中默认开启的采集器(#501)
    - 调整 Git 模式下文件夹[存放策略]()(#509)
- 推行新的版本号机制(#484)
    - 新的版本号形式为 1.2.3，此处 `1` 为 major 版本号，`2` 为 minor 版本号，`3` 为 patch 版本号
    - 以 minor 版本号的奇偶性来判定是稳定版（偶数）还是非稳定版（奇数）
    - 同一个 minor 版本号上，会有多个不同的 patch 版本号，主要用于问题修复以及功能调整
    - 新功能预计会发布在非稳定版上，待新功能稳定后，会发布新的稳定版本。如 1.3.x 新功能稳定后，会发布 1.4.0 稳定版，以合并 1.3.x 上的新功能
    - 非稳定版不支持直接升级，比如，不能升级到 1.3.x 这样的版本，只能直接安装非稳定版

### Breaking Changes {#cl-1.2.0-break-changes}

**老版本的 DataKit 通过 `datakit --version` 已经无法推送新升级命令**，直接使用如下命令：

- Linux/Mac:

```shell
{{ InstallCmd 0 (.WithPlatform "unix") }}
```

- Windows

```powershell
{{ InstallCmd 0 (.WithPlatform "windows") }}
```

---

## 1.1.9-rc7.1(2021/12/22)

- 修复 MySQL 采集器因局部采集失败导致的数据问题。

---

## 1.1.9-rc7(2021/12/16)

- Pipeline 总体进行了较大的重构(#339)：

    - 添加 `if/elif/else` [语法](pipeline#1ea7e5aa)
    - 暂时移除 `expr()/json_all()` 函数
    - 优化时区处理，增加 `adjust_timezone()` 函数
    - 各个 Pipeline 函数做了整体测试加强

- DataKit DaemonSet：

    - Git 配置 DaemonSet [ENV 注入](datakit-daemonset-deploy#00c8a780)(#470)
    - 默认开启采集器移除容器采集器，以避免一些重复的采集问题(#473)

- 其它：
    - DataKit 支持自身事件上报（以日志形式）(#463)
    - [ElasticSearch](elasticsearch) 采集器指标集下增加 `indices_lifecycle_error_count` 指标（注意： 采集该指标，需在 ES [增加 `ilm` 角色](elasticsearch#852abae7)）
    - DataKit 安装完成后自动增加 [cgroup 限制](datakit-conf#4e7ff8f3)
    - 部分跟中心对接的接口升级到了 v2 版本，故对接**非 SAAS 节点**的 Datakit，如果升级到当前版本，其对应的 DataWay 以及 Kodo 也需要升级，否则部分接口会报告 404 错误

### Breaking Changes

处理 JSON 数据时，如果最顶层是数组，需要使用下标方式进行选择，例如 JSON

```
[
	{"abc": 123},
	{"def": true}
]
```

经过 Pipeline 处理完之后，如果取第一个元素的 `abc` 资源，之前版本做法是：

```
[0].abc
```

当前版本需改为：

```
# 在前面加一个 . 字符
.[0].abc
```

---

## 1.1.9-rc6.1(2021/12/10)

- 修复 ElasticSearch 以及 Kafka 采集报错问题(#486)

---

## 1.1.9-rc6(2021/11/30)

- 针对 Pipeline 做了一点紧急修复：
    - 移除 `json_all()` 函数，这个函数对于异常的 JSON 有严重的数据问题，故选择禁用之(#457)
    - 修正 `default_time()` 函数时区设置问题(#434)
- 解决 [`Prom`](prom) 采集器在 Kubernetes 环境下 HTTPS 访问问题(#447)
- DataKit DaemonSet 安装的 [yaml 文件](https://static.guance.com/datakit/datakit.yaml){:target="_blank"} 公网可直接下载

---

## 1.1.9-rc5.1(2021/11/26)

- 修复 DDTrace 采集器因脏数据挂掉的问题

---

## 1.1.9-rc5(2021/11/23)

- 增加 [Pythond(alpha)](pythond) ，便于用 Python3 编写自定义采集器(#367)
<!-- - 支持 source map 文件处理，便于 RUM 采集器收集 JavaScript 调用栈信息(#267) -->
- [SkyWalking V3](skywalking) 已支持到 8.5.0/8.6.0/8.7.0 三个版本(#385)
- DataKit 初步支持[磁盘数据缓存(alpha)](datakit-conf#caa0869c)(#420)
- DataKit 支持选举状态上报(#427)
- DataKit 支持 Scheck 状态上报(#428)
- 调整 DataKit 使用入门文档，新的分类更便于找到具体文档

---

## 1.1.9-rc4.3(2021/11/19)

- 修复容器日志采集器因 Pipeline 配置失当无法启动的问题

---

## 1.1.9-rc4.2(2021/11/18)

- 紧急修复(#446)
    - 修复 Kubernetes 模式下 stdout 日志输出 level 异常
    - 修复选举模式下，未选举上的 MySQL 采集器死循环问题
    - DaemonSet 文档补全

---

## 1.1.9-rc4.1(2021/11/16)

- 修复 Kubernetes Pod 采集 namespace 命名空间问题(#439)

---

## 1.1.9-rc4(2021/11/09)

- 支持[通过 Git 来管理](datakit-conf#90362fd0) 各种采集器配置（`datakit.conf` 除外）以及 Pipeline(#366)
- 支持[全离线安装](datakit-offline-install#7f3c40b6)(#421)
<!--
- eBPF-network
    - 增加[DNS 数据采集]()(#418)
    - 增强内核适配性，内核版本要求已降低至 Linux 4.4+(#416) -->
- 增强数据调试功能，采集到的数据支持写入本地文件，同时发送到中心(#415)
- K8s 环境中，默认开启的采集器支持通过环境变量注入 tags，详见各个默认开启的采集器文档(#408)
- DataKit 支持[一键上传日志](datakit-tools-how-to#0b4d9e46)(#405)
<!-- - MySQL 采集器增加[SQL 语句执行性能指标]()(#382) -->
- 修复安装脚本中 root 用户设定的 bug(#430)
- 增强 Kubernetes 采集器：
    - 添加通过 Annotation 配置 [Pod 日志采集](kubernetes-podlogging)(#380)
    - 增加更多 Annotation key，[支持多 IP 情况](kubernetes-prom#b8ba2a9e)(#419)
    - 支持采集 Node IP(#411)
    - 优化 Annotation 在采集器配置中的使用(#380)
- 云同步增加[华为云与微软云支持](hostobject#031406b2)(#265)

---

## 1.1.9-rc3(2021/10/26)

- 优化 [Redis 采集器](redis) DB 配置方式(#395)
- 修复 [Kubernetes](kubernetes) 采集器 tag 取值为空的问题(#409)
- 安装过程修复 Mac M1 芯片支持(#407)
- [eBPF-network](net_ebpf) 修复连接数统计错误问题(#387)
- 日志采集新增[日志数据获取方式](logstreaming)，支持 [Fluentd/Logstash 等数据接入](logstreaming)(#394/#392/#391)
- [ElasticSearch](elasticsearch) 采集器增加更多指标采集(#386)
- APM 增加 [Jaeger 数据](jaeger)接入(#383)
- [Prometheus Remote Write](prom_remote_write)采集器支持数据切割调试
- 优化 [Nginx 代理](proxy#a64f44d8)功能
- DQL 查询结果支持 [CSV 文件导出](datakit-dql-how-to#2368bf1d)

---

## 1.1.9-rc2(2021/10/14)

- 新增[采集器](prom_remote_write)支持 Prometheus Remote Write 将数据同步给 DataKit(#381)
- 新增[Kubernetes Event 数据采集](kubernetes#49edf2c4)(#296)
- 修复 Mac 因安全策略导致安装失败问题(#379)
- [prom 采集器](prom) 调试工具支持从本地文件调试数据切割(#378)
- 修复 [etcd 采集器](etcd)数据问题(#377)
- DataKit Docker 镜像增加 arm64 架构支持(#365)
- 安装阶段新增环境变量 `DK_HOSTNAME` [支持](datakit-install#f9858758)(#334)
- [Apache 采集器](apache) 增加更多指标采集 (#329)
- DataKit API 新增接口 [`/v1/workspace`](apis#2a24dd46) 以获取工作空间信息(#324)
    - 支持 DataKit 通过命令行参数[获取工作空间信息](datakit-tools-how-to#88b4967d)

---

## 1.1.9-rc1.1(2021/10/09)

- 修复 Kubernetes 选举问题(#389)
- 修复 MongoDB 配置兼容性问题

---

## 1.1.9-rc1(2021/09/28)

- 完善 Kubernetes 生态下 [Prometheus 类指标采集](kubernetes-prom)(#368/#347)
- [eBPF-network](net_ebpf) 优化
- 修复 DataKit/DataWay 之间连接数泄露问题(#290)
- 修复容器模式下 DataKit 各种子命令无法执行的问题(#375)
- 修复日志采集器因 Pipeline 错误丢失原始数据的问题(#376)
- 完善 DataKit 端 [DCA](dca) 相关功能，支持在安装阶段[开启 DCA 功能](datakit-install#f9858758)。
- 下线浏览器拨测功能

---

## 1.1.9-rc0(2021/09/23)

- [日志采集器](logging)增加特殊字符（如颜色字符）过滤功能（默认关闭）(#351)
- [完善容器日志采集](container#6a1b31bb)，同步更多现有普通日志采集器功能（多行匹配/日志等级过滤/字符编码等）(#340)
- [主机对象](hostobject)采集器字段微调(#348)
- 新增如下几个采集器
    - [eBPF-network](net_ebpf)(alpha)(#148)
    - [Consul](consul)(#303)
    - [etcd](etcd)(#304)
    - [CoreDNS](coredns)(#305)
- 选举功能已经覆盖到如下采集器：(#288)
    - [Kubernetes](kubernetes)
    - [Prom](prom)
    - [GitLab](gitlab)
    - [NSQ](nsq)
    - [Apache](apache)
    - [InfluxDB](influxdb)
    - [ElasticSearch](elasticsearch)
    - [MongoDB](mongodb)
    - [MySQL](mysql)
    - [Nginx](nginx)
    - [PostgreSQL](postgresql)
    - [RabbitMQ](rabbitmq)
    - [Redis](redis)
    - [Solr](solr)

<!--
- [DCA](dca) 相关功能完善
	- 独立端口分离(#341)
	- 远程重启功能调整(#345)
	- 白名单功能(#244) -->

---

## 1.1.8-rc3(2021/09/10)

- DDTrace 增加 [resource 过滤](ddtrace#224e2ccd)功能(#328)
- 新增 [NSQ](nsq) 采集器(#312)
- K8s DaemonSet 部署时，部分采集器支持通过环境变量来变更默认配置，以[CPU 为例](cpu#1b85f981)(#309)
- 初步支持 [SkyWalkingV3](skywalking)(alpha)(#335)

### Bugs

- [RUM](rum) 采集器移除全文字段，减少网络开销(#349)
- [日志采集器](logging)增加对文件 truncate 情况的处理(#271)
- 日志字段切割错误字段兼容(#342)
- 修复[离线下载](datakit-offline-install)时可能出现的 TLS 错误(#330)

### 改进

- 日志采集器一旦配置成功，则触发一条通知日志，表明对应文件的日志采集已经开启(#323)

---

## 1.1.8-rc2.4(2021/08/26)

- 修复安装程序开启云同步导致无法安装的问题

---

## 1.1.8-rc2.3(2021/08/26)

- 修复容器运行时无法启动的问题

---

## 1.1.8-rc2.2(2021/08/26)

- 修复 [`hostdir`](hostdir) 配置文件不存在问题

---

## 1.1.8-rc2.1(2021/08/25)

- 修复 CPU 温度采集导致的无数据问题
- 修复 statsd 采集器退出崩溃问题(#321)
- 修复代理模式下自动提示的升级命令问题

---

## 1.1.8-rc2(2021/08/24)

- 支持同步 Kubernetes labels 到各种对象上（pod/service/...）(#279)
- `datakit` 指标集增加数据丢弃指标(#286)
- [Kubernetes 集群自定义指标采集](kubernetes-prom) 优化(#283)
- [ElasticSearch](elasticsearch) 采集器完善(#275)
- 新增[主机目录](hostdir)采集器(#264)
- [CPU](cpu) 采集器支持单个 CPU 指标采集(#317)
- [DDTrace](ddtrace) 支持多路由配置(#310)
- [DDTrace](ddtrace#fb3a6e17) 支持自定义业务 tag 提取(#316)
- [主机对象](hostobject)上报的采集器错误，只上报最近 30s(含)以内的错误(#318)
- [DCA 客户端](dca)发布
- 禁用 Windows 下部分命令行帮助(#319)
- 调整 DataKit [安装形式](datakit-install)，[离线安装](datakit-offline-install)方式做了调整(#300)
    - 调整之后，依然兼容之前老的安装方式

### Breaking Changes

- 从环境变量 `ENV_HOSTNAME` 获取主机名的功能已移除（1.1.7-rc8 支持），可通过[主机名覆盖功能](datakit-install#987d5f91) 来实现
- 移除命令选项 `--reload`
- 移除 DataKit API `/reload`，代之以 `/restart`
- 由于调整了命令行选项，之前的查看 monitor 的命令，也需要 sudo 权限运行（因为要读取 *datakit.conf* 自动获取 Datakit 的配置）

---

## 1.1.8-rc1.1(2021/08/13)

- 修复 `ENV_HTTP_LISTEN` 无效问题，该问题导致容器部署（含 K8s DaemonSet 部署）时，HTTP 服务启动异常。

---

## 1.1.8-rc1(2021/08/10)

- 修复云同步开启时，无法上报主机对象的问题
- 修复 Mac 上新装 DataKit 无法启动的问题
- 修复 Mac/Linux 上非 `root` 用户操作服务「假成功」的问题
- 优化数据上传的性能
- [`proxy`](proxy) 采集器支持全局代理功能，涉及内网环境的安装、更新、数据上传方式的调整
- 日志采集器性能优化
- 文档完善

---

## 1.1.8-rc0(2021/08/03)

- 完善 [Kubernetes](kubernetes) 采集器，增加更多 Kubernetes 对象采集
- 完善[主机名覆盖功能](datakit-install#987d5f91)
- 优化 Pipeline 处理性能（约 15 倍左右，视不同 Pipeline 复杂度而定）
- 加强[行协议数据检查](apis#2fc2526a)
- `system` 采集器，增加 [`conntrack` 以及 `filefd`](system) 两个指标集
- `datakit.conf` 增加 IO 调参入口，便于用户对 DataKit 网络出口流量做优化（参见下面的 Breaking Changes）
- DataKit 支持[服务卸载和恢复](datakit-service-how-to#9e00a535)
- Windows 平台的服务支持通过[命令行管理](datakit-service-how-to#147762ed)
- DataKit 支持动态获取最新 DataWay 地址，避免默认 DataWay 被 DDos 攻击
- DataKit 日志支持[输出到终端](datakit-daemonset-deploy#00c8a780)（Windows 暂不不支持），便于 k8s 部署时日志查看、采集
- 调整 DataKit 主配置，各个不同配置模块化（详见下面的 Breaking Changes）
- 其它一些 bug 修复，完善现有的各种文档

### Breaking Changes

以下改动，在升级过程中会*自动调整*，这里只是提及具体变更，便于大家理解

- 主配置修改：增加如下几个模块

```toml
[io]
  feed_chan_size                 = 1024  # IO 管道缓存大小
  hight_frequency_feed_chan_size = 2048  # 高频 IO 管道缓存大小
  max_cache_count                = 1024  # 本地缓存最大值，原主配置中 io_cache_count [此数值与 max_dynamic_cache_count 同时小于等于零将无限使用内存]
  cache_dump_threshold         = 512   # 本地缓存推送后清理剩余缓存阈值 [此数值小于等于零将不清理缓存，如遇网络中断可导致内存大量占用]
  max_dynamic_cache_count      = 1024  # HTTP 缓存最大值，[此数值与 max_cache_count 同时小于等于零将无限使用内存]
  dynamic_cache_dump_threshold = 512   # HTTP 缓存推送后清理剩余缓存阈值，[此数值小于等于零将不清理缓存，如遇网络中断可导致内存大量占用]
  flush_interval               = "10s" # 推送时间间隔
  output_file                  = ""    # 输出 io 数据到本地文件，原主配置中 output_file

[http_api]
	listen          = "localhost:9529" # 原 http_listen
	disable_404page = false            # 原 disable_404page

[logging]
	log           = "/var/log/datakit/log"     # 原 log
	gin_log       = "/var/log/datakit/gin.log" # 原 gin.log
	level         = "info"                     # 原 log_level
	rotate        = 32                         # 原 log_rotate
	disable_color = false                      # 新增配置
```

---

## 1.1.7-rc9.1(2021/07/17)

### 发布说明

- 修复因文件句柄泄露，导致 Windows 平台上重启 DataKit 可能失败的问题

## 1.1.7-rc9(2021/07/15)

### 发布说明

- 安装阶段支持填写云服务商、命名空间以及网卡绑定
- 多命名空间的选举支持
- 新增 [InfluxDB 采集器](influxdb)
- Datakit DQL 增加历史命令存储
- 其它一些细节 bug 修复

---

## 1.1.7-rc8(2021/07/09)

### 发布说明

- 支持 MySQL [用户](mysql#15319c6c)以及[表级别](mysql#3343f732)的指标采集
- 调整 monitor 页面展示
    - 采集器配置情况和采集情况分离显示
    - 增加选举、自动更新状态显示
- 支持从 `ENV_HOSTNAME` 获取主机名，以应付原始主机名不可用的问题
- 支持 tag 级别的 [Trace](ddtrace) 过滤
- [容器采集器](container)支持采集容器内进程对象
- 支持通过 [cgroup 控制 DataKit CPU 占用](datakit-conf#4e7ff8f3)（仅 Linux 支持）
- 新增 [IIS 采集器](iis)

### Bug 修复

- 修复云同步脏数据导致的上传问题

---

## 1.1.7-rc7(2021/07/01) {#cl-1.1.7-rc7}

### 发布说明

- DataKit API 支持，且支持 [JSON Body](apis#75f8e5a2)
- 命令行增加功能：

    - [DQL 查询功能](datakit-dql-how-to#cb421e00)
    - [命令行查看 monitor](datakit-tools-how-to#44462aae)
    - [检查采集器配置是否正确](datakit-tools-how-to#519a9e75)

- 日志性能优化（对各个采集器自带的日志采集而言，目前仅针对 nginx/MySQL/Redis 做了适配，后续将适配其它各个自带日志收集的采集器）

- 主机对象采集器，增加 [`conntrack`](hostobject#2300b531) 和 [`filefd`](hostobject#697f87e2) 俩类指标
- 应用性能指标采集，支持[采样率设置](ddtrace#c59ce95c)
- K8s 集群 Prometheus 指标采集[通用方案](kubernetes-prom)

### Breaking Changes

- 在 *datakit.conf* 中配置的 `global_tags` 中，`host` tag 将不生效，此举主要为了避免大家在配置 host 时造成一些误解（即配置了 `host`，但可能跟实际的主机名不同，造成一些数据误解）

---

## 1.1.7-rc6(2021/06/17)

### 发布说明

- 新增[Windows 事件采集器](windows_event)
- 为便于用户部署 [RUM](rum) 公网 DataKit，提供禁用 DataKit 404 页面的选项
- [容器采集器](container)字段有了新的优化，主要涉及 pod 的 restart/ready/state 等字段
- [Kubernetes 采集器](kubernetes) 增加更多指标采集
- 支持在 DataKit 端对日志进行（黑名单）过滤
    - 注意：如果 DataKit 上配置了多个 DataWay 地址，日志过滤功能将不生效。

### Breaking Changes

对于没有语雀文档支持的采集器，在这次发布中，均已移除（各种云采集器，如阿里云监控数据、费用等采集）。如果有对这些采集器有依赖，不建议升级。

---

## 1.1.7-rc5(2021/06/16)

### 问题修复

修复 [DataKit API](apis) `/v1/query/raw` 无法使用的问题。

---

## 1.1.7-rc4(2021/06/11)

### 问题修复

禁用 Docker 采集器，其功能完全由[容器采集器](container) 来实现。

原因：

- Docker 采集器和容器采集器并存的情况下（DataKit 默认安装、升级情况下，会自动启用容器采集器），会导致数据重复
- 现有 Studio 前端、模板视图等尚不支持最新的容器字段，可能导致用户升级上来之后，看不到容器数据。本版本的容器采集器会冗余一份原 Docker 采集器中采集上来的指标，使得 Studio 能正常工作。

> 注意：如果在老版本中，有针对 Docker 的额外配置，建议手动移植到 [容器采集器](container) 中来。它们之间的配置基本上是兼容的。

---

## 1.1.7-rc3(2021/06/10)

### 发布说明

- 新增 [磁盘 S.M.A.R.T 采集器](smart)
- 新增 [硬件 温度采集器](sensors)
- 新增 [Prometheus 采集器](prom)

### 问题修复

- 修正 [Kubernetes 采集器](kubernetes)，支持更多 K8s 对象统计指标收集
- 完善[容器采集器](container)，支持 image/container/pod 过滤
- 修正 [Mongodb 采集器](mongodb)问题
- 修正 MySQL/Redis 采集器可能因为配置缺失导致崩溃的问题
- 修正[离线安装问题](datakit-offline-install)
- 修正部分采集器日志设置问题
- 修正 [SSH](ssh)/[Jenkins](jenkins) 等采集器的数据问题

---

## 1.1.7-rc2(2021/06/07)

### 发布说明

- 新增 [Kubernetes 采集器](kubernetes)
- DataKit 支持 [DaemonSet 方式部署](datakit-daemonset-deploy)
- 新增 [SQL Server 采集器](sqlserver)
- 新增 [PostgreSQL 采集器](postgresql)
- 新增 [statsd 采集器](statsd)，以支持采集从网络上发送过来的 statsd 数据
- [JVM 采集器](jvm) 优先采用 DDTrace/StatsD 采集
- 新增[容器采集器](container)，增强对 k8s 节点（Node）采集，以替代原有 [docker 采集器](docker)（原 docker 采集器仍可用）
- [拨测采集器](dialtesting)支持 Headless 模式
- [Mongodb 采集器](mongodb) 支持采集 Mongodb 自身日志
- DataKit 新增 DQL HTTP [API 接口](apis) `/v1/query/raw`
- 完善部分采集器文档，增加中间件（如 MySQL/Redis/ES 等）日志采集相关文档

---

## 1.1.7-rc1(2021/05/26)

### 发布说明

- 修复 Redis/MySQL 采集器数据异常问题
- MySQL InnoDB 指标重构，具体细节参考 [MySQL 文档](mysql#e370e857)

---

## 1.1.7-rc0(2021/05/20)

### 发布说明

新增采集器：

- [Apache](apache)
- [Cloudprober 接入](cloudprober)
- [GitLab](gitlab)
- [Jenkins](jenkins)
- [Memcached](memcached)
- [Mongodb](mongodb)
- [SSH](ssh)
- [Solr](solr)
- [Tomcat](tomcat)

新功能相关：

- 网络拨测支持私有节点接入
- Linux 平台默认开启容器对象、日志采集
- CPU 采集器支持温度数据采集
- [MySQL 慢日志支持阿里云 RDS 格式切割](mysql#ee953f78)

其它各种 Bug 修复。

### Breaking Changes

[RUM 采集](rum)中数据类型做了调整，原有数据类型基本已经废弃，需[更新对应 SDK](/dataflux/doc/eqs7v2)。

---

## 1.1.6-rc7(2021/05/19)

### 发布说明

- 修复 Windows 平台安装、升级问题

---

## 1.1.6-rc6(2021/05/19)

### 发布说明

- 修复部分采集器（MySQL/Redis）数据处理过程中， 因缺少指标导致的数据问题
- 其它一些 bug 修复

---

## 1.1.6-rc5(2021/05/18)

### 发布说明

- 修复 HTTP API precision 解析问题，导致部分数据时间戳解析失败

---

## 1.1.6-rc4(2021/05/17)

### 发布说明

- 修复容器日志采集可能崩溃的问题

---

## 1.1.6-rc3(2021/05/13)

### 发布说明

本次发布，有如下更新：

- DataKit 安装/升级后，安装目录变更为

    - Linux/Mac: `/usr/local/datakit`，日志目录为 `/var/log/datakit`
    - Windows: `C:\Program Files\datakit`，日志目录就在安装目录下

- 支持 [`/v1/ping` 接口](apis#50ea0eb5)
- 移除 RUM 采集器，RUM 接口[默认已经支持](apis#f53903a9)
- 新增 monitor 页面：http://localhost:9529/monitor，以替代之前的 /stats 页面。reload 之后自动跳转到 monitor 页面
- 支持命令直接[安装 sec-checker](datakit-tools-how-to#01243fef) 以及[更新 ip-db](datakit-tools-how-to#ab5cd5ad)

---

## 1.1.6-rc2(2021/05/11)

### Bug 修复

- 修复容器部署情况下无法启动的问题

---

## 1.1.6-rc1(2021/05/10)

### 发布说明

本次发布，对 DataKit 的一些细节做了调整：

- DataKit 上支持配置多个 DataWay
- [云关联](hostobject#031406b2)通过对应 meta 接口来实现
- 调整 docker 日志采集的[过滤方式](docker#a487059d)
- [DataKit 支持选举](election)
- 修复拨测历史数据清理问题
- 大量文档[发布到语雀](https://www.yuque.com/dataflux/datakit){:target="_blank"}
- [DataKit 支持命令行集成 Telegraf](datakit-tools-how-to#d1b3b29b)
- DataKit 单实例运行检测
- DataKit [自动更新功能](datakit-update-crontab)

---

## 1.1.6-rc0(2021/04/30)

### 发布说明

本次发布，对 DataKit 的一些细节做了调整：

- Linux/Mac 安装完后，能直接在任何目录执行 `datakit` 命令，无需切换到 DataKit 安装目录
- Pipeline 增加脱敏函数 `cover()`
- 优化命令行参数，更加便捷
- 主机对象采集，默认过滤虚拟设备（仅 Linux 支持）
- Datakit 命令支持 `--start/--stop/--restart/--reload` 几个命令（需 root 权限），更加便于大家管理 DataKit 服务
- 安装/升级完成后，默认开启进程对象采集器（目前默认开启列表为 `cpu/disk/diskio/mem/swap/system/hostobject/net/host_processes`）
- 日志采集器 `tailf` 改名为 `logging`，原有的 `tailf` 名称继续可用
- 支持接入 Security 数据
- 移除 Telegraf 安装集成。如果需要 Telegraf 功能，可查看 :9529/man 页面，有专门针对 Telegraf 安装使用的文档
- 增加 Datakit How To 文档，便于大家初步入门（:9529/man 页面可看到）
- 其它一些采集器的指标采集调整

---

## v1.1.5-rc2(2021/04/22)

### Bug 修复

- 修复 Windows 上 `--version` 命令请求线上版本信息的地址错误
- 调整华为云监控数据采集配置，放出更多可配置信息，便于实时调整
- 调整 Nginx 错误日志（error.log）切割脚本，同时增加默认日志等级的归类

---

## v1.1.5-rc1(2021/04/21)

### Bug 修复

- 修复 `tailf` 采集器配置文件兼容性问题，该问题导致 `tailf` 采集器无法运行

---

## v1.1.5-rc0(2021/04/20)

### 发布说明

本次发布，对采集器做了较大的调整。

### Breaking Changes

涉及的采集器列表如下：

| 采集器          | 说明                                                                                                                                                                                      |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cpu`           | DataKit 内置 CPU 采集器，移除 Telegraf CPU 采集器，配置文件保持兼容。另外，Mac 平台暂不支持 CPU 采集，后续会补上                                                                          |
| `disk`          | DataKit 内置磁盘采集器                                                                                                                                                                    |
| `docker`        | 重新开发了 docker 采集器，同时支持容器对象、容器日志以及容器指标采集（额外增加对 K8s 容器采集）                                                                                           |
| `elasticsearch` | Datakit 内置 ES 采集器，同时移除 Telegraf 中的 ES 采集器。另外，可在该采集器中直接配置采集 ES 日志                                                                                        |
| `jvm`           | Datakit 内置 JVM 采集器                                                                                                                                                                   |
| `kafka`         | Datakit 内置 Kafka 指标采集器，可在该采集器中直接采集 Kafka 日志                                                                                                                          |
| `mem`           | Datakit 内置内存采集器，移除 Telegraf 内存采集器，配置文件保持兼容                                                                                                                        |
| `mysql`         | Datakit 内置 MySQL 采集器，移除 Telegraf MySQL 采集器。可在该采集器中直接采集 MySQL 日志                                                                                                  |
| `net`           | Datakit 内置网络采集器，移除 Telegraf 网络采集器。在 Linux 上，对于虚拟网卡设备，默认不再采集（需手动开启）                                                                               |
| `nginx`         | Datakit 内置 NGINX 采集器，移除 Telegraf NGINX 采集器。可在该采集器中直接采集 NGINX 日志                                                                                                  |
| `oracle`        | Datakit 内置 Oracle 采集器。可在该采集器中直接采集 Oracle 日志                                                                                                                            |
| `rabbitmq`      | Datakit 内置 RabbitMQ 采集器。可在该采集器中直接采集 RabbitMQ 日志                                                                                                                        |
| `redis`         | Datakit 内置 Redis 采集器。可在该采集器中直接采集 Redis 日志                                                                                                                              |
| `swap`          | Datakit 内置内存 swap 采集器                                                                                                                                                              |
| `system`        | Datakit 内置 system 采集器，移除 Telegraf system 采集器。内置的 system 采集器新增三个指标： `load1_per_core/load5_per_core/load15_per_core`，便于客户端直接显示单核平均负载，无需额外计算 |

以上采集器的更新，非主机类型的采集器，绝大部分涉均有指标集、指标名的更新，具体参考各个采集器文档。

其它兼容性问题：

- 出于安全考虑，采集器不再默认绑定所有网卡，默认绑定在 `localhost:9529` 上。原来绑定的 `0.0.0.0:9529` 已失效（字段 `http_server_addr` 也已弃用），可手动修改 `http_listen`，将其设定成 `http_listen = "0.0.0.0:9529"`（此处端口可变更）
- 某些中间件（如 MySQL/Nginx/Docker 等）已经集成了对应的日志采集，它们的日志采集可以直接在对应的采集器中配置，无需额外用 `tailf` 来采集了（但 `tailf` 仍然可以单独采集这些日志）
- 以下采集器，将不再有效，请用上面内置的采集器来采集
    - `dockerlog`：已集成到 docker 采集器
    - `docker_containers`：已集成到 docker 采集器
    - `mysqlMonitor`：以集成到 mysql 采集器

### 新增特性

- 拨测采集器（`dialtesting`）：支持中心化任务下发，在 Studio 主页，有单独的拨测入口，可创建拨测任务来试用
- 所有采集器的配置项中，支持配置环境变量，如 `host="$K8S_HOST"`，便于容器环境的部署
- http://localhost:9529/stats 新增更多采集器运行信息统计，包括采集频率（`frequency`）、每次上报数据条数（`avg_size`）、每次采集消耗（`avg_collect_cost`）等。部分采集器可能某些字段没有，这个不影响，因为每个采集器的采集方式不同
- http://localhost:9529/reload 可用于重新加载采集器，比如修改了配置文件后，可以直接 `curl http://localhost:9529/reload` 即可，这种形式不会重启服务，类似 Nginx 中的 `-s reload` 功能。当然也能在浏览器上直接访问该 reload 地址，reload 成功后，会自动跳转到 stats 页面
- 支持在 http://localhost:9529/man 页面浏览 DataKit 文档（只有此次新改的采集器文档集成过来了，其它采集器文档需在原来的帮助中心查看）。默认情况下不支持远程查看 DataKit 文档，可在终端查看（仅 Mac/Linux 支持）：

```shell
	# 进入采集器安装目录，输入采集器名字（通过 `Tab` 键选择自动补全）即可查看文档
	$ ./datakit -cmd -man
	man > nginx
	(显示 Nginx 采集文档)
	man > mysql
	(显示 MySQL 采集文档)
	man > Q               # 输入 Q 或 exit 退出
```

---

## v1.1.4-rc2(2021/04/07)

### Bug 修复

- 修复阿里云监控数据采集器（`aliyuncms`）频繁采集导致部分其它采集器卡死的问题。

---

## v1.1.4-rc1(2021/03/25)

### 改进

- 进程采集器 `message` 字段增加更多信息，便于全文搜索
- 主机对象采集器支持自定义 tag，便于云属性同步

---

## v1.1.4-rc0(2021/03/25)

### 新增功能

- 增加文件采集器、拨测采集器以及 HTTP 报文采集器
- 内置支持 ActiveMQ/Kafka/RabbitMQ/gin（Gin HTTP 访问日志）/Zap（第三方日志框架）日志切割

### 改进

- 丰富 `http://localhost:9529/stats` 页面统计信息，增加诸如采集频率（`n/min`），每次采集的数据量大小等
- DataKit 本身增加一定的缓存空间（重启即失效），避免偶然的网络原因导致数据丢失
- 改进 Pipeline 日期转换函数，提升准确性。另外增加了更多 Pipeline 函数（`parse_duration()/parse_date()`）
- trace 数据增加更多业务字段（`project/env/version/http_method/http_status_code`）
- 其它采集器各种细节改进

---

## v1.1.3-rc4(2021/03/16)

### Bug 修复

- 进程采集器：修复用户名缺失导致显示空白的问题，对用户名获取失败的进程，以 `nobody` 当做其用户名。

---

## v1.1.3-rc3(2021/03/04)

### Bug 修复

- 修复进程采集器部分空字段（进程用户以及进程命令缺失）问题
- 修复 Kubernetes 采集器内存占用率计算可能 panic 的问题

---

## v1.1.3-rc2(2021/03/01)

### Bug 修复

- 修复进程对象采集器 `name` 字段命名问题，以 `hostname + pid` 来命名 `name` 字段
- 修正华为云对象采集器 Pipeline 问题
- 修复 Nginx/MySQL/Redis 日志采集器升级后的兼容性问题

---

## v1.1.3-rc1(2021/02/26)

### 新增功能

- 增加内置 Redis/Nginx
- 完善 MySQL 慢查询日志分析

### 功能改进

- 进程采集器由于单次采集耗时过长，对采集器的采集频率做了最小值（30s）限制
- 采集器配置文件名称不再严格限制，任何形如 `xxx.conf` 的文件，都是合法的文件命名
- 更新版本提示判断，如果 git 提交码跟线上不一致，也会提示更新
- 容器对象采集器（`docker_containers`），增加内存/CPU 占比字段（`mem_usage_percent/cpu_usage`）
- K8s 指标采集器（`kubernetes`），增加 CPU 占比字段（`cpu_usage`）
- Tracing 数据采集完善对 service type 处理
- 部分采集器支持自定义写入日志或者指标（默认指标）

### Bug 修复

- 修复 Mac 平台上，进程采集器获取默认用户名无效的问题
- 修正容器对象采集器，获取不到*已退出容器*的问题
- 其它一些细节 bug 修复

### Breaking Changes

- 对于某些采集器，如果原始指标中带有 `uint64` 类型的字段，新版本会导致字段不兼容，应该删掉原有指标集，避免类型冲突

    - 原来对于 uint64 的处理，将其自动转成了 string，这会导致使用过程中困扰。实际上可以更为精确的控制这个整数移除的问题
    - 对于超过 max-int64 的 uint 整数，采集器会丢弃这样的指标，因为目前 influx1.7 不支持 uint64 的指标

- 移除部分原 `dkctrl` 命令执行功能，配置管理功能后续不再依赖该方式实现

---

## v1.1.2(2021/02/03)

### 功能改进

- 容器安装时，必须注入 `ENV_UUID` 环境变量
- 从旧版本升级后，会自动开启主机采集器（原 *datakit.conf* 会备份一个）
- 添加缓存功能，当出现网络抖动的情况下，不至于丢失采集到的数据（当长时间网络瘫痪的情况下，数据还是会丢失）
- 所有使用 `tailf` 采集的日志，必须在 Pipeline 中用 `time` 字段来指定切割出来的时间字段，否则日志存入时间字段会跟日志实际时间有出入

### Bug 修复

- 修复 Zipkin 中时间单位问题
- 主机对象出采集器中添加 `state` 字段

---

## v1.1.1(2021/02/01)

### Bug 修复

- 修复 Mysql Monitor 采集器 `status/variable` 字段均为 string 类型的问题。回退至原始字段类型。同时对 int64 溢出问题做了保护。
- 更改进程采集器部分字段命名，使其跟主机采集器命名一致

---

## v1.1.0(2021/01/29)

### 发布说明

本版本主要涉及部分采集器的 bug 修复以及 Datakit 主配置的调整。

### Breaking Changes

- 采用新的版本号机制，原来形如 `v1.0.0-2002-g1fe9f870` 这样的版本号将不再使用，改用 `v1.2.3` 这样的版本号
- 原 DataKit 顶层目录的 `datakit.conf` 配置移入 `conf.d` 目录
- 原 `network/net.conf` 移入 `host/net.conf`
- 原 `pattern` 目录转移到 `pipeline` 目录下
- 原 grok 中内置的 pattern，如 `%{space}` 等，都改成大写形式 `%{SPACE}`。**之前写好的 grok 需全量替换**
- 移除 `datakit.conf` 中 `uuid` 字段，单独用 `.id` 文件存放，便于统一 DataKit 所有配置文件
- 移除 Ansible 采集器事件数据上报

### Bug 修复

- 修复 `prom`、`oraclemonitor` 采集不到数据的问题
- `self` 采集器将主机名字段 hostname 改名成 host，并置于 tag 上
- 修复 `mysqlMonitor` 同时采集 MySQL 和 MariaDB 类型冲突问题
- 修复 SkyWalking 采集器日志不切分导致磁盘爆满问题

### 特性

- 新增采集器/主机黑白名单功能（暂不支持正则）
- 重构主机、进程、容器等对象采集器采集器
- 新增 Pipeline/Grok 调试工具
- `-version` 参数除了能看当前版本，还将提示线上新版本信息以及更新命令
- 支持 DDTrace 数据接入
- `tailf` 采集器新日志匹配改成正向匹配
- 其它一些细节问题修复
- 支持 Mac 平台的 CPU 数据采集

<!--
[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)
[:fontawesome-solid-flag-checkered:](index.md#legends "支持选举")

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

# 外链的添加方式
[some text](http://external-host.com){:target="_blank"}

## x.x.x(YY/MM/DD) {#cl-x.x.x}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-x.x.x-new}
### 问题修复 {#cl-x.x.x-fix}
### 功能优化 {#cl-x.x.x-opt}
### 兼容调整 {#cl-x.x.x-brk}
-->
<!-- markdown-link-check-enable -->
