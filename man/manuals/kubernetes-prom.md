{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Kubernetes 集群中自定义 Exporter 指标采集

## 介绍

该方案可以在 Kubernetes 集群中通过配置，采集集群中的自定义的 Pod 的 Exporter 数据。

目前只支持 Promtheus 格式的数据。

通过在 Kubernetes Pod 添加指定的 Annotation，实现 Exporter 功能。Annotation 格式内容如下：

- Key 固定为 `datakit/prom.instances`
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

配置文件支持通配符：

- `$IP`：通配 Pod 的内网 IP，形如 `172.16.0.3`，无需额外配置
- `$NAMESPACE`：Pod Namespace
- `$PODNAME`：Pod Name

## 操作过程

假设 Pod 名称为 `dummy-abc`

- 登录到 Kubernetes 所在主机

- 打开 `dummy-abc.yaml`，添加 Annotation 规范如下：

```
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
kubectl apply -f dummy-abc.yaml
```

至此，Annotation 已经添加完成。

DataKit Kubernetes 会自动忽略相同配置（通配符替换之后的配置），避免重复采集。
