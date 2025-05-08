# Datakit 日志采集全指南

## 前言 {#intro}

日志采集是<<<custom_key.brand_name>>> Datakit 重要的一项，它将主动采集或被动接收的日志数据加以处理，最终上传到<<<custom_key.brand_name>>>中心。

按照数据来源可以分为 “日志文件数据” 和 “网络流数据” 两种。分别对应以下：

- “日志文件数据”：
    - 本地磁盘文件：即存放在磁盘上的固定文件，通过 logging 采集器进行采集
    - 容器标准输出：容器 Runtime 将标准输出（stdout/stderr）落盘存储，由容器采集器访问 Runtime 获取落盘文件的路径并进行采集
    - 容器内日志文件：通过 Volume/VolumeMount 方法将文件挂载到外部，使得 Datakit 可以访问到

- “网络流数据”：
    - tcp/udp 数据：Datakit 监听对应的 tcp/udp 端口，接收对应的日志数据
    - logstreaming：通过 http 协议接收数据

以上几种数据类型从使用场景到配置方式截然不同，下文会详细描述。

## 使用场景和配置方式 {#config}

### 本地磁盘文件采集 {#config-local-files}

本章节只描述主机部署 Datakit 的场景，如果是 Kubernetes 部署不建议采集本地文件。

首先需要启动一个 logging 采集器，进入 Datakit 安装目录下的 `conf.d/log` 目录，复制 `logging.conf.sample` 并命名为 `logging.conf`。示例如下：

``` toml
  [[inputs.logging]]
   # 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定
   # 推荐使用绝对路径且指定文件后缀
   # 尽量缩小范围，避免采集到压缩包文件或二进制文件
   logfiles = [
    "/var/log/*.log",           # 文件路径下所有 log 文件
    "/var/log/*.txt",           # 文件路径下所有 txt 文件
    "/var/log/sys*",            # 文件路径下所有以 sys 前缀的文件
    "/var/log/syslog",           # unix 格式文件路径
    "c:/path/space 空格中文路径/some.txt", # windows 风格文件路径
   ]
  
   ## socket 目前支持两种协议：tcp/udp。建议开启内网端口防止安全隐患
   ## socket 和 log 目前只能选择其中之一，不能既通过文件采集，又通过 socket 采集
   sockets = [
    "tcp://0.0.0.0:9540"
    "udp://0.0.0.0:9541"
   ]
  
  # 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
   ignore = [""]
   
   # 数据来源，如果为空，则默认使用 'default'
   source = ""
   
   # 新增标记 tag，如果为空，则默认使用 $source
   service = ""
   
   # pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 pipeline
   pipeline = ""
   
   # 选择编码，如果编码有误会导致数据无法查看。默认为空即可
   # `utf-8`,`gbk`
   character_encoding = ""
   
   ## 设置正则表达式，例如 ^\d{4}-\d{2}-\d{2} 行首匹配 yyyy-mm-dd 时间格式
   ## 符合此正则匹配的数据，将被认定为有效数据，否则会累积追加到上一条有效数据的末尾
   ## 使用 3 个单引号 '''this-regexp''' 避免转义
   ## 正则表达式链接：https://golang.org/pkg/regexp/syntax/#hdr-syntax
   # multiline_match = '''^\s'''
  
   ## 是否开启自动多行模式，开启后会在 patterns 列表中匹配适用的多行规则
   auto_multiline_detection = true
   ## 配置自动多行的 patterns 列表，内容是多行规则的数组，即多个 multiline_match，如果为空则使用默认规则详见文档
   auto_multiline_extra_patterns = []
  
   ## 是否删除 ansi 转义码，例如标准输出的文本颜色等
   remove_ansi_escape_codes = false
  
   ## 忽略不活跃的文件，例如文件最后一次修改是 20 分钟之前，距今超出 10m，则会忽略此文件
   ## 时间单位支持 "ms", "s", "m", "h"
   ignore_dead_log = "1h"
  
   ## 是否从文件首部开始读取
   from_beginning = false
  
   # 自定义 tags
   [inputs.logging.tags]
   # some_tag = "some_value"
   # more_tag = "some_other_value"
```

这是一份基础的 logging.conf，其中 `multiline_match`、`pipeline` 等字段的作用详见后文的功能介绍。

### 容器标准输出采集 {#config-stdout}

采集容器应用的标准输出，也是最常见的方式，可以使用类似 `docker logs` 或 `crictl logs` 查看。

控制台输出（即 stdout/stderr）通过容器 Runtime 落盘到文件，Datakit 会自动获取到该容器的 `logpath` 进行采集。

如果要自定义采集的配置，可以通过添加容器环境变量或 Kubernetes Pod Annotation 的方式。

- 自定义配置的 key 有以下几种情况：
    - 容器环境变量的 key 固定为 `datakit_logs_config`
    - Pod Annotation 的 key 有两种写法：
        - `datakit/<container_name>.logs`，其中 `<container_name>` 需要替换为当前 Pod 的容器名，这在多容器环境下会用到
        - `datakit/logs` 会对该 Pod 的所有容器都适用

<!-- markdownlint-disable md046 -->
???+ info

    如果一个容器存在环境变量 `datakit_logs_config`，同时又能找到它所属 Pod 的 Annotation `datakit/logs`，按照就近原则，以容器环境变量的配置为准。
<!-- markdownlint-enable -->

- 自定义配置的 value 如下：

``` json
[
 {
  "disable" : false,
  "source" : "<your-source>",
  "service" : "<your-service>",
  "pipeline": "<your-pipeline.p>",
  "remove_ansi_escape_codes": false,
  "from_beginning"     : false,
  "tags" : {
   "<some-key>" : "<some_other_value>"
  }
 }
]
```

字段说明：

| 字段名                     | 取值             | 说明                                                                                                                                                                       |
| -----                      | ----             | ----                                                                                                                                                                       |
| `disable`                  | true/false       | 是否禁用该容器的日志采集，默认是 `false`                                                                                                                                   |
| `type`                     | `file`/不填      | 选择采集类型。如果是采集容器内文件，必须写成 `file`。默认为空是采集 `stdout/stderr`                                                                                        |
| `path`                     | 字符串           | 配置文件路径。如果是采集容器内文件，必须填写 Volume 的 path，注意不是容器内的文件路径，是容器外能访问到的路径。默认采集 `stdout/stderr` 不用填                             |
| `source`                   | 字符串           | 日志来源，参见[容器日志采集的 source 设置](../integrations/container.md#config-logging-source)                                                                                      |
| `service`                  | 字符串           | 日志隶属的服务，默认值为日志来源（source）                                                                                                                                 |
| `pipeline`                 | 字符串           | 适用该日志的 pipeline 脚本，默认值为与日志来源匹配的脚本名（`<source>.p`）                                                                                                 |
| `remove_ansi_escape_codes` | true/false       | 是否删除日志数据的颜色字符                                                                                                                                                 |
| `from_beginning`           | true/false       | 是否从文件首部采集日志                                                                                                                                                     |
| `multiline_match`          | 正则表达式字符串 | 用于[多行日志匹配](../integrations/logging.md#multiline)时的首行识别，例如 `"multiline_match":"^\\d{4}"` 表示行首是 4 个数字，在正则表达式规则中 `\d` 是数字，前面的 `\` 是用来转义 |
| `character_encoding`       | 字符串           | 选择编码，如果编码有误会导致数据无法查看，支持 `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""。默认为空即可                                                       |
| `tags`                     | key/value 键值对 | 添加额外的 tags，如果已经存在同名的 key 将以此为准（[:octicons-tag-24: version-1.4.6](changelog.md#cl-1.4.6) ）                                                    |

完整示例如下：

<!-- markdownlint-disable md046 -->
=== "容器环境变量"

    ``` shell
      $ cat dockerfile
      from pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04 as base
      run mkdir -p /opt
      run echo 'i=0; \n\
      while true; \n\
      do \n\
        echo "$(date +"%y-%m-%d %h:%m:%s") [$i] bash for loop examples. hello, world! testing output."; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
      done \n'\
      >> /opt/s.sh
      cmd ["/bin/bash", "/opt/s.sh"]
      
      ## 构建镜像
      $ docker build -t testing/log-output:v1 .
      
      ## 启动容器，添加环境变量 datakit_logs_config
      $ docker run --name log-output -env datakit_logs_config='[{"disable":false,"source":"log-source","service":"log-service"}]' -d testing/log-output:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="log-demo.yaml"
    apiversion: apps/v1
    kind: deployment
    metadata:
     name: log-demo-deployment
     labels:
      app: log-demo
    spec:
     replicas: 1
     selector:
      matchlabels:
       app: log-demo
     template:
      metadata:
       labels:
        app: log-demo
       annotations:
        ## 添加配置，且指定容器为 log-output
        datakit/log-output.logs: |
         [{
           "disable": false,
           "source": "log-output-source",
           "service": "log-output-service",
           "tags" : {
            "some_tag": "some_value"
           }
         }]
      spec:
       containers:
       - name: log-output
        image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
        args:
        - /bin/sh
        - -c
        - >
         i=0;
         while true;
         do
          echo "$(date +'%f %h:%m:%s') [$i] bash for loop examples. hello, world! testing output.";
          i=$((i+1));
          sleep 1;
         done
    ```
    
    执行 Kubernetes 命令，应用该配置：
    
    ``` shell
    $ kubectl apply -f log-output.yaml
    #...
    ```

---

???+ attention

    - 如无必要，不要轻易在环境变量和 Pod Annotation 中配置 Pipeline，一般情况下，通过 `source` 字段自动推导即可。
    - 如果是在配置文件或终端命令行添加 Environment/Annotations，两边是英文状态双引号，需要添加转义字符。


`multiline_match` 的值是双重转义，4 根斜杠才能表示实际的 1 根，例如 `\"multiline_match\":\"^\\\\d{4}\"` 等价 `"multiline_match":"^\d{4}"`，示例：

```shell
$ kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"log-source\",\"service\":\"log-service\",\"pipeline\":\"test.p\",\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
#...
```


如果一个 Pod/容器日志已经在采集中，此时再通过 `kubectl annotate` 命令添加配置不生效。


### 容器内日志文件采集 {#config-container-files}

对于容器内部的日志文件，和控制台输出日志的区别是需要指定文件路径，其他配置项大同小异。

同样是添加容器环境变量或 Kubernetes Pod Annotation 的方式，key 和 value 基本一致，详见前文。

完整示例如下：

<!-- markdownlint-disable md046 -->
=== "容器环境变量"

    ``` shell
      $ cat dockerfile
      from pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04 as base
      run mkdir -p /opt
      run echo 'i=0; \n\
      while true; \n\
      do \n\
        echo "$(date +"%y-%m-%d %h:%m:%s") [$i] bash for loop examples. hello, world! testing output." >> /tmp/opt/1.log; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
      done \n'\
      >> /opt/s.sh
      cmd ["/bin/bash", "/opt/s.sh"]
      
      ## 构建镜像
      $ docker build -t testing/log-to-file:v1 .
      
      ## 启动容器，添加环境变量 datakit_logs_config，注意字符转义
      ## 指定非 stdout 路径，"type" 和 "path" 是必填字段，且需要创建采集路径的 Volume
      ## 例如采集 `/tmp/opt/1.log` 文件，需要添加 `/tmp/opt` 的匿名 Volume
      $ docker run --env datakit_logs_config="[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/1.log\",\"source\":\"log-source\",\"service\":\"log-service\"}]" -v /tmp/opt -d testing/log-to-file:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="logging.yaml"
    apiversion: apps/v1
    kind: deployment
    metadata:
     name: log-demo-deployment
     labels:
      app: log-demo
    spec:
     replicas: 1
     selector:
      matchlabels:
       app: log-demo
     template:
      metadata:
       labels:
        app: log-demo
       annotations:
        ## 添加配置，且指定容器为 logging-demo
        ## 同时配置了 file 和 stdout 两种采集。注意要采集 "/tmp/opt/1.log" 文件，需要先给 "/tmp/opt" 添加 emptydir Volume
        datakit/logging-demo.logs: |
         [
          {
           "disable": false,
           "type": "file",
           "path":"/tmp/opt/1.log",
           "source": "logging-file",
           "tags" : {
            "some_tag": "some_value"
           }
          },
          {
           "disable": false,
           "source": "logging-output"
          }
         ]
      spec:
       containers:
       - name: logging-demo
        image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
        args:
        - /bin/sh
        - -c
        - >
         i=0;
         while true;
         do
          echo "$(date +'%f %h:%m:%s') [$i] bash for loop examples. hello, world! testing output.";
          echo "$(date +'%f %h:%m:%s') [$i] bash for loop examples. hello, world! testing output." >> /tmp/opt/1.log;
          i=$((i+1));
          sleep 1;
         done
        volumemounts:
        - mountpath: /tmp/opt
         name: datakit-vol-opt
       volumes:
       - name: datakit-vol-opt
        emptydir: {}
    ```
    
    执行 Kubernetes 命令，应用该配置：
    
    ``` shell
    $ kubectl apply -f logging.yaml
    #...
    ```
    
    对于容器内部的日志文件，在 Kubernetes 环境中还可以通过添加 sidecar 实现采集，参见[这里](../integrations/logfwd.md)。

### TCP/UDP 数据接收 {#config-tcpudp}

将 logging.conf 中 `logfiles` 注释掉，并配置 `sockets`，例如：

```yaml
[[inputs.logging]]
 sockets = [
  "tcp://127.0.0.1:9540"
  "udp://127.0.0.1:9541"
 ]

 source = "socket-testing"
```

上面这份配置会同时监听 tcp 端口 9540，和 udp 端口 9541，从这两个端口收到的日志数据，source 统一是 `socket-testing`。

使用 Linux 的 `nc` 命令测试如下：

```shell
# 发送 tcp 数据，按 ctrl-c 退出
$ nc 127.0.0.1 9540
This is a tcp message-1
This is a tcp message-2
This is a tcp message-3

# 发送一条 udp 数据
$ echo "This is a udp message" | nc -w 3 -v -u 127.0.0.1 9541
Connection to 127.1 (127.0.0.1) 9531 port [udp/*] succeeded!
```

> 注意：UDP 数据因为缺少上下文，所以不适用多行匹配。

### HTTP 数据接收 {#config-logstreaming}

启动一个 HTTP server，接收日志文本数据，上报到<<<custom_key.brand_name>>>。HTTP URL 固定为：`/v1/write/logstreaming`，即 `http://datakit_ip:port/v1/write/logstreaming`

> 注：如果 Datakit 以 DaemonSet 方式部署在 Kubernetes 中，可以使用 Service 方式访问，地址为 `http://datakit-service.datakit:9529`

<!-- markdownlint-disable md046 -->
=== "主机安装"

    进入 Datakit 安装目录下的 `conf.d/log` 目录，复制 `logstreaming.conf.sample` 并命名为 `logstreaming.conf`。示例如下：
      
    ```toml
    [inputs.logstreaming]
      ignore_url_tags = false
    
      ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
      ## buffer is the size of jobs' buffering of worker channel.
      ## threads is the total number fo goroutines at running time.
      # [inputs.logstreaming.threads]
      #   buffer = 100
      #   threads = 8
    
      ## Storage config a local storage space in hard dirver to cache trace data.
      ## path is the local file path used to cache data.
      ## capacity is total space size(MB) used to store data.
      # [inputs.logstreaming.storage]
      #   path = "./log_storage"
      #   capacity = 5120
    ```
    
    配置好后，[重启 Datakit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

#### logstreaming 支持参数 {#logstreaming-args}

Logstreaming 支持在 http url 中，通过添加参数可实现对日志数据的灵活操作。以下是支持的参数及其作用：

- `type`：数据格式，目前只支持 `influxdb` 和 `firelens`。
    - 当 `type` 为 `inflxudb` 时（`/v1/write/logstreaming?type=influxdb`），说明数据本身就是行协议格式（默认 precision 是 `s`），将只添加内置 tags，不再做其他操作
    - 当 `type` 为 `firelens` 时 (`/v1/write/logstreaming?type=firelens`)，数据格式应是 JSON 格式的多条日志
    - 当此值为空时，会对数据做分行和 Pipeline 等处理
- `source`：标识数据来源，即行协议的 measurement。例如 `nginx` 或者 `redis`（`/v1/write/logstreaming?source=nginx`）
    - 当 `type` 是 `influxdb` 时，此值无效
    - 默认为 `default`
- `service`：添加 service 标签字段，例如（`/v1/write/logstreaming?service=nginx_service`）
    - 默认为 `source` 参数值。
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/logstreaming?pipeline=nginx.p`）
- `tags`：添加自定义 tag，以英文逗号 `,` 分割，例如 `key1=value1` 和 `key2=value2`（`/v1/write/logstreaming?tags=key1=value1,key2=value2`）

#### logstreaming 使用案例 {#logstreaming-cases}

- Fluentd 使用 influxdb output [文档](https://github.com/fangli/fluent-plugin-influxdb){:target="_blank"}
- Fluentd 使用 http output [文档](https://docs.fluentd.org/output/http){:target="_blank"}
- logstash 使用 influxdb output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-influxdb.html){:target="_blank"}
- logstash 使用 http output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-http.html){:target="_blank"}

只需要将 output host 配置为 log-streaming url（`http://datakit_ip:port/v1/write/logstreaming`）并添加对应参数即可。

## 日志采集细节 {#logging-collect-details}

### 发现日志文件 {#discovery}

#### glob 配置文件路径 {#glob-discovery}

日志采集使用 glob 规则更方便地指定日志文件，以及自动发现和文件过滤。

| 通配符   | 描述                               | 正则示例       | 匹配示例                    | 不匹配                        |
| :--      | ---                                | ---            | ---                         | ----                          |
| `*`      | 匹配任意数量的任何字符，包括无     | `law*`         | `law, laws, lawyer`         | `groklaw, la, aw`             |
| `?`      | 匹配任何单个字符                   | `?at`          | `cat, cat, bat, bat`        | `at`                          |
| `[abc]`  | 匹配括号中给出的一个字符           | `[cb]at`       | `cat, bat`                  | `cat, bat`                    |
| `[a-z]`  | 匹配括号中给出的范围中的一个字符   | `letter[0-9]`  | `letter0, letter1, letter9` | `letters, letter, letter10`   |
| `[!abc]` | 匹配括号中未给出的一个字符         | `[!c]at`       | `bat, bat, cat`             | `cat`                         |
| `[!a-z]` | 匹配不在括号内给定范围内的一个字符 | `letter[!3-5]` | `letter1…`                  | `letter3 … letter5, letterxx` |

另需说明，除上述 glob 标准规则外，采集器也支持 `**` 进行递归地文件遍历，如示例配置所示。更多 grok 介绍，参见[这里](https://rgb-24bit.github.io/blog/2018/glob.html){:target="_blank"}。

#### 发现本地日志文件 {#local-file-discovery}

普通日志文件是最常见的一种，它们是进程直接将可读的记录数据写到磁盘文件，像著名的 “log4j” 框架或者执行 `echo "this is log" >> /tmp/log` 命令都会产生日志文件。

这种日志的文件路径大部分情况都是固定的，像 MySQL 在 Linux 平台的日志路径是 `/var/log/mysql/mysql.log`，如果运行 Datakit mysql 采集器，默认会去找个路径找寻日志文件。但是日志存储路径是可配的，Datakit 无法兼顾所有情况，所以必须支持手动指定文件路径。

在 Datakit 中使用 glob 模式配置文件路径，它使用通配符来定位文件名（当然也可以不使用通配符）。

举个例子，现在有以下的文件：

```shell
$ tree /tmp

/tmp
├── datakit
│   ├── datakit-01.log
│   ├── datakit-02.log
│   └── datakit-03.log
└── mysql.d
  └── mysql
    └── mysql.log

3 directories, 4 files
```

在 Datakit logging 采集器中可以通过配置 `logfiles` 参数项，指定要采集的日志文件，比如：

- 采集 `datakit` 目录下所有文件，glob 为 `/tmp/datakit/*`
- 采集所有带有 `datakit` 名字的文件，对应的 glob 为 `/tmp/datakit/datakit-*log`
- 采集 `mysql.log`，但是中间有 `mysql.d` 和 `mysql` 两层目录，有好几种方法定位到 `mysql.log` 文件：
    - 直接指定：`/tmp/mysql.d/mysql/mysql.log`
    - 单星号指定：`/tmp/*/*/mysql.log`，这种方法基本用不到
    - 双星号（`double star`）：`/tmp/**/mysql.log`，使用双星号 `**` 代替中间的多层目录结构，是较为简洁、常用的一种方式

在配置文件中使用 glob 指定文件路径后，Datakit 会定期在磁盘中搜寻符合规则的文件，如果发现没有在采集列表中，便将其添加并进行采集。


#### 定位容器标准输出日志 {#discovery-container-stdout-logs}

在容器中输出日志有两种方式：

- 一是直接写到挂载的磁盘目录，这种方式在主机看来和上述的 “普通日志文件” 相同，都是在磁盘固定位置的文件
- 另一种方式是输出到 stdout/stderr，由容器的 Runtime 来收集并管理落盘，这也是较为常见的方式。这个落盘路径通过访问 Runtime api 可以获取到

Datakit 通过连接 docker 或 containerd 的 sock 文件，访问它们的 api 获取指定容器的 `logpath`，类似在命令行执行 `docker inspect --format='{{`{{.logpath}}`}}' <container_id>`：

```shell
$ docker inspect --format='{{`{{.logpath}}`}}' <container_id>

/var/lib/docker/containers/<container_id>/<container_id>-json.log
```

获取到容器 `logpath` 后，使用这个路径和相应配置创建日志采集。


#### 定位容器内文件路径 {#discovery-container-file-logs}

要采集容器内文件，需要现将文件目录挂载到外部，此处只描述 Kubernetes 场景的容器内文件路径定位方案。

以前文 “容器内日志文件采集” 为例，采集容器内 `/tmp/opt/1.log` 文件，先给目录 `/tmp/opt` 创建一个 `emptyDir` Volume，此时在宿主机上执行命令（以 containerd Runtime 为例）：

```shell
$ crictl inspect <container_id>
#...
```

能看到有一个 Mount 如下：

```json
  "mounts": [
    {
      "containerpath": "/tmp/opt",
      "hostpath": "/var/lib/kubelet/pods/<pod_id>/volumes/kubernetes.io~empty-dir/datakit-vol-opt",
      "propagation": "propagation_private",
      "selinuxrelabel": false
    }
  ]
```

其中 `containerpath` 就是挂载的容器内目录，`hostpath` 是它在宿主机上的目录，在这个目录下有一个名为 `1.log` 的文件，就是要采集的日志文件。

Datakit 会从容器 Runtime 找到这个 `hostpath`，根据配置的采集文件是 `1.log` 拼接成 `/var/lib/kubelet/pods/<pod_id>/volumes/kubernetes.io~empty-dir/datakit-vol-opt/1.log`。

又因为 Datakit 会将宿主机根目录挂载到容器的 `/rootfs` 路径下，所以最终变成 `/rootfs/var/lib/kubelet/pods/<pod_id>/volumes/kubernetes.io~empty-dir/datakit-vol-opt/1.log`，这就是 Datakit 实际要采集的文件。

> Datakit 以 Kubernetes DaemonSet 方式部署时，会创建一个 VolumeMount，将宿主机的根目录 `/` 挂载到 Datakit 容器的 `/rootfs` 路径下，使得 Datakit 进程能只读访问宿主机文件。

### 文件过滤 {#skip-logs}

在配置文件路径时使用 glob 规则，可能会找到很多文件，有的文件长时间不更新，不希望打开和采集它。需要用到 `ignore_dead_log` 这个配置项。

`ignore_dead_log` 是一个时间配置，可以写成 `1h` 表示 1 小时，或 `20m` 表示 20 分钟。

Datakit 发现到这个文件时，会获取它的最后修改时间（`modify time`），使用当前时间计算得出该文件的最后修改时长。如果超过 1 小时或 20 分钟没更新，就忽略这个文件，不进行采集。

当下一次又发现这个文件，再次执行上面的判断。只有当该文件有变动时，才创建采集。

对于正在打开和采集的文件，如果超过时长没有更新，采集会退出并关闭文件资源，避免长时间占用。

> 容器标准输出和容器内文件的 `ignore_dead_log` 统一为 `1h` 即 1 小时。

### 采集的起始位置 {#start-pos}

文件采集的起始位置非常重要：如果始终从 head 采集可能采集重复，如果从 tail 采集可能会丢失部分数据。

Datakit 对采集的起始位置有几个判断方案，按照优先级依次是：

1. 从文件上次的采集位置继续采集
  Datakit 根据文件信息计算哈希值，在 `/usr/local/datakit/cache/logtail.history` 记录该文件的当前采集 position。如果能准去找到这个 position，且该值小于等于文件 size（说明没有被 truncated），从这个 position 继续采集。

1. 配置 `from_beginning` 强制从 head 采集
  当 `from_beginning` 为 `true`，且没有找到该文件的 position 记录时，从 head 开始采集。

1. 配置 `from_beginning_threshold_size` 根据文件大小判断是否从 head 采集
  `from_beginning_threshold_size` 是一个整数值。Datakit 发现到新文件时，如果文件 size 小于 `from_beginning_threshold_size` 就从文件 head 开始采集，否则从 tail 采集。默认值是 `2e7`（即 20 mb）。

1. 从 tail 开始采集
  默认情况下，从 tail 采集文件。

什么情况下该使用 `from_beginning`？

- 需要完整地采集某个日志文件，例如备份旧数据。需要注意，对于旧数据通常长时间不更新文件，可能被 `ignore_dead_log` 过滤掉。
- 在非 Linux 系统下，日志数据量很大，例如 10 秒钟写入超过 20 MB（`from_beginning_threshold_size`） 的场景，需要开启 `from_beginning`。

对于第二点进行补充：

Datakit 日志采集每 10 秒搜寻一次新文件，如果刚好卡在这个时间边界，使得 Datakit 下次搜寻到这个文件时（即 10 秒后）已经超过 20 MB，按照策略应该从 tail 采集，这种情况需要强制开启 `from_beginning` 从 head 采集。同时因为 Linux 的 Inotify 事件通知机制存在，理论上当文件产生，Datakit 立刻就能捕获到并开启采集，这个间隔很短不会产生超过 20 MB 的数据，所以也不必须开启 `from_beginning`。


### 数据编码转换 {#encoding}

Datakit 支持采集 `utf-8` 和 `gbk` 编码的文件。

### 多行数据拼接 {#multi-line}

通过识别多行日志的第一行特征，即可判定某行日志是不是一条新的日志。如果不符合这个特征，即认为当前行日志只是前一条多行日志的追加。

举例说明：一般情况下日志都是顶格写，但有些日志文本不是顶格写的，比如程序崩溃时的调用栈日志，这种日志文本就是多行日志。

在 Datakit 中，通过正则表达式来识别多行日志特征，正则匹配上的日志行，就是一条新的日志的开始；后续所有不匹配的日志行，都认为是这条新日志的追加，直到遇到另一行匹配正则的新日志为止。

在 `logging.conf` 中，修改如下配置：

```toml
  multiline_match = '''这里填写具体的正则表达式''' # 正则表达式建议用三个“英文单引号”括起来
```

日志采集器中使用的正则表达式风格[参考](https://golang.org/pkg/regexp/syntax/#hdr-syntax){:target="_blank"}

假定原数据为：

```not-set
2020-10-23 06:41:56,688 info demo.py 1.0
2020-10-23 06:54:20,164 error /usr/local/lib/python3.6/dist-packages/flask/app.py exception on /0 [get]
traceback (most recent call last):
 file "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
  response = self.full_dispatch_request()
zerodivisionerror: division by zero
2020-10-23 06:41:56,688 info demo.py 5.0
```

`multiline_match` 配置为 `^\\d{4}-\\d{2}-\\d{2}.*` 时，（意即匹配形如 `2020-10-23` 这样的行首）

切割出的三个行协议点如下（行号分别是 1/2/8）。可以看到 `traceback ...` 这一段（第 3 ~ 6 行）没有单独形成一条日志，而是追加在上一条日志（第 2 行）的 `message` 字段中。

```not-set
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 info demo.py 1.0" 1611746438938808642
testing,filename=/tmp/094318188 message="2020-10-23 06:54:20,164 error /usr/local/lib/python3.6/dist-packages/flask/app.py exception on /0 [get]
traceback (most recent call last):
 file \"/usr/local/lib/python3.6/dist-packages/flask/app.py\", line 2447, in wsgi_app
  response = self.full_dispatch_request()
zerodivisionerror: division by zero
" 1611746441941718584
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 info demo.py 5.0" 1611746443938917265
```

除了手写多行匹配规则之外，还有自动多行模式，开启此模式后，每一行日志数据都会在多行列表中匹配。如果匹配成功，就将当前的多行规则权重加一，以便后面能更快速的匹配到，然后退出匹配循环；如果整个列表结束依然没有匹配到，则认为匹配失败。

匹配成功与失败，后续操作和正常的多行日志采集是一样的：匹配成功，会将现存的多行数据发送出去，并将本条数据填入；匹配失败，会追加到现存数据的尾端。

因为日志存在多个多行配置，它们的优先级如下：

1. `multiline_match` 不为空，只使用当前规则
1. 使用 source 到 `multiline_match` 的映射配置（只在容器日志配置存在 `logging_source_multiline_map`），如果使用 source 能找到对应的多行规则，只使用此规则
1. 开启 `auto_multiline_detection`，如果 `auto_multiline_extra_patterns` 不为空，会在这些多行规则中匹配
1. 开启 `auto_multiline_detection`，如果 `auto_multiline_extra_patterns` 为空，使用默认的自动多行匹配规则列表，即：

```not-set
// time.rfc3339, "2006-01-02t15:04:05z07:00"
`^\d+-\d+-\d+t\d+:\d+:\d+(\.\d+)?(z\d*:?\d*)?`,

// time.ansic, "mon jan _2 15:04:05 2006"
`^[a-za-z_]+ [a-za-z_]+ +\d+ \d+:\d+:\d+ \d+`,

// time.rubydate, "mon jan 02 15:04:05 -0700 2006"
`^[a-za-z_]+ [a-za-z_]+ \d+ \d+:\d+:\d+ [\-\+]\d+ \d+`,

// time.unixdate, "mon jan _2 15:04:05 mst 2006"
`^[a-za-z_]+ [a-za-z_]+ +\d+ \d+:\d+:\d+( [a-za-z_]+ \d+)?`,

// time.rfc822, "02 jan 06 15:04 mst"
`^\d+ [a-za-z_]+ \d+ \d+:\d+ [a-za-z_]+`,

// time.rfc822z, "02 jan 06 15:04 -0700" // rfc822 with numeric zone
`^\d+ [a-za-z_]+ \d+ \d+:\d+ -\d+`,

// time.rfc850, "monday, 02-jan-06 15:04:05 mst"
`^[a-za-z_]+, \d+-[a-za-z_]+-\d+ \d+:\d+:\d+ [a-za-z_]+`,

// time.rfc1123, "mon, 02 jan 2006 15:04:05 mst"
`^[a-za-z_]+, \d+ [a-za-z_]+ \d+ \d+:\d+:\d+ [a-za-z_]+`,

// time.rfc1123z, "mon, 02 jan 2006 15:04:05 -0700" // rfc1123 with numeric zone
`^[a-za-z_]+, \d+ [a-za-z_]+ \d+ \d+:\d+:\d+ -\d+`,

// time.rfc3339nano, "2006-01-02t15:04:05.999999999z07:00"
`^\d+-\d+-\d+[a-za-z_]+\d+:\d+:\d+\.\d+[a-za-z_]+\d+:\d+`,

// 2021-07-08 05:08:19,214
`^\d+-\d+-\d+ \d+:\d+:\d+(,\d+)?`,

// default java logging simpleformatter date format
`^[a-za-z_]+ \d+, \d+ \d+:\d+:\d+ (am|pm)`,

// 2021-01-31 - with stricter matching around the months/days
`^\d{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])`,
```

### 特殊字节码过滤 {#escaping}

日志可能会包含一些不可读的特殊字符（比如终端输出的颜色等），可以将 `remove_ansi_escape_codes` 设置为 `true` 对其删除过滤。

“特殊字符” 在此处代指数据中的颜色字符，比如以下命令会在命令行终端输出一个红色 `red` 单词：

```shell
$ red='\033[0;31m' && nc='\033[0m' && print "${red}red${nc}"
#...
```

如果不进行处理，那么采集到的日志是 `\033[0;31m`，不仅缺乏美观、占用存储，还可能对后续的数据处理产生负影响。所以要在此处筛除特殊颜色字符。

<!-- markdownlint-disable md046 -->
???+ attention

开源社区有许多案例，大部分都使用正则表达式进行实现，性能不是很出色。但是对于一个成熟的日志输出框架，一定有关闭颜色字符的方法，Datakit 更推荐这种做法，由日志产生端从避免打印颜色字符。

对于此类颜色字符，通常建议在日志输出框架中关闭，而不是由 Datakit 进行过滤。特殊字符的筛选和过滤是由正则表达式处理，可能覆盖不够全面，且有一定的性能开销。
<!-- markdownlint-enable -->

处理性能基准测试结果如下，仅供参考：

```text
goos: linux
goarch: arm64
pkg: ansi
benchmarkstrip
benchmarkstrip-2     653751       1775 ns/op       272 b/op     3 allocs/op
benchmarkstrip-4     673238       1801 ns/op       272 b/op     3 allocs/op
pass
ok   ansi   2.422s
```

每一条文本的处理耗时增加 1700 ns 不等。如果不开启此功能将无额外损耗。

### Pipeline 字段切割 {#pipeline}

[Pipeline](../pipeline/use-pipeline/index.md) 主要用于切割非结构化的文本数据，或者用于从结构化的文本中（如 JSON）提取部分信息。

对日志数据而言，主要提取两个字段：

- `time`：即日志的产生时间，如果没有提取 `time` 字段或解析此字段失败，默认使用系统当前时间
- `status`：日志的等级，如果没有提取出 `status` 字段，则默认将 `stauts` 置为 `unknown`

有效的 `status` 字段值如下（不区分大小写）：

| 日志可用等级          | 简写    | studio 显示值 |
| ------------          | :----   | ----          |
| `alert`               | `a`     | `alert`       |
| `critical`            | `c`     | `critical`    |
| `error`               | `e`     | `error`       |
| `warning`             | `w`     | `warning`     |
| `notice`              | `n`     | `notice`      |
| `info`                | `i`     | `info`        |
| `debug/trace/verbose` | `d`     | `debug`       |
| `ok`                  | `o`/`s` | `ok`          |

示例：假定文本数据如下：

```not-set
12115:m 08 jan 17:45:41.572 # server started, redis version 3.0.6
```

Pipeline 脚本：

```python
add_pattern("date2", "%{monthday} %{month} %{year}?%{time}")
grok(_, "%{int:pid}:%{word:role} %{date2:time} %{notspace:serverity} %{greedydata:msg}")
group_in(serverity, ["#"], "warning", status)
cast(pid, "int")
default_time(time)
```

最终结果：

```python
{
  "message": "12115:m 08 jan 17:45:41.572 # server started, redis version 3.0.6",
  "msg": "server started, redis version 3.0.6",
  "pid": 12115,
  "role": "m",
  "serverity": "#",
  "status": "warning",
  "time": 1610127941572000000
}
```

Pipeline 的几个注意事项：

- 如果 logging.conf 配置文件中 `pipeline` 为空，默认使用 `<source-name>.p`（假定 `source` 为 `nginx`，则默认使用 `nginx.p`）
- 如果 `<source-name>.p` 不存在，将不启用 Pipeline 功能
- 所有 Pipeline 脚本文件，统一存放在 Datakit 安装路径下的 `pipeline` 目录下

### 根据白名单保留指定字段 {#field-whitelist}

日志采集有以下基础字段：

| 字段名                   | 是否仅在容器日志存在 |
| -----------              | -----------          |
| `service`                |                      |
| `status`                 |                      |
| `filepath`               |                      |
| `host`                   |                      |
| `log_read_lines`         |                      |
| `container_id`           | 是                   |
| `container_name`         | 是                   |
| `namespace`              | 是                   |
| `pod_name`               | 是                   |
| `pod_ip`                 | 是                   |
| `deployment`/`daemonset` | 是                   |

在特殊场景下，很多基础字段不是必要的。现在提供一个白名单（whitelist）功能，只保留指定的字段。

字段白名单只支持环境变量配置，例如 `ENV_LOGGING_FIELD_WHITE_LIST = '["host", "service", "filepath", "container_name"]'`，具体细节如下：

- 如果 whitelist 为空，则添加所有基础字段
- 如果 whitelist 不为空，且值有效，例如 `["filepath", "container_name"]`，则只保留这两个字段
- 如果 whitelist 不为空，且全部是无效字段，例如 `["no-exist"]` 或 `["no-exist-key1", "no-exist-key2"]`，则这条数据被丢弃

对于其他来源的 tags 字段，有以下几种情况：

- whitelist 对 Datakit 的全局标签（`global tags`）不生效
- 通过 `ENV_ENABLE_DEBUG_FIELDS = "true"` 开启的 debug 字段不受影响，包括日志采集的 `log_read_offset` 和 `log_file_inode` 两个字段，以及 `pipeline` 的 debug 字段

<!-- markdownlint-disable md046 -->
???+ attention

字段白名单是一个全局配置，同时对容器日志和 logging 采集器生效。
<!-- markdownlint-enable -->

### 最大文件采集数 {#max-collecting-files}

通过环境变量 `ENV_LOGGING_MAX_OPEN_FILES` 设置最大文件采集数量。例如 `ENV_LOGGING_MAX_OPEN_FILES="1000"`，表示最多同时采集 1000 个文件。

避免了使用 glob 规则匹配太多文件，导致 Datakit 资源占用过高。

`ENV_LOGGING_MAX_OPEN_FILES` 是一个全局配置，对 logging 采集器和容器日志采集同时生效。它的默认值是 500。

## 采集细节描述 {#more-colect-details}

### 容器日志采集的 source 设置 {#source-for-container-log}

在容器环境下，日志来源（`source`）设置是一个很重要的配置项，它直接影响在页面上的展示效果。但如果挨个给每个容器的日志配置一个 source 未免残暴。如果不手动配置容器日志来源，Datakit 有如下规则（优先级递减）用于自动推断容器日志的来源：

> 所谓不手动指定容器日志来源，就是指在 Pod Annotation 中不指定，在 container.conf 中也不指定（目前 container.conf 中无指定容器日志来源的配置项）

- Kubernetes 指定的容器名：从容器的 `io.kubernetes.container.name` 这个 label 上取值
- 容器本身的名称：通过 `docker ps` 或 `crictl ps` 能看到的容器名
- `default`: 默认的 `source`

### 超长多行日志处理的限制 {#split-large-log}

Datakit 日志采集最多支持单行日志 512 kib，如果超过 512 kib 会截断，从下一个字节开始是新一条日志。

多行日志最多允许 `ENV_DATAWAY_MAX_RAW_BODY_SIZE` * 0.8，默认情况下这个值是 819 kib，如果超过此值同样会截断，并创建新一条日志。

### 磁盘文件的发现过程 {#scan-and-inotify}

Datakit 发现日志文件的过程，是定期（10 秒）根据 glob 规则扫描文件，和 inotify 事件通知机制。其中 inotify 只有 Datakit 部署在 Linux 环境下才支持。

### 文件翻转方案推荐 {#file-rotate}

日志文件的翻转（rotate）是常见行为，大部分日志输出框架都有这个功能。

一般情况下，文件翻转有 2 种情况：

- 以时间或数字作为文件名，始终创建新文件

例如，现在有一个日志文件是 `/opt/log/access-20241203-10.log`，当这个文件满足一定条件（通常是超过指定 size）不再更新，会创建一个新文件 `/opt/log/access-20241203-11.log` 继续写入。

这种翻转方式，推荐文件数量不要太多，保持在 5 个左右最佳。如果文件数量太多，会降低文件搜寻的性能。

推荐采集路径写成 `/opt/log/*.log`， Datakit 会监听目录或定期搜寻这个符合条件的 log 文件。对于采集结束，且不更新的旧文件，Datakit 会定期清理并释放资源。

- 文件名固定不变，通过重命名翻转文件

例如有文件 `/opt/log/access-0.log`，当文件写满后会重命名为 `/opt/log/access-1.log` 当做备份，同时再创建一个 `/opt/log/access-0.log` 继续写入。

这是 Docker `file-json` 的轮转方式，Docker 容器启动时可以通过参数 `--log-opt max-size=100m` 和 `--log-opt max-file=3` 调整文件大小和文件个数。

对于这种翻转方式，采集路径不能写成 `*.log`，否则会将备份文件也一起采集，导致重复采集。

采集路径要写成 `/opt/log/access-0.log`，只采集这一个文件。

> 此外还有一种翻转方式，类似 Linux 执行 `truncate -s 0` 将文件 size 截断为 0 字节，清空文件内容。不推荐使用这种翻转方式，它会彻底将文件内容清空，如果 Datakit 没有及时采集到末尾，末尾的数据就全部丢失了。


## FAQ {#faq}

### 远程文件采集方案 {#faq-remote-logs}

在 Linux 操作系统上，可通过 [nfs 方式](https://linuxize.com/post/how-to-mount-an-nfs-share-in-linux/){:target="_blank"}，将日志所在主机的文件路径，挂载到 Datakit 主机下，logging 采集器配置对应日志路径即可。

### macos 日志采集器报错 `operation not permitted` {#permission-on-macos}

在 macos 中，因为系统安全策略的原因，Datakit 日志采集器可能会无法打开文件，报错 `operation not permitted`，解决方法参考 [apple developer doc](https://developer.apple.com/documentation/security/disabling_and_enabling_system_integrity_protection){:target="_blank"}。

### 如何估算日志的总量 {#stat-log-usage}

由于日志的收费是按照条数来计费的，一般情况下，大部分的日志都是程序写到磁盘的，只能看到磁盘占用大小（比如每天 100gb 日志）。

一种可行的方式，可以用以下简单的 shell 来判断：

```shell
# 统计 1gb 日志的行数
head -c 1g path/to/your/log.txt | wc -l
```

有时候，要估计一下日志采集可能带来的流量消耗：

```shell
# 统计 1gb 日志压缩后大小（字节）
head -c 1g path/to/your/log.txt | gzip | wc -c
```

这里拿到的是压缩后的字节数，按照网络 bit 的计算方法（x8），其计算方式如下，以此可拿到大概的带宽消耗：

``` not-set
bytes * 2 * 8 /1024/1024 = xxx mbit
```

但实际上 Datakit 的压缩率不会这么高，因为 Datakit 不会一次性发送 1gb 的数据，而且分多次发送的，这个压缩率在 85% 左右（即 100mb 压缩到 15mb），故一个大概的计算方式是：

``` not-set
1gb * 2 * 8 * 0.15/1024/1024 = xxx mbit
```

<!-- markdownlint-disable md046 -->
??? info

此处 `*2` 考虑到了 [Pipeline 切割](../pipeline/use-pipeline/index.md)导致的实际数据膨胀，而一般情况下，切割完都是要带上原始数据的，故按照最坏情况考虑，此处以加倍方式来计算。
<!-- markdownlint-enable -->

### 日志目录的软链接问题 {#soft-links}

正常情况下，Datakit 会从容器 Runtime 获取日志文件的路径，然后采集该文件。

一些特殊环境，会对该日志所在目录做一个软链接，Datakit 无法提前获知软链接的目标，无法挂载该目录，导致找不到该日志文件，无法进行采集。

例如，现找到一个容器日志文件，路径是 `/var/log/pods/default_log-demo_f2617302-9d3a-48b5-b4e0-b0d59f1f0cd9/log-output/0.log`，但是在当前环境，`/var/log/pods` 是一个软连接指向 `/mnt/container_logs`，见下：

```shell
$ ls /var/log -lh
total 284k
lrwxrwxrwx 1 root root  20 oct 8 10:06 pods -> /mnt/container_logs/
```

Datakit 需要挂载 `/mnt/container_logs` hostpath 才能使得正常采集，例如在 `datakit.yaml` 中添加以下：

```yaml
  # 省略
  spec:
   containers:
   - name: datakit
    image: pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:1.16.0
    volumemounts:
    - mountpath: /mnt/container_logs
     name: container-logs
   # 省略
   volumes:
   - hostpath:
     path: /mnt/container_logs
    name: container-logs
```

这种情况不太常见，一般只有提前知道该路径有软连接，或查看 Datakit 日志发现采集报错才执行。

### 根据容器 image 来调整日志采集 {#config-logging-on-container-image}

默认情况下，Datakit 会收集所在主机上所有容器标准输出日志，这会采集较多的日志。可以通过 image 或 namespace 来过滤容器。

<!-- markdownlint-disable md046 -->
=== "主机安装"

    ``` toml
      ## 以 image 为例
      ## 当容器的 image 能够匹配 `datakit` 时，会采集此容器的日志
      container_include_log = ["image:datakit"]
    
      ## 忽略所有 kodo 容器
      container_exclude_log = ["image:kodo"]
    ```
    
    `container_include` 和 `container_exclude` 必须以属性字段开头，格式为一种[类正则的 glob 通配](https://en.wikipedia.org/wiki/glob_(programming)){:target="_blank"}：`"<字段名>:<glob 规则>"`
    
    现支持以下 4 个字段规则，这 4 个字段都是基础设施的属性字段：
    
    - image : `image:pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:1.18.0`
    - image_name : `image_name:pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit`
    - image_short_name : `image_short_name:datakit`
    - namespace : `namespace:datakit-ns`
    
    
    对于同一类规则（`image` 或 `namespace`），如果同时存在 `include` 和 `exclude`，需要同时满足 `include` 成立，且 `exclude` 不成立的条件。例如：
    
    ```toml
      ## 这会导致所有容器都被过滤
      ## 例如有一个容器 `datakit`，它满足 include，同时又满足 exclude，那么它会被过滤，不采集日志；如果一个容器 `nginx`，首先它不满足 include，它会被过滤掉不采集。
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
    
    - a：`hello/hello-http:latest`
    - b：`world/world-http:latest`
    - c：`pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:1.2.0`
    
    如果只希望采集 Pod a 的日志，那么配置 ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG 即可：
    
    ``` yaml
      - env:
       - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: image:hello* # 指定镜像名或其通配
    ```
    
    或以命名空间来配置：
    
    ``` yaml
      - env:
       - name: ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG
        value: namespace:foo # 指定命名空间的容器日志不采集
    ```

---

???+ tip "如何查看镜像"

    Docker：
    
    ``` shell
    $ docker inspect --format '{{`{{.config.image}}`}}' <container_id>
    #...
    ```
    
    Kubernetes Pod：
    
    ``` shell
    $ kubectl get pod -o=jsonpath="{.spec.containers[0].image}" <pod_name>
    #...
    ```
    
    ???+ attention
    
    通过全局配置的 `container_exclude_log` 优先级低于容器的自定义配置 `disable`。例如，配置了 `container_exclude_log = ["image:*"]` 不采集所有日志，如果有以下 Pod Annotation 还是会采集容器标准输出日志：
    
    ```json
    [{
      "disable": false,
      "source": "logging-output"
    }]
    ```
