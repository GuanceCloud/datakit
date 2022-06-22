# seata
---

## 视图预览

![](../imgs/seata-1.png)

## 版本支持

操作系统：Linux / Windows

## 前置条件

- [安装 Datakit](datakit-install.md)

## 部署实施

( Linux / Windows 环境相同)

### 1:原理

Seata支持在TC开启Metrics数据采集并输出到Prometheus监控系统中。

### 2:指标采集 

1、 配置在 TC 中配置开启 Metrics，打开 TC 中 Metrics 的配置项，Seata Server 已经包含了 metrics(seata-metrics-all)依赖, 但是默认是关闭状态，需要开启 metrics 的采集配置。（**这一步非常重要）**

`Seata 1.5.0+`中使用application.yaml

```
seata:
  metrics:
    enabled: true
    registry-type: compact
    exporter-list: prometheus
    exporter-prometheus-port: 9898

```

`Seata <1.5.0 `开启 metrics 

github地址配置：[https://github.com/seata/seata/blob/1.4.2/server/src/main/resources/file.conf.example](https://github.com/seata/seata/blob/1.4.2/server/src/main/resources/file.conf.example)

服务端配置metrics

【file.conf:】

```
metrics {
  enabled = true
  registryType = "compact"
  # multi exporters use comma divided
  exporterList = "prometheus"
  exporterPrometheusPort = 9898
}
```

【registry.conf :】

```
registry {
    type = "nacos"
    nacos {
      application = "应用名"
      serverAddr = "xxxx:端口"
      group = "分组"
      namespace = "命名空间"
      cluster = "default"
      username = "按照实际填写"
      password = "按照实际填写"
    }
}
config {
  type = "file"
  file {
    name="file:/root/seata-config/file.conf"
  }
}
```

2、 访问 [http://tc-server-ip:9898/metrics](http://tc-server-ip:9898/metrics)，看是否能访问到 metrics 数据

```
# HELP seata seata
# TYPE seata untyped
seata_transaction{meter="counter",role="tc",status="committed",} 1358.0 1551946035372
seata_transaction{meter="counter",role="tc",status="active",} 0.0 1551946035372
seata_transaction{meter="summary",role="tc",statistic="count",status="committed",} 6.0 1551946035372
seata_transaction{meter="summary",role="tc",statistic="total",status="committed",} 6.0 1551946035372
seata_transaction{meter="summary",role="tc",statistic="tps",status="committed",} 1.6163793103448276 1551946035372
seata_transaction{meter="timer",role="tc",statistic="count",status="committed",} 6.0 1551946035372
seata_transaction{meter="timer",role="tc",statistic="total",status="committed",} 910.0 1551946035372
seata_transaction{meter="timer",role="tc",statistic="max",status="committed",} 164.0 1551946035372
seata_transaction{meter="timer",role="tc",statistic="average",status="committed",} 151.66666666666666 1551946035372
```

得到以上类似数据证明mertric开启成功。（如果某些Transaction状态没有发生，例如rollback，那么对应的Metrics指标也不会存在（输出））

3、 开启 Datakit promtheus 插件，创建seata-prom.conf

```shell
cd /usr/local/datakit/conf.d/prom/
cp prom.conf.sample seata-prom.conf
```

4、 修改 /usr/local/datakit/conf.d/prom/seata-prom.conf 配置文件

```
[[inputs.prom]]
  urls = ["http://ip:9898/metrics"]
  ## 忽略对 url 的请求错误
  ignore_req_err = false
  ## 采集器别名
  source = "资源名称"
  metric_types = []
  ## 指标过滤，只收集 seata_相关指标
  metric_name_filter = ["seata_*"]
  measurement_prefix = "seata_"
  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"
  ## TLS 配置
  tls_open = false
  ## 自定义Tags
  [inputs.prom.tags]
  app = "自定义app名字"
  # more_tag = "some_other_value"
```

5.重启 Datakit

```shell
systemctl restart datakit
```


### 3:指标预览

![image.png](../imgs/seata-2.png)

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Seata>

## 指标详解

| **Metrics** | **描述** |
| --- | --- |
| seata.transaction(role=tc,meter=counter,status=active/committed/rollback) | 当前活动中/已提交/已回滚的事务总数 |
| seata.transaction(role=tc,meter=summary,statistic=count,status=committed/rollback) | 当前周期内提交/回滚的事务数 |
| seata.transaction(role=tc,meter=summary,statistic=tps,status=committed/rollback) | 当前周期内提交/回滚的事务TPS(transaction per second) |
| seata.transaction(role=tc,meter=timer,statistic=total,status=committed/rollback) | 当前周期内提交/回滚的事务耗时总和 |
| seata.transaction(role=tc,meter=timer,statistic=count,status=committed/rollback) | 当前周期内提交/回滚的事务数 |
| seata.transaction(role=tc,meter=timer,statistic=average,status=committed/rollback) | 当前周期内提交/回滚的事务平均耗时 |
| seata.transaction(role=tc,meter=timer,statistic=max,status=committed/rollback) | 当前周期内提交/回滚的事务最大耗时 |


## 常见问题排查

<[无数据上报排查](why-no-data.md)>
