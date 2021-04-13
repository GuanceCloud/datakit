{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：{{.AvailableArchs}}

# 简介

采集 Docker 服务数据和容器数据，分别以指标、对象和日志的方式上报到 DataFlux 中。

## 前置条件

- 已安装 docker 1.24（[docker 官方链接](https://www.docker.com/get-started)）

## 配置

### Docker服务监听端口配置

如果需要远程采集 Docker 容器信息，需要 Docker 开启相应的监听端口。

以 ubuntu 为例，需要在 `/etc/docker` 径路下打开或创建 `daemon.json` 文件，添加内容如下：

```
{
    "hosts":["tcp://0.0.0.0:2375","unix:///var/run/docker.sock"]
}
```

重启服务后，Docker 便可以监听 `2375` 端口。详情见[官方配置文档](https://docs.docker.com/config/daemon/#configure-the-docker-daemon)。

### 采集器配置

进入 DataKit 安装目录下的 `conf.d/docker` 目录，复制 `docker.conf.sample` 并命名为 `docker.conf`。示例如下：

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
    
    # [[inputs.docker.log_option]]
        # container_name_match = "<regexp-container-name>"
        # source = "<your-source>"
        # service = "<your-service>"
        # pipeline = "<this-is-pipeline>"
    
    [inputs.docker.tags]
        # tags1 = "value1"
```

- 通过对 `collect_metric` 等三个配置项的开关，选择是否要开启此类数据的采集
- 当 `include_exited` 为 `true` 会采集非运行状态的容器
- `log_option` 为数组配置，可以有多个。`container_name_match` 为正则表达式，如果容器名能匹配该正则，将使用 `source` 和 `service` 以及 `pipeline`
- 当 `pipeline` 为空值或该文件不存在时，将不使用 pipeline 功能
- 正则表达式[文档链接](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)

**使用 pipeline 功能时，取其中的 `time` 字段作为此条数据的产生时间。如果没有 `time` 字段或解析此字段失败，默认使用数据采集到的时间**

**数据必须含有 `status` 字段。如果使用 pipeline 功能时且得到有效的 `status`，否则默认使用“info”**

有效的 `status` 字段值（不区分大小写）：

| status 有效字段值                | 对应值     |
| :---                             | ---        |
| `f`, `emerg`                     | `emerg`    |
| `a`, `alert`                     | `alert`    |
| `c`, `critical`                  | `critical` |
| `e`, `error`                     | `error`    |
| `w`, `warning`                   | `warning`  |
| `n`, `notice`                    | `notice`   |
| `i`, `info`                      | `info`     |
| `d`, `debug`, `trace`, `verbose` | `debug`    |
| `o`, `s`, `OK`                   | `OK`       |

## 指标集

### `docker_containers`

-  标签

| 名称                  | 描述                             |
| :--                   | ---                              |
| `container_id`        | 容器id                           |
| `container_name`      | 容器名称                         |
| `image_name`          | 容器镜像名称                     |
| `docker_image`        | 镜像名称+版本号                  |
| `host`                | 主机名                           |
| `stats`               | 运行状态，running/exited/removed |
| `kube_container_name` | TODO                             |
| `kube_daemon_set`     | TODO                             |
| `kube_deployment`     | TODO                             |
| `kube_namespace`      | TODO                             |
| `kube_ownerref_kind`  | TODO                             |
| `pod_name`            | pod名称                          |
| `pod_phase`           | pod生命周期                      |

- 指标列表

| 名称                 | 描述 | 类型  | 单位    |
| :--                  | ---  | :-:   | :-:     |
| `from_kubernetes`    | TODO | bool  | -       |
| `cpu_usage_percent`  | TODO | float | percent |
| `mem_limit`          | TODO | int   | bytes   |
| `mem_usage`          | TODO | int   | bytes   |
| `mem_usage_percent`  | TODO | float | percent |
| `mem_failed_count`   | TODO | int   | bytes   |
| `network_bytes_rcvd` | TODO | int   | bytes   |
| `network_bytes_sent` | TODO | int   | bytes   |
| `block_read_byte`    | TODO | int   | bytes   |
| `block_write_byte`   | TODO | int   | bytes   |

#### Docker容器日志

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

| 名称             | 描述                          |
| :--              | ---                           |
| `container_name` | 容器名称                      |
| `image_name`     | 容器镜像名称                  |
| `stream`         | 数据流方式，stdout/stderr/tty |

- 指标列表

| 名称              | 描述 | 类型   | 单位 |
| :--               | ---  | :-:    | :-:  |
| `from_kubernetes` | TODO | bool   | -    |
| `service`         | TODO | string | -    |
| `status`          | TODO | string | -    |
| `message`         | TODO | string | -    |
