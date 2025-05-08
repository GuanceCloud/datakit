# Deploying Datakit in Docker
---

This document explains how to install Datakit in Docker.

## Startup {#start}

The container startup command is as follows:

> The content like `<XXX-YYY-ZZZ>` should be filled in according to the actual situation.

```shell
sudo docker run \
    --hostname "$(hostname)" \
    --workdir /usr/local/datakit \
    -v "<YOUR-HOST-DIR-FOR-CONF>":"/usr/local/datakit/conf.d/host-inputs-conf" \
    -v "/":"/rootfs" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e ENV_DATAWAY="https://openway.<<<custom_key.brand_main_domain>>>?token=<YOUR-WORKSPACE-TOKEN>" \
    -e ENV_DEFAULT_ENABLED_INPUTS='cpu,disk,diskio,mem,swap,system,net,host_processes,hostobject,container,dk' \
    -e ENV_GLOBAL_HOST_TAGS="<TAG1=A1,TAG2=A2>" \
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
    pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:{{.Version}}
```

Parameter explanations:

- **`--hostname`**: Sets the hostname of the host machine as the hostname for Datakit. If we need to run multiple Datakits on the same host, we can add appropriate suffixes, such as `--hostname "$(hostname)-dk1"`.
- **`--workdir`**: Sets the working directory of the container.
- **`-v`**: Various host file mounts:
    - Datakit has many configuration files, which can be prepared on the host machine and mounted into the container in one go using `-v` (the path in the container is the *conf.d/host-inputs-conf* directory).
    - Mounting the host's root directory into Datakit allows access to various host information (e.g., files in the `/proc` directory) to facilitate data collection by the default enabled collectors.
    - Mounting the *docker.sock* file into the Datakit container enables the container collector to collect data. The directory of this file may vary on different hosts and should be configured according to the actual situation.
- **`-e`**: Various environment variable configurations for Datakit during runtime, which function similarly to those in [DaemonSet deployment](datakit-daemonset-deploy.md#env-setting).
- **`--publish`**: Facilitates external sending of Trace and other data to the Datakit container. Here, we map Datakit's HTTP port to the external 19529, so when setting the address for sending trace data, pay attention to this port configuration.
- The running Datakit is set with a CPU limit of 2 cores and a memory limit of 1GiB.

Suppose we have configured the following collectors in the */host/conf/dir* directory:

- **APM**: Collectors such as [DDTrace](../integrations/ddtrace.md) and [OpenTelemetry](../integrations/opentelemetry.md).
- **Prometheus exporter**: In the current Docker environment, if some application containers expose their own metrics (typically in the form of `http://ip:9100/metrics`), we can expose their ports and then write a [*prom.conf*](../integrations/prom.md) to collect these metrics.
- **Log collection**: If some Docker containers write logs to a specific directory on the host machine, we can write a separate [log collection configuration](../integrations/logging.md#config) to collect these files. However, we need to mount these host directories into the Datakit container using `-v` beforehand. Additionally, the default enabled `container` collector will automatically collect the stdout logs of all containers.

After the container is started, we can directly execute the following command on the host to check the status of Datakit:

```shell
docker exec -it <container name or container ID> datakit monitor
```

Or we can also enter the container to view more information:

```shell
docker exec -it <container name or container ID> /bin/bash
```

## Container Collector Configuration {#input-container}

The `container` collector is started by default as we configured. If we need to make additional configurations for the `container` collector, we can:

- Add extra [configuration for the container collector](../integrations/container.md#config) in the */host/conf/dir* directory, and make sure to remove `container` from the `ENV_DEFAULT_ENABLED_INPUTS` list.
- Or add additional environment variable configurations in the Docker startup command, see [here](../integrations/container.md#__tabbed_1_2).

## Disk Cache {#wal}

Datakit defaults to enabling [WAL to cache data](datakit-conf.md#dataway-wal). If no additional host storage is specified, these unsent data will be discarded when the Datakit container deleted. We can mount an additional directory from the host machine to store this data:

```shell hl_lines="4"
sudo docker run \
    --hostname "$(hostname)" \
    --workdir /usr/local/datakit \
    -v "<YOUR-HOST-DIR-FOR-WAL-CACHE>":"/usr/local/datakit/cache/dw-wal" \
    ...
```
