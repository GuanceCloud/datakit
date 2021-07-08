{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# {{.InputName}}

Kubernetes 集群指标采集，主要用于收集各种资源指标

## k8s 集群内部署使用 datakit

参考 [datakit daemonset deploy](datakit-daemonset-deploy)
kube-state-metrics 指标参考 [kube-state-metrics](kube-state-metrics)

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

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
