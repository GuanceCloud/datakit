# DataKit Operator 注入 DDTrace

## 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#install)
1. 在 Operator 中增加如下 ConfigMap 配置

    ```json
    {
        "server_listen": "0.0.0.0:9543", // 服务监听地址
        "log_level": "info", // 日志级别

        "admission_inject_v2": { // 注入配置 v2(operator version >= v1.8.0)
            "ddtraces": [ // DDTrace 配置数组
                {
                  // 此处填写 DDTrace 注入配置。..
                },
                {
                  // 可再注入另一个 DDTrace 配置。..
                }
            ]                   
        },
    
        "admission_inject": { // 注入配置 v1
            "ddtrace": {
                // 此处填写 DDTrace 注入配置。..
                // 对 v1 版本的配置，operator 只支持注入一份 DDTrace 配置，建议使用 v2 更灵活一些
            }
        },
    }
    ```

    <!-- 1. 在 deployment 添加指定 Annotation `admission.datakit/java-lib.version: "{{.DDTraceJavaExtVersion}}"`，表示需要注入默认版本的 DDtrace Java Agent。 -->



    DDTrace 注入有如下可配置字段：

    | 字段                         | 类型    | 描述                                                   | 是否必填     | 示例值                           |
    | ------:                      | :-----: | :------                                                | :---:        | :--------                        |
    | `envs`                       | object  | 环境变量映射                                           | Y[^envs]     | 见下方示例                       |
    | `image`                      | string  | DDTrace 镜像地址                                       | Y[^image]    | 见下方示例                       |
    | `label_selectors`            | array   | 标签选择器数组                                         | Y[^selector] | `["app=nginx", "tier=frontend"]` |
    | `language`                   | string  | 支持的语言类型（可选 `java`）                          | Y[^lang]     | `"java"`                         |
    | `namespace_selectors`        | array   | 命名空间选择器，支持正则                               | Y[^selector] | `["prod-*", "test"]`             |
    | `resources`                  | object  | 资源限制配置                                           | N            | 见下方示例                       |
    | ~~`enabled_namespaces`~~     | object  | 选择要注入的 Kubernetes namespace 并设定对应的开发语言 | Y            | 1.7.0 中 `admission_inject_v2` 已弃用|
    | ~~`enabled_labelselectors`~~ | object  | 通过 Kubernetes label 选择要注入的目标                 | Y            | 1.7.0 中 `admission_inject_v2` 已弃用|

    [^selector]: 此处必须填写，否则 Operator 会拒绝注入。
    [^image]: 虽然 Operator 内置了默认的镜像地址，对于离线环境，用户一般需要将镜像拷贝到内网，进而需要使用内网的镜像地址。
    [^lang]: 此处选择的语言必须和对应的 DDTrace 镜像内容匹配，如果不匹配，会导致注入失效。
    [^envs]: 这些环境变量设置非常关键，直接影响最终的数据效果。这里支持的 `fieldRef` 支持列表，参见[这里](datakit-operator.md#downwardapi)

### `check_annotation` 配置项说明 {#check-annotation-config}

`check_annotation` 是一个重要的配置字段，用于控制 DataKit Operator 如何处理 Pod 上的**版本注解**（如 `admission.datakit/java-lib.version`）。该字段的取值和行为如下：

| 取值     | 行为说明                                                                 |
|----------|--------------------------------------------------------------------------|
| `false`  | **（默认值）** 忽略 Pod 上的**版本注解**检查，直接根据选择器规则进行注入 |
| `true`   | 启用**版本注解**检查，只有匹配的 Pod 上存在版本注解才会注入             |

#### 重要逻辑说明 {#important-logic}

1. **功能特定注解始终有效**：
   - `admission.datakit/ddtrace.enabled` **不受** `check_annotation` 配置影响
   - 无论 `check_annotation` 是 `true` 还是 `false`，都会检测 `admission.datakit/ddtrace.enabled`
   - 如果 `admission.datakit/ddtrace.enabled: "false"`，将直接拒绝注入

2. **版本注解受 `check_annotation` 控制**：
   - `admission.datakit/java-lib.version` **受** `check_annotation` 配置影响
   - 当 `check_annotation: true` 时，需要版本注解存在才会注入
   - 当 `check_annotation: false` 时，忽略版本注解检查

3. **全局注解始终有效**：
   - `admission.datakit/enabled` **不受** `check_annotation` 配置影响
   - 如果 `admission.datakit/enabled: "false"`，将完全拒绝任何注入（最高优先级）

支持的 DDTrace 相关 Annotation：

| Annotation                           | 功能描述                               | 取值             | 受 `check_annotation` 影响 | 说明                                                                 |
|--------------------------------------|----------------------------------------|------------------|---------------------------|----------------------------------------------------------------------|
| `admission.datakit/ddtrace.enabled`  | 控制 DDTrace 注入                     | `"true"`/`"false"` | **否**                   | `"true"`：允许注入；`"false"`：拒绝注入；未设置：根据规则匹配决定    |
| `admission.datakit/java-lib.version` | 指定 DDTrace Java Agent 版本          | 版本字符串       | **是**                    | 例如 `"1.12.0"`，用于覆盖配置中的默认镜像版本                        |
| `admission.datakit/enabled`          | 控制所有注入功能（最高优先级）         | `"true"`/`"false"` | **否**                   | `"false"`：完全拒绝任何注入，优先级最高                            |

#### 当 `check_annotation: true` 时 {#when-check-annotation-true}

需要同时满足以下条件才会执行注入：

1. **配置匹配**：Pod 必须匹配 `namespace_selectors` 和 `label_selectors` 规则
2. **功能注解允许**：`admission.datakit/ddtrace.enabled` 不为 `"false"`（如果存在）
3. **版本注解存在**：Pod 上必须存在版本注解（如 `admission.datakit/java-lib.version`）

#### 当 `check_annotation: false` 时 {#when-check-annotation-false}

需要满足以下条件才会执行注入：

1. **配置匹配**：Pod 必须匹配 `namespace_selectors` 和 `label_selectors` 规则
2. **功能注解允许**：`admission.datakit/ddtrace.enabled` 不为 `"false"`（如果存在）
3. **忽略版本注解**：即使没有版本注解也会注入

#### 使用场景示例 {#use-case-examples}

1. **严格版本控制的场景**（`check_annotation: true`）：

   ```json
   {
       "namespace_selectors": ["prod"],
       "label_selectors": ["app=backend"],
       "check_annotation": true,
       "image": "internal-registry/dd-lib-java:{{.DDTraceJavaExtVersion}}",
       "language": "java"
   }
   ```

   **注入条件**：
   - Pod 在 `prod` 命名空间且带有 `app=backend` 标签
   - Pod **没有** `admission.datakit/ddtrace.enabled: "false"`（如果存在）
   - Pod **必须有** `admission.datakit/java-lib.version` 注解

2. **批量注入的场景**（`check_annotation: false`）：

   ```json
   {
       "namespace_selectors": ["staging"],
       "label_selectors": ["env=test"],
       "check_annotation": false,
       "image": "internal-registry/dd-lib-java:latest",
       "language": "java"
   }
   ```

   **注入条件**：
   - Pod 在 `staging` 命名空间且带有 `env=test` 标签
   - Pod **没有** `admission.datakit/ddtrace.enabled: "false"`（如果存在）
   - **忽略** `admission.datakit/java-lib.version` 注解检查

3. **选择性拒绝的场景**：

   ```json
   {
       "namespace_selectors": ["prod"],
       "label_selectors": ["app=java-app"],
       "check_annotation": false,
       "image": "internal-registry/dd-lib-java:latest",
       "language": "java"
   }
   ```

   **注入逻辑**：
   - 所有匹配的 Pod 都会被注入
   - 如果某个 Pod 有 `admission.datakit/ddtrace.enabled: "false"`，该 Pod 将被排除
   - 版本注解 `admission.datakit/java-lib.version: "1.15.0"` 可用于覆盖镜像版本，但不会影响是否注入的决策

    如下是一个示例：

    ```json
    {
        "namespace_selectors": [],
        "check_annotation": false,
        "label_selectors": [],
        "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:{{.DDTraceJavaExtVersion}}",
        "language": "java",
        "envs": {
             "DD_AGENT_HOST":           "datakit-service.datakit.svc.cluster.local",
             "DD_TRACE_AGENT_PORT":     "9529",
             "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc.cluster.local",
             "DD_JMXFETCH_STATSD_PORT": "8125",
             "DD_SERVICE":              "{fieldRef:metadata.labels['service']}",
             "POD_NAME":                "{fieldRef:metadata.name}",
             "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
             "NODE_NAME":               "{fieldRef:spec.nodeName}",
             "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
         },
        "resources": {
            "requests": {
                "cpu":    "100m",
                "memory": "64Mi"
            },
            "limits": {
                "cpu":    "500m",
                "memory": "512Mi"
            }
        }
    }
    ```

## 特殊 Deployment 的处理 {#special-deployment}

上面 Operator 的配置是针对整个集群中的 DDTrace 注入配置，某些时候这种一刀切的方式不适合特定的某些 Deployment，为此我们可以单独为这些 Deployment 做一些 Annotation 标记。

Operator 能识别如下两个 Annotation：

- `admission.datakit/ddtrace.enabled`：在单个 Deployment 标准自己是否开启注入，填写 `"true"` 即开启注入，`"false"` 则屏蔽注入，屏蔽后，Operator 会忽略注入这个 Deployment
- `admission.datakit/java-lib.version`：指定特定的 DDTrace 版本

> **注解使用说明**：关于 `check_annotation` 配置如何影响版本注解的行为，请参考 [Annotation 配置注入](datakit-operator.md#annotation-injection) 和 [`check_annotation` 配置项说明](datakit-operator.md#check-annotation-config)。

### Annotation 示例 {#anno-demo}

给 Deployment 标注是否注入标记：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
      annotations:
        admission.datakit/ddtrace.enabled: "true"
    spec:
      containers:
      - name: my-app
        image: my-app:1.2.3
        ports:
        - containerPort: 80
```

给 Deployment 注入 `dd-java-lib` 特定版本号 [^replace-ddtrace-version]：

[^replace-ddtrace-version]: 此处替换的是 Operator ConfigMap 中同一个镜像地址的不同版本，此处不能切换不同镜像地址。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
      annotations:
        admission.datakit/java-lib.version: "{{.DDTraceJavaExtVersion}}"
    spec:
      containers:
      - name: my-app
        image: my-app:1.2.3
        ports:
        - containerPort: 80
```

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f my-app.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
my-app-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

$ kubectl get pod my-app-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}

datakit-lib-init
```
