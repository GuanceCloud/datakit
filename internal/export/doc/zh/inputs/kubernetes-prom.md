---
title     : 'Kubernetes Prometheus Exporter'
summary   : '采集 Kubernetes 集群中自定义 Pod 暴露出来的 Prometheus 指标'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

:fontawesome-brands-linux: :material-kubernetes:

---

## 介绍 {#intro}

**已废弃，相关功能移动到 [KubernetesPrometheus 采集器](kubernetesprometheus.md)。**

本页保留 `datakit/prom.instances` 的旧版配置示例及自动发现功能的环境变量兼容说明。

## 使用 Annotations 开放指标接口 {#annotations-of-prometheus}

需要在 Kubernetes deployment 上添加特定的 template annotations，来采集由其创建的 Pod 暴露出来的指标。Annotations 要求如下：

- Key 为固定的 `datakit/prom.instances`
- Value 为 [prom 采集器](prom.md)完整配置，例如：

```toml
[[inputs.prom]]
  urls   = ["http://$IP:9100/metrics"]
  source = "<your-service-name>"
  measurement_name = "<measurement-metrics>"
  interval = "30s"

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

<!-- markdownlint-disable MD046 -->
!!! tip

    Prom 采集器不会自动添加诸如 `namespace` 和 `pod_name` 等 tags，可以在上面的 config 中使用通配符添加额外 tags，例如：

    ``` toml
      [inputs.prom.tags]
        namespace = "$NAMESPACE"
        pod_name = "$PODNAME"
        node_name = "$NODENAME"
    ```
<!-- markdownlint-enable MD046 -->

### 操作步骤 {#annotations-of-prometheus-steps}

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
            urls   = ["http://$IP:9100/metrics"]
            source = "<your-service-name>"
            interval = "30s"
            [inputs.prom.tags]
              namespace = "$NAMESPACE"
              pod_name  = "$PODNAME"
              node_name = "$NODENAME"
```

<!-- markdownlint-disable MD046 -->
???+ note

    `annotations` 一定添加在 `template` 字段下，这样 *deployment.yaml* 创建的 Pod 才会携带 `datakit/prom.instances`。
<!-- markdownlint-enable MD046 -->

- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotations 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotations，并采集 `url` 上暴露出来的指标。

## 自动发现 Pod/Service 的 Prometheus 指标 {#auto-discovery-metrics-with-prometheus}

配置方法和注解参数见 [KubernetesPrometheus 采集器](kubernetesprometheus.md#auto-discovery-metrics-with-prometheus)。原有环境变量仍然兼容，已有配置无需修改。以下开关默认关闭：

- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`**：设置为 `"true"` 时，开启基于 Pod Annotations 的自动发现
- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`**：设置为 `"true"` 时，开启基于 Service Annotations 的自动发现

环境变量配置详情请参考 [container 文档](container.md#config-using-env)。

## 延伸阅读 {#more-readings}

- [Prometheus Exporter 数据采集](prom.md)
