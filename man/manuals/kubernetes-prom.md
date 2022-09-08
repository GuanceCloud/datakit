{{.CSS}}
# Prometheus Exportor 指标采集
---

:fontawesome-brands-linux: :material-kubernetes:

---

## 介绍 {#intro}

本文档介绍如何采集 Kubernetes 集群中自定义 Pod 暴露出来的 Prometheus 指标。

需要在 Kubernetes deployment 上添加特定的 template annotations，来采集由其创建的 Pod 暴露出来的指标。Annotation 要求如下：

- Key 为固定的 `datakit/prom.instances`
- Value 为 [prom 采集器](prom.md)完整配置，例如：

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
    # namespace = "$NAMESPACE"
    # pod_name = "$PODNAME"
    # node_name = "$NODENAME"
```

其中支持如下几个通配符：

- `$IP`：通配 Pod 的内网 IP
- `$NAMESPACE`：Pod Namespace
- `$PODNAME`：Pod Name
- `$NODENAME`：Pod 所在的 Node 名称

!!! tip

    Prom 采集器不会自动添加诸如 `namespace` 和 `pod_name` 等 tags，可以在上面的 config 中使用通配符添加额外 tags，例如：

    ``` toml
      [inputs.prom.tags]
        namespace = "$NAMESPACE"
        pod_name = "$PODNAME"
        node_name = "$NODENAME"
    ```

### 选择指定 Pod IP {#pod-ip}

某些情况下， Pod 上会存在多个 IP，此时仅仅通过 `$IP` 来获取 Exporter 地址是不准确的。支持通过配置 Annotation 选择 Pod IP。

- Key 为固定的 `datakit/prom.instances.ip_index`
- Value 是自然数，例如 `0` `1` `2` 等，是要使用的 IP 在整个 IP 数组（Pod IPs）中的位置下标。

如果没有此 Annotation Key，则使用默认 Pod IP。

## 操作步骤 {#steps}

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
        datakit/prom.instances.ip_index: 2
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
          
            [inputs.prom.tags] # 视情况开启下面的 Tags
            #namespace = "$NAMESPACE"
            #pod_name = "$PODNAME"
            #node_name = "$NODENAME"
```

???+ attention

    `annotations` 一定添加在 `template` 字段下，这样 *deployment.yaml* 创建的 Pod 才会携带 `datakit/prom.instances`。


- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotation 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotation，并采集 `url` 上暴露出来的指标。

## 延伸阅读 {#more-readings}

- [Prometheus Exportor 数据采集](prom.md)
