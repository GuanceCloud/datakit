{{.CSS}}

# DataKit 版本历史

## 1.1.7-rc2(2021/06/07)

### 发布说明

- 新增 [Kubernetes 采集器](kubernetes)
- 新增 [SQL Server 采集器](sqlserver)
- 新增 [PostgreSQL 采集器](postgresql)
- 新增 [statsd 采集器](statsd)，以支持采集从网络上发送过来的 statsd 数据
- [JVM 采集器](jvm) 优先采用 ddtrace + statsd 采集
- 新增[容器采集器](container)，增强对 k8s 节点（Node）采集，以替代原有 [docker 采集器](docker) 采集器（原 docker 采集器仍可用）
- [拨测采集器](dialtesting)支持 Headleass 模式
- [Mongodb 采集器](mongodb) 支持采集 Mongodb 自身日志
- DataKit 新增 DQL HTTP [API 接口](apis) `/v1/query/raw`
- 完善部分采集器文档，增加中间件（如 MySQL/Redis/ES 等）日志采集相关文档

----

## 1.1.7-rc1(2021/05/26)

### 发布说明

- 修复 Redis/MySQL 采集器数据异常问题
- MySQL InnoDB 指标重构，具体细节参考 [MySQL 文档](mysql#e370e857)

----

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

----

## 1.1.6-rc7(2021/05/19)

### 发布说明

- 修复 Windows 平台安装、升级问题

----

## 1.1.6-rc6(2021/05/19)

### 发布说明

- 修复部分采集器（MySQL/Redis）数据处理过程中， 因缺少指标导致的数据问题
- 其它一些 bug 修复

----

## 1.1.6-rc5(2021/05/18)

### 发布说明

- 修复 HTTP API precision 解析问题，导致部分数据时间戳解析失败
----

## 1.1.6-rc4(2021/05/17)

### 发布说明

- 修复容器日志采集可能奔溃的问题

----

## 1.1.6-rc3(2021/05/13)

### 发布说明

本次发布，有如下更新：

- DataKit 安装/升级后，安装目录变更为
	- Linux/Mac: `/usr/local/datakit`，日志目录为 `/var/log/datakit`
	- Windows: `C:\Program Files\datakit`，日志目录就在安装目录下

- 支持 [`/v1/ping` 接口](apis#50ea0eb5)
- 移除 RUM 采集器，RUM 接口[默认已经支持](apis#f53903a9)
- 新增 monitor 页面：http://localhost:9529/monitor，以替代之前的 /stats 页面。reload 之后自动跳转到 monitor 页面
- 支持命令直接[安装 sec-checker](datakit-how-to#01243fef) 以及[更新 ip-db](datakit-how-to#ab5cd5ad)

----

## 1.1.6-rc2(2021/05/11)

### Bug 修复

- 修复容器部署情况下无法启动的问题

----

## 1.1.6-rc1(2021/05/10)

### 发布说明

本次发布，对 DataKit 的一些细节做了调整：

- DataKit 上支持配置多个 DataWay
- [云关联](hostobject#031406b2)通过对应 meta 接口来实现
- 调整 docker 日志采集的[过滤方式](docker#a487059d)
- [DataKit 支持选举](election)
- 修复拨测历史数据清理问题
- 大量文档[发布到语雀](https://www.yuque.com/dataflux/datakit)
- [DataKit 支持命令行集成 Telegraf](datakit-how-to#d1b3b29b)
- DataKit 单实例运行检测
- DataKit [自动更新功能](datakit-update-crontab)

----

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

----

## v1.1.5-rc2(2021/04/22)

### Bug 修复

- 修复 Windows 上 `--version` 命令请求线上版本信息的地址错误
- 调整华为云监控数据采集配置，放出更多可配置信息，便于实时调整
- 调整 Nginx 错误日志（error.log）切割脚本，同时增加默认日志等级的归类

----

## v1.1.5-rc1(2021/04/21)

### Bug 修复

- 修复 tailf 采集器配置文件兼容性问题，该问题导致 tailf 采集器无法运行

----

## v1.1.5-rc0(2021/04/20)

### 发布说明

本次发布，对采集器做了较大的调整。

### Breaking Changes

涉及的采集器列表如下：

| 采集器          | 说明                                                                                                                                                                                      |
| -----           | ----                                                                                                                                                                                      |
| `cpu`           | DataKit 内置 CPU 采集器，移除 Telegraf CPU 采集器，配置文件保持兼容。另外，Mac 平台暂不支持 CPU 采集，后续会补上                                                                          |
| `disk`          | DataKit 内置磁盘采集器                                                                                                                                                                    |
| `docker`        | 重新开发了 docker 采集器，同时支持容器对象、容器日志以及容器指标采集（额外增加对 K8s 容器采集）                                                                                           |
| `elasticsearch` | DataKit 内置ES 采集器，同时移除 Telegraf 中的 ES 采集器。另外，可在该采集器中直接配置采集 ES 日志                                                                                         |
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
----

## v1.1.4-rc2(2021/04/07)

### Bug 修复

- 修复阿里云监控数据采集器（`aliyuncms`）频繁采集导致部分其它采集器卡死的问题。

----

## v1.1.4-rc1(2021/03/25)

### 改进

- 进程采集器 `message` 字段增加更多信息，便于全文搜索
- 主机对象采集器支持自定义 tag，便于云属性同步

----

## v1.1.4-rc0(2021/03/25)

### 新增功能
- 增加文件采集器、拨测采集器以及HTTP报文采集器
- 内置支持 ActiveMQ/Kafka/RabbitMQ/gin（Gin HTTP访问日志）/Zap（第三方日志框架）日志切割

### 改进
- 丰富 `http://localhost:9529/stats` 页面统计信息，增加诸如采集频率（`n/min`），每次采集的数据量大小等
- DataKit 本身增加一定的缓存空间（重启即失效），避免偶然的网络原因导致数据丢失
- 改进 Pipeline 日期转换函数，提升准确性。另外增加了更多 Pipeline 函数（`parse_duration()/parse_date()`）
- trace 数据增加更多业务字段（`project/env/version/http_method/http_status_code`）
- 其它采集器各种细节改进

----
## v1.1.3-rc4(2021/03/16)

### Bug 修复

- 进程采集器：修复用户名缺失导致显示空白的问题，对用户名获取失败的进程，以 `nobody` 当做其用户名。

----
## v1.1.3-rc3(2021/03/04)

### Bug 修复

- 修复进程采集器部分空字段（进程用户以及进程命令缺失）问题
- 修复 kubernetes 采集器内存占用率计算可能 panic 的问题

<!--
### 新增功能
- `http://datakit:9529/reload` 会自动跳转到 `http://datakit:9529/stats`，便于查看 reload 后 datakit 的运行情况
- `http://datakit:9529/reload` 页面增加每分钟采集频率（`frequency`）以及每次采集的数据量大小统计
- `kubernetes` 指标采集器增加 node 的内存使用率（`mem_usage_percent`）采集 -->

----
## v1.1.3-rc2(2021/03/01)

### Bug 修复

- 修复进程对象采集器 `name` 字段命名问题，以 `hostname + pid` 来命名 `name` 字段
- 修正华为云对象采集器 pipeline 问题
- 修复 Nginx/MySQL/Redis 日志采集器升级后的兼容性问题

----
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

----
## v1.1.2(2021/02/03)

### 功能改进

- 容器安装时，必须注入 `ENV_UUID` 环境变量
- 从旧版本升级后，会自动开启主机采集器（原 datakit.conf 会备份一个）
- 添加缓存功能，当出现网络抖动的情况下，不至于丢失采集到的数据（当长时间网络瘫痪的情况下，数据还是会丢失）
- 所有使用 tailf 采集的日志，必须在 pipeline 中用 `time` 字段来指定切割出来的时间字段，否则日志存入时间字段会跟日志实际时间有出入

### Bug 修复

- 修复 zipkin 中时间单位问题
- 主机对象出采集器中添加 `state` 字段

----
## v1.1.1(2021/02/01)

### Bug 修复
- 修复 mysqlmonitor 采集器 status/variable 字段均为 string 类型的问题。回退至原始字段类型。同时对 int64 溢出问题做了保护。
- 更改进程采集器部分字段命名，使其跟主机采集器命名一致

----
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
