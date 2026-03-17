# DataKit Operator

---

:material-kubernetes:

---

DataKit Operator 是 DataKit 在 Kubernetes 编排的联动项目，旨在协助 DataKit 更方便的部署，以及其他诸如验证、注入的功能。

## 概述 {#overview}

DataKit Operator 通过 Kubernetes Admission Controller 机制，为 Kubernetes 集群提供自动化注入功能，帮助用户更轻松地集成观测能力。主要功能包括：

- **DDTrace 注入**：自动为 Java 应用注入 APM 追踪代理
- **日志采集**：通过 logfwd Sidecar 自动采集容器日志
- **性能分析**：注入 Flameshot 或 Profiler 组件进行应用性能监控
- **配置管理**：支持全局配置和声明式配置两种注入方式

**核心优势**：

- **自动化部署**：无需手动修改应用 YAML，减少配置错误
- **批量管理**：通过命名空间和标签选择器实现批量注入
- **灵活配置**：支持 JSON 配置和 Annotation 精细控制
- **版本兼容**：保持向后兼容，支持平滑升级

## 先决条件 {#prerequisites}

- 推荐 Kubernetes v1.24.1 及以上版本，且能够访问互联网（下载 yaml 文件并拉取对应镜像）
- 确保启用 `MutatingAdmissionWebhook` 和 `ValidatingAdmissionWebhook` [控制器](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"}
- 确保启用了 `admissionregistration.k8s.io/v1` API

## 安装 {#install}

<!-- markdownlint-disable MD046 -->
=== "Deployment"

    下载 [*datakit-operator.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml){:target="_blank"}，步骤如下：
    
    ``` shell
    $ kubectl create namespace datakit
    $ wget https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml
    $ kubectl apply -f datakit-operator.yaml
    $ kubectl get pod -n datakit
    
    NAME                               READY   STATUS    RESTARTS   AGE
    datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
    ```

=== "Helm"

    前提条件

    * Kubernetes >= 1.14
    * Helm >= 3.0+

    ```shell
    $ helm install datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace
    ```

    查看部署状态：

    ```shell
    $ helm -n datakit list
    ```

    可以通过如下命令来升级：

    ```shell
    $ helm -n datakit get values datakit-operator -a -o yaml > values.yaml
    $ helm upgrade datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit \
        -f values.yaml
    ```

    可以通过如下命令来卸载：

    ```shell
    $ helm uninstall datakit-operator -n datakit
    ```

???+ attention

    - DataKit Operator 有严格的程序和 yaml 对应关系，如果使用一份过旧的 yaml 可能无法安装新版 DataKit Operator，请重新下载最新版 yaml。
    - 如果出现 `InvalidImageName` 报错，可以手动 pull 镜像。
<!-- markdownlint-enable -->

### 配置说明 {#jsonconfig}

DataKit Operator 配置是 JSON 格式，在 Kubernetes 中单独以 ConfigMap 存放，以环境变量方式加载到容器中。

<!-- markdownlint-disable MD046 -->
=== "DataKit Operator >= v1.8.0"

    从 DataKit-Operator v1.8.0 版本开始，推荐使用 `admission_inject_v2` 配置项。新配置采用数组结构，支持更灵活的配置方式。

    ```json
    {
        "server_listen": "0.0.0.0:9543", // operator 自身服务监听地址
        "log_level": "info",             // operator 自身日志级别
        "admission_inject_v2": {         // 注入配置 v2
            "ddtraces": [...],           // DDTrace 配置数组
            "logfwds": [...],            // 日志转发配置数组
            "flameshots": [...]          // 性能分析配置数组
        },
        "admission_mutate": {            // 配置变更
            "loggings": [...]            // 日志配置变更
        }
    }
    ```

=== "DataKit Operator < v1.8.0"

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level":     "info",
        "admission_inject": {
            "ddtrace": {...},
            "profiler": {...},
            "logfwd": {...}
        },
        "admission_mutate": {
            "loggings": [...]
        }
    }
    ```
<!-- markdownlint-enable -->

## 注入方式 {#datakit-operator-inject}

DataKit Operator 支持两种资源输入方式，分别是

1. selector 配置注入（指令式）

    通过修改 DataKit-Operator config，指定目标 Pod 的 Namespace 和 Selector，如果发现 Pod 符合条件，就执行注入。

    **优点**：不需要在目标 Pod 添加 Annotation（但是需要重启目标 Pod）

    **缺点**：范围不够精确，可能存在无效注入

1. Annotation 配置注入（声明式）

    在目标 Pod 添加 Annotation 开启自身的注入。

    **优点**：可以通过 Annotation 精确控制是否拒绝注入

    **缺点**：不能单独通过 Annotation 触发注入，仍需配置匹配规则，即在目标 Pod annotation 上开启注入外，还需要在 Operator 配置其它字段。

### Selector 配置注入 {#selectors-injection}

通过配置 `namespace_selectors` 和 `label_selectors` 可以实现批量注入。

在 `admission_inject_v2` 配置中，`namespace_selectors` 和 `label_selectors` 直接在数组项中配置，以 DDTrace 注入为例：

```json
{
    "admission_inject_v2": {
        "ddtraces": [
            {
                "namespace_selectors": ["testns"],
                "label_selectors":     ["app=log-output"],
                ...
            }
        ]
    }
}
```

- `namespace_selectors`：命名空间选择器数组，支持正则表达式匹配。如需精确匹配，请使用 `^` 和 `$` 将模式包围，例如 `^testns$`
- `label_selectors`：标签选择器数组，使用 Kubernetes Label Selector 语法

如果同时配置了这两个 selector，则目标 Pod 必须同时满足这两个条件。关于 label selector 的编写规范，可参考此[官方文档](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/#label-selectors){:target="_blank"}。

### Annotation 配置注入 {#annotation-injection}

在 Deployment 添加指定 Annotation 可以控制是否允许注入。注意 Annotation 要添加在 template 中。

支持的 Annotation 如下：

| Annotation                            | 功能描述            | 取值             | 优先级   |
| ------------                          | ----------          | ------           | -------- |
| `admission.datakit/ddtrace.enabled`   | 控制 ddtrace 注入   | `"true"/"false"` | 中       |
| `admission.datakit/logfwd.enabled`    | 控制 logfwd 注入    | `"true"/"false"` | 中       |
| `admission.datakit/flameshot.enabled` | 控制 flameshot 注入 | `"true"/"false"` | 中       |
| `admission.datakit/enabled`           | 控制所有注入功能    | `"true"/"false"` | **最高** |

示例：

```yaml
    annotations:
    admission.datakit/ddtrace.enabled: "true"
    admission.datakit/logfwd.enabled: "true"
```

<!-- markdownlint-disable MD046 -->
???+ tip

    Annotation 可用于拒绝注入（设置为 `"false"` 即可），但作为主动注入时，需做如下配置：

    1. 在 DataKit-Operator 配置中设置匹配规则（`namespace_selectors`/`label_selectors`）和相应的配置字段
    1. Pod 匹配配置中的配置 selectors
<!-- markdownlint-enable -->

### `check_annotation` 配置项说明 {#check-annotation-config}

`check_annotation` 是一个配置字段，用于控制 DataKit Operator 如何处理 Pod 上的**版本注解**。该字段的取值和行为如下：

| 取值     | 行为说明                                                                 |
|----------|--------------------------------------------------------------------------|
| `false`  | **（默认值）** 忽略 Pod 上的**版本注解**检查，直接根据选择器规则进行注入 |
| `true`   | 启用**版本注解**检查，只有匹配的 Pod 上存在版本注解才会注入             |

#### 注解类型说明 {#annotation-types}

DataKit Operator 支持两种类型的注解，它们的行为不同：

**1. 启用/禁用注解（不受 `check_annotation` 影响）**
这些注解用于控制是否启用或禁用特定功能，**不受 `check_annotation` 配置影响**：

| Annotation                            | 功能描述            | 取值             | 优先级   |
| ------------                          | ----------          | ------           | -------- |
| `admission.datakit/enabled`           | 控制所有注入功能    | `"true"/"false"` | **最高** |
| `admission.datakit/ddtrace.enabled`   | 控制 ddtrace 注入   | `"true"/"false"` | 中       |
| `admission.datakit/logfwd.enabled`    | 控制 logfwd 注入    | `"true"/"false"` | 中       |
| `admission.datakit/flameshot.enabled` | 控制 flameshot 注入 | `"true"/"false"` | 中       |

**2. 版本注解（受 `check_annotation` 影响）**
这些注解用于指定组件版本，**受 `check_annotation` 配置控制**：

| Annotation                                  | 功能描述                       | 取值            |
| --------------------------------------      | ------------------------------ | ------------    |
| `admission.datakit/java-lib.version`        | 指定 DDTrace Java Agent 版本   | 版本字符串      |
| `admission.datakit/python-lib.version`      | 指定 DDTrace Python Agent 版本 | 版本字符串      |
| `admission.datakit/java-profiler.version`   | 指定 Java Profiler 版本        | 版本字符串      |
| `admission.datakit/python-profiler.version` | 指定 Python Profiler 版本      | 版本字符串      |
| `admission.datakit/golang-profiler.version` | 指定 Golang Profiler 版本      | 版本字符串      |
| `admission.datakit/logfwd.instances`        | 指定 logfwd Sidecar 版本       | JSON 配置字符串 |

#### 注入逻辑说明 {#injection-logic}

**核心规则**：

- `admission.datakit/enabled:"false"` 拒绝所有注入（最高优先级）
- 功能特定启用注解（如 `admission.datakit/ddtrace.enabled: "false"`）拒绝该功能注入
- 当 `check_annotation: true` 时，需要对应的版本注解存在才会注入
- 当 `check_annotation: false` 时，忽略版本注解检查

**注入条件对比**：

| 条件                          | `check_annotation: true` | `check_annotation: false` |
|-------------------------------|--------------------------|---------------------------|
| 配置匹配（selector 规则）     | ✓ 必须满足              | ✓ 必须满足               |
| 启用注解不为 `"false"`        | ✓ 必须满足              | ✓ 必须满足               |
| 版本注解存在                  | ✓ 必须存在              | ✗ 可忽略                 |

#### 使用场景示例 {#use-case-examples}

1. **批量注入**（`check_annotation: false`）：
   适合需要为大量 Pod 自动注入的场景，无需在每个 Pod 上添加版本注解。

2. **精确控制**（`check_annotation: true`）：
   适合需要严格控制版本、只对明确标记的 Pod 进行注入的场景。



## 支持的注入功能列表 {#supported-operator}

| 功能           | 简述                                                                                      |
| ---            | ---                                                                                       |
| DDtrace Agent  | 注入 DDTrace 组件，参见[这里](operator-ddtrace.md)                                        |
| logfwd         | 注入 logfwd 组件采集容器内日志， 参见[这里](operator-logfwd.md)                           |
| Flameshot      | 注入 Flameshot 组件动态采集应用 Profiling， 参见[这里](operator-flameshot.md)             |
| async-profiler | 注入 async-profiler 定期采集 Java 应用的 Profiling， 参见[这里](operator-asyncprofile.md) |
| py-spy         | 注入 py-spy 采集 Python 应用的 Profiling， 参见[这里](operator-pyspy.md)                  |
| logging        | 注入日志采集配置， 参见[这里](operator-logging.md)                                        |

## Downward API {#downwardapi}

在 DataKit Operator [:octicons-tag-24: v1.4.2](operator-changelog.md#cl-1.4.2) 及以后版本，`envs` 支持 Kubernetes Downward API 的 [环境变量取值字段](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef)。现支持以下几种：

| 字段                                       | 描述                                            | 示例                                 |
| ------:                                    | :------                                         | :------                              |
| `{fieldRef.metadata.name}`                 | Pod 的名称                                      | `nginx-123`                          |
| `{fieldRef.metadata.namespace}`            | Pod 的命名空间                                  | middleware                           |
| `{fieldRef.metadata.uid}`                  | Pod 的唯一 ID                                   | 12345678-1234-1234-1234-123456789abc |
| `{fieldRef.metadata.annotations['<KEY>']}` | Pod 的注解 `<KEY>` 的值                         | metadata.annotations['myannotation'] |
| `{fieldRef.metadata.labels['<KEY>']}`      | Pod 的标签 `<KEY>` 的值                         | metadata.labels['app']               |
| `{fieldRef.spec.serviceAccountName}`       | Pod 的服务账号名称                              | default                              |
| `{fieldRef.spec.nodeName}`                 | Pod 运行时所处的节点名称                        | node-01                              |
| `{fieldRef.status.hostIP}`                 | Pod 所在节点的主 IP 地址                        | 192.168.1.1                          |
| `{fieldRef.status.hostIPs}`                | status.hostIP 的双协议栈版本                    | ["192.168.1.1", "2001:db8::1"]       |
| `{fieldRef.status.podIP}`                  | Pod 的主 IP 地址                                | 10.0.0.1                             |
| `{resourceFieldRef:limits.cpu}`            | Pod 第一个容器的 CPU Limit（单位 millicores）   | 500                                  |
| `{resourceFieldRef:limits.memory}`         | Pod 第一个容器的 Memory Limit（单位 MiB）       | 1024                                 |
| `{resourceFieldRef:requests.cpu}`          | Pod 第一个容器的 CPU Request（单位 millicores） | 200                                  |
| `{resourceFieldRef:requests.memory}`       | Pod 第一个容器的 Memory Request（单位 MiB）     | 512                                  |

举个例子，现有一个 Pod 名称是 `nginx-123`，namespace 是 `middleware`，要给它注入环境变量 `POD_NAME` 和 `POD_NAMESPACE`，参考以下：

```json
{
    "admission_inject_v2": {
        "ddtraces": [
            {
                "namespace_selectors": ["middleware"],
                "language":            "java",
                "image":               "dd-lib-java-init:latest",
                "envs": {
                    "POD_NAME":      "{fieldRef:metadata.name}",
                    "POD_NAMESPACE": "{fieldRef:metadata.namespace}"
                }
            }
        }
    }
}
```

最终在该 Pod 可以看到：

``` shell
$ env | grep POD
POD_NAME=nginx-123
POD_NAMESPACE=middleware
```

<!-- markdownlint-disable MD046 -->
???+ note

    如果该 Value 占位符无法识别，会以纯字符串添加到环境变量。例如 `"POD_NAME": "{fieldRef:metadata.PODNAME}"`，这是错误的写法，在环境变量是 `POD_NAME={fieldRef:metadata.PODNAME}`。

### 关于 `{resourceFieldRef:*}` 的重要说明 {#resourcefieldref-important-notes}

`{resourceFieldRef:*}` 占位符用于引用 Pod 中**第一个容器**的资源限制（limits）和请求（requests）。使用时有以下注意事项：

1. **资源检查**：如果 Pod 的第一个容器没有配置对应的资源限制或请求，使用该占位符的环境变量将**不会被注入**。例如：
   - 如果容器没有设置 `limits.cpu`，那么 `{resourceFieldRef:limits.cpu}` 环境变量将被忽略
   - 如果容器没有设置 `requests.memory`，那么 `{resourceFieldRef:requests.memory}` 环境变量将被忽略

1. **单位说明**：
   - CPU 单位是 **millicores (m)**，例如 `500` 表示 500m（即 0.5 CPU）
   - 内存单位是 **MiB**，例如 `1024` 表示 1024Mi（即 1GiB）

1. **仅支持第一个容器**：`{resourceFieldRef:*}` 只能引用 Pod 中第一个容器的资源，不支持引用其他容器的资源。

1. **使用示例**：

```json
{
    "envs": {
        "APP_CPU_LIMIT": "{resourceFieldRef:limits.cpu}",
        "APP_MEMORY_REQUEST": "{resourceFieldRef:requests.memory}"
    }
}
```

1. **验证方法**：可以通过查看注入后的 Pod 环境变量来确认资源占位符是否正确解析：

```shell
kubectl exec <pod-name> -- env | grep APP_
APP_CPU_LIMIT=500
APP_MEMORY_REQUEST=512
```
<!-- markdownlint-enable -->

## FAQ {#faq}

### 如何禁用特定 Pod 的注入？ {#disable-inject}

给该 Pod 添加 Annotation `"admission.datakit/enabled": "false"`，将不再为它执行任何操作，此优先级最高。

### 工作原理是什么？ {#principles}

DataKit-Operator 使用 Kubernetes Admission Controller 功能进行资源注入，详细机制请查看[官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}

### 在 AWS EKS 环境需要注意什么？ {#aws-eks}

在 AWS EKS 环境部署，可能导致 DataKit-Operator 不生效，需要在安全组开启 `9543` 端口。

### 故障排查指南 {#debug}

| 问题 | 可能原因 | 解决方案 |
|--- |--- |---|
| 注入不生效 | Webhook 未正确配置 | 检查 `MutatingAdmissionWebhook` 和 `ValidatingAdmissionWebhook` |
| 镜像拉取失败 | 镜像地址或权限问题 | 验证镜像地址，检查镜像仓库访问权限 |
| 端口不可达 | 网络或安全组配置 | 开放 `9543` 端口，检查网络策略 |
