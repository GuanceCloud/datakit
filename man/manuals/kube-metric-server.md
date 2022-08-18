# Kubernetes with Metric Server
---

操作系统支持：Linux

## 视图预览

Kubernetes 性能指标展示：包括 pod 数量、deployment 数量、job 数量、endpoint 数量、service 数量、CPU、内存、Pod 分布等。<br />

![image.png](imgs/input-kube-metric-server-01.png)<br />
![image.png](imgs/input-kube-metric-server-02.png)<br />
![image.png](imgs/input-kube-metric-server-03.png)<br />
![1650000133(1).png](imgs/input-kube-metric-server-04.png)<br />
![1650000156(1).png](imgs/input-kube-metric-server-05.png)<br />
![1650000171(1).png](imgs/input-kube-metric-server-06.png)<br />
![1650000207(1).png](imgs/input-kube-metric-server-07.png)

## 安装部署

说明：示例 Kubernetes 版本为：1.22.6

### 前置条件

- Kubernetes 集群。
- 采集 Kubernetes Pod 指标数据，[需要 Kubernetes 安装 Metrics-Server 组件](https://github.com/kubernetes-sigs/metrics-server#installation)。


### 配置实施

#### 部署 Metric-Server  (必选)

新建 metric-server.yaml ，在 kubernetes 集群执行

```
kubectl apply -f metric-server.yaml 
```

metric-server.yaml  完整内容如下：

```
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-view: "true"
  name: system:aggregated-metrics-reader
rules:
- apiGroups:
  - metrics.k8s.io
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  - nodes/stats
  - namespaces
  - configmaps
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server-auth-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server:system:auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    k8s-app: metrics-server
  name: system:metrics-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:metrics-server
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: https
  selector:
    k8s-app: metrics-server
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    k8s-app: metrics-server
  name: metrics-server
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: metrics-server
  strategy:
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      labels:
        k8s-app: metrics-server
    spec:      
      containers:
      - args:
        - --cert-dir=/tmp
        - --secure-port=4443
        - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
        - --kubelet-use-node-status-port
        - --metric-resolution=15s
        - --kubelet-insecure-tls 
        # image: k8s.gcr.io/metrics-server/metrics-server:v0.5.2
        image: bitnami/metrics-server:0.5.2
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /livez
            port: https
            scheme: HTTPS
          periodSeconds: 10
        name: metrics-server
        ports:
        - containerPort: 4443
          name: https
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /readyz
            port: https
            scheme: HTTPS
          initialDelaySeconds: 20
          periodSeconds: 10
        resources:
          requests:
            cpu: 100m
            memory: 200Mi
        securityContext:
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
        volumeMounts:
        - mountPath: /tmp
          name: tmp-dir
      nodeSelector:
        kubernetes.io/os: linux
      priorityClassName: system-cluster-critical
      serviceAccountName: metrics-server
      volumes:
      - emptyDir: {}
        name: tmp-dir
---
apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  labels:
    k8s-app: metrics-server
  name: v1beta1.metrics.k8s.io
spec:
  group: metrics.k8s.io
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: metrics-server
    namespace: kube-system
  version: v1beta1
  versionPriority: 100

```

#### Daemonset 部署 DataKit (必选)

登录[观测云](https://console.guance.com/)，【集成】->【DataKit】-> 【Kubernetes】，下载 `datakit.yaml`（命名无要求）。

1、 修改 `datakit.yaml` 中的 dataway 配置

进入【管理】模块，找到下图中 token。

![1648545757(1).png](imgs/input-kube-metric-server-08.png)<br />替换 datakit.yaml 文件中的 ENV_DATAWAY 环境变量的 value 值中的 <your-token>。

```
        - name: ENV_DATAWAY
          value: https://openway.guance.com?token=<your-token>
```

在 datakit.yaml 文件中的 ENV_GLOBAL_TAGS 环境变量值最后增加 cluster_name_k8s=k8s-prod，其中  k8s-prod 为指标设置的全局 tag，即指标所在的集群名称。

```
        - name: ENV_GLOBAL_TAGS
          value: host=__datakit_hostname,host_ip=__datakit_ip,cluster_name_k8s=k8s-prod
```


2、 增加 ENV_NAMESPACE 环境变量 

修改 `datakit.yaml`，增加 ENV_NAMESPACE 环境变量，这个环境变量是为了区分不同集群的选举，多个集群 value 值不能相同。

```
        - name: ENV_NAMESPACE
          value: xxx
```

3、 定义ConfigMap

『注意』下载的 datakit.yaml 并没有 ConfigMap，定义的 ConfigMap 可一起放到 datakit.yaml 。

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: datakit-conf
  namespace: datakit
data:
    #### container
    container.conf: |-  
      [inputs.container]
        docker_endpoint = "unix:///var/run/docker.sock"
        containerd_address = "/var/run/containerd/containerd.sock"

        enable_container_metric = true
        enable_k8s_metric = true
        enable_pod_metric = true

        ## Containers logs to include and exclude, default collect all containers. Globs accepted.
        container_include_log = []
        container_exclude_log = ["image:pubrepo.jiagouyun.com/datakit/logfwd*", "image:pubrepo.jiagouyun.com/datakit/datakit*"]

        exclude_pause_container = true

        ## Removes ANSI escape codes from text strings
        logging_remove_ansi_escape_codes = false

        kubernetes_url = "https://kubernetes.default:443"

        ## Authorization level:
        ##   bearer_token -> bearer_token_string -> TLS
        ## Use bearer token for authorization. ('bearer_token' takes priority)
        ## linux at:   /run/secrets/kubernetes.io/serviceaccount/token
        ## windows at: C:\var\run\secrets\kubernetes.io\serviceaccount\token
        bearer_token = "/run/secrets/kubernetes.io/serviceaccount/token"
        # bearer_token_string = "<your-token-string>"

        [inputs.container.tags]
          # some_tag = "some_value"
          # more_tag = "some_other_value"
```

[inputs.container]参数说明

- enable_container_metric：是否开启 container 指标采集，请设置为true。
- enable_k8s_metric：是否开启 kubernetes 指标采集。
- enable_pod_metric：是否开启 Pod 指标采集。
- container_include_log：须要采集的容器日志。
- container_exclude_log：不须要采集的容器日志。

`container_include_log` 和 `container_exclude_log` 必须以 `image` 开头，格式为 `"image:<glob规则>"`，表示 glob 规则是针对容器 image 生效。[Glob 规则](https://en.wikipedia.org/wiki/Glob_(programming))是一种轻量级的正则表达式，支持 `*` `?` 等基本匹配单元

4、 使用ConfigMap

在 datakit.yaml 文件中的 volumeMounts 下面增加：

```
        - mountPath: /usr/local/datakit/conf.d/container/container.conf
          name: datakit-conf
          subPath: container.conf
```

5、 部署Datakit

```
kubectl apply -f datakit.yaml
```

#### 日志采集 

默认自动收集输出到控制台的日志。

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 kubernetes 指标都会带有 tag1 = "val1" 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)>
```
          [inputs.kubernetes.tags]
           #tag1 = "val1"
           #tag2 = "valn"   
```


## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Kubernetes Kubelet 监控视图>

## 异常检测

暂无

## 指标详解

#### `kube_cronjob`

Kubernetes cron job 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`cronjob`|Name must be unique within a namespace.|
|`namespace`|Namespace defines the space within each name must be unique.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`count`|Number of cronjobs|int|count|
|`duration_since_last_schedule`|The duration since the last time the cronjob was scheduled.|int|s|
|`spec_suspend`|This flag tells the controller to suspend subsequent executions.|bool|-|

#### `kube_daemonset`

Kubernetes Daemonset 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`daemonset`|Name must be unique within a namespace.|
|`namespace`|Namespace defines the space within each name must be unique.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`count`|Number of daemonsets|int|count|
|`daemons_unavailable`|The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds).|int|count|
|`desired`|The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod).|int|count|
|`misscheduled`|The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod.|int|count|
|`ready`|The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready.|int|count|
|`scheduled`|The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod.|int|count|
|`updated`|The total number of nodes that are running updated daemon pod.|int|count|

#### `kube_endpoint`

Kubernetes Endpoints 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`endpoint`|Name must be unique within a namespace.|
|`namespace`|Namespace defines the space within each name must be unique.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`address_available`|Number of addresses available in endpoint.|int|count|
|`address_not_ready`|Number of addresses not ready in endpoint.|int|count|
|`count`|Number of endpoints|int|count|

#### `kubernetes`

Kubernetes count 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`namespace`|namespace|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`cluster_role`|RBAC cluster role count|int|-|
|`cronjob`|cronjob count|int|-|
|`deployment`|deployment count|int|-|
|`job`|job count|int|-|
|`node`|node count|int|-|
|`pod`|pod count|int|-|
|`replica_set`|replica_set count|int|-|
|`service`|service count|int|-|

#### `kube_deployment`

Kubernetes Deployment 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`deployment`|Name must be unique within a namespace.|
|`namespace`|Namespace defines the space within each name must be unique.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`condition`|The current status conditions of a deployment|int|count|
|`count`|Number of deployments|int|count|
|`paused`|Indicates that the deployment is paused (true or false).|bool|-|
|`replicas`|Total number of non-terminated pods targeted by this deployment (their labels match the selector).|int|count|
|`replicas_available`|Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.|int|count|
|`replicas_unavailable`|Total number of unavailable pods targeted by this deployment.|int|count|
|`replicas_updated`|Total number of non-terminated pods targeted by this deployment that have the desired template spec.|int|count|
|`rollingupdate_max_surge`|The maximum number of pods that can be scheduled above the desired number of pods. |int|count|
|`rollingupdate_max_unavailable`|The maximum number of pods that can be unavailable during the update.|int|count|

#### `kube_node`

Kubernetes Node 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`node`|Name must be unique within a namespace. (depercated)|
|`node_name`|Name must be unique within a namespace.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`age`|The time in seconds since the creation of the node|int|s|
|`count`|Number of nodes|int|count|
|`cpu_allocatable`|The allocatable CPU of a node that is available for scheduling.|int|-|
|`cpu_capacity`|The CPU capacity of a node.|int|-|
|`ephemeral_storage_allocatable`|The allocatable ephemeral-storage of a node that is available for scheduling.|int|-|
|`memory_allocatable`|The allocatable memory of a node that is available for scheduling.|int|-|
|`memory_capacity`|The memory capacity of a node.|int|-|
|`pods_allocatable`|The allocatable pods of a node that is available for scheduling.|int|-|
|`pods_capacity`|The pods capacity of a node.|int|-|

#### `kube_pod`

Kubernetes pod 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`namespace`|Namespace defines the space within each name must be unique.|
|`pod`|Name must be unique within a namespace.|
|`pod_name`|Name must be unique within a namespace. (depercated)|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`count`|Number of pods|int|count|
|`cpu_usage`|The percentage of cpu used|float|percent|
|`memory_usage_bytes`|The number of memory used in bytes|float|B|
|`ready`|Describes whether the pod is ready to serve requests.|int|count|

#### `kube_replicaset`

Kubernetes replicaset 指标数据

- 标签


| 标签名 | 描述    |
|  ----  | --------|
|`deployment`|The name of the deployment which the object belongs to.|
|`namespace`|Namespace defines the space within each name must be unique.|
|`replica_set`|Name must be unique within a namespace.|

- 指标列表


| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|`count`|Number of replicasets|int|count|
|`fully_labeled_replicas`|The number of fully labeled replicas per ReplicaSet.|int|count|
|`replicas`|Replicas is the most recently oberved number of replicas.|int|count|
|`replicas_desired`|Replicas is the number of desired replicas.|int|count|
|`replicas_ready`|The number of ready replicas for this replica set.|int|count|


## 最佳实践

暂无

## 故障排查
<[无数据上报排查](why-no-data.md)>

