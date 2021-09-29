{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes 集群中自定义 Exporter 指标采集

## 介绍

本文档介绍如何采集 Kubernetes 集群中自定义 Pod 暴露出来的 Prometheus 指标。

可以在 Kubernetes Pod 上添加特定的 Annotation 来采集其暴露出来的指标。Annotation 要求如下：

- Key 为 `datakit/prom.instances`
- Value 为 [prom 采集器](prom)完整配置，例如：

```toml
[[inputs.prom]]
  ## Exporter 地址
  url = "http://$IP:9100/metrics"

  source = "<your-service-name>"
  metric_types = ["counter", "gauge"]
  # metric_name_filter = ["cpu"]
  # measurement_prefix = ""
  # measurement_name = "prom"

  interval = "10s"

  #tags_ignore = ["xxxx"]

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

假设 Pod 名称为 `prom-example`

- 登录到 Kubernetes 所在主机

- 打开 `prom-example.yaml`，添加 Annotation 规范如下：

```yaml
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

- 使用新的 yaml 创建资源

```shell
kubectl apply -f prom-example.yaml
```

至此，Annotation 已经添加完成。DataKit 稍后会读取到这个 Annotation，并采集 `url` 上暴露出来的指标。
