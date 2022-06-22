{{.CSS}}
# JVM
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

这里我们提供俩类 JVM 指标采集方式，一种方案是 Jolokia，一种是 ddtrace。如何选择的方式，我们有如下建议：

- 如果采集诸如 Kafka 等 java 开发的中间件 JVM 指标，我们推荐 Jolokia 方案。 ddtrace 偏重于链路追踪（APM），且有一定的运行开销，对于中间件而言，链路追踪意义不大。
- 如果采集自己开发的 java 应用 JVM 指标，我们推荐 ddtrace 方案，除了能采集 JVM 指标外，还能实现链路追踪（APM）数据采集

## 通过 ddtrace 采集 JVM 指标

DataKit 内置了 [statsd 采集器](statsd.md)，用于接收网络上发送过来的 statsd 协议的数据。此处我们利用 ddtrace 来采集 JVM 的指标数据，并通过 statsd 协议发送给 DataKit。

### 准备 statsd 配置

这里推荐使用如下的 statsd 配置来采集 ddtrace JVM 指标。将其拷贝到 `conf.d/statsd` 目录下，并命名为 `ddtrace-jvm-statsd.conf`：

```toml
[[inputs.statsd]]
  protocol = "udp"

  ## Address and port to host UDP listener on
  service_address = ":8125"

  ## separator to use between elements of a statsd metric
  metric_separator = "_"

  drop_tags = ["runtime-id"]
  metric_mapping = [
    "jvm_:jvm",
    "datadog_tracer_:ddtrace",
  ]

  # 以下配置无需关注...

  delete_gauges = true
  delete_counters = true
  delete_sets = true
  delete_timings = true

  ## Percentiles to calculate for timing & histogram stats
  percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]

  ## Parses tags in the datadog statsd format
  ## http://docs.datadoghq.com/guides/dogstatsd/
  parse_data_dog_tags = true

  ## Parses datadog extensions to the statsd format
  datadog_extensions = true

  ## Parses distributions metric as specified in the datadog statsd format
  ## https://docs.datadoghq.com/developers/metrics/types/?tab=distribution#definition
  datadog_distributions = true

  ## Number of UDP messages allowed to queue up, once filled,
  ## the statsd server will start dropping packets
  allowed_pending_messages = 10000

  ## Number of timing/histogram values to track per-measurement in the
  ## calculation of percentiles. Raising this limit increases the accuracy
  ## of percentiles but also increases the memory usage and cpu time.
  percentile_limit = 1000

  ## Max duration (TTL) for each metric to stay cached/reported without being updated.
  #max_ttl = "1000h"

  [inputs.statsd.tags]
  # some_tag = "your-tag-value" 
  # some_other_tag = "your-other-tag-value"
```

关于这里的配置说明：

- `service_address` 此处设置成 `:8125`，指 ddtrace 将 jvm 指标发送出来的目标地址
- `drop_tags` 此处我们将 `runtime-id` 丢弃，因为其可能导致时间线爆炸。如确实需要该字段，将其从 `drop_tags` 中移除即可
- `metric_mapping` 在 ddtrace 发送出来的原始数据中，有俩类指标，它们的指标名称分别以 `jvm_` 和 `datadog_tracer_` 开头，故我们将它们统一规约到俩类指标集中，一个是 `jvm`，一个是 `ddtrace` 自身运行指标

### 启动 java 应用

一种可行的 JVM 部署方式如下：

```shell
java -javaagent:dd-java-agent.jar \
	-Ddd.profiling.enabled=true \
	-Ddd.logs.injection=true \
	-Ddd.trace.sample.rate=1 \
	-Ddd.service=my-app \
	-Ddd.env=staging \
	-Ddd.agent.host=localhost \
	-Ddd.agent.port=9529 \
	-Ddd.jmxfetch.enabled=true \
	-Ddd.jmxfetch.check-period=1000 \
	-Ddd.jmxfetch.statsd.host=127.0.0.1  \
	-Ddd.jmxfetch.statsd.port=8125 \
	-Ddd.version=1.0 \
	-jar your-app.jar
```

注意：

- 关于 `dd-jave-agent.jar` 包的下载，参见 [这里](ddtrace.md)
- 建议给如下几个字段命名：
	- `service` 用于表示该 JVM 数据来自哪个应用
	- `env` 用于表示该 JVM 数据来自某个应用的哪个环境（如 prod/testing/preprod 等）

- 此处几个选项的意义：
	- `-Ddd.jmxfetch.check-period` 表示采集频率，单位为毫秒
	- `-Ddd.jmxfetch.statsd.host=127.0.0.1` 表示 DataKit 上 statsd 采集器的连接地址
	- `-Ddd.jmxfetch.statsd.port=8125` 表示 DataKit 上 statsd 采集器的 UDP 连接端口，默认为 8125
	- `-Ddd.trace.health.xxx` ddtrace 自身指标数据采集和发送设置
  - 如果要开启链路追踪（APM）可追加如下两个参数（DataKit HTTP 地址）
		- `-Ddd.agent.host=localhost`
		- `-Ddd.agent.port=9529`

开启后，大概能采集到如下指标：

- `buffer_pool_direct_capacity`
- `buffer_pool_direct_count`
- `buffer_pool_direct_used`
- `buffer_pool_mapped_capacity`
- `buffer_pool_mapped_count`
- `buffer_pool_mapped_used`
- `cpu_load_process`
- `cpu_load_system`
- `gc_eden_size`
- `gc_major_collection_count`
- `gc_major_collection_time`
- `gc_metaspace_size`
- `gc_minor_collection_count`
- `gc_minor_collection_time`
- `gc_old_gen_size`
- `gc_survivor_size`
- `heap_memory_committed`
- `heap_memory_init`
- `heap_memory_max`
- `heap_memory`
- `loaded_classes`
- `non_heap_memory_committed`
- `non_heap_memory_init`
- `non_heap_memory_max`
- `non_heap_memory`
- `os_open_file_descriptors`
- `thread_count`

其中每个指标有如下 tags （实际 tags 受 java 启动参数以及 statsd 配置影响）

- `env`
- `host`
- `instance`
- `jmx_domain`
- `metric_type`
- `name`
- `service`
- `type`
- `version`

## 通过 Jolokia 采集 JVM 指标

JVM 采集器可以通过 JMX 来采取很多指标，并将指标采集到观测云，帮助分析 Java 运行情况。

### 前置条件

安装或下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-jvm/1.6.2/jolokia-jvm-1.6.2-agent.jar){:target="_blank"}。DataKit 安装目录下的 `data` 目录中已经有下载好的 Jolokia jar 包。通过如下方式开启 Java 应用： 

```shell
java -javaagent:/path/to/jolokia-jvm-agent.jar=port=8080,host=localhost -jar your_app.jar
```

### 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

### 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 视图预览

JVM性能指标展示：CPU负载、直接缓冲区、线程数量、堆内存、GC次数、类加载数等。

![image](imgs/input-jvm-1.png)

## 安装部署

说明：示例 通过ddtrace采集jvm指标，通过Datakit内置的statsd接收ddtrace发送过来的jvm指标

### 前置条件

-  服务器 <[安装 Datakit](../datakit/datakit-install.md)>

### 配置实施

#### 指标采集 (必选)

1、开启ddtrace， 复制sample文件，不需要修改ddtrace.conf

```
cd /usr/local/datakit/conf.d/ddtrace
cp ddtrace.conf.sample ddtrace.conf
```

2、开启statsd， 复制sample文件，不需要修改statsd.conf

```
cd /usr/local/datakit/conf.d/statsd
cp statsd.conf.sample statsd.conf
```

3、开通外网访问(非必选)

如果远程服务器需要访问datakit或者datakit提供给本服务器内的容器中的应用调用，需要把datakit.conf文件中的listen = "localhost:9529"改成listen = "0.0.0.0:9529"

```
vi /usr/local/datakit/conf.d/datakit.conf
```

![image](imgs/input-jvm-2.png)

4、重启Datakit

```
systemctl restart datakit
```

指标预览(启动应用后才能上报数据)

![image](imgs/input-jvm-3.png)


## 启动应用

#### JAVA_OPTS声明

```
java  ${JAVA_OPTS} -jar your-app.jar
```

JAVA_OPTS示例

```
-javaagent:/usr/local/datakit/data/dd-java-agent.jar \
 -Ddd.service=<your-service>   \
 -Ddd.env=dev  \
 -Ddd.agent.port=9529  
```

参数说明

```
-Ddd.env：应用的环境类型，选填 
-Ddd.tags：自定义标签，选填    
-Ddd.service：JVM数据来源的应用名称，控制台显示“应用名” 必填  
-Ddd.agent.host=localhost    DataKit地址，选填  
-Ddd.agent.port=9529         DataKit端口，必填  
-Ddd.version:版本，选填 
-Ddd.jmxfetch.check-period 表示采集频率，单位为毫秒，默认true，选填   
-Ddd.jmxfetch.statsd.host=127.0.0.1 statsd 采集器的连接地址同DataKit地址，选填  
-Ddd.jmxfetch.statsd.port=8125 表示DataKit上statsd采集器的UDP连接端口，默认为 8125，选填   
-Ddd.trace.health.metrics.statsd.host=127.0.0.1  自身指标数据采集发送地址同DataKit地址，选填 
-Ddd.trace.health.metrics.statsd.port=8125  自身指标数据采集发送端口，选填   
-Ddd.service.mapping:应用调用的redis、mysql等别名，选填
```

#### jar使用方式

使用java -jar的方式启动jar，默认连接本机上的datakit，如需要连接远程服务器上的datakit，请使用-Ddd.agent.host和-Ddd.jmxfetch.statsd.host指定ip

```
 java -javaagent:/usr/local/datakit/data/dd-java-agent.jar \
 -Ddd.service=<your-service>   \
 -Ddd.env=dev  \
 -Ddd.agent.port=9529  
 -jar <your-app.jar>
```

#### Docker使用方式

1、Dockerfile中的ENTRYPOINT启动参数使用环境变量JAVA_OPTS

```
FROM openjdk:8u292-jdk

ENV jar your-app.jar
ENV workdir /data/app/
RUN mkdir -p ${workdir}
COPY ${jar} ${workdir}
WORKDIR ${workdir}

ENTRYPOINT ["sh", "-ec", "exec java  ${JAVA_OPTS} -jar ${jar} "]
```

2、制作镜像
```
docker build -t <your-app-image:v1> .
```

3、上传dd-java-agent.jar放到/tmp/work目录

4、docker run启动容器

请修改172.16.0.215为您的Datakit的ip地址，替换9299为您应用的端口，替换your-app为您的应用名，替换your-app-image:v1为您的镜像名

```
docker run  -v /tmp/work:/tmp/work -e JAVA_OPTS="-javaagent:/tmp/work/dd-java-agent.jar -Ddd.service=your-app  -Ddd.env=dev  -Ddd.agent.host=172.16.0.215 -Ddd.agent.port=9529  -Ddd.jmxfetch.statsd.host=172.16.0.215  " --name your-app -d -p 9299:9299 your-app-image:v1
```

#### Kubernetes使用方式

1、Dockerfile中的ENTRYPOINT启动参数使用环境变量JAVA_OPTS

```
FROM openjdk:8u292

ENV jar your-app.jar
ENV workdir /data/app/
RUN mkdir -p ${workdir}
COPY ${jar} ${workdir}
WORKDIR ${workdir}
ENTRYPOINT ["sh", "-ec", "exec java ${JAVA_OPTS} -jar ${jar}"]
```

2、制作镜像

```
docker build -t 172.16.0.215:5000/dk/your-app-image:v1 . 
```

3、上传harbor仓库

```
 docker push 172.16.0.215:5000/dk/your-app-image:v1  
```

4、编写应用的deployment.yml
JAVA_OPTS示例说明：-Ddd.tags=container_host:$(PODE_NAME)是把环境变量PODE_NAME的值，传到标签container_host中。 /usr/dd-java-agent/agent/dd-java-agent.jar使用了共享存储的路径，使用了pubrepo.jiagouyun.com/datakit/dk-sidecar:1.0镜像提供dd-java-agent.jar。

```
-javaagent:/usr/dd-java-agent/agent/dd-java-agent.jar -Ddd.service=<your-app-name> 
-Ddd.tags=container_host:$(PODE_NAME)  
-Ddd.env=dev  
-Ddd.agent.host=172.16.0.215 
-Ddd.agent.port=9529  
-Ddd.jmxfetch.statsd.host=172.16.0.215
```

新建your-app-deployment-yaml文件，完整示例内容如下，使用时请替换9299为您应用的端口，替换your-app-name为您的服务名，替换30001为您的应用对外暴露的端口，替换172.16.0.215:5000/dk/your-app-image:v1为您的镜像名：

```bash
apiVersion: v1
kind: Service
metadata:
  name: your-app-name
  labels:
    app: your-app-name
spec:
  selector:
    app: your-app-name
  ports:
    - protocol: TCP
      port: 9299
      nodePort: 30001
      targetPort: 9299
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-app-name
  labels:
    app: your-app-name
spec:
  replicas: 1
  selector:
    matchLabels:
      app: your-app-name
  template:
    metadata:
      labels:
        app: your-app-name
    spec:
      containers:
      - env:
        - name: PODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: JAVA_OPTS
          value: |-
            -javaagent:/usr/dd-java-agent/agent/dd-java-agent.jar -Ddd.service=<your-app-name> -Ddd.tags=container_host:$(PODE_NAME)  -Ddd.env=dev  -Ddd.agent.port=9529   
        - name: DD_AGENT_HOST
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.hostIP
        name: your-app-name
        image: 172.16.0.215:5000/dk/your-app-image:v1    
        #command: ["sh","-c"]
        ports:
        - containerPort: 9299
          protocol: TCP
        volumeMounts:
        - mountPath: /usr/dd-java-agent/agent
          name: ddagent
      initContainers:
      - command:
        - sh
        - -c
        - set -ex;mkdir -p /ddtrace/agent;cp -r /usr/dd-java-agent/agent/* /ddtrace/agent;
        image: pubrepo.jiagouyun.com/datakit/dk-sidecar:1.0
        imagePullPolicy: Always
        name: ddtrace-agent-sidecar
        volumeMounts:
        - mountPath: /ddtrace/agent
          name: ddagent
      restartPolicy: Always
      volumes:
      - emptyDir: {}
        name: ddagent
      
```

5、部署应用

```
 kubectl apply -f your-app-deployment-yaml
```

## 场景视图

场景 - 新建场景 - JVM 监控场景

## 异常检测

暂无

## 指标详解

### `java_runtime`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |


- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `Uptime` | The total runtime. | int | ms |


### `java_memory`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |


- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `HeapMemoryUsagecommitted` | The total Java heap memory committed to be used. | int | B |
  | `HeapMemoryUsageinit` | The initial Java heap memory allocated. | int | B |
  | `HeapMemoryUsagemax` | The maximum Java heap memory available. | int | B |
  | `HeapMemoryUsageused` | The total Java heap memory used. | int | B |
  | `NonHeapMemoryUsagecommitted` | The total Java non-heap memory committed to be used. | int | B |
  | `NonHeapMemoryUsageinit` | The initial Java non-heap memory allocated. | int | B |
  | `NonHeapMemoryUsagemax` | The maximum Java non-heap memory available. | int | B |
  | `NonHeapMemoryUsageused` | The total Java non-heap memory used. | int | B |
  | `ObjectPendingFinalizationCount` | The count of object pending finalization. | int | count |


### `java_garbage_collector`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |
  | `name` | the name of GC generation |


- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `CollectionCount` | The number of GC that have occurred. | int | count |
  | `CollectionTime` | The approximate GC collection time elapsed. | int | B |


### `java_threading`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |


- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `DaemonThreadCount` | The count of daemon thread. | int | count |
  | `PeakThreadCount` | The peak count of thread. | int | count |
  | `ThreadCount` | The count of thread. | int | count |
  | `TotalStartedThreadCount` | The total count of started thread. | int | count |


### `java_class_loading`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |


- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `LoadedClassCount` | The count of loaded class. | int | count |
  | `TotalLoadedClassCount` | The total count of loaded class. | int | count |
  | `UnloadedClassCount` | The count of unloaded class. | int | count |


### `java_memory_pool`

- 标签
  | 标签名 | 描述 |
  | --- | --- |
  | `jolokia_agent_url` | jolokia agent url path |
  | `name` | the name of space |

- 指标列表
  | 指标 | 描述 | 数据类型 | 单位 |
  | --- | --- | --- | --- |
  | `PeakUsagecommitted` | The total peak Java memory pool committed to be used | int | B |
  | `PeakUsageinit` | The initial peak Java memory pool allocated | int | B |
  | `PeakUsagemax` | The maximum peak Java  memory pool available. | int | B |
  | `PeakUsageused` | The total peak Java memory pool used. | int | B |
  | `Usagecommitted` | The total Java memory pool committed to be used | int | B |
  | `Usageinit` | The initial Java memory pool allocated | int | B |
  | `Usagemax` | The maximum Java  memory pool available. | int | B |
  | `Usageused` | The total Java memory pool used. | int | B |


## 延伸阅读

- [DDTrace Java 示例](ddtrace-java)
- [SkyWalking](skywalking)
- [Opentelemetry Java 示例](opentelemetry-java)

## 最佳实践

<[JVM可观测最佳实践](../best-practices/integrations/jvm.md)>

## 故障排查

<[无数据上报排查](why-no-data.md)>
