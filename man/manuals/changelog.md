{{.CSS}}

# DataKit 版本历史

## 1.2.4(2022/01/12)

- 修复日志 API 接口指标丢失问题(#551)
- 修复 [eBPF](ebpf) 网络流量统计部分丢失问题(#556)
- 修复采集器配置文件中 `$` 字符通配问题(#550)
- Pipeline *if* 语句支持空值比较，便于 Grok 切割判断(#538)

---

## 1.2.3(2022/01/10)

- 修复 datakit.yaml 格式错误问题(#544)
- 修复 [MySQL 采集器](mysql)选举问题(#543)
- 修复因 Pipeline 不配置导致日志不采集的问题(#546)

---

## 1.2.2(2022/01/07)

- [容器采集器](container)更新：
	- 修复日志处理效率问题(#540)
	-	优化配置文件黑白名单配置(#536)
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

## 1.2.0(2021/12/30)

### 采集器更新

- 重构 Kubernetes 云原生采集器，将其整合进[容器采集器](container)。原有 Kubernetes 采集器不再生效(#492)
- [Redis 采集器](redis)
	- 支持配置 [Redis 用户名](redis#852abae7)(#260)
	- 增加 [Latency](redis#1355d1f8) 以及 [Cluster](redis#786114c8) 指标集(#396)
- [Kafka 采集器](kafka)增强，支持 topic/broker/consumer/connnetion 等维度的指标(#397)
- 新增 [ClickHouse](clickhousev1) 以及 [Flink](flinkv1) 采集器(#458/#459)
- [主机对象采集器](hostobject)
	- 支持从 [`ENV_CLOUD_PROVIDER`](hostobject#224e2ccd) 读取云同步配置(#501)
	- 优化磁盘采集，默认不会再采集无效磁盘（比如总大小为 0 的一些磁盘）(#505)
- [日志采集器](logging) 支持接收 TCP/UDP 日志流(#503)
- [Prom 采集器](prom) 支持多 URL 采集(#506)
- 新增 [eBPF](ebpf) 采集器，它集成了 L4-network/DNS/Bash 等 eBFP 数据采集(507)
- [ElasticSearch采集器](elasticsearch) 增加 [Open Distro](https://opendistro.github.io/for-elasticsearch/) 分支的 ElasticSearch 支持(#510)

### Bug 修复

- 修复 [Statsd](statsd)/[Rabbitmq](rabbitmq) 指标问题(#497)
- 修复 [Windows Event](windows_event) 采集数据问题(#521)

### 其它

- [Pipeline](pipeline)
	- 增强 Pipeline 并行处理能力
	- 增加 [`set_tag()`](pipeline#6e8c5285) 函数(#444)
	- 增加 [`drop()`](pipeline#6e8c5285) 函数(#498)
- Git 模式
	- 在 DaemonSet 模式下的 Git，支持识别 `ENV_DEFAULT_ENABLED_INPUTS` 并将其生效，非 DaemonSet 模式下，会自动开启 datakit.conf 中默认开启的采集器(#501)
	- 调整 Git 模式下文件夹[存放策略]()(#509)
- 推行新的版本号机制(#484)
	- 新的版本号形式为 1.2.3，此处 `1` 为 master 版本号，`2` 为 minor 版本号，`3` 为 mini 版本号
	- 以 minor 版本号的奇偶性来判定是稳定版（偶数）还是非稳定版（奇数）
	- 同一个 minor 版本号上，会有多个不同的 mini 版本号，主要用于问题修复以及功能调整
	- 新功能预计会发布在非稳定版上，待新功能稳定后，会发布新的稳定版本。如 1.3.x 新功能稳定后，会发布 1.4.0 稳定版，以合并 1.3.x 上的新功能
	- 非稳定版不支持直接升级，比如，不能升级到 1.3.x 这样的版本，只能直接安装非稳定版
	- **老版本的 DataKit 通过 `datakit --version` 已经无法推送新升级命令**，直接使用如下命令：
		- Linux/Mac: `DK_UPGRADE=1 bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"`
		- Windows: `$env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;`

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
	- DataKit 安装完成后自动增加 [cgroup 限制](datakit-conf-how-to#9e364a84)
	- 部分跟中心对接的接口升级到了 v2 版本，故对接**非 SAAS 节点**的 DataKit，如果升级到当前版本，其对应的 DataWay 以及 Kodo 也需要升级，否则部分接口会报告 404 错误

### Breaking Changes

处理 json 数据时，如果最顶层是数组，需要使用下标方式进行选择，例如 JSON

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
	- 移除 `json_all()` 函数，这个函数对于异常的 json 有严重的数据问题，故选择禁用之(#457)
	- 修正 `default_time()` 函数时区设置问题(#434)
- 解决 [prom](prom) 采集器在 Kubernetes 环境下 HTTPS 访问问题(#447)
- DataKit DaemonSet 安装的 [yaml 文件](https://static.guance.com/datakit/datakit.yaml) 公网可直接下载
---

## 1.1.9-rc5.1(2021/11/26)

- 修复 ddtrace 采集器因脏数据挂掉的问题

---

## 1.1.9-rc5(2021/11/23)

- 增加 [pythond(alpha)](pythond) ，便于用 Python3 编写自定义采集器(#367)
<!-- - 支持 source map 文件处理，便于 RUM 采集器收集 JavaScript 调用栈信息(#266) -->
- [SkyWalking V3](skywalking) 已支持到 8.5.0/8.6.0/8.7.0 三个版本(#385)
- DataKit 初步支持[磁盘数据缓存(alpha)](datakit-conf-how-to#9dc84d15)(#420)
- DataKit 支持选举状态上报(#427)
- DataKit 支持 scheck 状态上报(#428)
- 调整 DataKit 使用入门文档，新的分类更便于找到具体文档

----

## 1.1.9-rc4.3(2021/11/19)

- 修复容器日志采集器因 pipeline 配置失当无法启动的问题

----

## 1.1.9-rc4.2(2021/11/18)

- 紧急修复(#446)
	- 修复 Kubernetes 模式下 stdout 日志输出 level 异常
	- 修复选举模式下，未选举上的 MySQL 采集器死循环问题
	- DaemonSet 文档补全

----

## 1.1.9-rc4.1(2021/11/16)

- 修复 Kubernetes Pod 采集 namespace 命名空间问题(#439)

-----

## 1.1.9-rc4(2021/11/09)

- 支持[通过 Git 来管理](datakit-conf-how-to#5dd2079e) 各种采集器配置（`datakit.conf` 除外）以及 Pipeline(#366)
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

------

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
- 修复容器模式下DataKit 各种子命令无法执行的问题(#375)
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
	- [Gitlab](gitlab)
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

----

## 1.1.8-rc3(2021/09/10)

- ddtrace 增加 [resource 过滤](ddtrace#224e2ccd)功能(#328)
- 新增 [NSQ](nsq) 采集器(#312)
- K8s daemonset 部署时，部分采集器支持通过环境变量来变更默认配置，以[CPU为例](cpu#1b85f981)(#309)
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

- 修复 [hostdir](hostdir) 配置文件不存在问题

---

## 1.1.8-rc2.1(2021/08/25)

- 修复 CPU 温度采集导致的无数据问题
- 修复 statsd 采集器退出奔溃问题(#321)
- 修复代理模式下自动提示的升级命令问题

---

## 1.1.8-rc2(2021/08/24)

- 支持同步 Kubernetes labels 到各种对象上（pod/service/...）(#279)
- `datakit` 指标集增加数据丢弃指标(#286)
- [Kubernetes 集群自定义指标采集](kubernetes-prom) 优化(#283)
- [ElasticSearch](elasticsearch) 采集器完善(#275)
- 新增[主机目录](hostdir)采集器(#264)
- [CPU](cpu) 采集器支持单个 CPU 指标采集(#317)
- [ddtrace](ddtrace) 支持多路由配置(#310)
- [ddtrace](ddtrace#fb3a6e17) 支持自定义业务 tag 提取(#316)
- [主机对象](hostobject)上报的采集器错误，只上报最近 30s(含)以内的错误(#318)
- [DCA客户端](dca)发布
- 禁用 Windows 下部分命令行帮助(#319)
- 调整 DataKit [安装形式](datakit-install)，[离线安装](datakit-offline-install)方式做了调整(#300)
	- 调整之后，依然兼容之前老的安装方式

### Breaking Changes

- 从环境变量 `ENV_HOSTNAME` 获取主机名的功能已移除（1.1.7-rc8 支持），可通过[主机名覆盖功能](datakit-install#987d5f91) 来实现
- 移除命令选项 `--reload`
- 移除 DataKit API `/reload`，代之以 `/restart`
- 由于调整了命令行选项，之前的查看 monitor 的命令，也需要 sudo 权限运行（因为要读取 datakit.conf 自动获取 DataKit 的配置）

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
- 加强[行协议数据检查](apis#f54b954f)
- `system` 采集器，增加 [`conntrack`以及`filefd`](system) 两个指标集
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
  feed_chan_size                 = 1024  # IO管道缓存大小
  hight_frequency_feed_chan_size = 2048  # 高频IO管道缓存大小
  max_cache_count                = 1024  # 本地缓存最大值，原主配置中 io_cache_count [此数值与max_dynamic_cache_count同时小于等于零将无限使用内存]
  cache_dump_threshold         = 512   # 本地缓存推送后清理剩余缓存阈值 [此数值小于等于零将不清理缓存，如遇网络中断可导致内存大量占用]
  max_dynamic_cache_count      = 1024  # HTTP缓存最大值，[此数值与max_cache_count同时小于等于零将无限使用内存]
  dynamic_cache_dump_threshold = 512   # HTTP缓存推送后清理剩余缓存阈值，[此数值小于等于零将不清理缓存，如遇网络中断可导致内存大量占用]
  flush_interval               = "10s" # 推送时间间隔
  output_file                  = ""    # 输出io数据到本地文件，原主配置中 output_file

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
- datakit DQL 增加历史命令存储
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
- 支持通过 [cgroup 控制 DataKit CPU 占用](datakit-conf-how-to#9e364a84)（仅 Linux 支持）
- 新增 [IIS 采集器](iis)

### Bug 修复

- 修复云同步脏数据导致的上传问题

---

## 1.1.7-rc7(2021/07/01)

### 发布说明

- DataKit API 支持，且支持 [JSON Body](apis#75f8e5a2)
- 命令行增加功能：

  - [DQL 查询功能](datakit-dql-how-to#cb421e00)
  - [命令行查看 monitor](datakit-tools-how-to#44462aae)
  - [检查采集器配置是否正确](datakit-tools-how-to#519a9e75)

- 日志性能优化（对各个采集器自带的日志采集而言，目前仅针对 nginx/MySQL/Redis 做了适配，后续将适配其它各个自带日志收集的采集器）

- 主机对象采集器，增加 [conntrack](hostobject#2300b531) 和 [filefd](hostobject#697f87e2) 俩类指标
- 应用性能指标采集，支持[采样率设置](ddtrace#c59ce95c)
- K8s 集群 Prometheus 指标采集[通用方案](kubernetes-prom)

### Breaking Changes

- 在 datakit.conf 中配置的 `global_tags` 中，`host` tag 将不生效，此举主要为了避免大家在配置 host 时造成一些误解（即配置了 `host`，但可能跟实际的主机名不同，造成一些数据误解）

---

## 1.1.7-rc6(2021/06/17)

### 发布说明

- 新增[Windows 事件采集器](windows_event)
- 为便于用户部署 [RUM](rum) 公网 DataKit，提供禁用 DataKit 404 页面的选项
- [容器采集器](container)字段有了新的优化，主要涉及 pod 的 restart/ready/state 等字段
- [Kubernetes 采集器](kubernetes) 增加更多指标采集
- 支持在 DataKit 端[对日志进行（黑名单）过滤](https://www.yuque.com/dataflux/doc/ilhawc#wGemu)
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
- 修正 MySQL/Redis 采集器可能因为配置缺失导致奔溃的问题
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
- [JVM 采集器](jvm) 优先采用 ddtrace + statsd 采集
- 新增[容器采集器](container)，增强对 k8s 节点（Node）采集，以替代原有 [docker 采集器](docker)（原 docker 采集器仍可用）
- [拨测采集器](dialtesting)支持 Headleass 模式
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
- [Gitlab](gitlab)
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

- 修复容器日志采集可能奔溃的问题

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
- 大量文档[发布到语雀](https://www.yuque.com/dataflux/datakit)
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
- datakit 命令支持 `--start/--stop/--restart/--reload` 几个命令（需 root 权限），更加便于大家管理 DataKit 服务
- 安装/升级完成后，默认开启进程对象采集器（目前默认开启列表为 `cpu/disk/diskio/mem/swap/system/hostobject/net/host_processes`）
- 日志采集器 `tailf` 改名为 `logging`，原有的 `tailf` 名称继续可用
- 支持接入 Security 数据
- 移除 Telegraf 安装集成。如果需要 Telegraf 功能，可查看 :9529/man 页面，有专门针对 Telegraf 安装使用的文档
- 增加 datakit-how-to 文档，便于大家初步入门（:9529/man 页面可看到）
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

- 修复 tailf 采集器配置文件兼容性问题，该问题导致 tailf 采集器无法运行

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
| `elasticsearch` | DataKit 内置 ES 采集器，同时移除 Telegraf 中的 ES 采集器。另外，可在该采集器中直接配置采集 ES 日志                                                                                        |
| `jvm`           | DataKit 内置 JVM 采集器                                                                                                                                                                   |
| `kafka`         | DataKit 内置 Kafka 指标采集器，可在该采集器中直接采集 Kafka 日志                                                                                                                          |
| `mem`           | DataKit 内置内存采集器，移除 Telegraf 内存采集器，配置文件保持兼容                                                                                                                        |
| `mysql`         | DataKit 内置 MySQL 采集器，移除 Telegraf MySQL 采集器。可在该采集器中直接采集 MySQL 日志                                                                                                  |
| `net`           | DataKit 内置网络采集器，移除 Telegraf 网络采集器。在 Linux 上，对于虚拟网卡设备，默认不再采集（需手动开启）                                                                               |
| `nginx`         | DataKit 内置 Nginx 采集器，移除 Telegraf Ngxin 采集器。可在该采集器中直接采集 Nginx 日志                                                                                                  |
| `oracle`        | DataKit 内置 Oracle 采集器。可在该采集器中直接采集 Oracle 日志                                                                                                                            |
| `rabbitmq`      | DataKit 内置 RabbitMQ 采集器。可在该采集器中直接采集 RabbitMQ 日志                                                                                                                        |
| `redis`         | DataKit 内置 Redis 采集器。可在该采集器中直接采集 Redis 日志                                                                                                                              |
| `swap`          | DataKit 内置内存 swap 采集器                                                                                                                                                              |
| `system`        | DataKit 内置 system 采集器，移除 Telegraf system 采集器。内置的 system 采集器新增三个指标： `load1_per_core/load5_per_core/load15_per_core`，便于客户端直接显示单核平均负载，无需额外计算 |

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
- 修复 kubernetes 采集器内存占用率计算可能 panic 的问题

<!--
### 新增功能
- `http://datakit:9529/reload` 会自动跳转到 `http://datakit:9529/stats`，便于查看 reload 后 datakit 的运行情况
- `http://datakit:9529/reload` 页面增加每分钟采集频率（`frequency`）以及每次采集的数据量大小统计
- `kubernetes` 指标采集器增加 node 的内存使用率（`mem_usage_percent`）采集 -->

---

## v1.1.3-rc2(2021/03/01)

### Bug 修复

- 修复进程对象采集器 `name` 字段命名问题，以 `hostname + pid` 来命名 `name` 字段
- 修正华为云对象采集器 pipeline 问题
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

- 移除部分原 dkctrl 命令执行功能，配置管理功能后续不再依赖该方式实现

---

## v1.1.2(2021/02/03)

### 功能改进

- 容器安装时，必须注入 `ENV_UUID` 环境变量
- 从旧版本升级后，会自动开启主机采集器（原 datakit.conf 会备份一个）
- 添加缓存功能，当出现网络抖动的情况下，不至于丢失采集到的数据（当长时间网络瘫痪的情况下，数据还是会丢失）
- 所有使用 tailf 采集的日志，必须在 pipeline 中用 `time` 字段来指定切割出来的时间字段，否则日志存入时间字段会跟日志实际时间有出入

### Bug 修复

- 修复 zipkin 中时间单位问题
- 主机对象出采集器中添加 `state` 字段

---

## v1.1.1(2021/02/01)

### Bug 修复

- 修复 mysqlmonitor 采集器 status/variable 字段均为 string 类型的问题。回退至原始字段类型。同时对 int64 溢出问题做了保护。
- 更改进程采集器部分字段命名，使其跟主机采集器命名一致

---

## v1.1.0(2021/01/29)

### 发布说明

本版本主要涉及部分采集器的 bug 修复以及 datakit 主配置的调整。

### Breaking Changes

- 采用新的版本号机制，原来形如 `v1.0.0-2002-g1fe9f870` 这样的版本号将不再使用，改用 `v1.2.3` 这样的版本号
- 原 DataKit 顶层目录的 `datakit.conf` 配置移入 `conf.d` 目录
- 原 `network/net.conf` 移入 `host/net.conf`
- 原 `pattern` 目录转移到 `pipeline` 目录下
- 原 grok 中内置的 pattern，如 `%{space}` 等，都改成大写形式 `%{SPACE}`。**之前写好的 grok 需全量替换**
- 移除 `datakit.conf` 中 `uuid` 字段，单独用 `.id` 文件存放，便于统一 DataKit 所有配置文件
- 移除 ansible 采集器事件数据上报

### Bug 修复

- 修复 `prom`、`oraclemonitor` 采集不到数据的问题
- `self` 采集器将主机名字段 hostname 改名成 host，并置于 tag 上
- 修复 `mysqlMonitor` 同时采集 MySQL 和 MariaDB 类型冲突问题
- 修复 Skywalking 采集器日志不切分导致磁盘爆满问题

### 特性

- 新增采集器/主机黑白名单功能（暂不支持正则）
- 重构主机、进程、容器等对象采集器采集器
- 新增 pipeline/grok 调试工具
- `-version` 参数除了能看当前版本，还将提示线上新版本信息以及更新命令
- 支持 DDTrace 数据接入
- `tailf` 采集器新日志匹配改成正向匹配
- 其它一些细节问题修复
- 支持 Mac 平台的 CPU 数据采集
