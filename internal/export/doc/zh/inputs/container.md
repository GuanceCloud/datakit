---
title: 'Kubernetes'
summary: '采集 Container 和 Kubernetes 的指标、对象和日志数据，上报到观测云。'
__int_icon:    'icon/kubernetes/'  
tags:
  - 'KUBERNETES'
  - '容器'
dashboard:
  - desc: 'Kubernetes Cluster Overview 监控视图'
    path: 'dashboard/zh/kubernetes'
  - desc: 'Kubernetes Nodes Overview 监控视图'
    path: 'dashboard/zh/kubernetes_nodes_overview'
  - desc: 'Kubernetes Services 监控视图'
    path: 'dashboard/zh/kubernetes_services'
  - desc: 'Kubernetes Deployments 监控视图'
    path: 'dashboard/zh/kubernetes_deployment'
  - desc: 'Kubernetes DaemonSets 监控视图'
    path: 'dashboard/zh/kubernetes_daemonset'
  - desc: 'Kubernetes StatefulSets 监控视图'
    path: 'dashboard/zh/kubernetes_statefulset'
  - desc: 'Kubernetes Pods Overview 监控视图'
    path: 'dashboard/zh/kubernetes_pods_overview'
  - desc: 'Kubernetes Pods Detail 监控视图'
    path: 'dashboard/zh/kubernetes_pod_detail'
  - desc: 'Kubernetes Events 监控视图'
    path: 'dashboard/zh/kubernetes_events'
 
monitor:
  - desc: 'Kubernetes'
    path: 'monitor/zh/kubernetes'
---



{{.AvailableArchs}}

---

采集 Container 和 Kubernetes 的指标、对象和日志数据，上报到观测云。

## 采集器配置 {#config}

### 前置条件 {#requrements}

- 目前 container 支持 Docker、Containerd、CRI-O 容器运行时
    - 版本要求：Docker v17.04 及以上版本，Containerd v1.5.1 及以上，CRI-O 1.20.1 及以上
- 采集 Kubernetes 数据需要 DataKit 以 [DaemonSet 方式部署](../datakit/datakit-daemonset-deploy.md)。

<!-- markdownlint-disable MD046 -->
???+ info

    - 容器采集支持 Docker 和 Containerd 两种运行时[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7)，且默认都开启采集。

=== "主机安装"

    如果是纯 Docker 或 Containerd 环境，那么 DataKit 只能安装在宿主机上。
    
    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 *{{.InputName}}.conf.sample* 并命名为 *{{.InputName}}.conf*。示例如下：
    
    ``` toml
    {{ CodeBlock .InputSample 4 }}
    ```

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

    环境变量额外说明：
    
    - ENV_INPUT_CONTAINER_TAGS：如果配置文件（*container.conf*）中有同名 tag，将会被这里的配置覆盖掉。
    
    - ENV_INPUT_CONTAINER_LOGGING_EXTRA_SOURCE_MAP：指定替换 source，参数格式是「正则表达式=new_source」，当某个 source 能够匹配正则表达式，则这个 source 会被 new_source 替换。如果能够替换成功，则不再使用 `annotations/labels` 中配置的 source（[:octicons-tag-24: Version-1.4.7](../datakit/changelog.md#cl-1.4.7)）。如果要做到精确匹配，需要使用 `^` 和 `$` 将内容括起来。比如正则表达式写成 `datakit`，不仅可以匹配 `datakit` 字样，还能匹配到 `datakit123`；写成 `^datakit$` 则只能匹配到的 `datakit`。
    
    - ENV_INPUT_CONTAINER_LOGGING_SOURCE_MULTILINE_MAP_JSON：用来指定 source 到多行配置的映射，如果某个日志没有配置 `multiline_match`，就会根据它的 source 来此处查找和使用对应的 `multiline_match`。因为 `multiline_match` 值是正则表达式较为复杂，所以 value 格式是 JSON 字符串，可以使用 [json.cn](https://www.json.cn/){:target="_blank"} 辅助编写并压缩成一行。

???+ attention

    - 对象数据采集间隔是 5 分钟，指标数据采集间隔是 60 秒，不支持配置
    - 采集到的日志，单行（包括经过 `multiline_match` 处理后）最大长度为 32MB，超出部分会被截断且丢弃

### Docker 和 Containerd sock 文件配置 {#sock-config}

如果 Docker 或 Containerd 的 sock 路径不是默认的，则需要指定一下 sock 文件路径，根据 DataKit 不同部署方式，其方式有所差别，以 Containerd 为例：

=== "主机部署"

    修改 container.conf 的 `endpoints` 配置项，将其设置为对应的 sock 路径。

=== "Kubernetes"

    更改 *datakit.yaml* 的 volumes `containerd-socket`，将新路径 mount 到 Datakit 中，同时配置环境变量 `ENV_INPUT_CONTAINER_ENDPOINTS`：

    ``` yaml hl_lines="3 4 7 14"
    # 添加 env
    - env:
      - name: ENV_INPUT_CONTAINER_ENDPOINTS
        value: ["unix:///path/to/new/containerd/containerd.sock"]
    
    # 修改 mountPath
      - mountPath: /path/to/new/containerd/containerd.sock
        name: containerd-socket
        readOnly: true
    
    # 修改 volumes
    volumes:
    - hostPath:
        path: /path/to/new/containerd/containerd.sock
      name: containerd-socket
    ```
<!-- markdownlint-enable -->

环境变量 `ENV_INPUT_CONTAINER_ENDPOINTS` 是追加到现有的 endpoints 配置，最终实际 endpoints 配置可能有很多项，采集器会去重然后逐一连接、采集。

默认的 endpoints 配置是：

```yaml
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
  ] 
```

使用环境变量 `ENV_INPUT_CONTAINER_ENDPOINTS` 为 `["unix:///path/to/new//run/containerd.sock"]`，最终 endpoints 配置如下：

```yaml
  endpoints = [
    "unix:///var/run/docker.sock",
    "unix:///var/run/containerd/containerd.sock",
    "unix:///var/run/crio/crio.sock",
    "unix:///path/to/new//run/containerd.sock",
  ] 
```

采集器会连接和采集这些容器运行时，如果 sock 文件不存在，会在第一次连接失败时输出报错日志，不影响后续采集。

### Prometheus Exporter 指标采集 {#k8s-prom-exporter}

<!-- markdownlint-disable MD024 -->
如果 Pod/容器有暴露 Prometheus 指标，有两种方式可以采集，参见[这里](kubernetes-prom.md)


### 日志采集 {#logging-config}

日志采集的相关配置详见[此处](container-log.md)。

---

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象 {#object}

{{ range $i, $o := .Measurements }}

{{if eq $o.Type "object"}}

### `{{$o.Name}}`

{{$o.Desc}}

- 标签

{{$o.TagsMarkdownTable}}

- 指标列表

{{$o.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

{{ range $i, $l := .Measurements }}

{{if eq $l.Type "logging"}}

### `{{$l.Name}}`

{{$l.Desc}}

- 标签

{{$l.TagsMarkdownTable}}

- 字段列表

{{$l.FieldsMarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable -->

## 联动 Dataway Sink 功能 {#link-dataway-sink}

Dataway Sink [详见文档](../deployment/dataway-sink.md)。

所有的 Kubernetes 资源采集，都会添加与 CustomerKey 匹配的 Label。例如 CustomerKey 是 `name`，DaemonSet、Deployment、Pod 等资源，会在自己当前的 Labels 中找到 `name`，并将其添加到 tags。

容器会添加其所属 Pod 的 Customer Labels。


## FAQ {#faq}

### 根据 Pod Namespace 过滤指标采集 {#config-metric-on-pod-namespace}

在启用 Kubernetes Pod 指标采集（`enable_pod_metric = true`）后，Datakit 将采集集群中所有 Pod 的指标数据。由于这可能会生成大量数据，因此可以通过 Pod 的 `namespace` 字段来过滤指标采集，从而仅采集特定命名空间中的 Pod 指标。

通过配置 `pod_include_metric` 和 `pod_exclude_metric`，可以控制哪些命名空间的 Pod 会被包含或排除在指标采集之外。

<!-- markdownlint-disable md046 -->
=== "主机安装"

    ``` toml
      ## 当 Pod 的 namespace 能够匹配 `datakit` 时，采集该 Pod 的指标
      pod_include_metric = ["namespace:datakit"]
    
      ## 忽略所有 namespace 是 `kodo` 的 Pod
      pod_exclude_metric = ["namespace:kodo"]
    ```
    
    - `include` 和 `exclude` 配置项必须以字段名开头，格式为类似于 [glob 通配符](https://en.wikipedia.org/wiki/glob_(programming)) 的表达式：`"<字段名>:<glob 规则>"`。
    - 目前，`namespace` 字段是唯一支持的过滤字段。例如：`namespace:datakit-ns`。
    
    如果同时设置了 `include` 和 `exclude` 配置，Pod 必须满足以下条件：
    
    - 必须满足 `include` 的规则
    - 且不满足 `exclude` 的规则
    
    例如，以下配置会导致所有 Pod 都被过滤掉：
    
    ```toml
      ## 只采集 `namespace:datakit` 的 Pod，排除所有命名空间
      pod_include_metric = ["namespace:datakit"]
      pod_exclude_metric = ["namespace:*"]
    ```

=== "Kubernetes"

    对于 Kubernetes 环境，可以通过以下环境变量来进行配置：
    
    - `ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC`
    - `ENV_INPUT_CONTAINER_POD_EXCLUDE_METRIC`
    
    例如，如果希望只采集 `namespace` 为 `kube-system` 的 Pod 指标，可以设置 `ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC` 环境变量，如下所示：
    
    ```yaml
      - env:
          - name: ENV_INPUT_CONTAINER_POD_INCLUDE_METRIC
            value: namespace:kube-system  # 指定需要采集的命名空间
    ```
    
    通过这种方式，可以灵活地控制 Datakit 采集的 Pod 指标范围，避免采集不需要的数据，从而优化系统性能和资源利用率。

<!-- markdownlint-disable MD013 -->
### :material-chat-question: NODE_LOCAL 需要新的权限 {#rbac-nodes-stats}
<!-- markdownlint-enable -->

`ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL` 模式只推荐 DaemonSet 部署时使用，该模式需要访问 kubelet，所以需要在 RBAC 添加 `nodes/stats` 权限。例如：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: [""]
  resources: ["nodes", "nodes/stats"]
  verbs: ["get", "list", "watch"]
```

此外，Datakit Pod 还需要开启 `hostNetwork: true` 配置项。

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 采集 PersistentVolumes 和 PersistentVolumeClaims 需要新的权限 {#rbac-pv-pvc}
<!-- markdownlint-enable -->

Datakit 在 1.25.0[:octicons-tag-24: Version-1.25.0](../datakit/changelog.md#cl-1.25.0) 版本支持采集 Kubernetes PersistentVolume 和 PersistentVolumeClaim 的对象数据，采集这两种资源需要新的 RBAC 权限，详细见下：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: [""]
  resources: ["persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
```

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Kubernetes YAML 敏感字段屏蔽 {#yaml-secret}
<!-- markdownlint-enable -->

Datakit 会采集 Kubernetes Pod 或 Service 等资源的 yaml 配置，并存储到对象数据的 `yaml` 字段中。如果该 yaml 中包含敏感数据（例如密码），Datakit 暂不支持手动配置屏蔽敏感字段，推荐使用 Kubernetes 官方的做法，即使用 ConfigMap 或者 Secret 来隐藏敏感字段。

例如，现在需要在 env 中添加一份密码，正常情况下是这样：

```yaml
    containers:
    - name: mycontainer
      image: redis
      env:
        - name: SECRET_PASSWORD
      value: password123
```

在编排 yaml 配置会将密码明文存储，这是很不安全的。可以使用 Kubernetes Secret 实现隐藏，方法如下：

创建一个 Secret：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: username123
  password: password123
```

执行：

```shell
kubectl apply -f mysecret.yaml
```

在 env 中使用 Secret：

```yaml
    containers:
    - name: mycontainer
      image: redis
      env:
        - name: SECRET_PASSWORD
      valueFrom:
          secretKeyRef:
            name: mysecret
            key: password
            optional: false
```

详见[官方文档](https://kubernetes.io/zh-cn/docs/concepts/configuration/secret/#using-secrets-as-environment-variables){:target="_blank"}。

## 延伸阅读 {#more-reading}

- [eBPF 采集器：支持容器环境下的流量采集](ebpf.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
- [Kubernetes 下 DataKit 的几种配置方式](../datakit/k8s-config-how-to.md)
