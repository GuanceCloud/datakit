---
title     : 'Kubernetes Prometheus Discovery'
summary   : '支持发现 Kubernetes 中的 Prometheus 指标暴露并采集'
tags      :
  - 'PROMETHEUS'
  - 'KUBERNETES'
__int_icon: 'icon/kubernetes'
---

[:octicons-tag-24: Version-1.34.0](../datakit/changelog.md#cl-1.34.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

## 概述 {#overview}

KubernetesPrometheus 是一个只能应用在 Kubernetes 的采集器，它根据自定义配置实现自动发现 Prometheus 服务并进行采集，极大简化了使用过程。

本采集器需要对 Kubernetes 一定的熟悉程度，例如能通过 kubectl 命令查看 Services、Pods 等资源的各项属性。

简述本采集器的实现方式，能更好的了解和使用采集器。 KubernetesPrometheus 的实现主要有以下几步：

1. 对 Kubernetes APIServer 注册事件通知机制，能及时获知各类资源的创建、更新和删除情况
1. 当某个资源（例如 Pod）被创建时，KubernetesPrometheus 会接收到通知，根据配置文件决定是否对该 Pod 进行采集
1. 如果该 Pod 符合条件，依照配置文件的占位符找到 Pod 对应属性（例如 Port 等），构建一个访问地址
1. KubernetesPrometheus 会访问该地址，将数据进行解析和添加标签
1. 如果该 Pod 发生更新或删除，KubernetesPrometheus 采集器会终止对该 Pod 的采集，再根据具体情况判断是否开启新采集


### 配置说明 {#input-config-added}

- 以下是一份最基础的配置，它只有 2 个配置项——选择自发现的目标是 Pod，和指定目标 Port。它实现了对所有 Pods 的 Prometheus 数据采集，即使它们并没有 export Prometheus 数据：

```yaml
[[inputs.kubernetesprometheus.instances]]
  role       = "pod"
  port       = "__kubernetes_pod_container_nginx_port_metrics_number"
```

- 对上面的配置进行补充，不再采集所有 Pods，而是按照 Namespace 和 Selector 指定某一类 Pod。如配置所示，现在只采集 Namespace 是 `middleware` 且存在 Label 是 `app=nginx` 的 Pods：

```yaml
[[inputs.kubernetesprometheus.instances]]
  role       = "pod"
  namespaces = ["middleware"]
  selector   = "app=nginx"

  port       = "__kubernetes_pod_container_nginx_port_metrics_number"
```

- 再次对配置进行补充，这次要添加一些标签。标签值是动态的，根据目标 Pod 的属性取值。此处添加了 4 个标签：

```yaml
[[inputs.kubernetesprometheus.instances]]
  role       = "pod"
  namespaces = ["middleware"]
  selector   = "app=nginx"

  port       = "__kubernetes_pod_container_nginx_port_metrics_number"

  [inputs.kubernetesprometheus.instances.custom]
    [inputs.kubernetesprometheus.instances.custom.tags]
      instance         = "__kubernetes_mate_instance"
      host             = "__kubernetes_mate_host"
      pod_name         = "__kubernetes_pod_name"
      pod_namespace    = "__kubernetes_pod_namespace"
```

- 如果目标 Pod 的 Prometheus 服务是 https 协议，还需要额外配置认证证书，这些证书已经提前挂载进 Datakit 容器中：

```yaml
[[inputs.kubernetesprometheus.instances]]
  role       = "pod"
  namespaces = ["middleware"]
  selector   = "app=nginx"

  scheme     = "https"
  port       = "__kubernetes_pod_container_nginx_port_metrics_number"

  [inputs.kubernetesprometheus.instances.custom]
    [inputs.kubernetesprometheus.instances.custom.tags]
      instance         = "__kubernetes_mate_instance"
      host             = "__kubernetes_mate_host"
      pod_name         = "__kubernetes_pod_name"
      pod_namespace    = "__kubernetes_pod_namespace"

  [inputs.kubernetesprometheus.instances.auth]
    [inputs.kubernetesprometheus.instances.auth.tls_config]
      insecure_skip_verify = false
      ca_certs = ["/opt/nginx/ca.crt"]
      cert     = "/opt/nginx/peer.crt"
      cert_key = "/opt/nginx/peer.key"
```

- 最后，这是一份完整的配置，它包含了所有的配置项：

```yaml
[[inputs.kubernetesprometheus.instances]]
  role       = "pod"
  namespaces = ["middleware"]
  selector   = "app=nginx"

  scrape     = "true"
  scheme     = "https"
  port       = "__kubernetes_pod_container_nginx_port_metrics_number"
  path       = "/metrics"
  params     = ""

  [inputs.kubernetesprometheus.instances.custom]
    measurement        = "pod-nginx"
    job_as_measurement = false
    [inputs.kubernetesprometheus.instances.custom.tags]
      instance         = "__kubernetes_mate_instance"
      host             = "__kubernetes_mate_host"
      pod_name         = "__kubernetes_pod_name"
      pod_namespace    = "__kubernetes_pod_namespace"

  [inputs.kubernetesprometheus.instances.auth]
    bearer_token_file      = "/var/run/secrets/kubernetes.io/serviceaccount/token"
    [inputs.kubernetesprometheus.instances.auth.tls_config]
      insecure_skip_verify = false
      ca_certs = ["/opt/nginx/ca.crt"]
      cert     = "/opt/nginx/peer.crt"
      cert_key = "/opt/nginx/peer.key"
```

此外还有一类全局配置，它是最顶层的配置，主要负责一些功能开启或关闭，以及给全部 instances 添加标签：

```yaml
[inputs.kubernetesprometheus]
  node_local      = true   # 开启 NodeLocal 模式，将采集分散到各个节点
  scrape_interval = "30s"  # 指定采集间隔，默认是 30 秒

  enable_discovery_of_prometheus_pod_annotations     = false  # 开启预定义的 Pod Annotations 配置
  enable_discovery_of_prometheus_service_annotations = false  # 开启预定义的 Service Annotations 配置
  enable_discovery_of_prometheus_pod_monitors        = false  # 开启 Prometheus PodMonitors CRD 功能
  enable_discovery_of_prometheus_service_monitors    = false  # 开启 Prometheus ServiceMonitors CRD 功能

  [inputs.kubernetesprometheus.global_tags]
    instance = "__kubernetes_mate_instance"
    host     = "__kubernetes_mate_host"
 
  [[inputs.kubernetesprometheus.instances]]
  # ..other
```

`global_tags` 会给全部 instance 添加 tags，只支持 `__kubernetes_mate_instance` 和 `__kubernetes_mate_host` 两个占位符，占位符功能请查看后文。

<!-- markdownlint-disable MD046 -->
???+ attention

    不需要手动配置 IP，采集器会使用默认 IP，具体是：

    - `node` 使用 InternalIP
    - `Pod` 使用 Pod IP
    - `Service` 使用对应 Endpoints 的 Address IP（多个）
    - `Endpoints` 使用对应 Address IP（多个）

    另外需要注意，端口不能绑定到环回地址，否则外部无法访问。
<!-- markdownlint-enable -->

假设这个 Pod IP 是 `172.16.10.10`，容器 nginx 的 metrics 端口是 9090。

最终 KubernetesPrometheus 采集器会创建一个目标地址是 `http://172.16.10.10:9090/metrics` 的 Prometheus 采集，解析数据后添加标签 `pod_name` 和 `pod_namespace`，指标集名称是 `pod-nginx`。

如果存在另一个 Pod 也符合 namespace 和 selector 配置，那么也采集它。

## 配置详解 {#input-config}

KubernetesPrometheus 采集器主要使用占位符进行配置，只保留最基础的、实现采集的必要配置（例如 port、path 等），现在逐个说明各个配置项的含义。

以上文的配置为例，其中：

### 主配置 {#input-config-main}

| 配置项      | 是否必要    | 默认值     | 描述                                                                                                                                                                                            | 是否支持占位符 |
| ----------- | ----------- | -----      | -----------                                                                                                                                                                                     | -----          |
| `role`      | Yes         | 无         | 指定采集的资源类型，只能是 `node`、`pod`、`service` 和 `endpoints` 任意一个                                                                                                                     | No             |
| `namespace` | No          | 无         | 限定这个资源所属的命名空间，它是个数组，可以写多个，例如 `["kube-system", "testing"]`                                                                                                           | No             |
| `selector`  | No          | 无         | labels 查询和过滤，它的范围更小、更精确。它是一个字符串，支持 `'=', '==', '!='`，例如 `key1=value1,key2=value2`，同时它支持 Glob 匹配模式。详见[后文](kubernetesprometheus.md#selector-example) | No             |
| `scrape`    | No          | "true"     | 判定是否要采集。当它为空字符串或为 `true` 时，会执行采集，否则不采集                                                                                                                            | Yes            |
| `scheme`    | No          | "http"     | 默认值是 `http`，如果采集需要用到证书，应改为 `https`                                                                                                                                           | Yes            |
| `port`      | Yes         | 无         | 目标地址的端口，需要手动配置                                                                                                                                                                    | Yes            |
| `path`      | No          | "/metrics" | http 访问路径，默认值是 `/metrics`                                                                                                                                                              | Yes            |
| `params`    | No          | 无         | http 访问参数，是一个字符串，例如 `name=nginx&package=middleware`                                                                                                                               | No             |

### 添加 HTTP Headers {#input-config-http-headers}

支持配置多个 Key/Value，在 HTTP 请求中添加它们。例如：

```yaml
  [inputs.kubernetesprometheus.instances]
    # other..
    [inputs.kubernetesprometheus.instances.http_headers]
      "Authorization" = "Bearer XXXXX"
      "X-testing-key" = "value"
```

### 定制化配置 {#input-config-custom}

| 配置项               | 是否必要    | 默认值                             | 描述                                                                          |
| -----------          | ----------- | -----                              | -----------                                                                   |
| `measurement`        | No          | 对指标字段名的第一条下划线切割所得 | 配置指标集名称                                                                |
| `job_as_measurement` | No          | false                              | 是否使用数据中的 `job` 标签值当做指标集名称                                   |
| `tags`               | No          | 无                                 | 添加标签，注意标签的 key 不支持占位符，value 支持占位符，详见后文的占位符描述 |

<!-- markdownlint-disable MD046 -->
???+ attention

    KubernetesPrometheus 采集器会添加 Datakit 的 `global_tags`[:octicons-tag-24: Version-1.65.1](../datakit/changelog.md#cl-1.65.1)。
<!-- markdownlint-enable -->

### 权限和验证 {#input-config-auth}

- `bearer_token_file` 配置 token 文件路径，通常和 `insecure_skip_verify` 一起用
- `tls_config` 配置证书相关，子配置项分别是 `insecure_skip_verify`、`ca_certs`、`cert`、`cert_key`，需要注意 `ca_certs` 是个数组配置

## 占位符说明 {#placeholders}

占位符是整个采集方案中非常重要的一部分。它本身是一个字符串，指向了资源的某个属性。

占位符主要有 2 类，即“固定匹配”和“通配匹配”：

- 固定匹配，类似 `__kubernetes_pod_name`，它是唯一的，只指向该 Pod Name，简单明了
- 通配匹配，用来配置一些自定义资源名，在下文中以 `%s` 代替。例如，Pod 有个 Label 是 `app=nginx`，需要把 `nginx` 取出来当做标签，要这样配置：

```yaml
    [inputs.kubernetesprometheus.instances.custom.tags]
      app = "__kubernetes_pod_label_app"
```

为什么要有这一步？

因为这个 label 的值不是固定的，根据 Pod 不同会有变化。在这个 Pod 是 `app=nginx`，在另一个 Pod 可能是 `app=redis`，如果要用同一份配置采集这 2 个 Pod，必然要对它们进行标签区分，就可以用这种配置方式。

占位符更多用在 `annotation` 和 `label` 的选择上，另外配置 port 也用到占位符。例如，Pod 有个容器叫 nginx，该容器有个 port 叫 `metrics`，现在采集这个端口，可以写成 `__kubernetes_pod_container_nginx_port_metrics_number`。

以下是全局占位符和各类资源（`node`、`pod`、`service`、`endpoints`）支持的占位符。

### 全局占位符 {#placeholders-global}

全局占位符是所有 Role 通用，多用来指定一些特殊标签。

<!-- markdownlint-disable MD049 -->
| Name                       | Description                                                           | 使用范围                                                                              |
| -----------                | -----------                                                           | -----                                                                                 |
| __kubernetes_mate_instance | 采集目标的 instance，即 `IP:PORT`                                     | 仅支持在 global_tags/custom.tags 使用，例如 `instance = "__kubernetes_mate_instance"` |
| __kubernetes_mate_host     | 采集目标的 host，即 `IP`。如果该值是 `localhost` 或环回地址将不再添加 | 仅支持在 global_tags/custom.tags 使用，例如 `host = "__kubernetes_mate_host"`         |
<!-- markdownlint-enable -->

### Node Role {#placeholders-node}

此类资源的采集地址是 InternalIP，对应 JSONPath 是 `.status.addresses[*].address ("type" is "InternalIP")`。

<!-- markdownlint-disable MD049 -->
| Name                                    | Description                          | 对应的 JSONPath                                       |
| -----------                             | -----------                          | -----                                                 |
| __kubernetes_node_name                  | Node 名称                            | .metadata.name                                        |
| __kubernetes_node_label_%s              | Node 标签                            | .metadata.labels['%s']                                |
| __kubernetes_node_annotation_%s         | Node 注解                            | .metadata.annotations['%s']                           |
| __kubernetes_node_address_Hostname      | Node 主机名                          | .status.addresses[*].address ("type" is "Hostname")   |
| __kubernetes_node_kubelet_endpoint_port | Node 的 kubelet 端口，一般都是 10250 | .status.daemonEndpoints.kubeletEndpoint.Port          |
<!-- markdownlint-enable -->

### Pod Role {#placeholders-pod}

此类资源的采集地址是 PodIP，对应 JSONPath 是 `.status.podIP`。

<!-- markdownlint-disable MD049 -->
| Name                                         | Description                                                                                                                | 对应的 JSONPath                                                |
| -----------                                  | -----------                                                                                                                | -----                                                          |
| __kubernetes_pod_name                        | Pod 名称                                                                                                                   | .metadata.name                                                 |
| __kubernetes_pod_namespace                   | Pod 命名空间                                                                                                               | .metadata.namespace                                            |
| __kubernetes_pod_label_%s                    | Pod 标签，例如 `_kubernetes_pod_label_app`                                                                                 | .metadata.labels['%s']                                         |
| __kubernetes_pod_annotation_%s               | Pod 注解，例如 `_kubernetes_pod_annotation_prometheus.io/port`                                                             | .metadata.annotations['%s']                                    |
| __kubernetes_pod_node_name                   | Pod 所属的 Node                                                                                                            | .spec.nodeName                                                 |
| __kubernetes_pod_container_%s_port_%s_number | 指定 container 的指定 port，例如 `__kubernetes_pod_container_nginx_port_metrics_number` 指向 `nginx` 容器的 `metrics` 端口 | .spec.containers[*].ports[*].containerPort ("name" equal "%s") |
<!-- markdownlint-enable -->

对于 __kubernetes_pod_container_%s_port_%s_number 举例：

现有 Pod nginx，它有 2 个容器，分别是 nginx 和 logfwd，现在要采集 nginx 容器的 8080 端口（假设配置中 8080 端口叫做 metrics），那么可以配置为：

`__kubernetes_pod_container_nginx_port_metrics_number`（注意 nginx 和 metrics 把 %s 替换了）

### Service Role {#placeholders-service}

Service 资源没有 IP 属性，所以使用跟它对应的 Endpoints Address IP 属性（存在多个），JSONPath 是 Endpoints `.subsets[*].addresses[*].ip`。

<!-- markdownlint-disable MD049 -->
| Name                                      | Description                                                                         | 对应的 JSONPath                                         |
| -----------                               | -----------                                                                         | -----                                                   |
| __kubernetes_service_name                 | Service 名称                                                                        | .metadata.name                                          |
| __kubernetes_service_namespace            | Service 命名空间                                                                    | .metadata.namespace                                     |
| __kubernetes_service_label_%s             | Service 标签                                                                        | .metadata.labels['%s']                                  |
| __kubernetes_service_annotation_%s        | Service 注解                                                                        | .metadata.annotations['%s']                             |
| __kubernetes_service_port_%s_port         | 指定 port（基本用不到，大部分场景都使用 targetPort）                                | .spec.ports[*].port ("name" equal "%s")                 |
| __kubernetes_service_port_%s_targetport   | 指定 targetPort                                                                     | .spec.ports[*].targetPort ("name" equal "%s")           |
| __kubernetes_service_target_kind          | Service 中没有 target，这是指向对应 endpoints 的 targetRef，取它的 `kind` 字段      | Endpoints: .subsets[*].addresses[*].targetRef.kind      |
| __kubernetes_service_target_name          | Service 中没有 target，这是指向对应 endpoints 的 targetRef，取它的 `name` 字段      | Endpoints: .subsets[*].addresses[*].targetRef.name      |
| __kubernetes_service_target_namespace     | Service 中没有 target，这是指向对应 endpoints 的 targetRef，取它的 `namespace` 字段 | Endpoints: .subsets[*].addresses[*].targetRef.namespace |
| __kubernetes_service_target_pod_name      | Deprecated, 请使用 `__kubernetes_service_target_name`                               | Endpoints: .subsets[*].addresses[*].targetRef.name      |
| __kubernetes_service_target_pod_namespace | Deprecated, 请使用 `__kubernetes_service_target_namespace`                          | Endpoints: .subsets[*].addresses[*].targetRef.namespace |
<!-- markdownlint-enable -->

### Endpoints Role {#placeholders-endpoints}

此类资源的采集地址是 Address IP（存在多个），对应 JSONPath 是 `.subsets[*].addresses[*].ip`。

<!-- markdownlint-disable MD049 -->
| Name                                                | Description                                                          | 对应的 JSONPath                               |
| -----------                                         | -----------                                                          | -----                                         |
| __kubernetes_endpoints_name                         | Endpoints 名称                                                       | .metadata.name                                |
| __kubernetes_endpoints_namespace                    | Endpoints 命名空间                                                   | .metadata.namespace                           |
| __kubernetes_endpoints_label_%s                     | Endpoints 标签                                                       | .metadata.labels['%s']                        |
| __kubernetes_endpoints_annotation_%s                | Endpoints 注解                                                       | .metadata.annotations['%s']                   |
| __kubernetes_endpoints_address_node_name            | Endpoints Address 的 Node 名称                                       | .subsets[*].addresses[*].nodeName             |
| __kubernetes_endpoints_address_target_kind          | targetRef 的 `kind` 字段                                             | .subsets[*].addresses[*].targetRef.kind       |
| __kubernetes_endpoints_address_target_name          | targetRef 的 `name` 字段                                             | .subsets[*].addresses[*].targetRef.name       |
| __kubernetes_endpoints_address_target_namespace     | targetRef 的 `namespace` 字段                                        | .subsets[*].addresses[*].targetRef.namespace  |
| __kubernetes_endpoints_address_target_pod_name      | Deprecated, 请使用 `__kubernetes_endpoints_address_target_name`      | .subsets[*].addresses[*].targetRef.name       |
| __kubernetes_endpoints_address_target_pod_namespace | Deprecated, 请使用 `__kubernetes_endpoints_address_target_namespace` | .subsets[*].addresses[*].targetRef.namespace  |
| __kubernetes_endpoints_port_%s_number               | 指定 port 名称，例如 `__kubernetes_endpoints_port_metrics_number`    | .subsets[*].ports[*].port ("name" equal "%s") |
<!-- markdownlint-enable -->

## 实际案例 {#example}

以下例子会创建一个 Service 和 Deployment，使用 KubernetesPrometheus 采集对应的 Pod。步骤如下：

1. 创建 Service 和 Deployment

```yaml
apiVersion: v1
kind: Service
metadata:
  name: prom-svc
  namespace: testing
  labels:
    app.kubernetes.io/name: prom
spec:
  selector:
    app.kubernetes.io/name: prom
  ports:
  - name: metrics
    protocol: TCP
    port: 8080
    targetPort: 30001
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prom-server
  namespace: testing
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: prom
  replicas: 1
  template:
    metadata:
      labels:
        app.kubernetes.io/name: prom
    spec:
      containers:
      - name: prom-server
        image: pubrepo.guance.com/datakit-dev/prom-server:v2
        imagePullPolicy: IfNotPresent
        env:
        - name: ENV_PORT
          value: "30001"
        - name: ENV_NAME
          value: "promhttp"
        ports:
        - name: metrics
          containerPort: 30001
```

1. 创建 ConfigMap 和 KubernetesPrometheus 配置

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:
    kubernetesprometheus.conf: |-
      [inputs.kubernetesprometheus]
        [[inputs.kubernetesprometheus.instances]]
          role       = "service"
          namespaces = ["testing"]
          selector   = "app.kubernetes.io/name=prom"

          scrape     = "true"
          scheme     = "http"
          port       = "__kubernetes_service_port_metrics_targetport"
          path       = "/metrics"
          params     = ""

          [inputs.kubernetesprometheus.instances.custom]
            measurement        = "prom-svc"
            job_as_measurement = false
            [inputs.kubernetesprometheus.instances.custom.tags]
              svc_name      = "__kubernetes_service_name"
              pod_name      = "__kubernetes_service_target_name"
              pod_namespace = "__kubernetes_service_target_namespace"
```

1. 在 Datakit yaml 中应用 `kubernetesprometheus.conf` 文件

``` yaml
        # ..other..
        volumeMounts:
        - mountPath: /usr/local/datakit/conf.d/kubernetesprometheus/kubernetesprometheus.conf
          name: datakit-conf
          subPath: kubernetesprometheus.conf
          readOnly: true
```

1. 最后启动 Datakit，在日志中能看到 `create prom url xxxxx for testing/prom-svc` 的内容，并在观测云页面看到 `prom-svc` 指标集。


---

## FAQ {#faq}

### 记录 UP 指标 {#collector-up}

up 指标是一个内置的重要指标，用于表示某个目标的健康状态，主要用于监控目标是否处于“正常”运行状态。KubernetesPrometheus 采集器默认开启 up 指标，以便提供关于目标服务是否可用的实时数据。

up 指标属于 `collector` 指标集，包含三个标签字段：`job`、`host` 和 `instance`，以及一个关键指标字段 `up`。其含义如下：

- `job`：采集任务的名称。这个标签有两种命名方式：
    - 手动在 Kubernetes Prometheus 配置中指定 `measurement`，且 `measurement` 不为空。例如 `kube-etcd`
    - 由目标服务的 Namespace 和 Name 拼接而成。例如 `middleware/etcd-abc`
- `instance`：采集目标的实例，表示为 `IP:PORT`。例如 `172.16.10.10:9090`
- `host`：采集目标的主机，即目标的 `IP` 地址。如果该值是 localhost 或环回地址（如 `127.0.0.1`），则该值为空字符串
- up：表示目标的运行状态
    - 当值为 1 时，表示目标正在运行并且可访问（即目标是 “up”）
    - 当值为 0 时，表示目标不可用，通常意味着目标无法访问或发生故障（即目标是 “down”）

此外，up 指标会自动添加 Datakit 的全局 `election_tags`。

### Selector 描述与示例 {#selector-example}

`selector` 是 `kubectl` 命令中常用的参数。例如，要查找标签（Labels）中包含 `tier=control-plane` 和 `component=kube-controller-manager` 的 Pod，可以使用以下命令：

```shell
$ kubectl get pod -n kube-system  --selector tier=control-plane,component=kube-controller-manager
NAMESPACE     NAME                      READY   STATUS    RESTARTS   AGE
kube-system   kube-controller-manager   1/1     Running   0          15d
```

`--selector` 参数与 `selector` 配置项功能相同。有关 `selector` 的更多使用方法，请参考[官方文档](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/){:target="_blank"}。

另外，Datakit 对 `selector` 的功能进行了扩展，使其支持 **Glob 匹配模式**。有关 Glob 的详细语法，请参考[Glob 模式文档](https://developers.tetrascience.com/docs/common-glob-pattern#glob-pattern-syntax)。以下是一些示例：

[:octicons-tag-24: Version-1.65.1](../datakit/changelog.md#cl-1.65.1)

- **`selector="app=middleware*"`**：匹配任意以 `middleware` 开头的值，例如 `middleware-etcd` 或 `middleware-coredns`。
- **`selector="app=middleware-{nginx,redis}"`**：匹配 `middleware-nginx` 和 `middleware-redis`，等同于 `app in (middleware-nginx, middleware-redis)`。
- **`selector="app=middleware-[123]"`**：匹配 `middleware-1`、`middleware-2` 和 `middleware-3` 中的任意一个。

<!-- markdownlint-disable MD046 -->
???+ attention
    在此处 Glob 模式中不支持 `!` 排除符。例如，`app=middleware-[!0123]` 会在解析阶段报错。这是因为在 Selector 语法中，`!` 是关键字符（例如用于 `app!=nginx`），因此不能用于 Glob 模式。
<!-- markdownlint-enable -->

### Bearer Token 验证 {#http-bearer-token}

通常情况下，使用 Bearer Token 验证有两个前提，分别是开启 `https` 和 `insecure_skip_verify` 为 `true`。

配置 Bearer Token 有两种方式：

- 如果 Token 是字符串，可以在 `http_headers` 手动填写，例如：

```yaml
    [inputs.kubernetesprometheus.instances.http_headers]
      "Authorization" = "Bearer XXXXX"
```

- 如果 Token 以文件形式存放，要在 `bearer_token_file` 指定文件路径，见示例。KubernetesPrometheus 采集器会自动读取文件内容添加到 `Authorization` Header，注意，如果手动配置 `http_headers` 存在 `Authorization`，`bearer_token_file` 会失效。

```yaml
    [inputs.kubernetesprometheus.instances.auth]
      bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
```
