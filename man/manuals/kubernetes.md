{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Kubernetes 集群指标采集，主要用于收集各种资源指标

## k8s 集群外部署使用 datakit

### 创建监控的 ServiceAccount 账号

> 注意：该账户拥有只读权限

- 创建 `account.yaml` 编排文件, 文件内容如下:

```yaml
# create ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - secrets
  - nodes
  - pods
  - services
  - resourcequotas
  - replicationcontrollers
  - limitranges
  - persistentvolumeclaims
  - persistentvolumes
  - namespaces
  - endpoints
  verbs:
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - list
  - watch
- apiGroups:
  - apps
  resources:
  - statefulsets
  - daemonsets
  - deployments
  - replicasets
  verbs:
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - cronjobs
  - jobs
  verbs:
  - list
  - watch
- apiGroups:
  - autoscaling
  resources:
  - horizontalpodautoscalers
  verbs:
  - list
  - watch
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - policy
  resources:
  - poddisruptionbudgets
  verbs:
  - list
  - watch
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - list
  - watch
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  - volumeattachments
  verbs:
  - list
  - watch
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - mutatingwebhookconfigurations
  - validatingwebhookconfigurations
  verbs:
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  - ingresses
  verbs:
  - list
  - watch
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - list
  - watch
---
# ClusterRoleBinding
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: datakit
subjects:
- kind: ServiceAccount
  name: datakit
  namespace: datakit
roleRef:
  kind: ClusterRole
  name: datakit
  apiGroup: rbac.authorization.k8s.io
```

- 由集群管理员执行以下命令创建监控只读权限的账户

```sh
## 执行编排yaml
kubectl apply -f account.yaml

## 确认创建成功
kubectl get sa -n datakit
NAME              SECRETS   AGE
datakit           1         3d13h
default           1         3d13h
```

- 获取服务地址

```sh
kubectl config view -o jsonpath='{"Cluster name\tServer\n"}{range .clusters[*]}{.name}{"\t"}{.cluster.server}{"\n"}{end}'
```

注意：以上得到的集群服务地址，将用于 kubernetes 采集器的 `url` 配置项中

- 获取认证 token 和证书

```sh
## 获取token
kubectl get secrets -n datakit -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='datakit')].data.token}"| base64 --decode > token

## 获取CA证书
kubectl get secrets -n datakit -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='datakit')].data.ca\\.crt}" | base64 --decode > ca_crt.pem

## 确认结果
ls -l 
-rw-r--r--  1 liushaobo  staff   1066  6  1 15:48 ca.crt
-rw-r--r--  1 liushaobo  staff    953  6  1 15:48 token
```

注意：以上得到的结果文件路径，将分别用于 kubernetes 采集器 `bearer_token` 和 `tls_ca`配置项中：

```
bearer_token = "/path/to/your/token"
tls_ca = "/path/to/your/ca_crt.pem"
```

### 部署 kube-state-metrics

- 创建 `kube-state-metrics.yaml` 编排文件, 文件内容如下:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.1.0
  name: kube-state-metrics
  namespace: datakit
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: kube-state-metrics
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kube-state-metrics
        app.kubernetes.io/version: 2.1.0
    spec:
      containers:
      - image: pubrepo.jiagouyun.com/metrics-server/kube-state-metrics:v2.1.0
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          timeoutSeconds: 5
        name: kube-state-metrics
        ports:
        - containerPort: 8080
          name: http-metrics
        - containerPort: 8081
          name: telemetry
        readinessProbe:
          httpGet:
            path: /
            port: 8081
          initialDelaySeconds: 5
          timeoutSeconds: 5
        securityContext:
          runAsUser: 65534
      nodeSelector:
        kubernetes.io/os: linux
      serviceAccountName: datakit
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/name: kube-state-metrics
    app.kubernetes.io/version: 2.1.0
  name: kube-state-metrics
  namespace: datakit
spec:
  type: NodePort
  ports:
  - name: http-metrics
    port: 8080
    targetPort: http-metrics
    nodePort: 30022
  - name: telemetry
    port: 8081
    targetPort: telemetry
  selector:
    app.kubernetes.io/name: kube-state-metrics
```

- 执行创建

```
kubectl apply -f kube-state-metrics.yaml
```

- 获取 kube-state-metrics Exporter地址

```
## 获取任意node ip (用该地址替换下文中的kube-state-metrics Exporter地址ip)
kubectl get node -o wide
```

## 配置

### kubernetes.conf 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

### kube-state-metric.conf 配置

进入 DataKit 安装目录下的 `conf.d/prom` 目录，创建 `kube_state_metric.conf` 文件，内容如下：

注意：修改 kube-state-metrics Exporter 地址

```
[[inputs.prom]]
  ## kube-state-metrics Exporter 地址
  url = "http://172.16.2.41:30828/metrics"

  # 只采集 counter 和 gauge 类型的指标
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  metric_name_filter = [
    "kube_daemonset_status_current_number_scheduled",
    "kube_daemonset_status_desired_number_scheduled",
    "kube_daemonset_status_number_available",
    "kube_daemonset_status_number_misscheduled",
    "kube_daemonset_status_number_ready",
    "kube_daemonset_status_number_unavailable",
    "kube_daemonset_updated_number_scheduled",

    "kube_deployment_spec_paused",
    "kube_deployment_spec_strategy_rollingupdate_max_unavailable",
    "kube_deployment_spec_strategy_rollingupdate_max_surge",
    "kube_deployment_status_replicas",
    "kube_deployment_status_replicas_available",
    "kube_deployment_status_replicas_unavailable",
    "kube_deployment_status_replicas_updated",
    "kube_deployment_status_condition",
    "kube_deployment_spec_replicas",

    "kube_endpoint_address_available",
    "kube_endpoint_address_not_ready",

    "kube_persistentvolumeclaim_status_phase",
    "kube_persistentvolumeclaim_resource_requests_storage_bytes",
    "kube_persistentvolumeclaim_access_mode",

    "kube_persistentvolume_status_phase",
    "kube_persistentvolume_capacity_bytes",

    "kube_secret_type",

    "kube_replicaset_status_replicas",
    "kube_replicaset_status_fully_labeled_replicas",
    "kube_replicaset_status_ready_replicas",
    "kube_replicaset_status_observed_generation",

    "kube_statefulset_status_replicas",
    "kube_statefulset_status_replicas_current",
    "kube_statefulset_status_replicas_ready",
    "kube_statefulset_status_replicas_updated",
    "kube_statefulset_status_observed_generation",
    "kube_statefulset_replicas",

    "kube_hpa_spec_max_replicas",
    "kube_hpa_spec_min_replicas",
    "kube_hpa_spec_target_metric",
    "kube_hpa_status_current_replicas",
    "kube_hpa_status_desired_replicas",
    "kube_hpa_status_condition",

    "kube_cronjob_status_active",
    "kube_cronjob_spec_suspend",
    "kube_cronjob_status_last_schedule_time",

    "kube_job_status_succeeded",
    "kube_job_status_failed",
    "kube_job_status_active",
    "kube_job_complete",
  ]

  interval = "10s"

  ## 自定义指标集名称
  [[inputs.prom.measurements]]
    ## daemonset
    prefix = "kube_daemonset_"
    name = "kube_daemonset"

  [[inputs.prom.measurements]]
    ## daemonset
    prefix = "kube_deployment_"
    name = "kube_deployment"

  [[inputs.prom.measurements]]
    ## endpoint
    prefix = "kube_endpoint_"
    name = "kube_endpoint"

  [[inputs.prom.measurements]]
    ## persistentvolumeclaim
    prefix = "kube_persistentvolumeclaim_"
    name = "kube_persistentvolumeclaim"

  [[inputs.prom.measurements]]
    ## persistentvolumeclaim
    prefix = "kube_persistentvolume_"
    name = "kube_persistentvolume"

  [[inputs.prom.measurements]]
    ## secret
    prefix = "kube_secret_"
    name = "kube_secret"

  [[inputs.prom.measurements]]
    ## replicaset
    prefix = "kube_replicaset_"
    name = "kube_replicaset"

  [[inputs.prom.measurements]]
    ## hpa
    prefix = "kube_hpa_"
    name = "kube_hpa"

  [[inputs.prom.measurements]]
    ## cronjob
    prefix = "kube_cronjob_"
    name = "kube_cronjob"

  [[inputs.prom.measurements]]
    ## job
    prefix = "kube_job_"
    name = "kube_job"

  ## 自定义Tags
  [inputs.prom.tags]
    #tag1 = "value1"
    #tag2 = "value2"
```

配置好后，重启 DataKit 即可。

### 开启 DataKit 选举

关于 Kubernetes 数据采集，我们建议开启选举来加以保护，编辑 datakit.conf，将如下配置开启即可：

```
enable_election = true
```

关于 DataKit 选举，参见[这里](election)。

## k8s 集群内部署使用 datakit

参考 [datakit daemonset deploy](datakit-daemonset-deploy)


## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

### kube_daemonset

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| daemonset | daemonset name |
| namespace | namespace       |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  status_current_number_scheduled   | The number of nodes running at least one daemon pod and are supposed to     |  int   |  -  |
|  status_desired_number_scheduled   | The number of nodes that should be running the daemon pod     |  int   |  -  |
|  status_number_available   | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available     |  int   |  -  |
|  status_desired_number_scheduled   | The number of nodes that should be running the daemon pod     |  int   |  -  |
|  status_number_misscheduled   | The number of nodes running a daemon pod but are not supposed to     |  int   |  -  |
|  status_number_ready   | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready     |  int   |  -  |
|  status_number_unavailable   | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available     |  int   |  -  |
|  updated_number_scheduled   | The total number of nodes that are running updated daemon pod     |  int   |  -  |

### kube_deployment

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| condition | condition |
| deployment | deployment name |
| namespace | namespace      |
| status | status      |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  spec_paused   | Whether the deployment is paused and will not be processed by the deployment controller    |  int   |  -  |
|  spec_replicas   | The number of replicas per deployment    |  int   |  -  |
|  spec_strategy_rollingupdate_max_surge   | Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a deployment    |  int   |  -  |
|  spec_strategy_rollingupdate_max_unavailable   | Maximum number of unavailable replicas during a rolling update of a deployment    |  int   |  -  |
|  status_number_misscheduled   | Number of desired pods for a deployment    |  int   |  -  |
|  status_condition   | The current status conditions of a deployment     |  int   |  -  |
|  status_replicas   | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available     |  int   |  -  |
|  status_replicas_available   | The number of available replicas per deployment   |  int   |  -  |
|  status_replicas_unavailable   | The number of unavailable replicas per deployment    |  int   |  -  |
|  status_replicas_updated   | The number of updated replicas per deployment     |  int   |  -  |

### kube_endpoint

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| endpoint | endpoint name |
| namespace | namespace      |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  address_available   | Number of addresses available in endpoint    |  int   |  -  |
|  address_not_ready   | Number of addresses not ready in endpoint    |  int   |  -  |

### kube_persistentvolume

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| persistentvolume | persistentvolume name |
| phase | phase      |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  capacity_bytes   | Persistentvolume capacity in bytes  |  int   |  -  |
|  status_phase   | The phase indicates if a volume is available, bound to a claim, or released by a claim    |  int   |  -  |

### kube_persistentvolumeclaim

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| access_mode | access mode |
| namespace | namespace |
| persistentvolumeclaim | persistentvolumeclaim name |
| phase | phase      |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  access_mode   | The access mode(s) specified by the persistent volume claim  |  int   |  -  |
|  resource_requests_storage_bytes   | The capacity of storage requested by the persistent volume claim   |  int   |  -  |
|  status_phase   | The phase the persistent volume claim is currently in   |  int   |  -  |

### kube_replicaset

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| namespace | namespace |
| replicaset | replicaset name |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  status_fully_labeled_replicas   | The number of fully labeled replicas per ReplicaSet |  int   |  -  |
|  status_observed_generation   | Number of desired pods for a ReplicaSet   |  int   |  -  |
|  status_ready_replicas   | The number of ready replicas per ReplicaSet  |  int   |  -  |
|  status_replicas   | The number of replicas per ReplicaSet   |  int   |  -  |

### kube_secret

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| namespace | namespace |
| secret | secret name |
| type | type |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  type   | Type about secret |  int   |  -  |

### kube_cronjob

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| namespace | namespace |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  status_active   | Active holds pointers to currently running jobs |  int   |  -  |
|  spec_suspend   | Suspend flag tells the controller to suspend subsequent executions  |  int   |  -  |
|  status_last_schedule_time   | LastScheduleTime keeps information of when was the last time the job was successfully scheduled  |  int   |  -  |

### kube_job

-  标签

| 标签名 | 描述    |
|  ----  | --------|
| namespace | namespace |

- 指标列表

| 指标 | 描述| 数据类型 | 单位   |
| ---- |---- | :---:    | :----: |
|  status_active   | The number of actively running pods |  int   |  -  |
|  status_failed   | The number of pods which reached Phase Failed  |  int   |  -  |
|  status_succeeded   | The number of pods which reached Phase Succeeded  |  int   |  -  |
|  complete   | The job has completed its execution   |  int   |  -  |
=======
### `kube_daemonset`

-  标签

| 标签名      | 描述           |
| ----        | --------       |
| `daemonset` | daemonset name |
| `namespace` | namespace      |

- 指标列表

| 指标                              | 描述                                                                                                                   | 数据类型 | 单位   |
| ----                              | ----                                                                                                                   | :---:    | :----: |
| `status_current_number_scheduled` | The number of nodes running at least one daemon pod and are supposed to                                                | int      | -      |
| `status_desired_number_scheduled` | The number of nodes that should be running the daemon pod                                                              | int      | -      |
| `status_number_available`         | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available | int      | -      |
| `status_desired_number_scheduled` | The number of nodes that should be running the daemon pod                                                              | int      | -      |
| `status_number_misscheduled`      | The number of nodes running a daemon pod but are not supposed to                                                       | int      | -      |
| `status_number_ready`             | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready     | int      | -      |
| `status_number_unavailable`       | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available        | int      | -      |
| `updated_number_scheduled`        | The total number of nodes that are running updated daemon pod                                                          | int      | -      |

### `kube_deployment`

-  标签

| 标签名       | 描述            |
| ----         | --------        |
| `condition`  | condition       |
| `deployment` | deployment name |
| `namespace`  | namespace       |
| `status`     | status          |

- 指标列表

| 指标                                          | 描述                                                                                                                          | 数据类型 | 单位   |
| ----                                          | ----                                                                                                                          | :---:    | :----: |
| `spec_paused`                                 | Whether the deployment is paused and will not be processed by the deployment controller                                       | int      | -      |
| `spec_replicas`                               | The number of replicas per deployment                                                                                         | int      | -      |
| `spec_strategy_rollingupdate_max_surge`       | Maximum number of replicas that can be scheduled above the desired number of replicas during a rolling update of a deployment | int      | -      |
| `spec_strategy_rollingupdate_max_unavailable` | Maximum number of unavailable replicas during a rolling update of a deployment                                                | int      | -      |
| `status_number_misscheduled`                  | Number of desired pods for a deployment                                                                                       | int      | -      |
| `status_condition`                            | The current status conditions of a deployment                                                                                 | int      | -      |
| `status_replicas`                             | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available               | int      | -      |
| `status_replicas_available`                   | The number of available replicas per deployment                                                                               | int      | -      |
| `status_replicas_unavailable`                 | The number of unavailable replicas per deployment                                                                             | int      | -      |
| `status_replicas_updated`                     | The number of updated replicas per deployment                                                                                 | int      | -      |

### `kube_endpoint`

-  标签

| 标签名      | 描述          |
| ----        | --------      |
| `endpoint`  | endpoint name |
| `namespace` | namespace     |

- 指标列表

| 指标                | 描述                                      | 数据类型 | 单位   |
| ----                | ----                                      | :---:    | :----: |
| `address_available` | Number of addresses available in endpoint | int      | -      |
| `address_not_ready` | Number of addresses not ready in endpoint | int      | -      |

### `kube_persistentvolume`

-  标签

| 标签名             | 描述                  |
| ----               | --------              |
| `persistentvolume` | persistentvolume name |
| `phase`            | phase                 |

- 指标列表

| 指标             | 描述                                                                                   | 数据类型 | 单位   |
| ----             | ----                                                                                   | :---:    | :----: |
| `capacity_bytes` | Persistentvolume capacity in bytes                                                     | int      | -      |
| `status_phase`   | The phase indicates if a volume is available, bound to a claim, or released by a claim | int      | -      |

### `kube_persistentvolumeclaim`

-  标签

| 标签名                  | 描述                       |
| ----                    | --------                   |
| `access_mode`           | access mode                |
| `namespace`             | namespace                  |
| `persistentvolumeclaim` | persistentvolumeclaim name |
| `phase`                 | phase                      |

- 指标列表

| 指标                              | 描述                                                             | 数据类型 | 单位   |
| ----                              | ----                                                             | :---:    | :----: |
| `access_mode`                     | The access mode(s) specified by the persistent volume claim      | int      | -      |
| `resource_requests_storage_bytes` | The capacity of storage requested by the persistent volume claim | int      | -      |
| `status_phase`                    | The phase the persistent volume claim is currently in            | int      | -      |

### `kube_replicaset`

-  标签

| 标签名       | 描述            |
| ----         | --------        |
| `namespace`  | namespace       |
| `replicaset` | replicaset name |

- 指标列表

| 指标                            | 描述                                                | 数据类型 | 单位   |
| ----                            | ----                                                | :---:    | :----: |
| `status_fully_labeled_replicas` | The number of fully labeled replicas per ReplicaSet | int      | -      |
| `status_observed_generation`    | Number of desired pods for a ReplicaSet             | int      | -      |
| `status_ready_replicas`         | The number of ready replicas per ReplicaSet         | int      | -      |
| `status_replicas`               | The number of replicas per ReplicaSet               | int      | -      |

### `kube_secret`

-  标签

| 标签名      | 描述        |
| ----        | --------    |
| `namespace` | namespace   |
| `secret`    | secret name |
| `type`      | type        |

- 指标列表

| 指标   | 描述              | 数据类型 | 单位   |
| ----   | ----              | :---:    | :----: |
| `type` | Type about secret | int      | -      |

### `kube_cronjob`

-  标签

| 标签名      | 描述      |
| ----        | --------  |
| `namespace` | namespace |

- 指标列表

| 指标                        | 描述                                                                                            | 数据类型 | 单位   |
| ----                        | ----                                                                                            | :---:    | :----: |
| `status_active`             | Active holds pointers to currently running jobs                                                 | int      | -      |
| `spec_suspend`              | Suspend flag tells the controller to suspend subsequent executions                              | int      | -      |
| `status_last_schedule_time` | LastScheduleTime keeps information of when was the last time the job was successfully scheduled | int      | -      |

### `kube_job`

-  标签

| 标签名      | 描述      |
| ----        | --------  |
| `namespace` | namespace |

- 指标列表

| 指标               | 描述                                             | 数据类型 | 单位   |
| ----               | ----                                             | :---:    | :----: |
| `status_active`    | The number of actively running pods              | int      | -      |
| `status_failed`    | The number of pods which reached Phase Failed    | int      | -      |
| `status_succeeded` | The number of pods which reached Phase Succeeded | int      | -      |
| `complete`         | The job has completed its execution              | int      | -      |
