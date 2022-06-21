# Kubernetes Scheduler
---


## 视图预览
Scheduler 性能指标展示：调度队列 pending pod 数、进入调度队列 pod 速率、http 请求数量、CPU、Memory、Goroutines等。<br />![1651892867(1).png](imgs/input-kube-scheduler-01.png)<br />
![1651892881(1).png](imgs/input-kube-scheduler-02.png)

## 版本支持
操作系统：Linux<br />Kubernetes 版本：1.18+

## 前置条件

- Kubernetes 集群 <[安装 Datakit](https://www.yuque.com/dataflux/integrations/kubernetes)>

## 安装配置
说明：示例 Kubernetes 版本为 1.22.6，DataKit 版本为 1.2.17，各个不同版本指标可能存在差异。

### 部署实施

#### 指标采集 (必选)

1. ConfigMap 增加 scheduler.conf 配置

在部署 DataKit 使用的 datakit.yaml 文件中，ConfigMap 资源中增加 scheduler.conf。
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:    
    #### scheduler ##下面是新增部分
    scheduler.conf: |-    
        [[inputs.prom]]
          ## Exporter地址或者文件路径（Exporter地址要加上网络协议http或者https）
          ## 文件路径各个操作系统下不同
          ## Windows example: C:\\Users
          ## UNIX-like example: /usr/local/
          urls = ["https://172.16.0.229:10259/metrics"]

          ## 采集器别名
          source = "prom-scheduler"

          ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
          # 默认只采集 counter 和 gauge 类型的指标
          # 如果为空，则不进行过滤
          metric_types = ["counter", "gauge"]

          ## 指标名称过滤
          # 支持正则，可以配置多个，即满足其中之一即可
          # 如果为空，则不进行过滤
          #metric_name_filter = [""]

          ## 指标集名称前缀
          # 配置此项，可以给指标集名称添加前缀
          measurement_prefix = ""

          ## 指标集名称
          # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
          # 如果配置measurement_name, 则不进行指标名称的切割
          # 最终的指标集名称会添加上measurement_prefix前缀
          measurement_name = "prom_scheduler"

          ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
          interval = "10s"

          ## 过滤tags, 可配置多个tag
          # 匹配的tag将被忽略
          # tags_ignore = ["xxxx"]

          ## TLS 配置
          tls_open = true
          #tls_ca = ""
          #tls_cert = ""
          #tls_key = ""

          ## 自定义指标集名称
          # 可以将包含前缀prefix的指标归为一类指标集
          # 自定义指标集名称配置优先measurement_name配置项
          #[[inputs.prom.measurements]]
          #  prefix = ""
          #  name = ""

          ## 自定义认证方式，目前仅支持 Bearer Token
          [inputs.prom.auth]
           type = "bearer_token"
          # token = "xxxxxxxx"
           token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"

          ## 自定义Tags
           ## 自定义Tags
          [inputs.prom.tags]
            instance = "172.16.0.229:10259"   
```
参数说明：
- urls：scheduler metrics 地址
- source：采集器别名
- metric_types：指标类型过滤
- metric_name_filter：指标名称过滤
- measurement_prefix：指标集名称前缀
- measurement_name：指标集名称
- interval：采集间隔
- tls_open：是否忽略安全验证 (如果是 https，请设置为 true，并设置相应证书)，此处为 true
- tls_ca：ca 证书路径
- type：自定义认证方式，scheduler 使用 bearer_token 认证
-  token_file：认证文件路径
- [inputs.prom.tags]：请参考插件标签


2. 挂载 scheduler.conf

在 datakit.yaml 文件的 volumeMounts 下面增加下面内容。
```
        - mountPath: /usr/local/datakit/conf.d/prom/scheduler.conf
          name: datakit-conf
          subPath: scheduler.conf  
```


3. 重启 DataKit  
```
kubectl delete -f datakit.yaml
kubectl apply -f datakit.yaml
```

指标预览<br />![1651892600(1).png](imgs/input-kube-scheduler-03.png)

#### 插件标签 (必选）
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，scheduler 指标都会带有 app = oa 的标签，可以进行快速查询
- 采集 scheduler  指标，必填的 key 是 instance，值是 scheduler  metrics 的 ip + 端口
- 相关文档 <[DataFlux Tag 应用最佳实践](https://www.yuque.com/dataflux/bp/tag)>

```
           ## 自定义Tags
          [inputs.prom.tags]
            instance = "172.16.0.229:10259"   
```
   <br />如果增加了自定义 tag，重启 Datakit 。
```
kubectl delete -f datakit.yaml
kubectl apply -f datakit.yaml
```

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Kubernetes Scheduler 监控视图>

## 指标详解
| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| scheduler_pending_pods | Number of pending pods, by the queue type. | int | <br /> |
| scheduler_queue_incoming_pods_total | Number of pods added to scheduling queues by event and queue type. | int | <br /> |
| rest_client_requests_total | Number of HTTP requests, partitioned by status code, method, and host. | int | <br /> |
| process_resident_memory_bytes | Resident memory size in bytes. | B | <br /> |
| process_cpu_seconds_total | Total user and system CPU time spent in seconds. | float | <br /> |
| go_goroutines | Number of goroutines that currently exist. | int | <br /> |


## 常见问题排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>

## 进一步阅读
暂无

