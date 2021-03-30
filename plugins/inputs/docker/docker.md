
---
title: Docker 容器对象采集
catalog: SuperCloud IT 监控
thumbnail: attachments/docker.png
tags: 对象采集
---

## 简介

采集 docker 容器数据作为对象上报到 DataFlux 中。

## 前置条件

- 已安装 docker 1.24（[docker 官方链接](https://www.docker.com/get-started)）
- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）

## 配置

### Docker 服务容器监听端口配置

如果需要远程采集 docker 容器信息，需要 docker 开启相应的监听端口。

以 ubuntu 为例，需要在 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下：
```
{
    "hosts":["tcp://0.0.0.0:2375","unix:///var/run/docker.sock"]
}
```
重启服务后，docker 便可以监听 `2375` 端口。详情见[官方配置文档](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

### Container 容器信息采集器配置

进入 DataKit 安装目录下的 `conf.d/docker` 目录，复制 `docker_containers.conf.sample` 并命名为 `docker_containers.conf`。示例如下：

``` toml
[[inputs.docker_containers]]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"

    # 采集间隔时长，数字+单位，有效的时间单位 "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # 不可以小于等于 0，必须在 5m 到 1h 之间，否则使用默认 5m
    # 必填
    interval = "5m"

    # 是否采集所有容器，包括 Exited 状态
    all = false

    ## Optional TLS Config
    # tls_ca = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    ## 使用 containerID link kubernetes pods
    [inputs.docker_containers.kubernetes]
    #   ## URL for the kubelet
      url = "http://127.0.0.1:10255"
    #
    #   ## Use bearer token for authorization. ('bearer_token' takes priority)
    #   ## If both of these are empty, we'll use the default serviceaccount:
    #   ## at: /run/secrets/kubernetes.io/serviceaccount/token
    #   # bearer_token = "/path/to/bearer/token"
    #   ## OR
    #   # bearer_token_string = "abc_123"
    #
    #   ## Optional TLS Config
    #   # tls_ca = /path/to/cafile
    #   # tls_cert = /path/to/certfile
    #   # tls_key = /path/to/keyfile
    #   ## Use TLS but skip chain & host verification
    #   # insecure_skip_verify = false
```

#### kubernetes容器对象采集

如果配置了 `inputs.docker_containers.kubernetes`，将查询对应 kubernetes 服务的所有 pod 信息，找到与 `container_id` 对应的 pod，取其中几个字段。

## 数据字段

| 名称           | 描述                       | 类型   |
| :--            | ---                        | ---    |
| name           | tags                       | string |
| container_id   | fields                     | string |
| images_name    | fields                     | string |
| container_name | fields                     | string |
| restart_count  | fields                     | int |
| status         | fields                     | string |
| created_time   | fields（时间戳，单位秒）   | int    |
| start_time     | fields（时间戳，单位毫秒） | int    |
| message        | fields                     | string |

**此容器的资源占用情况，除`cpu_usage`和`mem_usage_percent`以外，其他字段的单位是字节**

| 名称              | 描述                | 类型  |
| :--               | ---                 | ---   |
| cpu_usage         | fields（cpu使用率） | float |
| mem_usage         | fields              | float |
| mem_limit         | fields              | float |
| mem_usage_percent | fields（mem使用率） | float |
| network_in        | fields              | float |
| network_out       | fields              | float |
| block_in          | fields              | float |
| block_out         | fields              | float |

**采集 kubernetes pod 新增字段**

| 名称           | 描述   | 类型   |
| :--            | ---    | ---    |
| pod_name       | fields | string |
| pod_namespace  | fields | string |


---
title: dockerlog 日志采集
catalog: SuperCloud IT 监控
thumbnail: attachments/docker.png
tags: IT运维,日志采集
---

## 简介

采集 docker 容器日志数据上报到 DataFlux 中。

## 前置条件

- 已安装 docker 且容器需带有 `localor`、`json-file` 或 `journald` 日志驱动程序
- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）

## 配置

### docker 配置

如果需要使用 tcp 远程连接 docker，需要 docker 开启相应的端口。

在 ubuntu 系统，需要在 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下：
```
{
    "hosts":["tcp://0.0.0.0:2375","unix:///var/run/docker.sock"]
}
```
重启服务后，docker 便可以监听 2375 端口。官方配置文档[链接](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

### datakit 配置

进入 DataKit 安装目录下的 `conf.d/docker` 目录，复制 `dockerlog.conf.sample` 并命名为 `dockerlog.conf`。示例如下：

``` toml
[[inputs.dockerlog]]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"

    # 数据来源。如果为空，则使用容器名称
    source = ""

    # pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 pipeline
    pipeline_path = ""

    # 是否需要从头开始读取日志，设置为 true 时从头读取
    # 设置为 false 时从尾部读取
    from_beginning = false

    # Docker API 调用超时
    timeout = "5s"

    # 包含和被排除的容器，使用 Globs 规则。注意此处是对容器名做操作，不是容器id
    # 当两个数组都为空时，即包含所有容器
    container_name_include = []
    container_name_exclude = []

    # 包含和被排除的容器状态，使用 Globs 规则
    # 当两个数组都为空时，即包含处于运行（running）状态的容器
    container_state_include = []
    container_state_exclude = []

    # 包含和被排除的容器标签，使用 Globs 规则
    # 当两个数组都为空时，即包含所有标签
    docker_label_include = []
    docker_label_exclude = []

    # 是否添加 source tags，用以区分数据来源
    # 当值为 true 时，会添加 source tags，值为该容器 ID 前 12 个字符，即该容器的默认主机名
    source_tag = false

    ## Optional TLS Config
    # tls_ca = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false

    # 自定义 tags
    # [inputs.dockerlog.tags]
    # tags1 = "tags1"
```

pipeline 配置和使用：
    [pipeline 文档](../../../20-datakit文本处理/index.md)

    - 如果配置文件中 `pipeline_path` 为空，默认使用 $source.p
    - 如果 $source.p 不存在，将不使用 pipeline 功能
    - 所有 pipeline 脚本文件，统一存放在 datakit 安装路径下的 pipeline 和 pattern 文件夹中，具体写法请看文档

Globs 规则简述：

| 通配符 | 描述                               | 例子         | 匹配                       | 不匹配                      |
| :--    | ---                                | ---          | ---                        | ----                        |
| *      | 匹配任意数量的任何字符，包括无     | Law*         | Law, Laws, Lawyer          | GrokLaw, La, aw             |
| ?      | 匹配任何单个字符                   | ?at          | Cat, cat, Bat, bat         | at                          |
| [abc]  | 匹配括号中给出的一个字符           | [CB]at       | Cat, Bat                   | cat, bat                    |
| [a-z]  | 匹配括号中给出的范围中的一个字符   | Letter[0-9]  | Letter0, Letter1 … Letter9 | Letters, Letter, Letter10   |
| [!abc] | 匹配括号中未给出的一个字符         | [!C]at       | Bat, bat, cat              | Cat                         |
| [!a-z] | 匹配不在括号内给定范围内的一个字符 | Letter[!3-5] | Letter1…                   | Letter3 … Letter5, Letterxx |

> Docker 日志插件使用[Docker Engine API](https://docs.docker.com/engine/api/v1.24/) 来获取正在运行的 Docker 容器的日志

## 采集指标

| 指标              | 类型   | 单位   |
| :--               | ---    | ---    |
| container_image   | tags   | string |
| container_name    | tags   | string |
| container_version | tags   | string |
| endpoint          | tags   | string |
| source            | tags   | string |
| stream            | tags   | string |
| container_id      | fields | string |
| __content         | fields | string |

## 示例输出

```
test_container,container_image=cf,container_name=test_container,container_version=unknown,endpoint=unix:///var/run/docker.sock,source=ccbd9fac6539,stream=stdout message="2020-09-08T07:41:02Z [testing] - test01",container_id="ccbd9fac6539811f44fc76e694852bd0a12c415b7764030af2e3c807916a86b3" 1599550862380752586
test_container,container_image=cf,container_name=test_container,container_version=unknown,endpoint=unix:///var/run/docker.sock,source=ccbd9fac6539,stream=stdout message="2020-09-08T07:41:03Z [testing] - test02",container_id="ccbd9fac6539811f44fc76e694852bd0a12c415b7764030af2e3c807916a86b3" 1599550863381871160
test_container,container_image=cf,container_name=test_container,container_version=unknown,endpoint=unix:///var/run/docker.sock,source=ccbd9fac6539,stream=stdout message="2020-09-08T07:41:04Z [testing] - test03",container_id="ccbd9fac6539811f44fc76e694852bd0a12c415b7764030af2e3c807916a86b3" 1599550864382903094
test_container,container_image=cf,container_name=test_container,container_version=unknown,endpoint=unix:///var/run/docker.sock,source=ccbd9fac6539,stream=stdout message="2020-09-08T07:41:05Z [testing] - test04",container_id="ccbd9fac6539811f44fc76e694852bd0a12c415b7764030af2e3c807916a86b3" 1599550865383325905
test_container,container_image=cf,container_name=test_container,container_version=unknown,endpoint=unix:///var/run/docker.sock,source=ccbd9fac6539,stream=stdout message="2020-09-08T07:41:06Z [testing] - test05",container_id="ccbd9fac6539811f44fc76e694852bd0a12c415b7764030af2e3c807916a86b3" 1599550866384325478
```
---
title: Docker 指标采集
catalog: SuperCloud IT 监控
thumbnail: attachments/docker.png
tags: IT运维,指标采集
---

## 简介

采集 docker 指标上报到 DataFlux 中

### 场景参考

#### Docker Overview视图

![内置视图](./attachments/docker01.png)

![内置视图](./attachments/docker02.png)

![内置视图](./attachments/docker03.png)

#### Docker container内置视图

![内置视图](./attachments/docker_inspect01.png)

![内置视图](./attachments/docker_inspect02.png)

[Docker视图模板下载](./attachments/Docker标准化场景2.28版.json)

## 前置条件

- 已安装 DataKit（[DataKit 安装文档](../../../02-datakit采集器/index.md)）


## 配置

docker采集器分为docker服务基本状态采集器与container容器信息采集器两个部分。

- docker服务基本状态采集器：docker.conf.sample
- container容器信息采集器：docker_containers.conf.sample


用户可以基于自身需求选择性开启。文档内提供的内置视图模板需要同时开启以上两个采集器，否则无法完整显示所有监控视图。

### docker服务基本状态采集器的配置

进入 DataKit 安装目录下的 conf.d/docker 目录，复制 docker.conf.sample 并命名为 docker.conf。示例如下：


设置：

```
# Read metrics about docker containers
[[inputs.docker]]
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Set to true to collect Swarm metrics(desired_replicas, running_replicas)
  ## Note: configure this in one of the manager nodes in a Swarm cluster.
  ## configuring in multiple Swarm managers results in duplication of metrics.
  gather_services = false

  ## Only collect metrics for these containers. Values will be appended to
  ## container_name_include.
  ## Deprecated (1.4.0), use container_name_include
  container_names = []

  ## Set the source tag for the metrics to the container ID hostname, eg first 12 chars
  source_tag = false

  ## Containers to include and exclude. Collect all if empty. Globs accepted.
  container_name_include = []
  container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  ## example: container_state_include = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
  ## example: container_state_exclude = ["created", "restarting", "running", "removing", "paused", "exited", "dead"]
  # container_state_include = []
  # container_state_exclude = []

  ## Timeout for docker list, info, and stats commands
  timeout = "5s"

  ## Whether to report for each container per-device blkio (8:0, 8:1...) and
  ## network (eth0, eth1, ...) stats or not
  perdevice = true

  ## Whether to report for each container total blkio and network stats or not
  total = false

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Which environment variables should we use as a tag
  tag_env = ["JAVA_HOME", "HEAP_SIZE"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
```

配置好后，重启 DataKit 即可生效

### container容器信息采集器的配置

进入 DataKit 安装目录下的 conf.d/docker 目录，复制 docker_containers.conf.sample 并命名为docker_containers.conf。示例如下：

设置：

```
[[inputs.docker_containers]]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # require, cannot be less than zero
    interval = "5s"

    # Is all containers
    all = true

    # Timeout for Docker API calls.
    timeout = "5s"

    ## Optional TLS Config
    # tls_ca = "/tmp/ca.pem"
    # tls_cert = "/tmp/cert.pem"
    # tls_key = "/tmp/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
```

配置好后，重启 DataKit 即可生效

## 采集指标

### docker 指标 

- 启用docker.conf采集

| 指标 | 描述 | 类型 |  单位 | Tag |
| :-- | ---- | ---- | ---- | ----- |
| n_used_file_descriptors |  | integer | - | unit，engine_host，server_version |
| n_cpus | 容器可运行的CPU内核数 | integer | - | unit，engine_host，server_version |
| n_containers | 容器数量 | integer | - | unit，engine_host，server_version |
| n_containers_running | 运行的容器数量 | integer | - | unit，engine_host，server_version |
| n_containers_stopped | 停止的容器数量 | integer | - | unit，engine_host，server_version |
| n_containers_paused | 暂停的容器数量 | integer | - | unit，engine_host，server_version |
| n_images | 镜像数量 | integer | - | unit，engine_host，server_version |
| n_listener_events | 事件监听数 | integer | - | unit，engine_host，server_version |
| n_goroutines | go并发线程数 | integer | - | unit，engine_host，server_version |
| memory_total | 内存总计 | integer | - | unit，engine_host，server_version |

### docker_swam 指标 

- 启用docker.conf采集

| 指标          | 描述 | 类型 | 单位 | Tag                                    |
| ------------- | ---- | ---- | ---- | -------------------------------------- |
| tasks_desired |      |      | -    | service_id，service_name，service_mode |
| tasks_running |      |      | -    | service_id，service_name，service_mode |

### docker_container_cpu 指标 

- 启用docker_containers.conf采集

| 指标                         | 描述 | 类型 | 单位 | Tag                                                          |
| ---------------------------- | ---- | ---- | ---- | ------------------------------------------------------------ |
| throttling_periods           |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| throttling_throttled_periods |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| throttling_throttled_time    |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| usage_in_kernelmode          |   内核模式资源使用   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| usage_in_usermode            |   用户模式资源使用   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| usage_system                 |   系统资源使用   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| usage_total                  |   使用资源总计   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| cpu_usage                |   CPU使用率   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |
| container_id                 |   容器ID   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，cpu |

### docker_container_mem 指标
- 启用docker_containers.conf采集

| 指标                      | 描述 | 类型 | 单位 | Tag                                                          |
| ------------------------- | ---- | ---- | ---- | ------------------------------------------------------------ |
| total_pgmajfault           |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| cache                     |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| mapped_file               |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_mapped_file         |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| pgpgout                   |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| rss                       |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_mapped_file         |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| writeback                 |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| unevictable               |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| pgpgin                    |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_unevictable         |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| pgmajfault                |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_rss                 |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_rss_huge            |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_writeback           |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_inactive_anon       |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| rss_huge                  |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| hierarchical_memory_limit |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_pgfault             |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_active_file         |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| active_anon               |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_active_anon         |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_pgpgout             |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_cache               |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| inactive_anon             |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| active_file               |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| pgfault                   |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| inactive_file             |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| total_pgpgin              |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| max_usage                 |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| usage                     |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| failcnt                   |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| limit                     |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| container_id              |      |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version |


### docker_container_net 指标
- 启用docker_containers.conf采集

| 指标         | 描述 | 类型 | 单位 | Tag                                                          |
| ------------ | ---- | ---- | ---- | ------------------------------------------------------------ |
| rx_dropped   |   丢弃的接收包   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| rx_bytes     |   接收的字节数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| rx_errors    |   接收的错误数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| tx_packets   |   发送的数据包   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| tx_dropped   |   丢弃的发送包   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| rx_packets   |   接收的数据包   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| tx_errors    |   发送的错误数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| tx_bytes     |   发送的字节数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |
| container_id |   容器ID   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，network |

### docker_container_blkio 指标
- 启用docker_containers.conf采集

| 指标                             | 描述 | 类型 | 单位 | Tag                                                          |
| -------------------------------- | ---- | ---- | ---- | ------------------------------------------------------------ |
| io_service_bytes_recursive_async |   容器卷的异步块I/O请求字节数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_service_bytes_recursive_read  |   容器卷的块读取字节数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_service_bytes_recursive_sync  |   容器卷的同步块I/O请求字节数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_service_bytes_recursive_total |   容器卷的块读写字节总数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_service_bytes_recursive_write |    容器卷的块写入字节数  |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_serviced_recursive_async       |   已服务异步块I/O请求数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_serviced_recursive_read        |   已服务块设备的读取请求数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_serviced_recursive_sync        |   已服务的同步块I/O请求数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_serviced_recursive_total       |   已服务的块读写请求总数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| io_serviced_recursive_write       |   已服务块设备的写入请求计数   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |
| container_id                     |   容器ID   |      | -    | engine_host，server_version，container_image，container_name，container_status，container_version，device |

### docker_container_health 指标 (容器必须开启 HEALTHCHECK)
- 启用docker_containers.conf采集

| 指标           | 描述 | 类型    | 单位 | Tag                                                          |
| -------------- | ---- | ------- | ---- | ------------------------------------------------------------ |
| health_status  |      | string  | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| failing_streak |      | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |

### docker_container_status 指标
- 启用docker_containers.conf采集

| 指标         | 描述 | 类型    | 单位 | Tag                                                          |
| ------------ | ---- | ------- | ---- | ------------------------------------------------------------ |
| container_id |  容器ID    |         | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| oomkilled    |   内存用尽kill   | boolean | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| pid          |   进程ID   | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| exitcode     |   退出代码   | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| started_at   |   容器开始时间   | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| finished_at  |   容器结束时间   | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |
| uptime_ns    |    容器运行时间  | integer | -    | engine_host，server_version，container_image，container_name，container_status，container_version |

### docker_devicemapper
- 启用docker_containers.conf采集

| 指标                               | 描述 | 类型 | 单位 | Tag                                    |
| ---------------------------------- | ---- | ---- | ---- | -------------------------------------- |
| pool_blocksize_bytes               |   存储池块大小   |      | -    | engine_host，server_version，pool_name |
| data_space_used_bytes              |   数据空间的使用字节数   |      | -    | engine_host，server_version，pool_name |
| data_space_total_bytes             |   数据空间的总字节数   |      | -    | engine_host，server_version，pool_name |
| data_space_available_bytes         |   数据空间的可用字节数   |      | -    | engine_host，server_version，pool_name |
| metadata_space_used_bytes          |   元数据空间的使用字节数   |      | -    | engine_host，server_version，pool_name |
| metadata_space_total_bytes         |   元数据空间的总字节数   |      | -    | engine_host，server_version，pool_name |
| metadata_space_available_bytes     |   元数据空间的可用字节数   |      | -    | engine_host，server_version，pool_name |
| thin_pool_minimum_free_space_bytes |   精简存储池的最小可用空间字节数   |      | -    | engine_host，server_version，pool_name |
