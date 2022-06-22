# Nacos
---

## 视图预览
Nacos 性能指标展示：Nacos 在线时长、Nacos config 长链接数、Nacos config 配置个数、Service Count、http请求次数等。
![image.png](imgs/input-nacos-1.png)
## 版本支持
操作系统：Linux / Windows
Nacos 版本：>= 0.8.0
## 前置条件

- Nacos 服务器 <[安装 Datakit](/datakit/datakit-install/)>
## 安装配置
说明：示例 Nacos 版本为 1.4.1。
### 部署实施
( Linux / Windows 环境相同)
#### 指标采集 (必选)

1. 配置 application.properties 文件，暴露 metrics 数据
```
management.endpoints.web.exposure.include=*
```

2. 重启 Nacos

集群方式和单例模式启动参数有差异，具体参考 [Nacos 官方文档](https://nacos.io/zh-cn/docs/quick-start.html)。

3. 校验

访问 {ip}:8848/nacos/actuator/prometheus，看是否能访问到 metrics 数据

4. 开启 Datakit promtheus 插件
```shell
cd /usr/local/datakit/conf.d/prom/
cp prom.conf.sample nacos-prom.conf
```

5. 修改 nacos-prom.conf 配置文件

主要参数说明

- urls：promethues 指标地址，这里填写 Nacos 暴露出来的指标 url
- source：采集器别名，建议写成`nacos`
- interval：采集间隔
- measurement_prefix：指标前缀，便于指标分类查询
- tls_open：TLS 配置
- metric_types：指标类型，不填，代表采集所有指标
- [inputs.prom.tags]：额外定义的 tag
```
[[inputs.prom]]
  urls = ["http://192.168.0.189:8848/nacos/actuator/prometheus"]
  ## 忽略对 url 的请求错误
  ignore_req_err = false
  ## 采集器别名
  source = "nacos"
  metric_types = []
  measurement_prefix = "nacos_"
  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"
  ## TLS 配置
  tls_open = false
  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

6. 重启 Datakit
```shell
systemctl restart datakit
```

7. Nacos 指标采集验证，使用命令 /usr/local/datakit/datakit -M |egrep "最近采集|nacos" 或者通过 url 查看 ${ip}:9529/monitor

![image.png](imgs/input-nacos-2.png)
指标预览
![image.png](imgs/input-nacos-3.png)
#### 
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Nacos 监控视图>
## 指标详解

### JVM metrics
| 指标 | 含义 |
| --- | --- |
| system_cpu_usage | CPU使用率 |
| system_load_average_1m | load |
| jvm_memory_used_bytes | 内存使用字节，包含各种内存区 |
| jvm_memory_max_bytes | 内存最大字节，包含各种内存区 |
| jvm_gc_pause_seconds_count | gc次数，包含各种gc |
| jvm_gc_pause_seconds_sum | gc耗时，包含各种gc |
| jvm_threads_daemon | 线程数 |

### Nacos 监控指标
| 指标 | 含义 |
| --- | --- |
| http_server_requests_seconds_count | http请求次数，包括多种(url,方法,code) |
| http_server_requests_seconds_sum | http请求总耗时，包括多种(url,方法,code) |
| nacos_timer_seconds_sum | Nacos config水平通知耗时 |
| nacos_timer_seconds_count | Nacos config水平通知次数 |
| nacos_monitor{name='longPolling'} | Nacos config长连接数 |
| nacos_monitor{name='configCount'} | Nacos config配置个数 |
| nacos_monitor{name='dumpTask'} | Nacos config配置落盘任务堆积数 |
| nacos_monitor{name='notifyTask'} | Nacos config配置水平通知任务堆积数 |
| nacos_monitor{name='getConfig'} | Nacos config读配置统计数 |
| nacos_monitor{name='publish'} | Nacos config写配置统计数 |
| nacos_monitor{name='ipCount'} | Nacos naming ip个数 |
| nacos_monitor{name='domCount'} | Nacos naming域名个数(1.x 版本) |
| nacos_monitor{name='serviceCount'} | Nacos naming域名个数(2.x 版本) |
| nacos_monitor{name='failedPush'} | Nacos naming推送失败数 |
| nacos_monitor{name='avgPushCost'} | Nacos naming平均推送耗时 |
| nacos_monitor{name='leaderStatus'} | Nacos naming角色状态 |
| nacos_monitor{name='maxPushCost'} | Nacos naming最大推送耗时 |
| nacos_monitor{name='mysqlhealthCheck'} | Nacos naming mysql健康检查次数 |
| nacos_monitor{name='httpHealthCheck'} | Nacos naming http健康检查次数 |
| nacos_monitor{name='tcpHealthCheck'} | Nacos naming tcp健康检查次数 |

### nacos 异常指标
| 指标 | 含义 |
| --- | --- |
| nacos_exception_total{name='db'} | 数据库异常 |
| nacos_exception_total{name='configNotify'} | Nacos config水平通知失败 |
| nacos_exception_total{name='unhealth'} | Nacos config server之间健康检查异常 |
| nacos_exception_total{name='disk'} | Nacos naming写磁盘异常 |
| nacos_exception_total{name='leaderSendBeatFailed'} | Nacos naming leader发送心跳异常 |
| nacos_exception_total{name='illegalArgument'} | 请求参数不合法 |
| nacos_exception_total{name='nacos'} | Nacos请求响应内部错误异常（读写失败，没权限，参数错误） |

更多nacos指标，可以参考 [Nacos官方网站-监控](https://nacos.io/zh-cn/docs/monitor-guide.html)
## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>

