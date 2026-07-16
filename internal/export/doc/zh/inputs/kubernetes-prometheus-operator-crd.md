---
title     : 'Prometheus CRD'
summary   : '支持 Prometheus-Operator CRD 并采集对应指标'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

## 介绍 {#intro}

本文档介绍如何让 DataKit 支持 Prometheus-Operator CRD 并采集对应指标。

## 描述 {#description}

Prometheus 有一套完善的 Kubernetes 应用指标采集方案，流程简述如下：

1. 在 Kubernetes 集群中创建 Prometheus-Operator
2. 根据需求，创建对应的 CRD 实例，该实例必须携带采集目标指标的必要配置，例如 `matchLabels` `port` `path` 等配置
3. Prometheus-Operator 会监听 CRD 实例，并根据其配置项开启指标采集

<!-- markdownlint-disable MD046 -->
???+ note

    Prometheus-Operator [官方链接](https://github.com/prometheus-operator/prometheus-operator){:target="_blank"} 和 [应用示例](https://alexandrev.medium.com/prometheus-concepts-servicemonitor-and-podmonitor-8110ce904908){:target="_blank"}。
<!-- markdownlint-enable MD046 -->

在此处，DataKit 扮演了第 3 步的角色。DataKit 会监听和发现 Prometheus-Operator CRD，根据配置开启指标采集，并最终上传到<<<custom_key.brand_name>>>。创建、更新或删除 PodMonitor/ServiceMonitor 时，对应采集任务会动态增加、重建或停止，不需要重启 DataKit。

目前 DataKit 支持 Prometheus-Operator 的 `monitoring.coreos.com/v1` 版本 `PodMonitor` 和 `ServiceMonitor`，支持的主要配置如下：

```markdown
- PodMonitor
    - selector
    - podTargetLabels
    - podMetricsEndpoints:
        - scheme
          port
          path
          params
          tlsConfig.insecureSkipVerify
    - namespaceSelector:
        any
        matchNames
- ServiceMonitor
    - selector
    - targetLabels
    - endpoints:
        - scheme
          port
          path
          params
          tlsConfig.insecureSkipVerify
    - namespaceSelector:
        any
        matchNames
```

注意：

- Monitor 中的 `interval` 不控制实际采集间隔，采集间隔统一由 KubernetesPrometheus 采集器的 `scrape_interval` 控制。
- `tlsConfig` 仅支持 `insecureSkipVerify`，不支持从 Kubernetes Secret/ConfigMap 获取证书。
- 暂不支持 `basicAuth`、`bearerTokenSecret`、`authorization` 等 Monitor 认证配置。
- ServiceMonitor 的 `podTargetLabels` 暂不支持。

`params` 支持以 `measurement` 字段来指定数据的指标集，例如：

```yaml
params:
    measurement:
    - new-measurement
```

## 启用和授权 {#enable-and-rbac}

在 KubernetesPrometheus 采集器配置中按需开启 PodMonitor 和 ServiceMonitor 自动发现：

```toml
[inputs.kubernetesprometheus]
  enable_discovery_of_prometheus_pod_monitors     = true
  enable_discovery_of_prometheus_service_monitors = true
```

DataKit 的 ServiceAccount 建议配置以下权限：

```yaml
- apiGroups: ["monitoring.coreos.com"]
  resources: ["podmonitors", "servicemonitors"]
  verbs: ["get", "list", "watch"]
```

DataKit 使用 `list` 获取现有 Monitor，并使用 `watch` 接收创建、更新和删除事件。若升级 DataKit 时仍使用只包含 `get`、`list` 的旧版 RBAC，采集不会失败：DataKit 会打印一条 WARN 日志，并自动降级为每 20 秒执行一次 `list`。降级模式下配置变更最多延迟约 20 秒生效，建议更新 RBAC 以获得实时变更通知。

## 示例 {#example}

以 Nacos 集群为例。

安装 Nacos：

<!-- markdownlint-disable MD014 -->
```shell
$ git clone https://github.com/nacos-group/nacos-k8s.git
$ cd nacos-k8s
$ chmod +x quick-startup.sh
$ ./quick-startup.sh
```
<!-- markdownlint-enable MD014 -->

*nacos/nacos-quick-start.yaml* 容器端口配置：

```yaml
containers:
  - name: k8snacos
    imagePullPolicy: Always
    image: nacos/nacos-server:latest
    ports:
      - containerPort: 8848
        name: client
      - containerPort: 9848
        name: client-rpc
      - containerPort: 9849
        name: raft-rpc
      - containerPort: 7848
        name: old-raft-rpc
```

- metrics 接口：`$IP:8848/nacos/actuator/prometheus`
- metrics port：8848

现在在 Kubernetes 集群中存在一个 Nacos metrics 服务可以采集指标。

### 创建 Prometheus-Operator CRD {#create-crd}

- 安装 Prometheus-Operator

<!-- markdownlint-disable MD014 -->
```shell
$ wget https://github.com/prometheus-operator/prometheus-operator/releases/download/v0.62.0/bundle.yaml
$ kubectl apply -f bundle.yaml
$ kubectl get crd

NAME                                        CREATED AT
alertmanagerconfigs.monitoring.coreos.com   2022-08-11T03:15:57Z
alertmanagers.monitoring.coreos.com         2022-08-11T03:15:57Z
podmonitors.monitoring.coreos.com           2022-08-11T03:15:57Z
probes.monitoring.coreos.com                2022-08-11T03:15:57Z
prometheuses.monitoring.coreos.com          2022-08-11T03:15:57Z
servicemonitors.monitoring.coreos.com       2022-08-11T03:15:57Z
thanosrulers.monitoring.coreos.com          2022-08-11T03:15:57Z
```
<!-- markdownlint-enable MD014 -->

- 创建 PodMonitor

<!-- markdownlint-disable MD014 -->
``` shell
$ cat pod-monitor.yaml

apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: nacos
  labels:
    app: nacos
spec:
  podMetricsEndpoints:
  - port: client
    path: /nacos/actuator/prometheus
  namespaceSelector:
    matchNames:
    - default
  selector:
    matchLabels:
      app: nacos

$ kubectl apply -f pod-monitor.yaml
```
<!-- markdownlint-enable MD014 -->

几个重要的配置项要和 Nacos 一致：

- namespace: default
- app: `nacos`
- port: client
- path: `/nacos/actuator/prometheus`

PodMonitor 的完整字段定义参见 [Prometheus Operator API 文档](https://prometheus-operator.dev/docs/api-reference/api/){:target="_blank"}；DataKit 实际支持范围以前文列表为准。

### 指标集和 tags {#measurement-and-tags}

详见 [指标集命名规则](kubernetesprometheus.md#measurement-naming-rules) 和 [自动添加的标签](kubernetesprometheus.md#auto-added-tags)。

### 验证 {#check}

启动 DataKit，使用 `datakit monitor -V` 或在<<<custom_key.brand_name>>>页面上查看，能找到以 `nacos_` 开头的指标集说明采集成功。
