# 在 Docker 中部署 Datakit
---

本文档介绍如何在 Docker 中安装 DataKit。

## 启动 {#start}

容器启动命令如下：

```shell
sudo docker run \
    --hostname "$(hostname)" \
    --workdir /usr/local/datakit \
    -v "/host/conf/dir":"/usr/local/datakit/conf.d/host-inputs-conf" \
    -v "/":"/rootfs" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e ENV_DATAWAY="https://openway.guance.com?token=<YOUR-WORKSPACE-TOKEN>" \
    -e ENV_DEFAULT_ENABLED_INPUTS='cpu,disk,diskio,mem,swap,system,net,host_processes,hostobject,container,dk' \
    -e ENV_GLOBAL_HOST_TAGS="tag1=a1,tag2=a2" \
    -e ENV_HTTP_LISTEN="0.0.0.0:9529" \
    -e HOST_PROC="/rootfs/proc" \
    -e HOST_SYS="/rootfs/sys" \
    -e HOST_ETC="/rootfs/etc" \
    -e HOST_VAR="/rootfs/var" \
    -e HOST_RUN="/rootfs/run" \
    -e HOST_DEV="/rootfs/dev" \
    -e HOST_ROOT="/rootfs" \
    --cpus 2 \
    --memory 1g \
    --privileged \
    --publish 19529:9529 \
    -d \
    pubrepo.guance.com/datakit/datakit:{{.Version}}
```

启动参数说明：

- **`--hostname`** 将宿主机的主机名作为 Datakit 运行的主机名，如果需要在当前宿主机上运行多个 Datakit，可以给它适当加一些后缀 `--hostname "$(hostname)-dk1"`
- **`--workdir`** 设置容器工作目录
- **`-v`**：各种宿主机文件挂载：
    - DataKit 中有很多配置文件，我们可以将其在宿主机上准备好，通过 `-v` 一次性整个挂载到容器中去（容器中的路径为 *conf.d/host-inputs-conf* 目录）
    - 此处将宿主机根目录挂载进 Datakit，目的是访问宿主机上的各种信息（比如 `/proc` 目录下的各种文件），便于默认开启的采集器采集数据
    - 将 *docker.sock* 文件挂载进 Datakit 容器，便于 container 采集器采集数据。不同宿主机该文件目录可能不同，需按照实际来配置
- **`-e`**：各种 Datakit 运行期的环境变量配置，这些环境变量功能跟 [DaemonSet 部署](datakit-daemonset-deploy.md#env-setting)时是一样的
- **`--publish`**：便于外部将 Trace 等数据发送给 Datakit 容器，此处我们将 Datakit 的 HTTP 端口映射到外面的 19529 上，诸如 trace 数据设置发送地址的时候，需关注这个端口设置。
- 此处对该运行的 Datakit 设置了 2C 的 CPU 和 1GiB 内存限制

假如我们在 */host/conf/dir* 目录下配置了如下一些采集器：

- **APM**：[DDTrace](../integrations/ddtrace.md)/[OpenTelemetry](../integrations/opentelemetry.md) 等采集器
- **Prometheuse exporter**：在当前 docker 环境中，某些应用容器暴露了自身指标（一般形如 `http://ip:9100/metrics`），那么我们可以将其端口暴露出来，然后编写 [*prom.conf*](../integrations/prom.md) 来采集这些指标
- **日志采集**：如果某些 Docker 容器将日志写入了宿主机的某个目录，我们可以单独编写[日志采集配置](../integrations/logging.md#config)来采集这些文件。不过事先我们需要通过 `-v` 将这些宿主机的目录挂载进 Datakit 容器。另外，默认开启的 `container` 采集器，会自动采集所有容器的 stdout 日志

容器启动后，可以在宿主几上直接执行如下命令查看 Datakit 的运行情况：

```shell
docker exec -it <容器名或容器 ID> datakit monitor
```

也可以直接进入容器，查看更多信息：

```shell
docker exec -it <容器名或容器 ID> /bin/bash
```

## container 采集器配置 {#input-container}

上面 container 采集器是默认启动的，如果要对容器采集器做一些额外配置，可以单独额外配置容器采集器。容器采集器支持通过 *.conf* 文件和 ENV 两种方式来调整其行为：

- 在 */host/conf/dir* 目录下额外[配置 container 采集器](../integrations/container.md#config)，同时，务必将 `container` 从 `ENV_DEFAULT_ENABLED_INPUTS` 列表中移除。
- 在 Docker 启动命令中，增加额外的环境变量配置，参见[这里](../integrations/container.md#__tabbed_1_2)
