{{.CSS}}
# Resin

- 文档发布日期：2022-06-12
- 操作系统支持：`Linux/Windows`

## 视图预览

Resin 性能指标展示：启动时间、堆内存、非堆内存、类、线程等。

![](imgs/input-resin-1.png) 

## 前置条件

- 安装 Resin 服务器

说明：示例 Resin 版本为 Windows 环境 Resin/4.0.66 (Windows），指标采集是通过 jolokia-jvm-agent 来采集 Resin 运行时指标。

### 指标采集 (必选)

- 配置 config/resin.properties

新增`jvm_args`，参数说明：

- javaagent：jolokia-jvm-agent
- port=9530 # jolokia-jvm-agent对外暴露的指标端口

```shell
jvm_args: -Xmx2048m -XX:MaxPermSize=256m -javaagent:C:/"Program Files"/datakit/data/jolokia-jvm-agent.jar=port=9530
```

- 重启resin。

双击 resin.exe 

- 开启 Datakit JVM 插件，复制 sample 文件

```shell
cd datakit/conf.d/jvm
cp jvm.conf.sample jvm.conf
```

- 修改 jvm.conf 配置文件

主要参数说明

- urls              : jolokia agent 访问地址
- interval          : 采集频率
- inputs.jvm.metric : jvm相关指标

jvm.conf 配置如下：

```toml
# {"version": "1.2.12", "desc": "do NOT edit this line"}
[[inputs.jvm]]
# default_tag_prefix      = ""
# default_field_prefix    = ""
# default_field_separator = "."

# username = ""
# password = ""
# response_timeout = "5s"

## Optional TLS config
# tls_ca   = "/var/private/ca.pem"
# tls_cert = "/var/private/client.pem"
# tls_key  = "/var/private/client-key.pem"
# insecure_skip_verify = false

## Monitor Intreval
interval   = "10s"

# Add agents URLs to query
urls = ["http://localhost:9530/jolokia/"]

## Add metrics to read
[[inputs.jvm.metric]]
name  = "resin_runtime"
mbean = "java.lang:type=Runtime"
paths = ["Uptime","StartTime","VmVersion","SpecName"]

[[inputs.jvm.metric]]
name  = "resin_memory"
mbean = "java.lang:type=Memory"
paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

[[inputs.jvm.metric]]
name     = "resin_garbage_collector"
mbean    = "java.lang:name=*,type=GarbageCollector"
paths    = ["CollectionTime", "CollectionCount"]
tag_keys = ["name"]

[[inputs.jvm.metric]]
name  = "resin_threading"
mbean = "java.lang:type=Threading"
paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]

[[inputs.jvm.metric]]
name  = "resin_class_loading"
mbean = "java.lang:type=ClassLoading"
paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]

[[inputs.jvm.metric]]
name     = "resin_memory_pool"
mbean    = "java.lang:name=*,type=MemoryPool"
paths    = ["Usage", "PeakUsage", "CollectionUsage"]
tag_keys = ["name"]

[inputs.jvm.tags]
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
```

- 重启 DataKit (如果需要开启日志，请配置日志采集再重启)

```shell
datakit service -R
```

- 查看 resin 采集器运行情况

```shell
datakit monitor -I resin
```

- 指标预览

指标采集上来后，在页面上大概能看到如下这些指标：

![](imgs/input-resin-2.png)

### 日志采集(非必选)

参数说明

- logfiles: 日志文件路径 (通常填写访问日志和错误日志)
- source:  日志来源
- service:  服务名称
- 相关[Pipeline 文档](../datakit/pipeline.md)

```toml
[[inputs.logging]]
## required
logfiles = [
  # 实际 resin 日志目录，视具体情况而定
  "D:/software_installer/resin-4.0.66/log/*.log",
]

# only two protocols are supported:TCP and UDP
# sockets = [
#	 "tcp://0.0.0.0:9530",
#	 "udp://0.0.0.0:9531",
# ]

## glob filteer
ignore = [""]

## your logging source, if it's empty, use 'default'
source = "resin"

## add service tag, if it's empty, use $source.
service = "resin"

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

重启 DataKit：

```shell
datakit service -R
```

日志预览：

![](imgs/input-resin-3.png)


### 插件标签 (非必选）

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 resin 指标都会带有 app = resin-test 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](https://www.yuque.com/dataflux/bp/tag)>
```
# 示例
[inputs.jvm.tags]
   app = "resin-test"
```
重启datakit
```
datakit --restart
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Resin 监控视图>

## 指标集

### `resin_runtime`

| 指标        | 描述                |
| ---         | ---                 |
| `Uptime`    | 在线时长            |
| `StartTime` | 启动时间            |
| `VmVersion` | 虚拟机版本          |
| `SpecName`  | Java 虚拟机规范名称 |


### `resin_memory`

| 指标                 | 描述     |
| ---                  | ---      |
| `HeapMemoryUsage`    | 堆内存   |
| `NonHeapMemoryUsage` | 非堆内存 |

### `resin_threading`

| 指标                      | 描述         |
| ---                       | ---          |
| `TotalStartedThreadCount` | 启动线程总数 |
| `ThreadCount`             | 活动线程数量 |
| `DaemonThreadCount`       | 守护线程数量 |
| `PeakThreadCount`         | 峰值         |

### `resin_class_loading`

| 指标                  | 描述         |
| ---                   | ---          |
| LoadedClassCount      | 已加载当前类 |
| UnloadedClassCount    | 已卸载类总数 |
| TotalLoadedClassCount | 已加载类总数 |

### `resin_memory_pool`

| 指标            | 描述             |
| ---             | ---              |
| Usage           | 已使用内存池     |
| PeakUsage       | 已使用内存池峰值 |
| CollectionUsage | 已使用内存池回收 |


### `resin_garbage_collector`

| 指标            | 描述   |
| ---             | ---    |
| CollectionTime  | GC时间 |
| CollectionCount | GC次数 |

## 更多文档

- [无数据上报排查](../datakit/why-no-data.md)
