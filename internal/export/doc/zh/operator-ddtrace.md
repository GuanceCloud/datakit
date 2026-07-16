# DataKit Operator 注入 DDTrace

DataKit Operator 在 Pod **创建时**通过 admission webhook 修改 Pod 模板：它添加 `datakit-lib-init` initContainer，将语言库复制到共享卷 `/datadog-lib`，再把该卷和必要的启动环境注入业务容器。它不会修改已经运行的 Pod；变更 Operator ConfigMap、镜像版本或 Deployment 注解后，需要通过发布/重启创建新 Pod 才会生效。

使用前请明确三件事：

1. 选择器决定哪些工作负载会被注入，空选择器可能扩大影响范围，应先在独立命名空间试点；
1. `image` 中的库版本必须匹配**业务容器**的语言运行时版本，而不是 initContainer 的运行时；
1. 注入库不等于目标应用一定能启动或产生链路，仍需验证业务容器的启动参数、环境变量和实际 trace 请求。

## 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#install)
1. 在 Operator 中增加如下 ConfigMap 配置

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level": "info",
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "namespace_selectors": ["staging"],
                    "label_selectors": ["app=example"],
                    "check_annotation": false,
                    "image": "<ddtrace-library-image>",
                    "language": "java",
                    "envs": {
                        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
                        "DD_TRACE_AGENT_PORT": "9529"
                    }
                }
            ]
        },
        "admission_inject": {
            "ddtrace": {}
        }
    }
    ```

    上例是可解析的 JSON。`admission_inject_v2`（Operator `v1.8.0+`）支持多个 DDTrace 配置，建议优先使用；`admission_inject` 是旧版兼容配置，通常只能表达一套 DDTrace 规则。不要把带 `//` 注释的 JSON 直接复制到 ConfigMap。

    DDTrace 注入有如下可配置字段：

    | 字段                         | 类型    | 描述                                                   | 是否必填     | 示例值                           |
    | ------:                      | :-----: | :------                                                | :---:        | :--------                        |
    | `envs`                       | object  | 环境变量映射                                           | Y[^envs]     | 见下方示例                       |
    | `image`                      | string  | DDTrace 镜像地址                                       | Y[^image]    | 见下方示例                       |
    | `label_selectors`            | array   | 标签选择器数组                                         | Y[^selector] | `["app=nginx", "tier=frontend"]` |
    | `language`                   | string  | 支持的语言类型（可选 `java`/`python`/`php`/`nodejs`）   | Y[^lang]     | `"nodejs"`                       |
    | `namespace_selectors`        | array   | 命名空间选择器，使用正则表达式                         | Y[^selector] | `["^prod-.*$", "^test$"]`       |
    | `resources`                  | object  | 资源限制配置                                           | N            | 见下方示例                       |
    | ~~`enabled_namespaces`~~     | object  | 选择要注入的 Kubernetes namespace 并设定对应的开发语言 | Y            | 1.7.0 中 `admission_inject_v2` 已弃用|
    | ~~`enabled_labelselectors`~~ | object  | 通过 Kubernetes label 选择要注入的目标                 | Y            | 1.7.0 中 `admission_inject_v2` 已弃用|

    [^selector]: 字段本身必须填写，否则 Operator 会拒绝注入。空数组会使选择范围变宽；上线前应明确它是否符合预期。
    [^image]: 安装模板中提供了默认镜像地址，对于离线环境，用户一般需要将镜像拷贝到内网，进而需要使用内网的镜像地址。
    [^lang]: 此处选择的语言必须和对应的 DDTrace 镜像内容匹配，如果不匹配，会导致注入失效。
    [^envs]: 这些环境变量设置非常关键，直接影响最终的数据效果。这里支持的 `fieldRef` 支持列表，参见[这里](datakit-operator.md#downwardapi)

    `language` 字段允许的值不代表任意版本都已有可用镜像。本页给出 Node.js 和 Python 的已核对版本映射；Java 使用下文示例并验证最终 JVM 命令已加载 Agent；PHP 仅应使用安装模板或发行说明中明确匹配 PHP 运行方式的镜像。不要用一种语言的镜像尝试注入另一种运行时。

### 先小范围试点 {#pilot}

建议先使用一个测试 namespace 和一个明确的 `app=<name>` 标签选择器，再逐步扩大范围。镜像标签应固定为具体版本，不要在生产规则中使用 `latest`。每次变更后至少验证：

```shell
kubectl get pod <pod-name> -o jsonpath='{.spec.initContainers[*].name}'
kubectl describe pod <pod-name>
```

第一个命令应包含 `datakit-lib-init`；第二个命令用于检查挂载、环境变量和 webhook 事件。随后还要从业务容器确认语言启动参数，并在 DataKit monitor 中确认 trace 请求。

### DDTrace Lib 注入方式和镜像选择 {#ddtrace-lib-image-selection}

DataKit Operator 注入 DDTrace 时，会添加名为 `datakit-lib-init` 的 initContainer，将 DDTrace 库拷贝到共享卷 `/datadog-lib`，再把该目录挂载到业务容器。镜像版本必须按**业务容器内的语言运行时版本**选择，而不是按 initContainer 的运行时选择。

镜像仓库按品牌区分。本文示例统一使用 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator`，发布到不同品牌站点时会替换为对应仓库地址。

#### Node.js 注入 {#ddtrace-nodejs-injection}

Node.js 注入会在业务容器中设置或追加 `NODE_OPTIONS`：

```shell
--require=/datadog-lib/node_modules/dd-trace/init
```

如果业务容器中已经存在 `NODE_OPTIONS`，Operator 会在原值后追加上述参数。Node.js 镜像必须和业务容器中的 Node.js 主版本匹配。

业务应用若自行覆盖 `NODE_OPTIONS`，会丢失注入参数并导致没有 trace。部署后可通过 `kubectl exec` 查看 `NODE_OPTIONS`，确认其中保留 `--require=/datadog-lib/node_modules/dd-trace/init`。

| 业务容器 Node.js 版本 | 推荐镜像 | 版本要求说明 |
| --- | --- | --- |
| Node.js 18 到 25 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:5.102.0` | 默认推荐版本，内置 `dd-trace@5.102.0`，要求 `node >=18 <26` |
| Node.js 16 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:4.55.0` | 使用 `dd-trace` 4.x 系列；不建议把 `5.102.0` 用于 Node.js 16 |
| Node.js 14 及以下 | 不作为默认支持范围 | 需要使用更旧的 `dd-trace` 主版本和对应镜像；建议优先升级 Node.js |

Node.js DDTrace 配置示例：

```json
{
    "namespace_selectors": ["default"],
    "label_selectors": [],
    "check_annotation": false,
    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-js-init:5.102.0",
    "language": "nodejs",
    "envs": {
        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
        "DD_TRACE_AGENT_PORT": "9529",
        "DD_SERVICE": "{fieldRef:metadata.labels['app']}",
        "POD_NAME": "{fieldRef:metadata.name}",
        "POD_NAMESPACE": "{fieldRef:metadata.namespace}",
        "NODE_NAME": "{fieldRef:spec.nodeName}",
        "DD_TAGS": "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
    }
}
```

#### Python 注入 {#ddtrace-python-injection}

Python 注入会在业务容器中设置或追加 `PYTHONPATH=/datadog-lib/`，让 Python 进程从 `/datadog-lib` 加载 DDTrace 相关库和注入 bootstrap。Python 的 DDTrace 包含 CPython ABI 相关 wheel，镜像版本需要和业务容器 Python 小版本匹配；版本不匹配时，可能出现 `ModuleNotFoundError`、native extension 加载失败或启动失败。

应用镜像或启动脚本如自行覆盖 `PYTHONPATH`，必须保留 `/datadog-lib/`，否则注入库无法被加载。启动后可执行 `python -c 'import ddtrace; print(ddtrace.__version__)'` 做最小加载验证。

| 业务容器 Python 版本 | 推荐镜像 | 版本要求说明 |
| --- | --- | --- |
| Python 3.7 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v2.21.12` | `ddtrace` 2.x 支持 Python 3.7；Python 3.7 不支持 `ddtrace` 3.x/4.x |
| Python 3.8 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v3.19.7` | `ddtrace` 3.x 支持 Python 3.8；`ddtrace` 4.x 要求 Python 3.9 及以上 |
| Python 3.9 到 3.14 | `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v4.8.3` | `ddtrace` 4.x 当前要求 `python >=3.9,<3.15` |

Python DDTrace 配置示例：

```json
{
    "namespace_selectors": ["default"],
    "label_selectors": [],
    "check_annotation": false,
    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-python-init:v4.8.3",
    "language": "python",
    "envs": {
        "DD_AGENT_HOST": "datakit-service.datakit.svc.cluster.local",
        "DD_TRACE_AGENT_PORT": "9529",
        "DD_SERVICE": "{fieldRef:metadata.labels['app']}",
        "POD_NAME": "{fieldRef:metadata.name}",
        "POD_NAMESPACE": "{fieldRef:metadata.namespace}",
        "NODE_NAME": "{fieldRef:spec.nodeName}",
        "DD_TAGS": "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
    }
}
```

> 版本选择依据来自 DDTrace 上游包的运行时要求：Node.js 以 npm `dd-trace` 的 `engines.node` 为准，Python 以 PyPI `ddtrace` 的 `Requires-Python` 和 wheel 支持为准。

#### Java 注入验证 {#ddtrace-java-injection}

Java 注入除库卷外还必须让 JVM 实际加载 `-javaagent`。应用启动后检查最终 Pod 的 `JAVA_TOOL_OPTIONS`、容器 command/args 或进程命令行，确认其中包含 Operator 注入的 Agent 路径；再确认 `DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT=9529`。仅看到 `datakit-lib-init` 成功并不能证明 JVM 已加载 Agent。

### `check_annotation` 配置项说明 {#check-annotation-config}

`check_annotation` 是一个重要的配置字段，用于控制 DataKit Operator 如何处理 Pod 上的**版本注解**（如 `admission.datakit/java-lib.version`、`admission.datakit/python-lib.version`、`admission.datakit/nodejs-lib.version`）。该字段的取值和行为如下：

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
   - `admission.datakit/<language>-lib.version` **受** `check_annotation` 配置影响
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
| `admission.datakit/python-lib.version` | 指定 DDTrace Python Lib 版本        | 版本字符串       | **是**                    | 例如 `"v3.19.7"`，用于覆盖配置中的默认镜像版本                       |
| `admission.datakit/nodejs-lib.version` | 指定 DDTrace Node.js Lib 版本       | 版本字符串       | **是**                    | 例如 `"5.102.0"`，用于覆盖配置中的默认镜像版本                       |
| `admission.datakit/enabled`          | 控制所有注入功能（最高优先级）         | `"true"`/`"false"` | **否**                   | `"false"`：完全拒绝任何注入，优先级最高                            |

#### 当 `check_annotation: true` 时 {#when-check-annotation-true}

需要同时满足以下条件才会执行注入：

1. **配置匹配**：Pod 必须匹配 `namespace_selectors` 和 `label_selectors` 规则
2. **功能注解允许**：`admission.datakit/ddtrace.enabled` 不为 `"false"`（如果存在）
3. **版本注解存在**：Pod 上必须存在版本注解（如 `admission.datakit/java-lib.version`、`admission.datakit/python-lib.version` 或 `admission.datakit/nodejs-lib.version`）

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
       "image": "internal-registry/dd-lib-java:<pinned-version>",
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
       "image": "internal-registry/dd-lib-java:<pinned-version>",
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

Operator 能识别如下 Annotation：

- `admission.datakit/ddtrace.enabled`：在单个 Deployment 标准自己是否开启注入，填写 `"true"` 即开启注入，`"false"` 则屏蔽注入，屏蔽后，Operator 会忽略注入这个 Deployment
- `admission.datakit/java-lib.version`：指定特定的 DDTrace Java Agent 版本
- `admission.datakit/python-lib.version`：指定特定的 DDTrace Python Lib 版本
- `admission.datakit/nodejs-lib.version`：指定特定的 DDTrace Node.js Lib 版本

> **注解使用说明**：关于 `check_annotation` 配置如何影响版本注解的行为，请参考 [Annotation 配置注入](datakit-operator.md#annotation-injection) 和[本页的 `check_annotation` 配置项说明](operator-ddtrace.md#check-annotation-config)。

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
