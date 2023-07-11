# Datakit Operator

---

:material-kubernetes:

---

## 概述和安装 {#datakit-operator-overview-and-install}

Datakit Operator 是 Datakit 在 Kubernetes 编排的联动项目，旨在协助 Datakit 更方便的部署，以及其他诸如验证、注入的功能。

目前 Datakit-Operator 提供以下功能：

- 注入 DDTrace SDK（Java/Python/JavaScript）以及对应环境变量信息，参见[文档](datakit-operator.md#datakit-operator-inject-lib)
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

[:octicons-tag-24: Datakit Operator v1.2.1]

Datakit Operator 配置是 JSON 格式，在 Kubernetes 中单独以 ConfigMap 存放，以环境变量方式加载到容器中。

默认配置如下：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image":   "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image":     "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125"
            }
        },
        "logfwd": {
            "images": {
                "logfwd_image": "pubrepo.guance.com/datakit/logfwd:1.5.8"
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

- `envs` 同样是多个 Key/Value，但是 Key 和 Value 不固定。Datakit Operator 会在目标容器中注入所有 Key/Value 环境变量。例如在 `envs` 中添加一个 `FAKE_ENV`：

```json
    "admission_inject": {
        "ddtrace": {
            "images": {
                "java_agent_image":   "pubrepo.guance.com/datakit-operator/dd-lib-java-init:v1.8.4-guance",
                "python_agent_image": "pubrepo.guance.com/datakit-operator/dd-lib-python-init:v1.6.2",
                "js_agent_image":     "pubrepo.guance.com/datakit-operator/dd-lib-js-init:v3.9.2"
            },
            "envs": {
                "DD_AGENT_HOST":           "datakit-service.datakit.svc",
                "DD_TRACE_AGENT_PORT":     "9529",
                "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc",
                "DD_JMXFETCH_STATSD_PORT": "8125",
                "FAKE_ENV":                "ok"
            }
        }
    }
```

所有注入 `ddtrace` agent 的容器，都会添加 `envs` 的 5 个环境变量。

## 使用 Datakit Operator 注入文件和程序 {#datakit-operator-inject-sidecar}

在大型 Kubernetes 集群中，批量修改配置是比较麻烦的事情。Datakit-Operator 会根据 Annotation 配置，决定是否对其修改或注入。

目前支持的功能有：

- 注入 `ddtrace` agent 和 environment 的功能
- 挂载 `logfwd` sidecar 并开启日志采集的功能

<!-- markdownlint-disable MD046 -->
???+ info

    只支持 v1 版本的 `deployments/daemonsets/cronjobs/jobs/statefulsets` 这五类 Kind，且因为 Datakit-Operator 实际对 PodTemplate 操作，所以不支持 Pod。 在本文中，以 `Deployment` 代替描述这五类 Kind。
<!-- markdownlint-enable -->

### 注入 `ddtrace` agent 和相关的环境变量 {#datakit-operator-inject-lib}

#### 使用说明 {#datakit-operator-inject-lib-usage}

1. 在目标 Kubernetes 集群，[下载和安装 Datakit-Operator](datakit-operator.md#datakit-operator-inject-lib)
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

[logfwd](logfwd.md#using) 是 Datakit 的专属日志采集应用，需要先在同一个 Kubernetes 集群中部署 Datakit，且达成以下两点：

1. Datakit 开启 `logfwdserver` 采集器，例如监听端口是 `9533`
2. Datakit service 需要开放 `9533` 端口，使得其他 Pod 能访问 `datakit-service.datakit.svc:9533`

#### 使用说明 {#datakit-operator-inject-logfwd-instructions}

1. 在目标 Kubernetes 集群，[下载和安装 Datakit-Operator](datakit-operator.md#datakit-operator-inject-lib)
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

参数说明，可参考 [logfwd 配置](logfwd.md#config)：

- `datakit_addr` 是 Datakit logfwdserver 地址
- `loggings` 为主要配置，是一个数组，可参考 [Datakit logging 采集器](logging.md)
    - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
    - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
    - `source` 数据来源，如果为空，则默认使用 'default'
    - `service` 新增标记 tag，如果为空，则默认使用 $source
    - `pipeline` Pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
    - `character_encoding` 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
    - `multiline_match` 多行匹配，详见 [Datakit 日志多行配置](logging.md#multiline)，注意因为是 JSON 格式所以不支持 3 个单引号的“不转义写法”，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
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

---

补充：

- Datakit-Operator 使用 Kubernetes Admission Controller 功能进行资源注入，详细机制请查看[官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}
