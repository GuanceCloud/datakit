---
title: 'Kubernetes CRD 方式配置容器日志采集'
summary: '基于 Kubernetes CRD 方式配置容器日志采集'
tags:
  - '日志'
  - '容器'
  - 'KUBERNETES'
__int_icon:    'icon/kubernetes/'
---

DataKit 通过 Kubernetes Custom Resource Definition (CRD) 提供了一种声明式的容器日志采集配置方式。用户可以通过创建 `ClusterLoggingConfig` 资源来自动配置 DataKit 的日志采集，无需手动修改 DataKit 配置文件或重启 DataKit。

## 前置要求 {#prerequisites}

- Kubernetes 集群版本 1.16+
- DataKit [:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0) 或更新版本
- 集群管理员权限（用于注册 CRD）

## 使用流程 {#usage-workflow}

1. 注册 Kubernetes CRD
1. 创建 CRD 资源，自动应用采集配置
1. 为 DataKit 配置 CRD 相关的 RBAC 权限，启动 DataKit 服务

### 注册 Kubernetes CRD {#register-kubernetes-crd}

使用以下 YAML 注册 `ClusterLoggingConfig` CRD：

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: clusterloggingconfigs.logging.datakits.io
  labels:
    app: datakit-logging-config
    version: v1alpha1
spec:
  group: logging.datakits.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              type: object
              required:
                - selector
              properties:
                selector:
                  type: object
                  properties:
                    namespaceRegex:
                      type: string
                    podRegex:
                      type: string
                    podLabelSelector:
                      type: string
                    containerRegex:
                      type: string
                podTargetLabels:
                  type: array
                  items:
                    type: string
                configs:
                  type: array
                  items:
                    type: object
                    required:
                      - source
                      - type
                    properties:
                      source:
                        type: string
                      type:
                        type: string
                      disable:
                        type: boolean
                      path:
                        type: string
                      multiline_match:
                        type: string
                      pipeline:
                        type: string
                      storage_index:
                        type: string
                      tags:
                        type: object
                        additionalProperties:
                          type: string
  scope: Cluster
  names:
    plural: clusterloggingconfigs
    singular: clusterloggingconfig
    kind: ClusterLoggingConfig
    shortNames:
      - logging
```

应用 CRD：

```bash
kubectl apply -f clusterloggingconfig-crd.yaml
```

验证 CRD 注册：

```bash
kubectl get crd clusterloggingconfigs.logging.datakits.io
```

### 创建 CRD 配置资源 {#create-crd-configuration-resource}

以下示例配置采集 `test01` 命名空间中所有以 `logging` 开头的 Pod 的日志文件：

```yaml
apiVersion: logging.datakits.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: nginx-logs
spec:
  selector:
    namespaceRegex: "^(test01)$"
    podRegex: "^(logging.*)$"

  podTargetLabels:
    - app
    - version

  configs:
    - source: "nginx-access"
      type: "file"
      path: "/var/log/nginx/access.log"
      pipeline: "nginx-access.p"
      tags:
        log_type: "access"
        component: "nginx"
        
    - source: "nginx-error"
      type: "file"
      path: "/var/log/nginx/error.log"
      pipeline: "nginx-error.p"
      tags:
        log_type: "error"
        component: "nginx"
```

应用配置：

```bash
kubectl apply -f logging-config.yaml
```

#### 配置详解 {#configuration-details}

- `selector` 选择器配置

选择器用于匹配目标 Pod 和容器，所有条件为 **AND** 关系。

| 字段               | 类型   | 必填   | 说明                                      | 示例                                 |
| ------             | ------ | ------ | ------                                    | ------                               |
| `namespaceRegex`   | string | 否     | 命名空间名称正则匹配                      | `"^(default\|nginx)$"`               |
| `podRegex`         | string | 否     | Pod 名称正则匹配                          | `"^(nginx-log-demo.*)$"`             |
| `podLabelSelector` | string | 否     | Pod 标签选择器（逗号分隔的 key=value 对） | `"app=nginx,environment=production"` |
| `containerRegex`   | string | 否     | 容器名称正则匹配                          | `"^(nginx\|app-container)$"`         |

选择器示例组合：

```yaml
selector:
  namespaceRegex: "^(production|staging)$"  # 匹配 production 或 staging 命名空间
  podLabelSelector: "app=web-server"        # 匹配包含 app=web-server 标签的 Pod
  containerRegex: "^(app|web)$"             # 匹配名为 app 或 web 的容器
```

- `podTargetLabels` Pod 标签传播

| 字段              | 类型     | 必填   | 说明                                 | 示例                                |
| ------            | ------   | ------ | ------                               | ------                              |
| `podTargetLabels` | []string | 否     | 从 Pod Labels 复制到日志标签的键列表 | `["app", "version", "environment"]` |

- `configs` 采集配置

| 字段              | 类型              | 必填     | 说明                                             | 示例                                           |
| ------            | ------            | ------   | ------                                           | ------                                         |
| `type`            | string            | 是       | 采集类型：`file` - 文件日志，`stdout` - 标准输出 | `"file"`                                       |
| `source`          | string            | 是       | 日志来源标识，用于区分不同日志流                 | `"nginx-access"`                               |
| `path`            | string            | 条件必填 | 日志文件路径（支持 glob 模式），type=file 时必填 | `"/var/log/nginx/*.log"`                       |
| `disable`         | boolean           | 否       | 是否禁用此采集配置                               | `false`                                        |
| `multiline_match` | string            | 否       | 多行日志起始行的正则表达式                       | `"^\\d{4}-\\d{2}-\\d{2}"`                      |
| `pipeline`        | string            | 否       | 日志解析管道配置文件名称                         | `"nginx-access.p"`                             |
| `storage_index`   | string            | 否       | 日志存储的索引名称                               | `"app-logs"`                                   |
| `tags`            | map[string]string | 否       | 附加到日志的标签键值对                           | `{"log_type": "access", "component": "nginx"}` |

### 添加相关 RBAC 配置 {#add-rbac-configuration}

在 DataKit 的 ClusterRole 中添加以下权限：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
  # 原有的其他权限
  - apiGroups: ["logging.datakits.io"]
    resources: ["clusterloggingconfigs"]
    verbs: ["get", "list", "watch"]
```

完整 RBAC 配置示例：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: datakit
rules:
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["nodes", "nodes/stats", "nodes/metrics", "namespaces", "pods", "pods/log", "events", "services", "endpoints", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "daemonsets", "statefulsets", "replicasets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: [ "get", "list", "watch"]
- apiGroups: ["monitoring.coreos.com"]
  resources: ["podmonitors", "servicemonitors"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["logging.datakits.io"]
  resources: ["clusterloggingconfigs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
```

## 示例应用 {#example-application}

以下是一个完整的 CRD 测试应用示例：

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: test01

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: logging-deployment
  namespace: test01
  labels:
    app: logging
    version: v1.0
    environment: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: logging
  template:
    metadata:
      labels:
        app: logging
        version: v1.0
        environment: test
    spec:
      containers:
      - name: demo
        image: ubuntu:22.04
        env:
        resources:
          limits:
            cpu: "200m"
            memory: "100Mi"
          requests:
            cpu: "100m"
            memory: "50Mi"
        command: ["/bin/bash", "-c", "--"]
        args:
        - |
          mkdir -p /tmp/opt/abc;
          i=1;
          while true; do
            echo "Writing logs to file ${i}.log";
            for ((j=1;j<=10000;j++)); do
              echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/abc/file_${i}.log;
              sleep 1;
            done;
            echo "Finished writing 10000 lines to file_${i}.log";
            i=$((i+1));
          done
```

应用部署：

```bash
kubectl apply -f test-application.yaml
kubectl apply -f logging-config.yaml
```

## FAQ {#faq}

- 是否支持 CRD 动态创建、变更或删除？

支持。DataKit 能够根据 CRD 的状态变化，动态调整日志与字段的采集行为。当 CRD 被创建或更新时，相关配置会自动应用到所有匹配的容器上；若 CRD 被删除，则当前正在使用该配置的日志采集任务会终止，但容器标准输出（stdout）仍会按默认配置继续采集。

- CRD 配置和 Pod Annotations 配置哪一个优先级更高？

Pod Annotations 配置优先级更高。如果一个容器同时匹配到 CRD 配置，并且其所属 Pod 包含 `datakit/logs` 注解配置，最终将采用 Pod Annotations 中的配置，CRD 配置将不生效。

- CRD 配置变更后多久生效？

配置变更的空窗期最长 1 分钟。

- 多个 ClusterLoggingConfig 匹配同一个 Pod 时如何处理？

理论上会应用最先创建的、ResourceVersion 最小的 ClusterLoggingConfig，尽量避免此类情况。

- 采集容器内的日志文件需要添加挂载吗？

从[:octicons-tag-24: Version-1.84.0](../datakit/changelog-2025.md#cl-1.84.0)开始，对于普通 Docker 模式或 Containerd 运行时（不包括 CRI-O），无需挂载即可采集容器内日志文件。对于 CRI-O 运行时，Docker 对路径使用了 tmpfs mount，需要添加 `emptyDir` 挂载。
