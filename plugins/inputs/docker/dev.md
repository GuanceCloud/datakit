## docker采集器开发文档

docker采集器有5个数据源。其中，docker自身服务数据2个，容器数据3个。

在使用方面，简化了配置文件，删除一些精细控制（比如忽略指定label的容器），示例配置：

```
[inputs.docker]
    # 指定连接方式，基础配置
    endpoint = "unix:///var/run/docker.sock"

    # 是否采集所有容器，默认为false时只采集正在运行的容器（running）
    all = false

    # 采集间隔
    # 指标采集和对象采集周期不同
    # 如果interval配置小于0，则不开启对应的采集
    collect_metrics_interval = "10s"
    collect_object_interval = "5m"

    ## tls配置相关
    # tls_ca = "/tmp/ca.pem"
    # tls_cert = "/tmp/cert.pem"
    # tls_key = "/tmp/key.pem"
    # insecure_skip_verify = false

    # 日志相关，见后文
    [[inputs.docker.log_pipeline]]
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

调用 docker API 获取容器数据，将其转换为指标。

需要在配置文件中指定，是否采集 k8s 容器指标，如果该容器名符合 k8s 容器命名规则，则根据配置选择是否采集。

比如配置项名为 `ignore_k8s_container = true`。

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

因数据源较多，指标列表待定。
