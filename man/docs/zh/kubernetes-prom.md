
# Prometheus Exporter 指标采集

---

:fontawesome-brands-linux: :material-kubernetes:

---

## 介绍 {#intro}

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

### 选择指定 Pod IP {#pod-ip}

某些情况下， Pod 上会存在多个 IP，此时仅仅通过 `$IP` 来获取 Exporter 地址是不准确的。支持通过配置 Annotations 选择 Pod IP。

- Key 为固定的 `datakit/prom.instances.ip_index`
- Value 是自然数，例如 `0` `1` `2` 等，是要使用的 IP 在整个 IP 数组（Pod IPs）中的位置下标。

如果没有此 Annotations Key，则使用默认 Pod IP。

### 操作步骤 {#steps}

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

<!-- markdownlint-disable MD046 -->
???+ attention

    `annotations` 一定添加在 `template` 字段下，这样 *deployment.yaml* 创建的 Pod 才会携带 `datakit/prom.instances`。
<!-- markdownlint-enable -->

- 使用新的 yaml 创建资源

```shell
kubectl apply -f deployment.yaml
```

至此，Annotations 已经添加完成。DataKit 稍后会读取到 Pod 的 Annotations，并采集 `url` 上暴露出来的指标。

## 自动发现 Pod/Service 的 Prometheus 指标 {#auto-discovery-metrics-with-prometheus}

[:octicons-tag-24: Version-1.5.10](changelog.md#cl-1.5.10)

根据 Pod 或 Service 的指定 Annotations，拼接一个 HTTP URL 并以此创建 Prometheus 指标采集。

此功能默认关闭，要先在 Datakit 开启此功能，按需添加以下两个环境变量，详见 [container 文档](container.md)：

- `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS`: `"true"`
- `ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS`: `"true"`

**注意，此功能可能会产生大量的时间线。**

### 示例 {#auto-discovery-metrics-with-prometheu-example}

以在 Service 添加 Annotations 为例。使用以下 yaml 配置，创建 Pod 和 Service，并在 Service 添加 `prometheus.io/scrape` 等 Annotations：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: ns-testing
  labels:
    app.kubernetes.io/name: proxy
spec:
  containers:
  - name: nginx
    image: nginx:stable
    ports:
      - containerPort: 80
        name: http-web-svc

---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  namespace: ns-testing
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
spec:
  selector:
    app.kubernetes.io/name: proxy
  ports:
  - name: name-of-service-port
    protocol: TCP
    port: 8080
    targetPort: http-web-svc
```

Datakit 会自动发现带有 `prometheus.io/scrape: "true"` 的 Service，并根据其另外的配置项，构建 prom 采集：

- `prometheus.io/scrape`：只采集为 "true "的 Service，必选项
- `prometheus.io/port`：指定 metrics 端口，必选项
- `prometheus.io/scheme`：根据 metrics endpoint 选择 `https` 和 `http`，默认是 `http`
- `prometheus.io/path`：配置 metrics path，默认是 `/metrics`

以上文的 Service yaml 配置为例，最终 Datakit 会访问 `http://nginx-service.ns-testing:8080/metrics` 采集 Prometheus 指标。

采集间隔为 1 分钟。

### 指标集和 tags {#measurement-and-tags}

自动发现 Pod/Service Prometheus，其指标集名称是由 Datakit 解析所得，默认会将指标名称以下划线 `_` 进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称。

例如以下的 Prometheus 原数据：

```not-set
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
```

以第一根下划线做区分，左边 `promhttp` 是指标集名称，右边 `metric_handler_errors_total` 是字段名。

Datakit 会添加额外 tag 用来在 Kubernetes 集群中定位这个资源：

- 对于 `Service` 会添加 `namespace` 和 `service_name` 两个 tag
- 对于 `Pod` 会添加 `namespace` 和 `pod_name` 两个 tag

## 延伸阅读 {#more-readings}

- [Prometheus Exporter 数据采集](prom.md)
