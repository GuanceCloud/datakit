## docker开发文档

docker采集器有5个数据源。其中，docker自身服务数据2个，容器数据3个。

在使用方面，简化了配置文件，删除一些精细控制（比如忽略指定label的容器），示例配置：

```toml
[inputs.docker]
    # Docker Endpoint
    # To use TCP, set endpoint = "tcp://[ip]:[port]"
    # To use environment variables (ie, docker-machine), set endpoint = "ENV"
    endpoint = "unix:///var/run/docker.sock"
    
    collect_metric = true
    collect_object = true
    collect_logging = true
    
    # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    collect_metric_interval = "10s"
    collect_object_interval = "5m"
    
    # Is all containers, Return all containers. By default, only running containers are shown.
    include_exited = false
    
    ## Optional TLS Config
    # tls_ca = "/path/to/ca.pem"
    # tls_cert = "/path/to/cert.pem"
    # tls_key = "/path/to/key.pem"
    ## Use TLS but skip chain & host verification
    # insecure_skip_verify = false
    
    #[[inputs.docker.log_option]]
        # container_name_match = "<regexp-container-name>"
        # source = "<your-source>"
        # service = "<your-service>"
        # pipeline = "<this-is-pipeline>"
    
    [inputs.docker.tags]
        # tags1 = "value1"
```

### docker服务本身的指标

采集 CPU、内存等使用情况，跟进程采集类似，初步调研没有现成的数据接口，可能需要读取 proc。

### docker服务本身的日志

采集docker服务本身的日志，而不是 `docker logs CONTAINER_ID` 容器日志。

日志采集结果跟 `journalctl` 命令相同，如下：

```
$> journalctl -u docker.service

Jun 12 16:33:14 ubuntu-server systemd[1]: Starting Docker Application Container Engine...
Jun 12 16:33:15 ubuntu-server dockerd[1126]: time="2020-06-12T08:33:15.299517668Z" level=info msg="Starting up"
Jun 12 16:33:15 ubuntu-server dockerd[1126]: time="2020-06-12T08:33:15.336971602Z" level=info msg="detected 127.0.0.53 nameserver, assuming systemd-resolved, so using resolv.conf: /run/systemd/resolve/resolv.conf"
```

大多数系统的 docker 服务日志都是 `journalctl` 输出（不包括mac）。目前尚未找到日志源文件路径，应该是写到 `systemd-journalctl` 服务中，可能需要抓取 `stdout`。

**需要使用 pipeline 对日志进行切割。**

### docker容器指标

调用 docker API 获取容器数据，将其转换为指标。需要采集以下几种数据：

- cpu
- mem
- kmen
- io
- net
- container
- images

如果该容器名符合 k8s 容器命名规则，则默认访问本机 k8s 服务，查找对应的容器信息，获取和补充以下数据：

- kube_container_name
- kube_daemon_set
- kube_deployment
- kube_namespace
- kube_ownerref_kind
- kube_ownerref_name
- kube_replica_set
- pod_name
- pod_phase

### docker容器对象

跟docker容器指标大致相同，只是数据发送到对象接口。

### docker container 日志

调用 docker API 获取容器日志，对齐进行 pipeline 切割后，发送到日志接口。

可以在配置文件中指定，容器名符合一定规则，该容器日志采用指定的 pipeline。示例配置文件如下：

```
[inputs.docker]
    # other

[[inputs.docker.log_pipeline]]
    # regexp
    container_name_match = "nginx-*"
    source = "nginxlog"
    service = "nginx"
    pipeline = "nginx.p"

```

如果容器名能够匹配 `container_name_match`，则对该容器日志进行 pipeline，并指定 `source` 和 `service`；
否则，不进行 pipeline，且 `source` 和 `service` 默认使用容器名。

### 指标数据

#### `docker`（docker服务指标，待补充）

| 名称 | 描述 | 类型 | 单位 |
| :--  | ---  | ---  | ---  |
| NULL |      |      |      |

#### `docker_containers`（docker容器指标）

-  标签

| 名称                | 描述                             |
| :--                 | ---                              |
| container_id        | 容器id                           |
| container_name      | 容器名称                         |
| image_name          | 容器镜像名称                     |
| docker_image        | 镜像名称+版本号                  |
| host                | 主机名                           |
| stats               | 运行状态，running/exited/removed |
| kube_container_name | TODO                             |
| kube_daemon_set     | TODO                             |
| kube_deployment     | TODO                             |
| kube_namespace      | TODO                             |
| kube_ownerref_kind  | TODO                             |
| pod_name            | pod名称                          |
| pod_phase           | pod生命周期                      |

- 指标列表

| 名称               | 描述 | 类型  | 单位    |
| :--                | ---  | :-:   | :-:     |
| from_kubernetes    | TODO | bool  | -       |
| cpu_usage_percent  | TODO | float | percent |
| mem_limit          | TODO | int   | bytes   |
| mem_usage          | TODO | int   | bytes   |
| mem_usage_percent  | TODO | float | percent |
| mem_failed_count   | TODO | int   | bytes   |
| network_bytes_rcvd | TODO | int   | bytes   |
| network_bytes_sent | TODO | int   | bytes   |
| block_read_byte    | TODO | int   | bytes   |
| block_write_byte   | TODO | int   | bytes   |

#### docker容器日志

指标集为配置文件 `inputs.docker.log_option` 字段 `source`，比如配置如下：

```toml
[[inputs.docker.log_option]]
    container_name_match = "nginx-version-*"
    source = "nginx"
    service = "nginx"
    pipeline = "nginx.p"
```

如果日志来源的容器，容器名能够匹配 `container_name_match` 正则，该容器日志的指标集为 `source` 字段。

如果没有匹配到或 `source` 为空，指标集为容器名

-  标签

| 名称           | 描述                          |
| :--            | ---                           |
| container_name | 容器名称                      |
| image_name     | 容器镜像名称                  |
| stream         | 数据流方式，stdout/stderr/tty |

- 指标列表

| 名称            | 描述 | 类型   | 单位 |
| :--             | ---  | :-:    | :-:  |
| from_kubernetes | TODO | bool   | -    |
| service         | TODO | string | -    |
| status          | TODO | string | -    |
| message         | TODO | string | -    |
