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
| `tags`               | key/value 键值对 | 添加额外的 tags，如果已经存在同名的 key 将以此为准（[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6) ）                                                     |

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
    $ docker run --name log-output -env DATAKIT_LOGS_CONFIG='[{"disable":false,"source":"log-source","service":"log-service"}]' -d testing/log-output:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="log-demo.yaml"
    apiVersion: v1
    kind: Pod
    metadata:
      name: log-demo
      annotations:
        ## 添加配置，且指定容器为 log-output
        datakit/log-output.logs: |
          [{
              "disable": false,
              "source":  "log-output-source",
              "service": "log-output-service",
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
    kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"log-source\",\"service\":\"log-service\",\"pipeline\":\"test.p\",\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
    ``` 

    如果一个 Pod/容器日志已经在采集中，此时再通过 `kubectl annotate` 命令添加配置不生效。

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
        echo "$(date +"%Y-%m-%d %H:%M:%S")  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
    done \n'\
    >> /opt/s.sh
    CMD ["/bin/bash", "/opt/s.sh"]

    ## 构建镜像
    $ docker build -t testing/log-to-file:v1 .

    ## 启动容器，添加环境变量 DATAKIT_LOGS_CONFIG，注意字符转义
    ## 指定非 stdout 路径，"type" 和 "path" 是必填字段，且需要创建采集路径的 volume
    ## 例如采集 `/tmp/opt/log` 文件，需要添加 `/tmp/opt` 的匿名 volume
    $ docker run --env DATAKIT_LOGS_CONFIG="[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/log\",\"source\":\"log-source\",\"service\":\"log-service\"}]" -v /tmp/opt  -d testing/log-to-file:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="logging.yaml"
    apiVersion: v1
    kind: Pod
    metadata:
      name: log-demo
      annotations:
        ## 添加配置，且指定容器为 logging-demo
        ## 同时配置了 file 和 stdout 两种采集。注意要采集 "/tmp/opt/log" 文件，需要先给 "/tmp/opt" 添加 emptyDir volume
        datakit/logging-demo.logs: |
          [
            {
              "disable": false,
              "type": "file",
              "path":"/tmp/opt/log",
              "source":  "logging-file",
              "tags" : {
                "some_tag": "some_value"
              }
            },
            {
              "disable": false,
              "source":  "logging-output"
            }
          ]
    spec:
      containers:
      - name: logging-demo
        image: pubrepo.guance.com/base/ubuntu:18.04
        args:
        - /bin/sh
        - -c
        - >
          i=0;
          while true;
          do
            echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.";
            echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log;
            i=$((i+1));
            sleep 1;
          done
        volumeMounts:
        - mountPath: /tmp/opt
          name: datakit-vol-opt
      volumes:
      - name: datakit-vol-opt
        emptyDir: {}
    ```

    执行 Kubernetes 命令，应用该配置：

    ``` shell
    $ kubectl apply -f logging.yaml
    ...
    ```
<!-- markdownlint-enable -->

对于容器内部的日志文件，在 Kubernetes 环境中还可以通过添加 sidecar 实现采集，参见[这里](logfwd.md)。

## 根据容器 image 来调整日志采集 {#logging-with-image-config}

默认情况下，DataKit 会收集所在机器/Node 上所有容器的 stdout/stderr 日志，这可能不是大家的预期行为。某些时候，我们希望只采集（或不采集）部分容器的日志，这里可以通过镜像名称或命名空间来间接指代目标容器。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    ``` toml
    ## 以 image 为例
    ## 当容器的 image 能够匹配 `datakit` 时，会采集此容器的日志
    container_include_log = ["image:datakit"]
    ## 忽略所有 kodo 容器
    container_exclude_log = ["image:kodo"]
    ```
    
    `container_include` 和 `container_exclude` 必须以属性字段开头，格式为一种[类正则的 Glob 通配](https://en.wikipedia.org/wiki/Glob_(programming)){:target="_blank"}：`"<字段名>:<glob 规则>"`

    现支持以下 4 个字段规则，这 4 个字段都是基础设施的属性字段：

    - image : `image:pubrepo.guance.com/datakit/datakit:1.18.0`
    - image_name : `image_name:pubrepo.guance.com/datakit/datakit`
    - image_short_name : `image_short_name:datakit`
    - namespace : `namespace:datakit-ns`

    对于同一类规则（`image` 或 `namespace`），如果同时存在 `include` 和 `exclude`，需要同时满足 `include` 成立，且 `exclude` 不成立的条件。例如：
    ```toml
    ## 这会导致所有容器都被过滤。如果有一个容器 `datakit`，它满足 include，同时又满足 exclude，那么它会被过滤，不采集日志；如果一个容器 `nginx`，首先它不满足 include，它会被过滤掉不采集。

    container_include_log = ["image_name:datakit"]
    container_exclude_log = ["image_name:*"]
    ```

    多种类型的字段规则有任意一条匹配，就不再采集它的日志。例如：
    ```toml
    ## 容器只需要满足 `image_name` 和 `namespace` 任意一个，就不再采集日志。

    container_include_log = []
    container_exclude_log = ["image_name:datakit", "namespace:datakit-ns"]
    ```

    `container_include_log` 和 `container_exclude_log` 的配置规则比较复杂，同时使用会有多种优先级情况。建议只使用 `container_exclude_log` 一种。


=== "Kubernetes"

    可通过如下环境变量 

    - ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
    - ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG

    来配置容器的日志采集。假设有 3 个 Pod，其 image 分别是：

    - A：`hello/hello-http:latest`
    - B：`world/world-http:latest`
    - C：`pubrepo.guance.com/datakit/datakit:1.2.0`

    如果只希望采集 Pod A 的日志，那么配置 ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG 即可：

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: image:hello*  # 指定镜像名或其通配
    ```

    或以命名空间来配置：

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG
        value: namesapce:foo  # 指定命名空间的容器日志不采集
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

    通过全局配置的 container_exclude_log 优先级低于容器的自定义配置 `disable`。例如，配置了 `container_exclude_log = ["image:*"]` 不采集所有日志，如果有 Pod Annotation 如下：

    ```json
    [
      {
          "disable": false,
          "type": "file",
          "path":"/tmp/opt/log",
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

### :material-chat-question: 日志目录的软链接问题 {#log-path-link}

正常情况下，Datakit 会从容器/Kubernetes API 找到日志文件的路径，然后采集该文件。

一些特殊环境，会对该日志所在目录做一个软连接，Datakit 无法提前获知软连接的目标，无法挂载该目录，导致找不到该日志文件，无法进行采集。

例如，现找到一个容器日志文件，路径是 `/var/log/pods/default_log-demo_f2617302-9d3a-48b5-b4e0-b0d59f1f0cd9/log-output/0.log`，但是在当前环境，`/var/log/pods` 是一个软连接指向 `/mnt/container_logs`，见下：

```shell
root@node-01:~# ls /var/log -lh
total 284K
lrwxrwxrwx 1 root root   20 Oct  8 10:06 pods -> /mnt/container_logs/
```

Datakit 需要挂载 `/mnt/container_logs` hostPath 才能使得正常采集，例如在 `datakit.yaml` 中添加以下：

```yaml
    # 省略
    spec:
      containers:
      - name: datakit
        image: pubrepo.guance.com/datakit/datakit:1.16.0
        volumeMounts:
        - mountPath: /mnt/container_logs
          name: container-logs
      # 省略
      volumes:
      - hostPath:
          path: /mnt/container_logs
        name: container-logs
```

这种情况不太常见，一般只有提前知道该路径有软连接，或查看 Datakit 日志发现采集报错才执行。

### :material-chat-question: 容器日志采集的 source 设置 {#config-logging-source}

在容器环境下，日志来源（`source`）设置是一个很重要的配置项，它直接影响在页面上的展示效果。但如果挨个给每个容器的日志配置一个 source 未免残暴。如果不手动配置容器日志来源，DataKit 有如下规则（优先级递减）用于自动推断容器日志的来源：

> 所谓不手动指定容器日志来源，就是指在 Pod Annotation 中不指定，在 container.conf 中也不指定（目前 container.conf 中无指定容器日志来源的配置项）

- Kubernetes 指定的容器名：从容器的 `io.kubernetes.container.name` 这个 label 上取值
- 容器本身的名称：通过 `docker ps` 或 `crictl ps` 能看到的容器名
- `default`: 默认的 `source`


## 延伸阅读 {#more-reading}

- [Pipeline：文本数据处理](../developers/pipeline/index.md)
- [DataKit 日志采集综述](datakit-logging.md)
