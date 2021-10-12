{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes 集群中自定义 Exporter 指标采集

## 介绍

本文档介绍如何采集 Kubernetes 集群中自定义 Pod 暴露出来的 Prometheus 指标。

需要在 Kubernetes deployment 上添加特定的 template annotations，来采集由其创建的 Pod 暴露出来的指标。Annotation 要求如下：

- Key 为固定的 `datakit/prom.instances`
- Value 为 [prom 采集器](prom)完整配置，例如：

```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://$IP:9100/metrics"

  source = "<your-service-name>"
  metric_types = ["counter", "gauge"]

  measurement_name = "prom"
  # metric_name_filter = ["cpu"]
  # measurement_prefix = ""
  #tags_ignore = ["xxxx"]

  interval = "10s"

  #[[inputs.prom.measurements]]
  # prefix = "cpu_"
  # name = "cpu"

  [inputs.prom.tags]
  namespace = "$NAMESPACE"
  pod_name = "$PODNAME"
```

其中支持如下几个通配符：

- `$IP`：通配 Pod 的内网 IP
- `$NAMESPACE`：Pod Namespace
- `$PODNAME`：Pod Name

## 操作过程

- 登录到 Kubernetes 所在主机

- 打开 `deployment.yaml`，添加 template annotations 示例如下：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prom-deployment
  labels:
    app: prom
spec:
  template:
    metadata:
      labels:
        app: prom
      annotations:
        datakit/prom.instances: |
          [[inputs.prom]]
            url = "http://$IP:9100/metrics"
          
            source = "<your-service-name>"
            metric_types = ["counter", "gauge"]
            # metric_name_filter = ["cpu"]
            # measurement_prefix = ""
            # measurement_name = "prom"
          
            interval = "10s"
          
            # tags_ignore = ["xxxx"]
          
            #[[inputs.prom.measurements]]
            # prefix = "cpu_"
            # name = "cpu"
          
            [inputs.prom.tags]
            namespace = "$NAMESPACE"
            pod_name = "$PODNAME"
```

> 注意， `annotations` 一定添加在 `template` 字段下，这样 deployment.yaml 创建的 Pod 才会携带 `datakit/prom.instances`。

- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotation 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotation，并采集 `url` 上暴露出来的指标。
