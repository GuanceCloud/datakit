# DataKit Operator 注入 DDTrace

## 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#install)
1. 在 Operator 中增加如下 ConfigMap 配置

    ```json
    {
        "server_listen": "0.0.0.0:9543", // 服务监听地址
        "log_level": "info", // 日志级别

        "admission_inject_v2": { // 注入配置 v2(operator version >= v1.7.0)
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

    如下是一个示例：

    ```json
    {
        "namespace_selectors": [],
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

上面 Operator 的配置是针对整个集群中的 DDTrace 注入配置，某些时候这种一刀切的方式不适合特定的某些 Deployment，为此我们可以单独为这些 Deployment 做一些 Annotation 标记：

Operator 能识别如下两个 Annotation：

- `admission.datakit/ddtrace.enabled`：在单个 Deployment 标准自己是否开启注入，填写 `"true"` 即开启注入，`"false"` 则屏蔽注入，屏蔽后，Operator 会忽略注入这个 Deployment
- `admission.datakit/java-lib.version`：指定特定的 DDTrace 版本

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
