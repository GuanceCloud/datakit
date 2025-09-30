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

本文档介绍如何采集 Kubernetes 集群中自定义 Pod 暴露出来的 Prometheus 指标，有两种方式：

- 通过 Annotations 方式将指标接口暴露给 DataKit
- 通过自动发现 Kubernetes Endpoint Services 到 Prometheus，将指标接口暴露给 DataKit

以下会详细说明两种方式的用法。

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
<!-- markdownlint-enable -->

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
<!-- markdownlint-enable -->

- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotations 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotations，并采集 `url` 上暴露出来的指标。

好的，我来帮您完善和补充这段文档，使其更加清晰、完整和专业。

---

## 自动发现 Pod/Service 的 Prometheus 指标 {#auto-discovery-metrics-with-prometheus}

**注意：此功能的完整文档和最新配置已迁移至 [KubernetesPrometheus 采集器 - “基于 Annotations 的 Prometheus 指标自动发现机制”](kubernetesprometheus.md)。本文档仅保留基础的环境变量配置说明，建议前往新文档查看详细配置示例和最佳实践。**

### 功能概述 {#auto-discovery-metrics-overview}

该功能能够自动发现 Kubernetes 集群中 Pod 或 Service 上的特定 Annotations，并根据注解内容动态生成 Prometheus 指标的采集配置。当 Pod 或 Service 被添加了预定义的注解后，DataKit 会自动拼接 HTTP URL 并创建对应的 Prometheus 指标采集任务，无需手动修改采集器配置。

### 开启方式 {#auto-discovery-metrics-enablement}

此功能默认关闭，需要通过以下环境变量在 DataKit 中启用。这两个环境变量作为全局开关，控制自动发现功能的启用状态：

- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`**：设置为 `"true"` 时，开启基于 Pod Annotations 的自动发现
- **`ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`**：设置为 `"true"` 时，开启基于 Service Annotations 的自动发现

环境变量配置详情请参考 [container 文档](container.md#config-using-env)。

### 功能特点 {#auto-discovery-metrics-features}

- **自动发现**：无需重启 DataKit，自动识别新建或更新的 Pod/Service
- **动态配置**：根据注解内容动态生成采集配置，灵活适应不同应用
- **资源过滤**：支持通过注解值对采集目标进行精确控制
- **配置继承**：Pod 级别的配置优先级高于 Service 级别

### 兼容性说明 {#auto-discovery-metrics-note}

原有的环境变量开启方式完全兼容，现有配置无需修改。但后续的功能增强和配置选项将在 [KubernetesPrometheus 采集器文档](kubernetesprometheus.md) 中更新，建议用户迁移至新的配置方式以获得更完整的功能支持。

## 延伸阅读 {#more-readings}

- [Prometheus Exporter 数据采集](prom.md)
