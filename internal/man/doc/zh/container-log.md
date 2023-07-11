# 容器日志

---

Datakit 支持采集 Kubernetes 和主机容器日志，从数据来源上，可以分为以下两种：

- 控制台输出：即容器应用的 stdout/stderr 输出，也是最常见的方式，可以使用类似 `docker logs` 或 `kubectl logs` 查看

- 容器内部文件：如果日志不输出到 stdout/stderr，那应该就是落盘存储到文件，采集这种日志的稍微麻烦一些需要做 mount 挂载

本文会详细介绍这两种采集方式。

## 控制台 stdout/stderr 日志采集 {#logging-stdout}

控制台输出（即 stdout/stderr）通过容器 runtime 落盘到文件，Datakit 会自动获取到该容器的 LogPath 进行采集。

如果要自定义采集的配置，可以通过添加容器环境变量或 Kubernetes Pod Annotation 的方式。

- 自定义配置的 Key 有以下几种情况：
    - 容器环境变量的 Key 固定为 `DATAKIT_LOGS_CONFIG`
    - Pod Annotation 的 Key 有两种写法：
        - `datakit/<container_name>.logs`，其中 `<container_name>` 需要替换为当前 Pod 的容器名，这在多容器环境下会用到
        - `datakit/logs` 会对该 Pod 的所有容器都适用


<!-- markdownlint-disable MD046 -->
???+ info

    如果一个容器存在环境变量 `DATAKIT_LOGS_CONFIG`，同时又能找到它所属 Pod 的 Annotation `datakit/logs`，按照就近原则，以容器环境变量的配置为准。
<!-- markdownlint-enable -->

- 自定义配置的 Value 如下：

``` json
[
  {
    "disable" : false,
    "source"  : "<your-source>",
    "service" : "<your-service>",
    "pipeline": "<your-pipeline.p>",
    "tags" : {
      "<some-key>" : "<some_other_value>"
    }
  }
]
```

字段说明：

| 字段名               | 取值             | 说明                                                                                                                                                                |
| -----                | ----             | ----                                                                                                                                                                |
| `disable`            | true/false       | 是否禁用该容器的日志采集，默认是 `false`                                                                                                                            |
| `type`               | `file`/不填      | 选择采集类型。如果是采集容器内文件，必须写成 `file`。默认为空是采集 `stdout/stderr`                                                                                 |
| `path`               | 字符串           | 配置文件路径。如果是采集容器内文件，必须填写 volume 的 path，注意不是容器内的文件路径，是容器外能访问到的路径。默认采集 `stdout/stderr` 不用填                      |
| `source`             | 字符串           | 日志来源，参见[容器日志采集的 source 设置](container.md#config-logging-source)                                                                                      |
| `service`            | 字符串           | 日志隶属的服务，默认值为日志来源（source）                                                                                                                          |
| `pipeline`           | 字符串           | 适用该日志的 Pipeline 脚本，默认值为与日志来源匹配的脚本名（`<source>.p`）                                                                                          |
| `multiline_match`    | 正则表达式字符串 | 用于[多行日志匹配](logging.md#multiline)时的首行识别，例如 `"multiline_match":"^\\d{4}"` 表示行首是 4 个数字，在正则表达式规则中 `\d` 是数字，前面的 `\` 是用来转义 |
| `character_encoding` | 字符串           | 选择编码，如果编码有误会导致数据无法查看，支持 `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""。默认为空即可                                                |
| `tags`               | key/value 键值对 | 添加额外的 tags，如果已经存在同名的 key 将以此为准（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ）                                                     |

完整示例如下：

<!-- markdownlint-disable MD046 -->
=== "容器环境变量"

    ``` shell
    $ cat Dockerfile
    FROM pubrepo.guance.com/base/ubuntu:18.04 AS base
    Run mkdir -p /opt
    Run echo 'i=0; \n\
    while true; \n\
    do \n\
        echo "$(date +"%Y-%m-%d %H:%M:%S")  [$i]  Bash For Loop Examples. Hello, world! Testing output."; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
    done \n'\
    >> /opt/s.sh
    CMD ["/bin/bash", "/opt/s.sh"]

    ## 构建镜像
    $ docker build -t testing/log-output:v1 .

    ## 启动容器，添加环境变量 DATAKIT_LOGS_CONFIG
    $ docker run --name log-output -env DATAKIT_LOGS_CONFIG='[{"disable":false,"source":"testing-source","service":"testing-service"}]' -d testing/log-output:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="log-output.yaml"
    apiVersion: v1
    kind: Pod
    metadata:
      name: log-output
      annotations:
        ## 添加配置，且指定容器为 log-output
        datakit/log-output.logs: |
          [{
              "disable": false,
              "source":  "testing-source-02",
              "service": "testing-service",
              "tags" : {
                "some_tag": "some_value"
              }
          }]"
    spec:
      containers:
      - name: log-output
        image: pubrepo.guance.com/base/ubuntu:18.04
        args:
        - /bin/sh
        - -c
        - >
          i=0;
          while true;
          do
            echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.";
            i=$((i+1));
            sleep 1;
          done
    ```

    执行 Kubernetes 命令，应用该配置：

    ``` shell
    $ kubectl apply -f log-output.yaml
    ...
    ```

???+ attention

    - 如无必要，不要轻易在环境变量和 Pod Annotation 中配置 Pipeline，一般情况下，通过 `source` 字段自动推导即可。
    - 如果是在配置文件或终端命令行添加 Env/Annotations，两边是英文状态双引号，需要添加转义字符。

    `multiline_match` 的值是双重转义，4 根斜杠才能表示实际的 1 根，例如 `\"multiline_match\":\"^\\\\d{4}\"` 等价 `"multiline_match":"^\d{4}"`，示例：

    ```shell
    kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
    ``` 

<!-- markdownlint-enable -->

## 容器内日志文件采集 {#logging-with-inside-config}

对于容器内部的日志文件，和控制台输出日志的区别是需要指定文件路径，其他配置项大同小异。

<!-- markdownlint-disable MD046 -->
???+ attention

    配置的文件路径，不是容器内的文件路径，是通过 volume mount 能在外部访问到的路径。
<!-- markdownlint-enable -->

同样是添加容器环境变量或 Kubernetes Pod Annotation 的方式，Key 和 Value 基本一致，详见前文。

完整示例如下：

<!-- markdownlint-disable MD046 -->
=== "容器环境变量"

    ``` shell
    $ cat Dockerfile
    FROM pubrepo.guance.com/base/ubuntu:18.04 AS base
    Run mkdir -p /opt
    Run echo 'i=0; \n\
    while true; \n\
    do \n\
        echo "$(date +"%Y-%m-%d %H:%M:%S")  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt01/log; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
    done \n'\
    >> /opt/s.sh
    CMD ["/bin/bash", "/opt/s.sh"]

    ## 构建镜像
    $ docker build -t testing/log-to-file:v1 .

    ## 启动容器，添加环境变量 DATAKIT_LOGS_CONFIG，注意字符转义
    ## 跟配置 stdout 不同，"type" 和 "path" 是必填字段
    ## 注意 "path" 的值是 "/tmp/opt02/log" 而不是 "/tmp/opt01/log"，"opt01" 是容器内路径，实际 volume 出来是 "opt02"
    $ docker run --env DATAKIT_LOGS_CONFIG="[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt02/log\",\"source\":\"testing-source\",\"service\":\"testing-service\"}]" -v /tmp/opt02:/tmp/opt01  -d testing/log-to-file:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="logging.yaml"
    apiVersion: v1
    kind: Pod
    metadata:
      name: logging
      annotations:
        ## 添加配置，且指定容器为 logging
        ## 同时配置了 file 和 stdout 两种采集，注意 "path" 的 "/tmp/opt02/log" 是 volume path
        datakit/logging.logs: |
          [
            {
              "disable": false,
              "type": "file",
              "path":"/tmp/opt02/log",
              "source":  "logging-file",
              "tags" : {
                "some_tag": "some_value"
              }
            },
            {
              "disable": false,
              "source":  "logging-output"
            }
          ]"
    spec:
      containers:
      - name: logging
        image: pubrepo.guance.com/base/ubuntu:18.04
        args:
        - /bin/sh
        - -c
        - >
          i=0;
          while true;
          do
            echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.";
            echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt01/log;
            i=$((i+1));
            sleep 1;
          done
        volumeMounts:
        - mountPath: /tmp/opt01
          name: opt
      volumes:
      - name: opt
        hostPath:
          path: /tmp/opt02
    ```

    执行 Kubernetes 命令，应用该配置：

    ``` shell
    $ kubectl apply -f logging.yaml
    ...
    ```
<!-- markdownlint-enable -->

对于容器内部的日志文件，在 Kubernetes 环境中还可以通过添加 sidecar 实现采集，参见[这里](logfwd.md)。

## 根据容器 image 来调整日志采集 {#logging-with-image-config}

默认情况下，DataKit 会收集所在机器/Node 上所有容器的 stdout/stderr 日志，这可能不是大家的预期行为。某些时候，我们希望只采集（或不采集）部分容器的日志，这里可以通过镜像名称来间接指代目标容器。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    ``` toml
    ## 当容器的 image 能够匹配 `hello*` 时，会采集此容器的日志
    container_include_logging = ["image:hello*"]
    ## 忽略所有容器
    container_exclude_logging = ["image:*"]
    ```
    
    `container_include` 和 `container_exclude` 必须以 `image` 开头，格式为一种[类正则的 Glob 通配](https://en.wikipedia.org/wiki/Glob_(programming)){:target="_blank"}：`"image:<glob 规则>"`

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

???+ attention

    通过全局配置的 container_exclude_logging 优先级低于容器的自定义配置 `disable`。例如，配置了 `container_exclude_logging = ["image:*"]` 不采集所有日志，如果有 Pod Annotation 如下：

    ```json
    [
      {
          "disable": false,
          "type": "file",
          "path":"/tmp/opt02/log",
          "source":  "logging-file",
          "tags" : {
            "some_tag": "some_value"
          }
      },
      {
          "disable": true,
          "source":  "logging-output"
      }
    ]"
    ```

    这份配置距离容器更近，优先级更高。配置的 `disable=fasle` 表明要采集日志文件，把上面的全局配置覆盖了。

    所以这个容器日志文件最终还是会采集，但是控制台输出 stdout/stderr 不采集，因为 `disable=true`。

<!-- markdownlint-enable -->

## FAQ {#faq}

### :material-chat-question: 容器日志采集的 source 设置 {#config-logging-source}

在容器环境下，日志来源（`source`）设置是一个很重要的配置项，它直接影响在页面上的展示效果。但如果挨个给每个容器的日志配置一个 source 未免残暴。如果不手动配置容器日志来源，DataKit 有如下规则（优先级递减）用于自动推断容器日志的来源：

> 所谓不手动指定容器日志来源，就是指在 Pod Annotation 中不指定，在 container.conf 中也不指定（目前 container.conf 中无指定容器日志来源的配置项）

- 容器本身的名称：通过 `docker ps` 或 `crictl ps` 能看到的容器名
- Kubernetes 指定的容器名：从容器的 `io.kubernetes.container.name` 这个 label 上取值
- `default`: 默认的 `source`

## 延伸阅读 {#more-reading}

- [Pipeline：文本数据处理](../developers/pipeline/index.md)
- [DataKit 日志采集综述](datakit-logging.md)
