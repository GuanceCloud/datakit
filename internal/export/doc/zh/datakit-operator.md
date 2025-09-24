# DataKit Operator

---

:material-kubernetes:

---

## 概述和安装 {#datakit-operator-overview-and-install}

DataKit Operator 是 DataKit 在 Kubernetes 编排的联动项目，旨在协助 DataKit 更方便的部署，以及其他诸如验证、注入的功能。

目前 DataKit-Operator 提供以下功能：

- 注入 DDTrace Java SDK 以及对应环境变量信息，参见[文档](datakit-operator.md#datakit-operator-inject-lib)
- 注入 Sidecar logfwd 服务以采集容器内日志，参见[文档](datakit-operator.md#datakit-operator-inject-logfwd)
- 支持 DataKit 采集器的任务选举，参见[文档](election.md#plugins-election)

先决条件：

- 推荐 Kubernetes v1.24.1 及以上版本，且能够访问互联网（下载 yaml 文件并拉取对应镜像）
- 确保启用 `MutatingAdmissionWebhook` 和 `ValidatingAdmissionWebhook` [控制器](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"}
- 确保启用了 `admissionregistration.k8s.io/v1` API

### 安装步骤 {#datakit-operator-install}

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
         --repo  https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
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
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        -n datakit \
        -f values.yaml
    ```

    可以通过如下命令来卸载：

    ```shell
    $ helm uninstall datakit-operator -n datakit
    ```
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note

    - DataKit-Operator 有严格的程序和 yaml 对应关系，如果使用一份过旧的 yaml 可能无法安装新版 DataKit-Operator，请重新下载最新版 yaml。
    - 如果出现 `InvalidImageName` 报错，可以手动 pull 镜像。
<!-- markdownlint-enable -->

## 配置说明 {#datakit-operator-jsonconfig}

[:octicons-tag-24: Version-1.4.2](changelog.md#cl-1.4.2)

DataKit Operator 配置是 JSON 格式，在 Kubernetes 中单独以 ConfigMap 存放，以环境变量方式加载到容器中。

默认配置如下：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "enabled_namespaces":     [],
            "enabled_labelselectors": [],
            "images": {
                "java_agent_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
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
                   "cpu":    "200m",
                   "memory": "128Mi"
                 }
            }
        },
        "logfwd": {
            "images": {
                "logfwd_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.82.0"
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
        },
        "profiler": {
            "images": {
                "java_profiler_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/async-profiler:latest",
                "python_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/py-spy:latest",
                "golang_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/go-pprof:latest"
            },
            "envs": {
                "DK_AGENT_HOST":  "datakit-service.datakit.svc",
                "DK_AGENT_PORT":  "9529",
                "DK_PROFILE_VERSION": "1.2.333",
                "DK_PROFILE_ENV": "prod",
                "DK_PROFILE_DURATION": "240",
                "DK_PROFILE_SCHEDULE": "0 * * * *"
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
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["test01"],
                "label_selectors":     ["app=logging"],
                "config":"[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"storage_index\":\"logging-index\"\"source\":\"logging-tmp\"},{\"disable\":true,\"type\":\"file\",\"path\":\"/var/log/opt/**/*.log\",\"source\":\"logging-var\"}]"
            }
        ]
    }
}
```

主要配置项是 `ddtrace`、`logfwd` 和 `profiler`，指定注入的镜像和环境变量。此外，ddtrace 还支持根据 `enabled_namespaces` 和 `enabled_selectors` 批量注入，详见后文的“注入方式”。

### 指定镜像地址 {#datakit-operator-config-images}

DataKit Operator 主要作用就是注入镜像和环境变量，使用 `images` 配置镜像地址。`images` 是多个 Key/Value，Key 是固定的，修改 Value 值指定镜像地址。

正常情况下，镜像统一存放在 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator`，对于一些特殊环境不方便访问此镜像库，可以使用以下方法（以 `dd-lib-java-init` 镜像为例）：

1. 在可以访问 `pubrepo.<<<custom_key.brand_main_domain>>>` 的环境中，pull 镜像 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.30.1-ext`，并将其转存到自己的镜像库，例如 `inside.image.hub/datakit-operator/dd-lib-java-init:v1.30.1-ext`
1. 修改 JSON 配置，将 `admission_inject`->`ddtrace`->`images`->`java_agent_image` 修改为 `inside.image.hub/datakit-operator/dd-lib-java-init:v1.30.1-ext`，应用此 yaml
1. 此后 DataKit Operator 会使用的新的 Java Agent 镜像路径

**DataKit Operator 不检查镜像，如果该镜像路径错误，Kubernetes 创建 Pod 会报错。**


### 添加环境变量 {#datakit-operator-config-envs}

所有需要注入的环境变量，都必须在配置文件指定，DataKit Operator 不默认添加任何环境变量。

环境变量配置项是 `envs`，由多个 Key/Value 组成：Key 是固定值；Value 可以是固定值，也可以是占位符，根据实际情况取值。

例如在 `envs` 中添加一个 `testing-env`：

```json
{
    "admission_inject": {
        "ddtrace": {
            "envs": {
                "DD_AGENT_HOST":       "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT": "9529",
                "testing-env":         "ok"
            }
        }
    }
}
```

所有注入 `ddtrace` agent 的容器，都会添加 `envs` 的 3 个环境变量。

在 DataKit Operator v1.4.2 及以后版本，`envs` 支持 Kubernetes Downward API 的 [环境变量取值字段](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef)。现支持以下几种：

- `metadata.name`：Pod 的名称
- `metadata.namespace`： Pod 的命名空间
- `metadata.uid`： Pod 的唯一 ID
- `metadata.annotations['<KEY>']`： Pod 的注解 `<KEY>` 的值（例如：metadata.annotations['myannotation']）
- `metadata.labels['<KEY>']`： Pod 的标签 `<KEY>` 的值（例如：metadata.labels['mylabel']）
- `spec.serviceAccountName`： Pod 的服务账号名称
- `spec.nodeName`： Pod 运行时所处的节点名称
- `status.hostIP`： Pod 所在节点的主 IP 地址
- `status.hostIPs`： 这组 IP 地址是 status.hostIP 的双协议栈版本，第一个 IP 始终与 status.hostIP 相同。 该字段在启用了 PodHostIPs 特性门控后可用。
- `status.podIP`： Pod 的主 IP 地址（通常是其 IPv4 地址）
- `status.podIPs`： 这组 IP 地址是 status.podIP 的双协议栈版本，第一个 IP 始终与 status.podIP 相同。

举个例子，现有一个 Pod 名称是 nginx-123，namespace 是 middleware，要给它注入环境变量 `POD_NAME` 和 `POD_NAMESPACE`，参考以下：

```json
{
    "admission_inject": {
        "ddtrace": {
            "envs": {
                "POD_NAME":      "{fieldRef:metadata.name}",
                "POD_NAMESPACE": "{fieldRef:metadata.namespace}"
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
<!-- markdownlint-enable -->

## 注入方式 {#datakit-operator-inject}

DataKit-Operator 支持两种资源输入方式，分别是“全局配置 namespaces 和 selectors”，以及在目标 Pod 添加指定 Annotation。它们的区别如下：

- 全局配置 namespace 和 selector：通过修改 DataKit-Operator config，指定目标 Pod 的 Namespace 和 Selector，如果发现 Pod 符合条件，就执行注入资源。
    - 优点：不需要在目标 Pod 添加 Annotation（但是需要重启目标 Pod）
    - 缺点：范围不够精确，可能存在无效注入

- 在目标 Pod 添加 Annotation：在目标 Pod 添加 Annotation，DataKit-Operator 会检查 Pod Annotation，如果符合条件就执行注入。
    - 优点：范围足够精确，不存在无效注入
    - 缺点：必须在目标 Pod 添加 Annotation，且需要重启目标 Pod

<!-- markdownlint-disable MD046 -->
???+ note

    截止到 DataKit-Operator v1.5.8，全局配置 namespaces 和 selectors 方式只在注入 DDtrace 生效，对于 logfwd 和 profiler 无效，后者仍需添加 annotation 注入。
<!-- markdownlint-enable -->


<!-- markdownlint-disable MD013 -->
### 全局配置 namespaces 和 selectors 配置 {#datakit-operator-config-ddtrace-enabled}
<!-- markdownlint-enable -->

`enabled_namespaces` 和 `enabled_labelselectors` 是 `ddtrace` 专属，它们是对象数组，需要指定 `namespace` 和 `language`。数组之间是“或”的关系，写法如下（详见后文的配置说明）：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "enabled_namespaces": [
                {
                    "namespace": "testns",  # 指定 namespace
                    "language": "java"      # 指定需要注入的 agent 语言
                }
            ],
            "enabled_labelselectors": [
                {
                    "labelselector": "app=log-output",  # 指定 labelselector
                    "language": "java"                  # 指定需要注入的 agent 语言
                }
            ]
            # other..
        }
    }
}
```

如果一个 Pod 即满足 `enabled_namespaces` 规则，又满足 `enabled_labelselectors`，以 `enabled_labelselectors` 配置为准（通常在 `language` 取值用到）。

关于 labelselector 的编写规范，可参考此[官方文档](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/#label-selectors){:target="_blank"}。

<!-- markdownlint-disable MD046 -->
???+ note

    - 在 Kubernetes 1.16.9 或更早版本，Admission 不记录 Pod Namespace，所以无法使用 `enabled_namespaces` 功能。
<!-- markdownlint-enable -->

### 添加 Annotation 配置注入 {#datakit-operator-config-annotation}

在 Deployment 添加指定 Annotation，表示需要注入 `ddtrace` 文件。注意 Annotation 要添加在 template 中。

其格式为：

- key 是 `admission.datakit/%s-lib.version`，`%s` 需要替换成指定的语言，目前支持 `java`
- value 是指定版本号。默认是 DataKit-Operator 配置 `java_agent_image` 指定的版本

例如添加 Annotation 如下：

```yaml
      annotations:
        admission.datakit/java-lib.version: "v1.36.2-ext"
```

表示这个 Pod 需要注入的镜像版本是 v1.36.2-ext，镜像地址取自配置 `admission_inject`->`ddtrace`->`images`->`java_agent_image`，替换镜像版本为"v1.36.2-ext"，即 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.36.2-ext`。

## DataKit Operator 注入 {#datakit-operator-inject-sidecar}

在大型 Kubernetes 集群中，批量修改配置是比较麻烦的事情。DataKit-Operator 会根据 Annotation 配置，决定是否对其修改或注入。

目前支持的功能有：

- 注入 `ddtrace` agent 和 environment 的功能
- 挂载 `logfwd` sidecar 并开启日志采集的功能
- 注入 [`async-profiler`](https://github.com/async-profiler/async-profiler){:target="_blank"}  采集 JVM 程序的 profile 数据 [:octicons-beaker-24: Experimental](index.md#experimental)
- 注入 [`py-spy`](https://github.com/benfred/py-spy){:target="_blank"} 采集 Python 应用的 profile 数据 [:octicons-beaker-24: Experimental](index.md#experimental)

<!-- markdownlint-disable MD046 -->
???+ info

    只支持 v1 版本的 `deployments/daemonsets/cronjobs/jobs/statefulsets` 这五类 Kind，且因为 DataKit-Operator 实际对 PodTemplate 操作，所以不支持 Pod。 在本文中，以 `Deployment` 代替描述这五类 Kind。
<!-- markdownlint-enable -->

### DDtrace Agent {#datakit-operator-inject-lib}

#### 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
1. 在 deployment 添加指定 Annotation `admission.datakit/java-lib.version: ""`，表示需要注入默认版本的 DDtrace Java Agent。

#### 用例 {#datakit-operator-inject-lib-example}

下面是一个 Deployment 示例，给 Deployment 创建的所有 Pod 注入 `dd-java-lib`：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
      annotations:
        admission.datakit/java-lib.version: ""
    spec:
      containers:
      - name: nginx
        image: nginx:1.22
        ports:
        - containerPort: 80
```

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f nginx.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

$ kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}

datakit-lib-init
```

### logfwd {#datakit-operator-inject-logfwd}

#### 前置条件 {#datakit-operator-inject-logfwd-prerequisites}

[logfwd](../integrations/logfwd.md#using) 是 DataKit 的专属日志采集应用，需要先在同一个 Kubernetes 集群中部署 DataKit，且达成以下两点：

1. DataKit 开启 `logfwdserver` 采集器，例如监听端口是 `9533`
2. DataKit service 需要开放 `9533` 端口，使得其他 Pod 能访问 `datakit-service.datakit.svc:9533`

#### 使用说明 {#datakit-operator-inject-logfwd-instructions}

1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
2. 在 deployment 添加指定 Annotation，表示需要挂载 logfwd sidecar。注意 Annotation 要添加在 template 中
    - key 统一是 `admission.datakit/logfwd.instances`
    - value 是一个 JSON 字符串，是具体的 logfwd 配置，示例如下：

``` json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
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

参数说明，可参考 [logfwd 配置](../integrations/logfwd.md#config)：

- `datakit_addr` 是 DataKit logfwdserver 地址
- `loggings` 为主要配置，是一个数组，可参考 [DataKit logging 采集器](../integrations/logging.md)
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `storage_index` 指定日志存储索引
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 $source
    - `pipeline` Pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
    - `multiline_match` 多行匹配，详见 [DataKit 日志多行配置](../integrations/logging.md#multiline)，注意因为是 JSON 格式所以不支持 3 个单引号的“不转义写法”，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
    - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

<!-- markdownlint-disable MD046 -->
???+ note

    注入 logfwd 时，DataKit Operator 默认复用相同路径的 volume，避免因为存在同样路径的 volume 而注入报错。

    路径末尾有斜线和无斜线的意义不同，例如 `/var/log` 和 `/var/log/` 是不同路径，不能复用。
<!-- markdownlint-enable -->

#### 用例 {#datakit-operator-inject-logfwd-example}

下面是一个 Deployment 示例，使用 shell 持续向文件写入数据，且配置该文件的采集：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: logging-deployment
  labels:
    app: logging
spec:
  replicas: 1
  selector:
    matchLabels:
      app: logging
  template:
    metadata:
      labels:
        app: logging
      annotations:
        admission.datakit/logfwd.instances: '[{"datakit_addr":"datakit-service.datakit.svc:9533","loggings":[{"logfiles":["/var/log/log-test/*.log"],"source":"deployment-logging","tags":{"key01":"value01"}}]}]'
    spec:
      containers:
      - name: log-container
        image: busybox
        args: [/bin/sh, -c, 'mkdir -p /var/log/log-test; i=0; while true; do printf "$(date "+%F %H:%M:%S") [%-8d] Bash For Loop Examples.\\n" $i >> /var/log/log-test/1.log; i=$((i+1)); sleep 1; done']
```

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f logging.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb       1/1     Running   0             4s

$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
log-container datakit-logfwd
```

最终可以在<<<custom_key.brand_name>>>日志平台查看日志是否采集。

### async-profiler {#inject-async-profiler}

#### 前置条件 {#async-profiler-prerequisites}

- 集群已安装 [DataKit](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/){:target="_blank"}。
- [开启 profile](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} 采集器。
- Linux 内核参数 [kernel.perf_event_paranoid](https://www.kernel.org/doc/Documentation/sysctl/kernel.txt){:target="_blank"} 值设置为 2 及以下。

<!-- markdownlint-disable MD046 -->
???+ note

    `async-profiler` 使用 [`perf_events`](https://perf.wiki.kernel.org/index.php/Main_Page){:target="_blank"} 工具来抓取 Linux 的内核调用堆栈，非特权进程依赖内核的相应设置，可以使用以下命令来修改内核参数：
    ```shell
    $ sudo sysctl kernel.perf_event_paranoid=1
    $ sudo sysctl kernel.kptr_restrict=0
    # 或者
    $ sudo sh -c 'echo 1 >/proc/sys/kernel/perf_event_paranoid'
    $ sudo sh -c 'echo 0 >/proc/sys/kernel/kptr_restrict'
    ```
<!-- markdownlint-enable -->

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加 annotation：`admission.datakit/java-profiler.version: "latest"`，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

接下来以一个名为 `movies-java` 的 `Deployment` 资源配置文件为例进行说明。

```yaml
kind: Deployment
metadata:
  name: movies-java
  labels:
    app: movies-java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-java
  template:
    metadata:
      name: movies-java
      labels:
        app: movies-java
      annotations:
        admission.datakit/java-profiler.version: "0.4.4"
    spec:
      containers:
        - name: movies-java
          image: zhangyicloud/movies-java:latest
          imagePullPolicy: IfNotPresent
          securityContext:
            seccompProfile:
              type: Unconfined
          env:
            - name: JAVA_OPTS
              value: ""

      restartPolicy: Always
```

应用配置文件并检查是否生效：

```shell
$ kubectl apply -f deployment-movies-java.yaml

$ kubectl get pods | grep movies-java
movies-java-784f4bb8c7-59g6s   2/2     Running   0          47s

$ kubectl describe pod movies-java-784f4bb8c7-59g6s | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    12m   kubelet            Created container datakit-profiler
  Normal  Started    12m   kubelet            Started container datakit-profiler
```

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `jps -q -J-XX:+PerfDisableSharedMem | head -n 20` 来查找容器中的 JVM 进程，出于性能的考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 `datakit-operator-config` 下的环境变量来配置 profiling 的行为。
    
    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-java-784f4bb8c7-59g6s -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

### py-spy {#inject-py-spy}

#### 前置条件 {#py-spy-prerequisites}

- 当前只支持 Python 官方解释器（`CPython`）

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加 annotation：`admission.datakit/python-profiler.version: "latest"`，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

接下来将以一个名为 "movies-python" 的 `Deployment` 资源配置文件为例进行说明。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-python
  labels:
    app: movies-python
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-python
  template:
    metadata:
      name: movies-python
      labels:
        app: movies-python
      annotations:
        admission.datakit/python-profiler.version: "latest"
    spec:
      containers:
        - name: movies-python
          image: zhangyicloud/movies-python:latest
          imagePullPolicy: Always
          command:
            - "gunicorn"
            - "-w"
            - "4"
            - "--bind"
            - "0.0.0.0:8080"
            - "app:app"
```

应用资源配置并验证是否生效：

```shell
$ kubectl apply -f deployment-movies-python.yaml

$ kubectl get pods | grep movies-python
movies-python-78b6cf55f-ptzxf   2/2     Running   0          64s
 
$ kubectl describe pod movies-python-78b6cf55f-ptzxf | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    98s   kubelet            Created container datakit-profiler
  Normal  Started    97s   kubelet            Started container datakit-profiler
```

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20` 来查找容器中的 `Python` 进程，出于性能考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 ConfigMap `datakit-operator-config`  下的环境变量来配置 profiling 的行为。

    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-python-78b6cf55f-ptzxf -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

## DataKit Operator 资源变动 {#datakit-operator-mutate-resource}

### 添加 DataKit Logging 采集所需的配置 {#add-logging-configs}

DataKit Operator 可以为指定的 Pod 自动添加 DataKit Logging 采集所需的配置，包括 `datakit/logs` 注解和对应的文件路径 volume/volumeMount，简化了手动配置的繁杂步骤。这样，用户无需手动干预每个 Pod 配置即可自动启用日志采集功能。

以下是一个配置示例，展示了如何通过 DataKit Operator 的 `admission_mutate` 配置来实现日志采集配置的自动注入：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        # 其他配置
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["middleware"],
                "label_selectors":     ["app=logging"],
                "config": "[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"source\":\"logging-tmp\"}]"
            }
        ]
    }
}
```

`admission_mutate.loggings`：这是一个对象数组，包含多个日志采集配置。每个日志配置包括以下字段：

- `namespace_selectors`：限定符合条件的 Pod 所在的 Namespacce。可以设置多个 Namespace，Pod 必须匹配至少一个 Namespace 才会被选中。与 `label_selectors` 是“或”的关系。
- `label_selectors`：限定符合条件的 Pod 的 label。Pod 必须匹配至少一个 label selector 才会被选中。与 `namespace_selectors` 是“或”的关系。
- `config`：这是一个 JSON 字符串，它将被添加到 Pod 的注解中，注解的 Key 是 `datakit/logs`。如果该 Key 已经存在，它不会被覆盖或重复添加。这个配置将告诉 DataKit 如何采集日志。

DataKit Operator 会自动解析 `config` 配置，并根据其中的路径（`path`）为 Pod 创建对应的 volume 和 volumeMount。

以上述 DataKit Operator 配置为例，如果发现某个 Pod 的 Namespace 是 `middleware`，或 Labels 匹配 `app=logging`，就在 Pod 新增注解和挂载。例如：

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    datakit/logs: '[{"disable":false,"type":"file","path":"/tmp/opt/**/*.log","source":"logging-tmp"}]'
  labels:
    app: logging
  name: logging-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /tmp/opt/log1;
      i=1;
      while true; do
        echo "Writing logs to file ${i}.log";
        for ((j=1;j<=10000000;j++)); do
          echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log1/file_${i}.log;
          sleep 1;
        done;
        echo "Finished writing 5000000 lines to file_${i}.log";
        i=$((i+1));
      done
    command:
    - /bin/bash
    - -c
    - --
    image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /tmp/opt
      name: datakit-logs-volume-0
  volumes:
  - emptyDir: {}
    name: datakit-logs-volume-0
```

这个 Pod 存在 label `app=logging`，能够匹配上，于是 DataKit Operator 就给它添加了 `datakit/logs` 注解，并且将路径 `/tmp/opt` 添加 EmptyDir 挂载。

DataKit 日志采集发现到 Pod 后，就会根据 `datakit/logs` 内容进行定制化采集。

### FAQ {#datakit-operator-faq}

- 怎样指定某个 Pod 不注入？给该 Pod 添加 Annotation `"admission.datakit/enabled": "false"`，将不再为它执行任何操作，此优先级最高。

- DataKit-Operator 使用 Kubernetes Admission Controller 功能进行资源注入，详细机制请查看[官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}

- 在 AWS EKS 环境部署，可能导致 DataKit-Operator 不生效，需要在安全组开启 `9543` 端口。
