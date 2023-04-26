
# 容器日志

---

容器/Pod 上的日志采集，相比主机上的日志采集，有很大的不同，其配置方式、采集机制都有差异。从数据来源来说，可以分为两部分：

- stdout/stderr 日志采集：即容器应用的 stdout/stderr 输出，也是最常见的方式，可以使用类似 `docker logs` 或 `kubectl logs` 查看
- 容器/Pod 内部文件采集：如果日志不输出到 stdout/stderr，那应该就是落盘存储到文件，采集这种日志的需要稍微复杂一点

本文会详细介绍这两种采集方式。

## stdout/stderr 日志采集 {#logging-stdout}

stdout/stderr 日志采集的主要配置有以下两种方式：

- 通过 Pod/容器 *镜像特征* 来配置日志采集
- 通过 Annotation/Label 来标注特定 Pod/容器的日志采集

### 根据容器 image 来调整日志采集 {#logging-with-image-config}

默认情况下，DataKit 会收集所在机器/Node 上所有容器/Pod 的 stdout/stderr 日志，这可能不是大家的预期行为。某些时候，我们希望只采集（或不采集）部分容器/Pod 的日志，这里可以通过镜像名称来间接指代目标容器/Pod。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    ``` toml
    ## 当容器的 image 能够匹配 `hello*` 时，会采集此容器的日志
    container_include_logging = ["image:hello*"]
    ## 忽略所有容器
    container_exclude_logging = ["image:*"]
    ```
    
    - `container_include` 和 `container_exclude` 必须以 `image` 开头，格式为一种[类正则的 Glob 通配](https://en.wikipedia.org/wiki/Glob_(programming)){:target="_blank"}：`"image:<glob规则>"`

=== "Kubernetes"

    可通过如下环境变量 

    - ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
    - ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG

    来配置容器的日志采集。假设有 3 个 Pod，其 image 分别是：

    - A：`hello/hello-http:latest`
    - B：`world/world-http:latest`
    - C：`registry.jiagouyun.com/datakit/datakit:1.2.0`

    如果只希望采集 Pod A 的日志，那么配置 ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG 即可：

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: image:hello*  # 指定镜像名或其通配
    ```

???+ tip "如何查看镜像"

    Docker：

    ``` shell
    docker inspect --format "{{`{{.Config.Image}}`}}" $CONTAINER_ID
    ```

    Kubernetes Pod：
    
    ``` shell
    echo `kubectl get pod -o=jsonpath="{.items[0].spec.containers[0].image}"`
    ```

<!-- markdownlint-enable -->

### 通过 Annotation/Label 调整容器日志采集 {#logging-with-annotation-or-label}

通过镜像只能大概配置日志采集（或不采集）的目标容器/Pod，更高级一些的配置则需要在容器/Pod 上追加对应的标注，通过这种定向的标注，可以用来配置诸如日志来源、多行配置、Pipeline 切割等功能。

在容器 Label 或 Pod yaml 中，标注有特定的 Key（`datakit/logs`），其 Value 是一个 JSON 字符串，示例如下：

``` json
[
  {
    "disable"  : false,
    "source"   : "testing-source",
    "service"  : "testing-service",
    "pipeline" : "test.p",
    "multiline_match" : "^\\d{4}-\\d{2}",
    "enable_diskcache" : false,

    # 用法和上文的 `根据 image 过滤容器` 完全相同，`image:` 后面填写正则表达式
    "only_images"  : ["image:<your_image_regexp>"],

    # 可以给该容器/Pod 日志打上额外的标签
    "tags" : {
      "some_tag" : "some_value",
      "more_tag" : "some_other_value"
    }
  }
]
```

Value 字段说明：

| 字段名             | 必填 | 取值             | 默认值 | 说明                                                                                                                                                             |
| -----              | ---- | ----             | ----   | ----                                                                                                                                                             |
| `disable`          | N    | true/false       | false  | 是否禁用该 pod/容器的日志采集                                                                                                                                    |
| `source`           | N    | 字符串           | 无     | 日志来源，参见[容器日志采集的 source 设置](container.md#config-logging-source)                                                                                   |
| `service`          | N    | 字符串           | 无     | 日志隶属的服务，默认值为日志来源（source）                                                                                                                       |
| `pipeline`         | N    | 字符串           | 无     | 适用该日志的 Pipeline 脚本，默认值为与日志来源匹配的脚本名（`<source>.p`）                                                                                       |
| `only_images`      | N    | 字符串数组       | 无     | 针对 Pod 内部多容器情景，如果填写了任何 image 通配，则只采集能匹配这些 image 的容器的日志，类似白名单功能；如果字段为空，即认为采集该 Pod 中所有容器的日志       |
| `enable_diskcache` | N    | true/false       | false  | 是否开启磁盘缓存，可以有效避免采集延迟，有一定的性能开销，建议只在日志量超过 3000 条/秒再开启                                                                    |
| `multiline_match`  | N    | 正则表达式字符串 | 无     | 用于[多行日志匹配](logging.md#multiline)时的首行识别，例如 `"multiline_match":"^\\d{4}"` 表示行首是4个数字，在正则表达式规则中`\d` 是数字，前面的 `\` 是用来转义 |
| `tags`             | N    | key/value 键值对 | 无     | 添加额外的 tags，如果已经存在同名的 key 将以此为准（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ）                                                  |

#### 配置示例 {#logging-annotation-label-example}

可以通过配置容器的 Label 或 Pod 的 Annotation 来指定日志采集配置。

<!-- markdownlint-disable MD046 -->
=== "Docker"

    Docker 容器添加 Label 的方法，参见[这里](https://docs.docker.com/engine/reference/commandline/run/#set-metadata-on-container--l---label---label-file){:target="_blank"}。

=== "Kubernetes"

    在 Kubernetes 可以在创建 Deployment 时，以 `template` 模式添加 Pod Annotation，例如：
    
    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: testing-log-deployment
      labels:
        app: testing-log
    spec:
      template:
        metadata:
          labels:
            app: testing-log
          annotations:
            datakit/logs: |
              [
                {
                  "disable": false,
                  "source": "testing-source",
                  "service": "testing-service",
                  "pipeline": "test.p",
                  "multiline_match": "^\\d{4}-\\d{2}",
                  "only_images": ["image:.*nginx.*", "image:.*my_app.*"],
                  "tags" : {
                    "some_tag" : "some_value"
                  }
                }
              ]
    ```

???+ attention

    - 如无必要，不要轻易在 Annotation/Label 中配置 Pipeline，一般情况下，通过 `source` 字段自动推导即可。
    - 如果是在配置文件或终端命令行添加 Labels/Annotations，两边是英文状态双引号，需要添加转义字符。

    `multiline_match` 的值是双重转义，4 根斜杠才能表示实际的 1 根，例如 `\"multiline_match\":\"^\\\\d{4}\"` 等价 `"multiline_match":"^\d{4}"`，示例：

    ```shell
    kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
    ``` 

<!-- markdownlint-enable -->

## 非 stdout/stderr 日志采集 {#logging-not-stdout}

如果容器/Pod 将它们将日志输出到文件，使用现有的 stdout/stderr 方式是采集不到的。

对于这类容器/Pod，最简单的办法是将该文件挂载到主机，然后 Datakit 开启和配置 [logging 采集器](logging.md)。如果不方便挂载文件，可以使用以下方法。

### 采集 Pod 内部日志 {#logging-with-sidecar-config}

如果 Pod 日志并未打到 stdout，则可以通过 Sidecar 形式可以采集 Pod 内部日志，参见[这里](logfwd.md)。

### 采集容器内部日志 {#logging-with-inside-config}

在该容器上添加 Label 指定文件路径，由 Datakit 采集对应的文件。

<!-- markdownlint-disable MD046 -->
???+ attention

    - 此方式只在 Datakit 主机部署时生效，Kubernetes DaemonSet 部署则不生效
    - 只支持 Docker runtime，暂不支持 containerd
    - 只支持 GraphDriver 是 `overlay2` 的容器

<!-- markdownlint-enable -->

在容器 Label 中，标注有特定的 Key（`datakit/logs/inside`），其 Value 是一个 JSON 字符串，示例如下：

``` json
[
  {
    "source"   : "testing-source",
    "service"  : "testing-service",
    "pipeline" : "test.p",
    "multiline_match" : "^\\d{4}-\\d{2}",
    "paths"    : [
        "/tmp/data*"
    ],
    "tags" : {
      "some_tag" : "some_value",
      "more_tag" : "some_other_value"
    }
  }
]
```

Value 字段说明：

| 字段名            | 必填 | 取值             | 默认值 | 说明                                                                                                                                                             |
| -----             | ---- | ----             | ----   | ----                                                                                                                                                             |
| `source`          | N    | 字符串           | 无     | 日志来源，参见[容器日志采集的 source 设置](container.md#config-logging-source)                                                                                   |
| `service`         | N    | 字符串           | 无     | 日志隶属的服务，默认值为日志来源（source）                                                                                                                       |
| `pipeline`        | N    | 字符串           | 无     | 适用该日志的 Pipeline 脚本，默认值为与日志来源匹配的脚本名（`<source>.p`）                                                                                       |
| `paths`           | N    | 字符串数组       | 无     | 配置多个文件路径，支持通配符，通配用法详见[此处](logging.md#grok-rules)                                                                                          |
| `multiline_match` | N    | 正则表达式字符串 | 无     | 用于[多行日志匹配](logging.md#multiline)时的首行识别，例如 `"multiline_match":"^\\d{4}"` 表示行首是4个数字，在正则表达式规则中`\d` 是数字，前面的 `\` 是用来转义 |
| `tags`            | N    | key/value 键值对 | 无     | 添加额外的 tags，如果已经存在同名的 key 将以此为准（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ）                                                  |

#### 配置示例 {#logging-inside-example}

创建 Dockerfile 内容如下：

```dockerfile
FROM ubuntu:18.04 AS base

RUN  echo 'i=0          \n\
while true              \n\
do                      \n\
    printf "[%d] Bash For Loop Examples. Hello, world! Testing output.\\n" $i >> /tmp/data1 \n\
    true $(( i++ ))     \n\
    sleep 1             \n\
done' >> /root/s.sh

LABEL datakit/logs/inside '[{"source":"testing-source","service":"testing-service","paths":["/tmp/data*"],"tags":{"some_tag":"some_value","more_tag":"some_other_value"}}]'

CMD ["/bin/bash","/root/s.sh"]
```

创建 image 和容器：

```shell
docker build -t testing-image:v1 .
docker run -d testing-image:v1
```

Datakit 在发现这个容器后，会根据其 `datakit/logs/inside` 的配置创建日志采集。

## FAQ {#faq}

### :material-chat-question: 容器日志的特殊字节码过滤 {#special-char-filter}

容器日志可能会包含一些不可读的字节码（比如终端输出的颜色等），可以

- 将 `logging_remove_ansi_escape_codes` 设置为 `true`
- DataKit DaemonSet 部署时，将 `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` 置为 `true`

此配置会影响日志的处理性能，基准测试结果如下：

``` shell
goos: linux
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/test
cpu: Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz
BenchmarkRemoveAnsiCodes
BenchmarkRemoveAnsiCodes-8        636033              1616 ns/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/test       1.056s
```

每一条文本的处理耗时将额外增加 `1616 ns` 不等。如果日志中不带有颜色等修饰，不要开启该功能。

### :material-chat-question: 容器日志采集的 source 设置 {#config-logging-source}

在容器环境下，日志来源（`source`）设置是一个很重要的配置项，它直接影响在页面上的展示效果。但如果挨个给每个容器的日志配置一个 source 未免残暴。如果不手动配置容器日志来源，DataKit 有如下规则（优先级递减）用于自动推断容器日志的来源：

> 所谓不手动指定容器日志来源，就是指在 Pod Annotation 中不指定，在 container.conf 中也不指定（目前 container.conf 中无指定容器日志来源的配置项）

- 容器名：一般从容器的 `io.kubernetes.container.name` 这个 label 上取值。如果不是 Kubernetes 创建的容器（比如只是单纯的 Docker 环境，那么此 label 没有，故不以容器名作为日志来源）
- short-image-name: 镜像名，如 `nginx.org/nginx:1.21.0` 则取 `nginx`。在非 Kubernetes 容器环境下，一般首先就是取（精简后的）镜像名
- `unknown`: 如果镜像名无效（如 `sha256:b733d4a32c...`），则取该未知值

## 延伸阅读 {#more-reading}

- [Pipeline：文本数据处理](../developers/pipeline.md)
- [DataKit 日志采集综述](datakit-logging.md)
