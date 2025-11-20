---
title     : 'Log Sidecar'
summary   : 'Sidecar 形式的日志采集'
tags:
  - 'KUBERNETES'
  - '日志'
  - '容器'
__int_icon      : 'icon/kubernetes'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


:material-kubernetes:

---

为了便于在 Kubernetes Pod 中采集应用容器的日志，提供一个轻量的日志采集客户端，以 sidecar 方式挂载到 Pod 中，并将采集到的日志发送给 DataKit。

## 使用 {#using}

分成两部分，一是配置 DataKit 开启相应的日志接收功能，二是配置和启动 logfwd 采集。

## 配置 {#config}

### DataKit 配置 {#datakit-conf}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    需要先开启 [logfwdserver](logfwdserver.md)，进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `logfwdserver.conf.sample` 并命名为 `logfwdserver.conf`。示例如下：

    ``` toml hl_lines="1"
    [inputs.logfwdserver] # 注意这里是 logfwdserver 的配置
      ## logfwd 接收端监听地址和端口
      address = "0.0.0.0:9533"

      [inputs.logfwdserver.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入 logfwdserver 采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### logfwd 使用和配置（1.86.0 及以后） {#config-1-86-0}
<!-- markdownlint-enable -->

> logfwd 推荐在 Kubernetes Serverless 环境使用，如果已经部署了 DaemonSet DataKit，再使用 logfwd 可能会数据重复。

自 logfwd 1.86.0 版本起，整体使用方式进一步简化，并移除了部分繁琐配置，主要新增能力如下：

- 支持通过 DataKit-Operator 拉取 `ClusterLoggingConfig` CRD，自动匹配 Pod 并热加载采集配置；
- 同时兼容手写环境变量配置 (`LOGFWD_LOG_CONFIGS`)，满足无 DataKit-Operator 或调试场景；
- 采集任务统一通过 WebSocket 与 DataKit `inputs.logfwdserver` 通信，连接失败时自动重连（每秒重试）；
- 自动补充 Pod 元数据（`pod_name`、`namespace`、`pod_ip`）及目标 Label，可与旧版 volume/挂载方案无缝复用。

#### 启动方式概览 {#config-1-86-0-overview}

| 场景                          | 关键变量                                          | 说明                                                                                                                      |
| :---                          | :---                                              | :---                                                                                                                      |
| 搭配 DataKit-Operator（推荐） | `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` + Pod metadata | 由 DataKit-Operator 返回匹配的 CRD JSON，logfwd 自动创建/刷新 tailer；需要在 `ClusterLoggingConfig` 中声明日志路径、Pipeline 等。 |
| 手动配置                      | `LOGFWD_LOG_CONFIGS`                              | 与旧版 JSON 语义一致，但通过环境变量传入，适合开发/过渡场景，可与 DataKit-Operator 共存（手动配置优先级更高）。                   |

仍需像旧版一样提前准备日志文件的共享 `volume`/`volumeMount`；logfwd 只做监听，不会创建挂载。

#### 全局环境变量 {#config-1-86-0-globals}

| 环境变量名                         | 配置项含义                                                                                                                                                                              |
| :---                               | :---                                                                                                                                                                                    |
| `LOGFWD_LOG_LEVEL`                 | 运行日志级别，默认 `info`，设为 `debug` 可查看更多调试输出。                                                                                                                            |
| `LOGFWD_DATAKIT_HOST`              | DataKit 实例地址（IP 或可解析域名）。                                                                                                                                                   |
| `LOGFWD_DATAKIT_PORT`              | DataKit `logfwdserver` 监听端口，例如 `9533`。                                                                                                                                          |
| `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` | DataKit-Operator Endpoint，形如 `datakit-operator.datakit.svc:443` 或 `https://datakit-operator.datakit.svc:443`，用于查询 CRD 配置；留空则不会尝试拉取。支持自动添加 `https://` 前缀。 |
| `LOGFWD_GLOBAL_SOURCE`             | 全局 `source`，优先级高于单条配置中的 `source` 字段。                                                                                                                                   |
| `LOGFWD_GLOBAL_SERVICE`            | 全局 `service`，若单条配置中未指定 `service`，则使用全局值；若全局值也为空，则回退为 `source`。                                                                                         |
| `LOGFWD_GLOBAL_STORAGE_INDEX`      | 全局 `storage_index`，优先级高于单条配置中的 `storage_index` 字段。                                                                                                                     |
| `LOGFWD_POD_NAME`                  | 自动写入 `pod_name` tag，通常通过 Downward API 注入。                                                                                                                                   |
| `LOGFWD_POD_NAMESPACE`             | 自动写入 `namespace` tag。                                                                                                                                                              |
| `LOGFWD_POD_IP`                    | 自动写入 `pod_ip` tag，便于定位容器实例。                                                                                                                                               |

> 提示：如果需要附加更多 tag，可在 Pod 中挂载 `/etc/podinfo/labels` 文件（由 Datakit-Operator 注入 logfwd sidecar 时会自动添加），logfwd 会解析并和 CRD 中的 `podTargetLabels` 对齐。

#### 采集配置 {#config-1-86-0-log-configs}

logfwd 支持两种配置方式，按优先级从高到低：

1. **手动配置（`LOGFWD_LOG_CONFIGS`）**：通过环境变量传入 JSON 字符串，结构与旧版 `loggings` 子项基本一致。存在手动配置时，logfwd 会立即创建 tailer，并在进程存活期间保持该配置；删除变量或清空内容后需重启容器以释放。
1. **DataKit-Operator CRD**：当指定 `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` 时，logfwd 每分钟调用一次 DataKit-Operator API，通过 MD5 校验配置内容判定是否需要热更新。配置变更后会自动重新创建 tailer，无需重启容器。

> **注意**：如果同时存在手动配置和 CRD 配置，且两者指向同一日志路径，会导致重复采集。建议优先使用 CRD 配置，手动配置仅用于调试或特殊场景。

`LOGFWD_LOG_CONFIGS` 字段结构示例如下：

``` json
[
  {
    "type": "file",
    "disable": false,
    "source": "nginx-access",
    "service": "nginx",
    "path": "/var/log/nginx/access.log",
    "pipeline": "nginx-access.p",
    "storage_index": "app-logs",
    "multiline_match": "^\\d{4}-\\d{2}-\\d{2}",
    "remove_ansi_escape_codes": false,
    "from_beginning": false,
    "character_encoding": "utf-8",
    "tags": {
      "env": "production",
      "team": "backend"
    }
  }
]
```

| 字段                       | 类型    | 必填     | 说明                                                                                                  | 示例                      |
| ------                     | ------  | ------   | ------                                                                                                | ------                    |
| `type`                     | string  | 是       | logfwd 采集类型只能是 `"file"`                                                                        | `"file"`                  |
| `disable`                  | boolean | 否       | 是否禁用此采集配置                                                                                    | `false`                   |
| `source`                   | string  | 是       | 日志来源标识，用于区分不同日志流                                                                      | `"nginx-access"`          |
| `service`                  | string  | 否       | 日志隶属的服务，默认值为日志来源（source）                                                            | `"nginx"`                 |
| `path`                     | string  | 条件必填 | 日志文件路径（支持 glob 模式），type=file 时必填                                                      | `"/var/log/nginx/*.log"`  |
| `multiline_match`          | string  | 否       | 多行日志起始行的正则表达式，注意 JSON 中需要转义反斜杠                                                | `"^\\d{4}-\\d{2}-\\d{2}"` |
| `pipeline`                 | string  | 否       | 日志解析管道配置文件名称（需在 DataKit 端配置）                                                       | `"nginx-access.p"`        |
| `storage_index`            | string  | 否       | 日志存储的索引名称                                                                                    | `"app-logs"`              |
| `remove_ansi_escape_codes` | boolean | 否       | 是否删除日志数据的 ANSI 转义字符（颜色代码等）                                                        | `false`                   |
| `from_beginning`           | boolean | 否       | 是否从文件首部开始采集日志（默认从文件末尾开始）                                                      | `false`                   |
| `character_encoding`       | string  | 否       | 字符编码，支持 `utf-8`, `utf-16le`, `utf-16be`, `gbk`, `gb18030` 或空字符串（自动检测）。默认为空即可 | `"utf-8"`                 |
| `tags`                     | object  | 否       | 额外的标签键值对，会附加到每条日志记录上                                                              | `{"env": "prod"}`         |


当配置了 `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` 时，logfwd 会根据 `LOGFWD_POD_NAMESPACE`、`LOGFWD_POD_NAME` 以及 `pod_labels`（可选，需挂载 `/etc/podinfo/labels` 文件）向 DataKit-Operator 发起请求。只要某条 `ClusterLoggingConfig` CRD 规则匹配当前 Pod，就会返回对应的 `configs` JSON 并触发热更新。

CRD 配置示例：

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^(default|production)$"
    podRegex: "^(nginx-.*)$"
    podLabelSelector: "app=nginx,env=production"
    containerRegex: "^(nginx|app)$"
  
  podTargetLabels:
    - app
    - version
    - team
  
  configs:
    - type: "file"
      source: "nginx-access"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
      storage_index: "app-logs"
      tags:
        log_type: "access"
        component: "nginx"
    
    - type: "file"
      source: "nginx-error"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
      storage_index: "app-logs"
      tags:
        log_type: "error"
        component: "nginx"
```

CRD 选择器说明：

| 字段               | 类型   | 必填   | 说明                                      | 示例                                 |
| ------             | ------ | ------ | ------                                    | ------                               |
| `namespaceRegex`   | string | 否     | 命名空间名称正则匹配（所有条件为 AND 关系）| `"^(default\|production)$"`          |
| `podRegex`         | string | 否     | Pod 名称正则匹配                          | `"^(nginx-.*)$"`                     |
| `podLabelSelector` | string | 否     | Pod 标签选择器（逗号分隔的 key=value 对） | `"app=nginx,environment=production"` |
| `containerRegex`   | string | 否     | 容器名称正则匹配                          | `"^(nginx\|app-container)$"`         |

`podTargetLabels`：指定需要从 Pod Labels 中提取并附加到日志的标签键列表。logfwd 会读取 `/etc/podinfo/labels` 文件（由 Downward API 或 DataKit-Operator 注入），提取匹配的标签并添加到日志的 `tags` 中。

配置热更新机制：

- logfwd 每分钟轮询一次 DataKit-Operator API
- 通过计算配置内容的 MD5 值判断是否有变更
- 配置变更后自动停止旧 tailer 并创建新 tailer，无需重启容器
- 配置变更通常在 1 分钟内生效

<!-- markdownlint-disable MD046 -->
??? topic

    - 需要在业务 Pod/sidecar 中预先用 `volumes`/`volumeMounts` 共享日志目录（如 `emptyDir`），否则 logfwd 无法访问日志文件。
    - `LOGFWD_LOG_CONFIGS` 与 CRD 配置相互独立，若两者指向同一路径会导致重复采集。
    - DataKit-Operator 支持为目标 Pod 自动注入 logfwd sidecar 及挂载，具体请查看 DataKit-Operator [文档](../datakit/datakit-operator.md)。
<!-- markdownlint-enable -->

#### ClusterLoggingConfig CRD 选择器支持 {#config-1-86-0-crd-selector}

logfwd 通过 DataKit-Operator 查询 `ClusterLoggingConfig` CRD 时，支持以下选择器字段用于匹配目标 CRD 及日志采集配置：

| 选择器字段          | 说明                                                                                 | 示例                                    |
| :------------------ | :-------------------------------------------------------------------                 | :-------------------------------------- |
| `namespaceRegex`    | 命名空间名称正则匹配，使用 logfwd 容器的 `LOGFWD_POD_NAMESPACE` 环境变量作为查询参数 | `"^(default)$"`                         |
| `podNameRegex`      | Pod 名称正则匹配，使用 logfwd 容器的 `LOGFWD_POD_NAME` 环境变量作为查询参数          | `"^(nginx-app.*)$"`                     |
| `podLabelSelector`  | Pod 标签选择器（前提是在 logfwd 容器内的 `/etc/podinfo/labels` 有 labels 内容）      | `"app=nginx,environment=production"`    |

<!-- markdownlint-disable MD046 -->
??? topic

    - logfwd **不支持** `containerRegex` 选择器。由于 logfwd 以 Pod Sidecar 方式运行，它只负责采集日志文件，无法区分容器名。
    - `podLabelSelector` 的使用依赖于 `/etc/podinfo/labels` 文件的存在。DataKit-Operator 在注入 logfwd sidecar 时会自动挂载该文件（通过 Downward API），如果该文件不存在或为空，`podLabelSelector` 将无法生效。
    - 所有选择器条件为 **AND** 关系，即所有指定的选择器都必须匹配，Pod 才会被选中。
<!-- markdownlint-enable -->

#### 示例：Kubernetes Pod 配置 {#config-1-86-0-example}

- 使用 DataKit-Operator CRD 配置

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-app
  namespace: default
  labels:
    app: nginx
    version: v1.0
spec:
  containers:
  - name: nginx
    image: nginx:latest
    volumeMounts:
    - name: nginx-logs
      mountPath: /var/log/nginx
  - name: logfwd
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{ .Version }}
    env:
    - name: LOGFWD_LOG_LEVEL
      value: "info"  # 可选：debug 可查看详细日志
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_DATAKIT_OPERATOR_ENDPOINT
      value: datakit-operator.datakit.svc:443
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: LOGFWD_POD_IP
      valueFrom:
        fieldRef:
          fieldPath: status.podIP
    volumeMounts:
    - name: podinfo
      mountPath: /etc/podinfo
      readOnly: true
    - name: nginx-logs
      mountPath: /var/log/nginx
      readOnly: true
  volumes:
  - name: podinfo
    downwardAPI:
      items:
      - path: "labels"
        fieldRef:
          fieldPath: metadata.labels
  - name: nginx-logs
    emptyDir: {}
```

对应的 `ClusterLoggingConfig` CRD 配置：

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^default$"
    podLabelSelector: "app=nginx"
  podTargetLabels:
    - app
    - version
  configs:
    - type: "file"
      source: "nginx-access"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
    - type: "file"
      source: "nginx-error"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
```

- 使用手动配置

若需要临时使用手动配置或调试，可添加 `LOGFWD_LOG_CONFIGS` 环境变量：

```yaml
spec:
  containers:
  - name: logfwd
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{ .Version }}
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    - name: LOGFWD_LOG_CONFIGS
      value: |
        [
          {
            "type": "file",
            "source": "app-logs",
            "path": "/var/log/app/*.log",
            "pipeline": "app.p",
            "from_beginning": false,
            "tags": {
              "env": "production"
            }
          }
        ]
    volumeMounts:
    - name: app-logs
      mountPath: /var/log/app
      readOnly: true
  volumes:
  - name: app-logs
    emptyDir: {}
```

其余挂载形态、`volumes`/`volumeMounts` 的写法、资源限制等与 1.86.0 之前保持一致，可继续参考下一节中的旧版示例。

### logfwd 使用和配置（1.86.0 之前版本） {#config-before-1-86-0}

logfwd 主配置是 JSON 格式，以下是配置示例：

``` json
[
    {
        "datakit_addr": "127.0.0.1:9533",
        "loggings": [
            {
                "logfiles":      ["<your-logfile-path>"],
                "ignore":        [],
                "storage_index": "<your-storage-index>",
                "source":        "<your-source>",
                "service":       "<your-service>",
                "pipeline":      "<your-pipeline.p>",
                "character_encoding": "",
                "multiline_match": "<your-match>",
                "tags": {}
            },
            {
                "logfiles": ["<your-logfile-path-2>"],
                "source": "<your-source-2>"
            }
        ]
    }
]
```

配置参数说明：

- `datakit_addr` 是 DataKit logfwdserver 地址，通常使用环境变量 `LOGFWD_DATAKIT_HOST` 和 `LOGFWD_DATAKIT_PORT` 进行配置

- `loggings` 为主要配置，是一个数组，子项也基本和 [logging](logging.md) 采集器相同。
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `storage_index` 指定日志存储索引
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 `$source`
    - `pipeline` Pipeline 脚本路径，如果为空将使用 `$source.p`，如果 `$source.p` 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` 选择编码，如果编码有误，会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
    - `multiline_match` 多行匹配，与 [logging](logging.md) 该项配置一样，注意因为是 JSON 格式所以不支持 3 个单引号的“不转义写法”，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
    - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

支持的环境变量：

| 环境变量名                       | 配置项含义                                                                                                              |
| :---                             | :---                                                                                                                    |
| `LOGFWD_DATAKIT_HOST`            | DataKit 地址                                                                                                            |
| `LOGFWD_DATAKIT_PORT`            | DataKit Port                                                                                                            |
| `LOGFWD_TARGET_CONTAINER_IMAGE`  | 配置目标容器的镜像名，例如 `nginx:1.22`，解析并添加相关的 tag（`image`、`image_name`、`image_short_name`、`image_tag`） |
| `LOGFWD_GLOBAL_SOURCE`           | 配置全局 source，优先级最高                                                                                             |
| `LOGFWD_GLOBAL_STORAGE_INDEX`    | 配置全局 storage_index，优先级最高                                                                                      |
| `LOGFWD_GLOBAL_SERVICE`          | 配置全局 service，优先级最高                                                                                            |
| `LOGFWD_POD_NAME`                | 指定 pod name，会 tags 中添加 `pod_name`                                                                                |
| `LOGFWD_POD_NAMESPACE`           | 指定 pod namespace，会 tags 中添加 `namespace`                                                                          |
| `LOGFWD_ANNOTATION_DATAKIT_LOGS` | 使用当前 Pod 的 Annotations `datakit/logs` 配置，优先级比 logfwd JSON 配置更高                                          |
| `LOGFWD_JSON_CONFIG`             | logfwd 主配置，即上文的 JSON 格式文本                                                                                   |

#### 安装和运行 {#install-run}

logfwd 在 Kubernetes 的部署配置分为两部分，一是 Kubernetes Pod 创建 `spec.containers` 的配置，包括注入环境变量和挂载目录。配置如下：

```yaml
spec:
  containers:
  - name: logfwd
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_ANNOTATION_DATAKIT_LOGS
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.annotations['datakit/logs']
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: LOGFWD_GLOBAL_SOURCE
      value: nginx-souce-test
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.85.0
    imagePullPolicy: Always
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
    volumeMounts:
    - mountPath: /opt/logfwd/config
      name: logfwd-config-volume
      subPath: config
    workingDir: /opt/logfwd
  volumes:
  - configMap:
      name: logfwd-config
    name: logfwd-config-volume
```

第二份配置为 logfwd 实际运行的配置，即前文提到的 JSON 格式的主配置，在 Kubernetes 中以 ConfigMap 形式存在。

根据 logfwd 配置示例，按照实际情况修改 `config`。`ConfigMap` 格式如下：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: logfwd-conf
data:
  config: |
    [
        {
            "loggings": [
                {
                    "logfiles": ["/var/log/1.log"],
                    "source": "log_source",
                    "tags": {}
                },
                {
                    "logfiles": ["/var/log/2.log"],
                    "source": "log_source2"
                }
            ]
        }
    ]
```

将两份配置集成到现有的 Kubernetes yaml 中，并使用 `volumes` 和 `volumeMounts` 将目录在 containers 内部共享，即可实现 logfwd 容器采集其他容器的日志文件。

> 注意，需要使用 `volumes` 和 `volumeMounts` 将应用容器（即示例中的 `count` 容器）的日志目录挂载和共享，以便在 logfwd 容器中能够正常访问到。`volumes` 官方说明[文档](https://kubernetes.io/docs/concepts/storage/volumes/){:target="_blank"}

完整示例如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: logfwd
spec:
  containers:
  - name: count
    image: busybox
    args:
    - /bin/sh
    - -c
    - >
      i=0;
      while true;
      do
        echo "$i: $(date)" >> /var/log/1.log;
        echo "$(date) INFO $i" >> /var/log/2.log;
        i=$((i+1));
        sleep 1;
      done
    volumeMounts:
    - name: varlog
      mountPath: /var/log
  - name: logfwd
    env:
    - name: LOGFWD_DATAKIT_HOST
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: status.hostIP
    - name: LOGFWD_DATAKIT_PORT
      value: "9533"
    - name: LOGFWD_ANNOTATION_DATAKIT_LOGS
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.annotations['datakit/logs']
    - name: LOGFWD_POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    - name: LOGFWD_POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:{{.Version}}
    imagePullPolicy: Always
    resources:
      requests:
        cpu: "200m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "2Gi"
    volumeMounts:
    - name: varlog
      mountPath: /var/log
    - mountPath: /opt/logfwd/config
      name: logfwd-config-volume
      subPath: config
    workingDir: /opt/logfwd
  volumes:
  - name: varlog
    emptyDir: {}
  - configMap:
      name: logfwd-config
    name: logfwd-config-volume

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: logfwd-config
data:
  config: |
    [
        {
            "loggings": [
                {
                    "logfiles": ["/var/log/1.log"],
                    "source": "log_source",
                    "tags": {
                        "flag": "tag1"
                    }
                },
                {
                    "logfiles": ["/var/log/2.log"],
                    "source": "log_source2"
                }
            ]
        }
    ]
```

### 性能测试 {#bench}

- 环境：

```shell
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
```

- 日志文件内容为 1000w 条 nginx 日志，文件大小 2.2GB：

```not-set
192.168.17.1 - - [06/Jan/2022:16:16:37 +0000] "GET /google/company?test=var1%20Pl HTTP/1.1" 401 612 "http://www.google.com/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36" "-"
```

- 结果：

耗时**95 秒**将所有日志读取和转发完毕，平均每秒读取 10w 条日志。

单核心 CPU 使用率峰值为 42%，以下是当时的 `top` 记录：

```shell
top - 16:32:46 up 52 days,  7:28, 17 users,  load average: 2.53, 0.96, 0.59
Tasks: 464 total,   2 running, 457 sleeping,   0 stopped,   5 zombie
%Cpu(s): 30.3 us, 33.7 sy,  0.0 ni, 34.3 id,  0.1 wa,  0.0 hi,  1.5 si,  0.0 st
MiB Mem :  15885.2 total,    985.2 free,   6204.0 used,   8696.1 buff/cache
MiB Swap:   2048.0 total,      0.0 free,   2048.0 used.   8793.3 avail Mem

    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND
1850829 root      20   0  715416  17500   8964 R  42.1   0.1   0:10.44 logfwd
```

## 延伸阅读 {#more-reading}

- [DataKit 日志采集综述](datakit-logging.md)
- [Socket 日志接入最佳实践](logging_socket.md)
- [Kubernetes 中指定 Pod 的日志采集配置](container-log.md#logging-with-annotation-or-label)
- [第三方日志接入](logstreaming.md)
- [Kubernetes 环境下 DataKit 配置方式介绍](../datakit/k8s-config-how-to.md)
- [以 DaemonSet 形式安装 DataKit](../datakit/datakit-daemonset-deploy.md)
- [在 DataKit 上部署 `logfwdserver`](logfwdserver.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
