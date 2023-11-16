# Datakit Operator

---

:material-kubernetes:

---

## 概述和安装 {#datakit-operator-overview-and-install}

Datakit Operator 是 Datakit 在 Kubernetes 编排的联动项目，旨在协助 Datakit 更方便的部署，以及其他诸如验证、注入的功能。

目前 Datakit-Operator 提供以下功能：

- 注入 DDTrace SDK（Java/Python/Node.js）以及对应环境变量信息，参见[文档](datakit-operator.md#datakit-operator-inject-lib)
- 注入 Sidecar logfwd 服务以采集容器内日志，参见[文档](datakit-operator.md#datakit-operator-inject-logfwd)
- 支持 Datakit 采集器的任务选举，参见[文档](election.md#plugins-election)

先决条件：

- 推荐 Kubernetes v1.24.1 及以上版本，且能够访问互联网（下载 yaml 文件并拉取对应镜像）
- 确保启用 `MutatingAdmissionWebhook` 和 `ValidatingAdmissionWebhook` [控制器](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"}
- 确保启用了 `admissionregistration.k8s.io/v1` API

### 安装步骤 {#datakit-operator-install}

下载 [*datakit-operator.yaml*](https://static.guance.com/datakit-operator/datakit-operator.yaml){:target="_blank"}，步骤如下：

``` shell
kubectl create namespace datakit

wget https://static.guance.com/datakit-operator/datakit-operator.yaml

kubectl apply -f datakit-operator.yaml

kubectl get pod -n datakit

NAME                               READY   STATUS    RESTARTS   AGE
datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - Datakit-Operator 有严格的程序和 yaml 对应关系，如果使用一份过旧的 yaml 可能无法安装新版 Datakit-Operator，请重新下载最新版 yaml。
    - 如果出现 `InvalidImageName` 报错，可以手动 pull 镜像。
<!-- markdownlint-enable -->

### 相关配置 {#datakit-operator-jsonconfig}

[:octicons-tag-24: Datakit Operator v1.4.2]

Datakit Operator 配置是 JSON 格式，在 Kubernetes 中单独以 ConfigMap 存放，以环境变量方式加载到容器中。

默认配置如下：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image":   "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.20.2-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image":     "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125",
                "POD_NAME":                "{fieldRef:metadata.name}",
                "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
                "NODE_NAME":               "{fieldRef:spec.nodeName}",
                "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
            }
        },
        "logfwd": {
            "images": {
                "logfwd_image": "pubrepo.guance.com/datakit/logfwd:1.15.2"
            }
        },
        "profiler": {
            "images": {
                "java_profiler_image":   "pubrepo.guance.com/datakit-operator/async-profiler:0.1.0",
                "python_profiler_image": "pubrepo.guance.com/datakit-operator/py-spy:0.1.0",
                "golang_profiler_image": "pubrepo.guance.com/datakit-operator/go-pprof:0.1.0"
            },
            "envs": {
                "DK_AGENT_HOST":  "datakit-service.datakit.svc",
                "DK_AGENT_PORT":  "9529",
                "DK_PROFILE_VERSION": "1.2.333",
                "DK_PROFILE_ENV": "prod",
                "DK_PROFILE_DURATION": "240",
                "DK_PROFILE_SCHEDULE": "0 * * * *"
            }
        }
    }
}
```

其中，`admission_inject` 允许对 `ddtrace` 和 `logfwd` 做更精细的配置：

- `images` 是多个 Key/Value，Key 是固定的，修改 Value 值实现自定义 image 路径。

<!-- markdownlint-disable MD046 -->
???+ info

    Datakit Operator 的 `ddtrace` agent 镜像统一存放在 `pubrepo.guance.com/datakit-operator`，对于一些特殊环境可能不方便访问此镜像库，支持修改环境变量，指定镜像路径，方法如下：
    
    1. 在可以访问 `pubrepo.guance.com` 的环境中，pull 镜像 `pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance`，并将其转存到自己的镜像库，例如 `inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance`
    1. 修改 JSON 配置，将 `admission_inject`->`ddtrace`->`images`->`java_agent_image` 修改为 `inside.image.hub/datakit-operator/dd-lib-java-init:v1.8.4-guance`，应用此 yaml
    1. 此后 Datakit Operator 会使用的新的 Java Agent 镜像路径
    
    **Datakit Operator 不检查镜像，如果该镜像路径错误，Kubernetes 在创建时会报错。**
    
    如果已经在 Annotation 的 `admission.datakit/java-lib.version` 指定了版本，例如 `admission.datakit/java-lib.version:v2.0.1-guance` 或 `admission.datakit/java-lib.version:latest`，会使用这个 `v2.0.1-guance` 版本。
<!-- markdownlint-enable -->

- `envs` 同样是多个 Key/Value，Datakit Operator 会在目标容器中注入所有 Key/Value 环境变量。例如在 `envs` 中添加一个 `FAKE_ENV`：

```json
{
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST": "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT": "9529",
                "FAKE_ENV": "ok"
            }
        }
    }
}
```

所有注入 `ddtrace` agent 的容器，都会添加 `envs` 的 3 个环境变量。

在 Datakit Operator v1.4.2 及以后版本，`envs` 支持 Kubernetes Downward API 的 [环境变量取值字段](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef)。现支持以下几种：

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

如果该写法无法识别，会将其转换成纯字符串添加到环境变量。例如 `"POD_NAME": "{fieldRef:metadata.PODNAME}"`，这是错误的写法，最终在环境变量是 `POD_NAME={fieldRef:metadata.PODNAME}`。

## 使用 Datakit Operator 注入文件和程序 {#datakit-operator-inject-sidecar}

在大型 Kubernetes 集群中，批量修改配置是比较麻烦的事情。Datakit-Operator 会根据 Annotation 配置，决定是否对其修改或注入。

目前支持的功能有：

- 注入 `ddtrace` agent 和 environment 的功能
- 挂载 `logfwd` sidecar 并开启日志采集的功能
- 注入 [`async-profiler`](https://github.com/async-profiler/async-profiler){:target="_blank"} *:octicons-beaker-24: Experimental* 采集 JVM 程序的 profile 数据
- 注入 [`py-spy`](https://github.com/benfred/py-spy){:target="_blank"} *:octicons-beaker-24: Experimental* 采集 Python 应用的 profile 数据

<!-- markdownlint-disable MD046 -->
???+ info

    只支持 v1 版本的 `deployments/daemonsets/cronjobs/jobs/statefulsets` 这五类 Kind，且因为 Datakit-Operator 实际对 PodTemplate 操作，所以不支持 Pod。 在本文中，以 `Deployment` 代替描述这五类 Kind。
<!-- markdownlint-enable -->

### 注入 `ddtrace` agent 和相关的环境变量 {#datakit-operator-inject-lib}

#### 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 Datakit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
2. 在 deployment 添加指定 Annotation，表示需要注入 `ddtrace` 文件。注意 Annotation 要添加在 template 中
    - key 是 `admission.datakit/%s-lib.version`，%s 需要替换成指定的语言，目前支持 `java`、`python` 和 `js`
    - value 是指定版本号。如果为空，将使用环境变量的默认镜像版本

#### 用例 {#datakit-operator-inject-lib-example}

下面是一个 Deployment 示例，给 Deployment 创建的所有 Pod 注入 `dd-js-lib`：

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
        admission.datakit/js-lib.version: ""
    spec:
      containers:
      - name: nginx
        image: nginx:1.22
        ports:
        - containerPort: 80
```

使用 yaml 文件创建资源：

```shell
kubectl apply -f nginx.yaml
```

验证如下：

```shell
kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}

datakit-lib-init
```

### 注入 logfwd 程序并开启日志采集 {#datakit-operator-inject-logfwd}

#### 前置条件 {#datakit-operator-inject-logfwd-prerequisites}

[logfwd](../integrations/logfwd.md#using) 是 Datakit 的专属日志采集应用，需要先在同一个 Kubernetes 集群中部署 Datakit，且达成以下两点：

1. Datakit 开启 `logfwdserver` 采集器，例如监听端口是 `9533`
2. Datakit service 需要开放 `9533` 端口，使得其他 Pod 能访问 `datakit-service.datakit.svc:9533`

#### 使用说明 {#datakit-operator-inject-logfwd-instructions}

1. 在目标 Kubernetes 集群，[下载和安装 Datakit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
2. 在 deployment 添加指定 Annotation，表示需要挂载 logfwd sidecar。注意 Annotation 要添加在 template 中
    - key 统一是 `admission.datakit/logfwd.instances`
    - value 是一个 JSON 字符串，是具体的 logfwd 配置，示例如下：

``` json
[
    {
        "datakit_addr": "datakit-service.datakit.svc:9533",
        "loggings": [
            {
                "logfiles": ["<your-logfile-path>"],
                "ignore": [],
                "source": "<your-source>",
                "service": "<your-service>",
                "pipeline": "<your-pipeline.p>",
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

- `datakit_addr` 是 Datakit logfwdserver 地址
- `loggings` 为主要配置，是一个数组，可参考 [Datakit logging 采集器](../integrations/logging.md)
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 $source
    - `pipeline` Pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
    - `multiline_match` 多行匹配，详见 [Datakit 日志多行配置](../integrations/logging.md#multiline)，注意因为是 JSON 格式所以不支持 3 个单引号的“不转义写法”，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
    - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

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
kubectl apply -f logging.yaml
```

验证如下：

```shell
$ kubectl get pod
$ NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb       1/1     Running   0             4s
$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
$ log-container datakit-logfwd
```

最终可以在观测云日志平台查看日志是否采集。

### 注入 `async-profiler` 工具采集 JVM 应用性能数据 {#inject-async-profiler}

#### 前置条件 {#async-profiler-prerequisites}

- 集群已安装 [Datakit](https://docs.guance.com/datakit/datakit-daemonset-deploy/){:target="_blank"}。
- [开启 profile](https://docs.guance.com/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} 采集器。
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
Datakit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。


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
        admission.datakit/java-profiler.version: "latest"
    spec:
      containers:
        - name: movies-java
          image: zhangyicloud/movies-java:latest
          imagePullPolicy: IfNotPresent
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

稍等几分钟后即可在观测云控制台 [应用性能检监测-Profiling](https://console.guance.com/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `jps -q -J-XX:+PerfDisableSharedMem | head -n 20` 来查找容器中的 JVM 进程，出于性能的考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 `datakit-operator-config` 下的环境变量来配置 profiling 的行为。
    
    | 环境变量 | 说明 | 默认值 |
    |----|--|-----|
    |  `DK_PROFILE_SCHEDULE`  | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` |  `0 * * * *`（每小时调度一次）   |
    | `DK_PROFILE_DURATION`   | 每次 profiling 持续的时间，单位秒 |   240（4 分钟） |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-java-784f4bb8c7-59g6s -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->



### 注入 `py-spy` 工具采集 Python 应用性能数据 {#inject-py-spy}

#### 前置条件 {#py-spy-prerequisites}

- 当前只支持 Python 官方解释器（`CPython`）

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加 annotation：`admission.datakit/python-profiler.version: "latest"`，然后应用该资源配置文件，
Datakit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

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
$
$ kubectl get pods | grep movies-python
movies-python-78b6cf55f-ptzxf   2/2     Running   0          64s
$ 
$ kubectl describe pod movies-python-78b6cf55f-ptzxf | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    98s   kubelet            Created container datakit-profiler
  Normal  Started    97s   kubelet            Started container datakit-profiler
```

稍等几分钟后即可在观测云控制台 [应用性能检监测-Profiling](https://console.guance.com/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20` 来查找容器中的 `Python` 进程，出于性能考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 ConfigMap `datakit-operator-config`  下的环境变量来配置 profiling 的行为。

    | 环境变量 | 说明 | 默认值 |
    |----|--|-----|
    |  `DK_PROFILE_SCHEDULE`  | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` |  `0 * * * *`（每小时调度一次）   |
    | `DK_PROFILE_DURATION`   | 每次 profiling 持续的时间，单位秒 |   240（4 分钟） |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-python-78b6cf55f-ptzxf -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

---

补充：

- Datakit-Operator 使用 Kubernetes Admission Controller 功能进行资源注入，详细机制请查看[官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}
