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

## 自动发现 Pod/Service 的 Prometheus 指标 {#auto-discovery-metrics-with-prometheus}

[:octicons-tag-24: Version-1.5.10](../datakit/changelog.md#cl-1.5.10)

根据 Pod 或 Service 的指定 Annotations，拼接一个 HTTP URL 并以此创建 Prometheus 指标采集。

此功能默认关闭，要先在 DataKit 开启此功能，按需添加以下两个环境变量，详见 [container 文档](container.md)：

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
    prometheus.io/port: "80"
spec:
  selector:
    app.kubernetes.io/name: proxy
  ports:
  - name: name-of-service-port
    protocol: TCP
    port: 8080
    targetPort: http-web-svc
```

DataKit 会自动发现带有 `prometheus.io/scrape: "true"` 的 Service，并通过 `selector` 找到匹配的 Pod，构建 prom 采集：

- `prometheus.io/scrape`：只采集为 "true "的 Service，必选项。
- `prometheus.io/port`：指定 metrics 端口，必选项。注意这个端口必须在 Pod 存在否则会采集失败。
- `prometheus.io/scheme`：根据 metrics endpoint 选择 `https` 和 `http`，默认是 `http`。
- `prometheus.io/path`：配置 metrics path，默认是 `/metrics`。
- `prometheus.io/param_measurement`：配置指标集名称，默认是当前 Pod 的父级 OwnerReference。

采集目标的 IP 地址是 `PodIP`。

<!-- markdownlint-disable MD046 -->
???+ note

    DataKit 并不是去采集 Service 本身，而且采集 Service 配对的 Pod。
<!-- markdownlint-enable -->

默认采集间隔为 1 分钟。

### 指标集和 tags {#measurement-and-tags}

自动发现 Pod/Service Prometheus，指标集命名有 4 种情况，按照优先级分别是：

1. 手动配置指标集

    - 在 Pod/Service Annotations 配置 `prometheus.io/param_measurement`，其值为指定的指标集名称，例如：

      ```yaml
      apiVersion: v1
      kind: Pod
      metadata:
        name: testing-prom
        labels:
          app.kubernetes.io/name: MyApp
        annotations:
          prometheus.io/scrape: "true"
          prometheus.io/port: "8080"
          prometheus.io/param_measurement: "pod-measurement"
      ```

      它的 Prometheus 数据指标集为 `pod-measurement`。

    - 如果是 Prometheus 的 PodMonitor/ServiceMonitor CRDs，可以使用 `params` 指定 `measurement`，例如：

      ```yaml
      params:
          measurement:
          - new-measurement
      ```

1. 由数据切割所得

    - 默认会将指标名称以下划线 `_` 进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称。

      例如以下的 Prometheus 原数据：

      ```not-set
      # TYPE promhttp_metric_handler_errors_total counter
      promhttp_metric_handler_errors_total{cause="encoding"} 0
      ```

      以第一根下划线做区分，左边 `promhttp` 是指标集名称，右边 `metric_handler_errors_total` 是字段名。

    - 为了保证字段名和原始 Prom 数据一致，container 采集器支持 “保留 prom 原始字段名”，开启方式如下：

        - 配置文件是 `keep_exist_prometheus_metric_name = true`
        - 环境变量是 `ENV_INPUT_CONTAINER_KEEP_EXIST_PROMETHEUS_METRIC_NAME = "true"`

      以上面的 `promhttp_metric_handler_errors_total` 数据为例，开启此功能后，指标集是 `promhttp`，但是字段名不再切割，会使用原始值 `promhttp_metric_handler_errors_total`。

DataKit 会添加额外 tag 用来在 Kubernetes 集群中定位这个资源：

- 对于 `Service` 会添加 `namespace` 和 `service_name` `pod_name` 三个 tag
- 对于 `Pod` 会添加 `namespace` 和 `pod_name` 两个 tag

### 采集当前 Kubernetes Prometheus 数据 {#kube-self-metrtic}

此功能属于试验性质，后续可能会修改。

DataKit 支持以简单的配置方式，使用环境变量 `ENV_INPUT_CONTAINER_ENABLE_K8S_SELF_METRIC_BY_PROM="true"` 开启采集 Kubernetes 集群的 Prometheus 数据。

数据源包括 APIServer、Controller 等，采集方式参考以下：

<!-- markdownlint-disable -->
| resource   | from                   | election | measurement      | method                                        |
| ---        | --                     | --       | --               | --                                            |
| APIServer  | Kubernetes             | true     | `kube-apiserver`   | `https://kubernetes.default.com:443/metrics`    |
| Controller | Kubernetes Static Pods | false    | `kube-controller`  | `https://127.0.0.1:10257/metrics`               |
| Scheduler  | Kubernetes Static Pods | false    | `kube-scheduler`   | `https://127.0.0.1:10259/metrics`               |
| Etcd       | Kubernetes Static Pods | false    | `kube-etcd`        | `https://127.0.0.1:2379/metrics` （需配置证书） |
| CoreDNS    | Kubernetes Pods        | true     | `kube-coredns`     | `http://Endpoint-IP:Port/metrics`               |
| Proxy      | Kubernetes Proxy       | false    | `kube-proxy`       | `http://127.0.0.1:10249/metrics`                |
| cAdvisor   | Kubelet cAdvisor       | false    | `kubelet-cadvisor` | `https://127.0.0.1:10250/metrics/cadvisor`      |
| Resource   | Kubelet Resource       | false    | `kubelet-resource` | `https://127.0.0.1:10250/metrics/resource`      |
<!-- markdownlint-enable -->

- Kubernetes 相关的服务（APIServer、Controller、Scheduler）使用 BearerToken 进行验证。

- 对于 Static Pods，DataKit 会先确认当前 Node 上是否存在此类 Pod，例如 `kubectl get pod -n kube-system -l tier=control-plane,component=kube-scheduler --field-selector spec.nodeName=Node-01` 查看当前 Node-01 是否存在 Scheduler 服务。如果存在，根据默认 url 去采集。

采集 Etcd 需要配置证书，操作步骤如下：

1. 询问 Kubernetes 管理员，Etcd 的证书存放路径。例如以下：

```shell
$ ls /etc/kubernetes/pki/etcd
ca.crt  ca.key  healthcheck-client.crt  healthcheck-client.key  peer.crt  peer.key  server.crt  server.key
```

1. 使用 Etcd 证书创建 Secret，命令如下：

```shell
$ kubectl create secret generic datakit-etcd-ssl --from-file=/etc/kubernetes/pki/etcd/ca.crt --from-file=/etc/kubernetes/pki/etcd/peer.crt --from-file=/etc/kubernetes/pki/etcd/peer.key -n datakit
secret/datakit-etcd-ssl created
```

1. 在 DataKit 的 yaml 添加以下配置：

```yaml
    spec:
      containers:
      - name: datakit
        env:
        - name: ENV_INPUT_CONTAINER_ENABLE_K8S_SELF_METRIC_BY_PROM    # 开启对 Kubernetes Prometheus 数据采集
          value: "true"
        - name: ENV_INPUT_CONTAINER_K8S_SELF_METRIC_CONFIG    # 指定配置和证书路径
          value: '{"etcd":{"ca_file":"/tmp/etcd/ca.crt","cert_file":"/tmp/etcd/peer.crt","key_file":"/tmp/etcd/peer.key"}}' 
        volumeMounts:
        - name: etcd-ssl    # 添加 volumeMount
          mountPath: /tmp/etcd   
          
      # ..other..

      volumes:
      - name: etcd-ssl    # 使用 Secret 创建 volume
        secret:
          secretName: datakit-etcd-ssl
```

环境变量 `ENV_INPUT_CONTAINER_K8S_SELF_METRIC_CONFIG` 是一个 JSON 配置，格式如下：

```json
{
    "etcd": {
        "ca_file":   "/tmp/etcd/ca.crt",
        "cert_file": "/tmp/etcd/peer.crt",
        "key_file":  "/tmp/etcd/peer.key"
    }
}
```

**注意，如果 DataKit 部署在云平台，不再支持采集 Kubernetes 系统服务和 Etcd 组件，因为云平台通常会隐藏这部分资源，无法再查询到。**

## 延伸阅读 {#more-readings}

- [Prometheus Exporter 数据采集](prom.md)
