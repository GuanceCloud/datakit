# Ingress Nginx (Prometheus) 
---

## 视图预览

Ingress性能指标展示：Ingress Controller的平均cpu使用率、平均内存使用、网络请求/响应合计、Ingress Config的加载次数、Ingress Config上次加载结果、Ingress的转发成功率等。

![](../imgs/ingress-nginx-prom-1.png)

## 安装部署

说明：示例 Ingress 版本为：willdockerhub/ingress-nginx-controller:v1.0.0(CentOS环境下kubeadmin部署)，各个不同版本指标可能存在差异

### 前置条件

- 登录[观测云](https://console.guance.com/)， 【集成】->【Datakit】-> 【Kubernetes】。

### 配置实施

#### 指标采集 (必选)

1、 获取部署Ingress的yaml文件

```
wget https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.0.0/deploy/static/provider/baremetal/deploy.yaml
```

2、 编辑deploy.yaml，把service的type设置成NodePort，并对外暴露10254端口，参考下图
```
vi deploy.yaml
```
![](../imgs/ingress-nginx-prom-2.png)

3、 开启Input

观测云接入Ingress指标数据，需要datakit开启prom插件，在prom插件配置中指定exporter的url，在kubernetes集群中采集Ingress Controller指标，推荐使用annotations增加注解的方式。编辑deploy.yaml文件，找到ingress-nginx-controller镜像所对应的Deployment ，增加annotations。

```
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:10254/metrics"
            source = "prom-ingress"
            metric_types = ["counter", "gauge"]
            # metric_name_filter = ["cpu"]
            # measurement_prefix = ""
            measurement_name = "prom_ingress"
            interval = "30s"
            tags_ignore = ["build","le","method","release","repository"]
            # tags_ignore = ["xxxx"]
            [[inputs.prom.measurements]]
              prefix = "nginx_ingress_controller_"
              name = "prom_ingress"
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
            
```

参数说明

- url:   Exporter URLs，多个url用逗号分割，示例["[http://127.0.0.1:9100/metrics",](http://127.0.0.1:9100/metrics",) "[http://127.0.0.1:9200/metrics"]](http://127.0.0.1:9200/metrics"])

- source:  采集器别名
- metric_types:  指标类型，可选值是counter, gauge, histogram, summary
- measurement_name:  指标集名称
- interval: 采集频率
- inputs.prom.measurements: 指标集为prefix的前缀归为name的指标集
- tags_ignore:  忽略的 tag。

4、 部署Ingress

```
kubectl apply -f deploy.yaml
```

指标预览

![](../imgs/ingress-nginx-prom-3.png)

## 场景视图

场景 - 新建仪表板 - Nginx Ingress Controller


## 异常检测

暂无

## 指标详解
如果配置了inputs.prom.measurements，观测云采集到的指标需要加上前缀才能与表格匹配。举例，下配置了前缀nginx_ingress_controller_，指标集是prom_ingress。
```
 [[inputs.prom.measurements]]
              prefix = "nginx_ingress_controller_"
              name = "prom_ingress"
```
nginx_ingress_controller_requests指标在观测云上的指标就是prom_ingress指标集下的requests指标。

| 指标 | 描述 | 数据类型 | 单位 |
| --- | --- | --- | --- |
| nginx_ingress_controller_requests | The total number of client requests | int | count |
| nginx_ingress_controller_nginx_process_connections | current number of client connections with state {active, reading, writing, waiting} | int | count |
| nginx_ingress_controller_success | Cumulative number of Ingress controller reload operations | int | count |
| nginx_ingress_controller_config_last_reload_successful | Whether the last configuration reload attempt was successful | int | count |
| nginx_ingress_controller_nginx_process_resident_memory_bytes | number of bytes of memory in use | float | B |
| nginx_ingress_controller_nginx_process_cpu_seconds_total | Cpu usage in seconds | float | B |
| nginx_process_resident_memory_bytes | number of bytes of memory in use | int | B |
| nginx_ingress_controller_request_duration_seconds_bucket | The request processing time in milliseconds | int | count |
| nginx_ingress_controller_request_size_sum | The request length (including request line, header, and request body) | int | count |
| nginx_ingress_controller_response_size_sum | The response length (including request line, header, and request body) | int | count |
| nginx_ingress_controller_ssl_expire_time_seconds | Number of seconds since 1970 to the SSL Certificate expire | int | count |


## 最佳实践

[Nginx Ingress可观测最佳实践](/best-practices/integrations/ingress-nginx.md)

## 故障排查

- [无数据上报排查](why-no-data.md)
