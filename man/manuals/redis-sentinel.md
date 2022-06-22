# Redis Sentinel
---

## 视图预览
Redis-sentinel 观测场景主要展示了 Redis 的集群、slaves、节点分布信息等。
![](imgs/input-redis-sentinel-1.png)

## 版本支持
操作系统：Linux / Windows
redis-sentinel-exporter >=0.1
## 前置条件

- 在 Redis 应用服务器上安装 DataKit <[安装 DataKit](/datakit/datakit-install/)>
## 安装部署
### 下载 redis-sentinel-exporter 指标采集器
下载地址 [https://github.com/lrwh/redis-sentinel-exporter/releases](https://github.com/lrwh/redis-sentinel-exporter/releases)
![](imgs/input-redis-sentinel-2.png)
### 启动 redis-sentinel-exporter 
```bash
java -Xmx64m -jar redis-sentinel-exporter-0.1.jar --spring.redis.sentinel.master=mymaster --spring.redis.sentinel.nodes="127.0.0.1:26379,127.0.0.1:26380,127.0.0.1:26381"
```
参数说明
spring.redis.sentinel.master ： 集群名称
spring.redis.sentinel.nodes ： 哨兵节点地址
### 配置实施
#### 指标采集 (必选)

1. 开启 Datakit prom 插件，复制 sample 文件
```bash
cd /usr/local/datakit/conf.d/prom/
cp prom.conf.sample redis-sentinel-prom.conf
```

2. 修改 `redis-sentinel-prom.conf` 配置文件
```toml
# {"version": "1.2.12", "desc": "do NOT edit this line"}

[[inputs.prom]]
## Exporter URLs
urls = ["http://localhost:6390/metrics"]

## 忽略对 url 的请求错误
ignore_req_err = false

## 采集器别名
source = "redis_sentinel"

## 采集数据输出源
# 配置此项，可以将采集到的数据写到本地文件而不将数据打到中心
# 之后可以直接用 datakit --prom-conf /path/to/this/conf 命令对本地保存的指标集进行调试
# 如果已经将 url 配置为本地文件路径，则 --prom-conf 优先调试 output 路径的数据
# output = "/abs/path/to/file"

## 采集数据大小上限，单位为字节
# 将数据输出到本地文件时，可以设置采集数据大小上限
# 如果采集数据的大小超过了此上限，则采集的数据将被丢弃
# 采集数据大小上限默认设置为32MB
# max_file_size = 0

## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
# 默认只采集 counter 和 gauge 类型的指标
# 如果为空，则不进行过滤
metric_types = []

## 指标名称过滤
# 支持正则，可以配置多个，即满足其中之一即可
# 如果为空，则不进行过滤
# metric_name_filter = ["cpu"]

## 指标集名称前缀
# 配置此项，可以给指标集名称添加前缀
# measurement_prefix = "redis_sentinel_"

## 指标集名称
# 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
# 如果配置measurement_name, 则不进行指标名称的切割
# 最终的指标集名称会添加上measurement_prefix前缀
measurement_name = "redis_sentinel"

## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
interval = "10s"

## 过滤tags, 可配置多个tag
# 匹配的tag将被忽略
# tags_ignore = ["xxxx"]

## TLS 配置
tls_open = false
# tls_ca = "/tmp/ca.crt"
# tls_cert = "/tmp/peer.crt"
# tls_key = "/tmp/peer.key"

## 自定义认证方式，目前仅支持 Bearer Token
# token 和 token_file: 仅需配置其中一项即可
# [inputs.prom.auth]
# type = "bearer_token"
# token = "xxxxxxxx"
# token_file = "/tmp/token"

## 自定义指标集名称
# 可以将包含前缀prefix的指标归为一类指标集
# 自定义指标集名称配置优先measurement_name配置项
#[[inputs.prom.measurements]]
#  prefix = "cpu_"
#  name = "cpu"

# [[inputs.prom.measurements]]
# prefix = "mem_"
# name = "mem"

## 自定义Tags
[inputs.prom.tags]
# some_tag = "some_value"
  # more_tag = "some_other_value"
```
主要参数说明

- urls：promethues 指标地址，这里填写 redis-sentinel-exporter 暴露出来的指标 url
- source：采集器别名
- interval：采集间隔
- measurement_prefix：指标前缀，便于指标分类查询
- tls_open：TLS 配置
- metric_types：指标类型，不填，代表采集所有指标
- [inputs.prom.tags]：额外定义的 tag

3. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```bash
systemctl restart datakit
```

4. redis-sentinel 指标采集验证，使用命令 /usr/local/datakit/datakit -M |egrep "最近采集|redis-sentinel" 或者通过 url 查看 ${ip}:9529/monitor

![](imgs/input-redis-sentinel-3.png)

5. 指标预览

![](imgs/input-redis-sentinel-4.png)
#### 日志采集 (非必选)

1. 修改 `redis.conf` 配置文件

参数说明

- files：日志文件路径 (通常填写访问日志和错误日志)
- ignore：要过滤的文件名
- pipeline：日志切割文件
- character_encoding：日志编码格式
- match：开启多行日志收集
- 相关文档 <[DataFlux pipeline 文本数据处理](/datakit/pipeline/)>
```
# {"version": "1.2.12", "desc": "do NOT edit this line"}

[[inputs.logging]]
  ## required
  logfiles = [
    "D:/software_installer/Redis-x64-3.2.100/log/sentinel_*_log.log",
  ]
  # only two protocols are supported:TCP and UDP
  # sockets = [
  #	 "tcp://0.0.0.0:9530",
  #	 "udp://0.0.0.0:9531",
  # ]
  ## glob filteer
  ignore = [""]

  ## your logging source, if it's empty, use 'default'
  source = "redis-sentinel"

  ## add service tag, if it's empty, use $source.
  service = "redis-sentinel"

  ## grok pipeline script name
  pipeline = ""

  ## optional status:
  ##   "emerg","alert","critical","error","warning","info","debug","OK"
  ignore_status = []

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  character_encoding = ""

  ## datakit read text from Files or Socket , default max_textline is 32k
  ## If your log text line exceeds 32Kb, please configure the length of your text, 
  ## but the maximum length cannot exceed 32Mb 
  # maximum_length = 32766

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  ## removes ANSI escape codes from text strings
  remove_ansi_escape_codes = false

  ## if file is inactive, it is ignored
  ## time units are "ms", "s", "m", "h"
  # ignore_dead_log = "1h"

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```

3. 重启 Datakit (如果需要开启自定义标签，请配置插件标签再重启)
```
systemctl restart datakit
```

4. redis-sentinel 指标采集验证，使用命令 /usr/local/datakit/datakit -M |egrep "最近采集|logging" 或者通过 url 查看 ${ip}:9529/monitor

![](imgs/input-redis-sentinel-5.png)

5. 日志预览

![](imgs/input-redis-sentinel-6.png)

6. 日志 pipeline 功能切割字段说明
- Redis 通用日志切割
原始日志为

```
[11412] 05 May 10:17:31.329 # Creating Server TCP listening socket *:26380: bind: No such file or directory
```

![](imgs/input-redis-sentinel-7.png)
![](imgs/input-redis-sentinel-8.png)
切割后的字段列表如下：

| 字段名 | 字段值 | 说明 |
| --- | --- | --- |
| `pid` | `122` | 进程id |
| `role` | `M` | 角色 |
| `service` | `*` | 服务 |
| `status` | `notice` | 日志级别 |
| `message` | `Creating Server TCP listening socket *:26380: bind: No such file or directory` | 日志内容 |
| `time` | `1557861100164000000` | 纳秒时间戳（作为行协议时间） |

#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 `redis-sentinel` 指标都会带有`service = "redis-sentinel"`的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.prom.tags]
		service = "redis-sentinel"
    # some_tag = "some_value"
    # more_tag = "some_other_value"
```
重启 Datakit
```
systemctl restart datakit
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Redis-sentinel 监控视图>
## 异常检测
无
## 指标详解
| 指标 | 含义 | 类型 |
| --- | --- | --- |
| redis_sentinel_known_sentinels | 哨兵实例数 | Gauge |
| redis_sentinel_known_slaves | 集群slaves实例数 | Gauge |
| redis_sentinel_cluster_type | 集群节点类型 | Gauge |
| redis_sentinel_link_pending_commands | 哨兵挂起命令数 | Gauge |
| redis_sentinel_odown_slaves | slave客观宕机 | Gauge |
| redis_sentinel_sdown_slaves | slave主观宕机 | Gauge |
| redis_sentinel_ok_slaves | 正在运行的slave数 | Gauge |
| redis_sentinel_ping_latency | 哨兵ping的延迟显示为毫秒 | Gauge |
| redis_sentinel_last_ok_ping_latency | 哨兵ping成功的秒数 | Gauge |

## 最佳实践
无
## 故障排查
<[无数据上报排查](/datakit/why-no-data/)>
