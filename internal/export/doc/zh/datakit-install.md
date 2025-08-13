
# 主机安装
---

本文介绍 DataKit 的基本安装。

## 注册/登陆<<<custom_key.brand_name>>> {#login-guance}

浏览器访问 [<<<custom_key.brand_name>>>注册入口](https://auth.<<<custom_key.brand_main_domain>>>/redirectpage/register){:target="_blank"}，填写对应信息之后，即可[登陆<<<custom_key.brand_name>>>](https://console.<<<custom_key.brand_main_domain>>>/pageloading/login){:target="_blank"}

## 获取安装命令 {#get-install}

登陆工作空间，点击左侧「集成」选择顶部「DataKit」，即可看到各种平台的安装命令。

> 注意，以下 Linux/Mac/Windows 安装程序，能自动识别硬件平台（arm/x86, 32bit/64bit），无需做硬件平台选择。

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    安装命令支持 `bash` 和 `ash`([:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)) :

    - `bash`

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") }}
    ```

    - `ash`

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithShell "ash") }}
    ```

    安装完成后，在终端会看到安装成功的提示。

=== "Windows"

    Windows 上安装需在 Powershell 命令行安装，且必须以管理员身份运行 Powershell。按下 Windows 键，输入 powershell 即可看到弹出的 powershell 图标，右键选择「以管理员身份运行」即可。
    
    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") }}
    ```
<!-- markdownlint-enable -->

### 安装精简版的 DataKit {#lite-install}

可以通过在安装命令中添加 `DK_LITE` 环境变量来安装精简版的 DataKit ([:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)) :

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "DK_LITE" "1" ) }}
    ```

=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithEnvs "DK_LITE" "1" ) }}
    ```

<!-- markdownlint-enable -->

精简版 DataKit 只包含以下采集器：

| 采集器名称                                                        | 说明                                           |
| ----------------------------------------------------------------- | ---------------------------------------------- |
| [CPU（`cpu`）](../integrations/cpu.md)                            | 采集主机的 CPU 使用情况                        |
| [Disk（`disk`）](../integrations/disk.md)                         | 采集磁盘占用情况                               |
| [磁盘 IO（`diskio`）](../integrations/diskio.md)                  | 采集主机的磁盘 IO 情况                         |
| [内存（`mem`）](../integrations/mem.md)                           | 采集主机的内存使用情况                         |
| [Swap（`swap`）](../integrations/swap.md)                         | 采集 Swap 内存使用情况                         |
| [System（`system`）](../integrations/system.md)                   | 采集主机操作系统负载                           |
| [Net（`net`）](../integrations/net.md)                            | 采集主机网络流量情况                           |
| [主机进程（`host_processes`）](../integrations/host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表      |
| [主机对象（`hostobject`）](../integrations/hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等） |
| [DataKit（`dk`）](../integrations/dk.md)                          | 采集 DataKit 自身运行指标收集                  |
| [用户访问监测 (`rum`)](../integrations/rum.md)                    | 用于收集用户访问监测数据                       |
| [网络拨测 (`dialtesting`)](../integrations/dialtesting.md)        | 采集网络拨测数据                               |
| [Prom 采集 (`prom`)](../integrations/prom.md)                     | 采集 Prometheus Exporters 暴露出来的指标数据   |
| [日志采集 (`logging`)](../integrations/logging.md)                | 采集文件日志数据                               |

### 安装 DataKit 的 eBPF Trace Linker 版本 {#elinker-install}

可以通过在安装命令中添加 `DK_ELINKER` 环境变量来安装用于 eBPF Span 的连接和 eBPF Trace 生成的 DataKit ELinker 版本（[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)）:

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "DK_ELINKER" "1" ) }}
    ```

=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithEnvs "DK_ELINKER" "1" ) }}
    ```
<!-- markdownlint-enable -->
DataKit ELinker 只包含以下采集器：

| 采集器名称                                                       | 说明                                                        |
| ---------------------------------------------------------------- | ----------------------------------------------------------- |
| [CPU（`cpu`）](../integrations/cpu.md)                           | 采集主机的 CPU 使用情况                                     |
| [Disk（`disk`）](../integrations/disk.md)                        | 采集磁盘占用情况                                            |
| [磁盘 IO（`diskio`）](../integrations/diskio.md)                 | 采集主机的磁盘 IO 情况                                      |
| [eBPF Trace Linker（`ebpftrace`）](../integrations/ebpftrace.md) | 接收 eBPF 链路 span 并连接这些 spans 来生成 trace id 等信息 |
| [Swap（`swap`）](../integrations/swap.md)                        | 采集 Swap 内存使用情况                                      |
| [System（`system`）](../integrations/system.md)                  | 采集主机操作系统负载                                        |
| [Net（`net`）](../integrations/net.md)                           | 采集主机网络流量情况                                        |
| [主机对象（`hostobject`）](../integrations/hostobject.md)        | 采集主机基础信息（如操作系统信息、硬件信息等）              |
| [DataKit（`dk`）](../integrations/dk.md)                         | 采集 DataKit 自身运行指标收集                               |

### 安装指定版本的 DataKit {#version-install}

可通过在安装命令中指定版本号来安装指定版本的 DataKit，如安装 1.2.3 版本的 DataKit：

```shell
{{ InstallCmd 0 (.WithPlatform "unix") (.WithVersion "-1.2.3") }}
```

Windows 下同理：

```powershell
{{ InstallCmd 0 (.WithPlatform "windows") (.WithVersion "-1.2.3") }}
```

## 额外支持的环境变量 {#extra-envs}

如果需要在安装阶段定义一些 DataKit 配置，可在安装命令中增加环境变量，在 `DK_DATAWAY` 前面追加即可。如追加 `DK_NAMESPACE` 设置：

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "DK_NAMESPACE" "[NAMESPACE]" ) }}
    ```

=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithEnvs "DK_NAMESPACE" "[NAMESPACE]" ) }}
    ```
<!-- markdownlint-enable -->

俩种环境变量的设置格式为：

```shell
# Windows: 多个环境变量之间以分号分割
$env:NAME1="value1"; $env:Name2="value2"

# Linux/Mac: 多个环境变量之间以空格分割
NAME1="value1" NAME2="value2"
```

安装脚本支持的环境变量如下（全平台支持）。

<!-- markdownlint-disable MD046 -->
???+ note

    [全离线安装](datakit-offline-install.md#offline)不支持这些环境变量设置。但可以通过[代理](datakit-offline-install.md#with-datakit)以及[设置本地安装地址](datakit-offline-install.md#with-nginx)方式来设置这些环境变量。
<!-- markdownlint-enable -->

### 最常用环境变量 {#common-envs}

- `DK_DATAWAY`：指定 DataWay 地址，目前 DataKit 安装命令已经默认带上
- `DK_GLOBAL_TAGS`：已弃用，改用 DK_GLOBAL_HOST_TAGS
- `DK_GLOBAL_HOST_TAGS`：支持安装阶段填写全局主机 tag，格式范例：`host=__datakit_hostname,host_ip=__datakit_ip`（多个 tag 之间以英文逗号分隔）
- `DK_GLOBAL_ELECTION_TAGS`：支持安装阶段填写全局选举 tag，格式范例：`project=my-porject,cluster=my-cluster`（多个 tag 之间以英文逗号分隔）
- `DK_CLOUD_PROVIDER`：支持安装阶段填写云厂商(目前支持如下几类云主机 `aliyun/aws/tencent/hwcloud/azure`)。**该功能已弃用**，DataKit 已经可以自动识别云主机类型。
- `DK_USER_NAME`：DataKit 服务运行时的用户名。默认为 `root`。更详情的说明见下面的 “注意事项”。
- `DK_DEF_INPUTS`：[默认开启的采集器](datakit-input-conf.md#default-enabled-inputs)配置。如果要禁用某些采集器，需手动将其屏蔽，比如，要禁用 `cpu` 和 `mem` 采集器，需这样指定：`-cpu,-mem`，即除了这两个采集器之外，其它默认采集器均开启。
- `DK_LITE`：安装精简版 DataKit 时，可设置该变量为 `1`。([:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0))

<!-- markdownlint-disable MD046 -->
???+ tip "禁用所有默认采集器 [:octicons-tag-24: Version-1.5.5](changelog.md#cl-1.5.5)"

    如果要禁用所有默认开启的采集器，可以将 `DK_DEF_INPUTS` 设置为 `-`，如

    ```shell
    DK_DEF_INPUTS="-" \
    DK_DATAWAY=https://openway.<<<custom_key.brand_main_domain>>>?token=<TOKEN> \
    bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

    另外，如果之前有安装过 DataKit，必须将之前的默认采集器配置都删除掉，因为 DataKit 在安装的过程中只能添加采集器配置，但不能删除采集器配置。

???+ note "注意事项"

    由于权限问题，如果通过 `DK_USER_NAME` 修改 DataKit 服务运行时的用户名为非 `root`，那么以下采集器将不可使用：

    - [eBPF](../integrations/ebpf.md){:target="_blank"}

    另外，需要注意以下几项：

    - 必须先手动创建好用户和用户组，用户名和用户组名称必须一致，再进行安装。不同 Linux 发行版创建的命令可能会有差异，以下命令仅供参考：

        === "CentOS/RedHat"

            ```sh
            # 创建系统用户组 datakit
            groupadd --system datakit

            # 创建系统用户 datakit，并将用户 datakit 添加进组 datakit 中（这里用户名和组名都是 datakit）
            adduser --system --no-create-home datakit -g datakit

            # 禁止用户名 datakit 用于登录（用于 CentOS/RedHat 系 Linux）
            usermod -s /sbin/nologin datakit
            ```

        === "Ubuntu/Debian"

            ```sh
            # 在 Ubuntu 上，同时创建用户并添加进用户组的命令可能会报错，这个时候需要分成两步

            # 创建系统用户组 datakit
            groupadd --system datakit

            # 创建系统用户 datakit
            adduser --system --no-create-home datakit
            
            # 将用户 datakit 添加进组 datakit
            usermod -a -G datakit datakit

            # 禁止用户名 datakit 用于登录（用于 Ubuntu/Debian 系 Linux）
            usermod -s /usr/sbin/nologin datakit
            ```

        === "其它 Linux"

            ```sh
            # 在其它 Linux 上，同时创建用户并添加进用户组的命令可能会报错，这个时候需要分成两步

            # 创建系统用户组 datakit
            groupadd --system datakit
            
            # 创建系统用户 datakit
            adduser --system --no-create-home datakit
            
            # 将用户 datakit 添加进组 datakit
            usermod -a -G datakit datakit
            
            # 禁止用户名 datakit 用于登录（用于其它 Linux）
            usermod -s /bin/false datakit
            ```

        ```sh
        # 安装 DataKit
        DK_USER_NAME="datakit" DK_DATAWAY="..." bash -c ...
        ```

<!-- markdownlint-enable -->

### DataKit 自身日志相关 {#env-logging}

- `DK_LOG_LEVEL`: 可选值 info/debug
- `DK_LOG`: 如果改成 stdout, 日志将不写文件，而是终端输出
- `DK_GIN_LOG`: 如果改成 stdout, 日志将不写文件，而是终端输出

### DataKit pprof 相关 {#env-pprof}

- `DK_ENABLE_PPROF`: 是否开启 `pprof`。[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) 已默认开启。
- `DK_PPROF_LISTEN`: `pprof` 服务监听地址

### DataKit 选举相关 {#env-election}

- `DK_ENABLE_ELECTION`: 开启选举，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可。（如 `True`/`False`）
- `DK_NAMESPACE`：支持安装阶段指定命名空间(选举用)

### HTTP/API 相关环境变量 {#env-http-api}

- `DK_HTTP_LISTEN`：支持安装阶段指定 DataKit HTTP 服务绑定的网卡（默认 `localhost`）
- `DK_HTTP_PORT`：支持安装阶段指定 DataKit HTTP 服务绑定的端口（默认 `9529`）
- `DK_RUM_ORIGIN_IP_HEADER`: RUM 专用
- `DK_DISABLE_404PAGE`: 禁用 DataKit 404 页面 (公网部署 DataKit RUM 时常用。如 `True`/`False`)
- `DK_INSTALL_IPDB`: 安装时指定 IP 库(当前仅支持 `iploc/geolite2`)
- `DK_UPGRADE_IP_WHITELIST`: 从 DataKit [1.5.9](changelog.md#cl-1.5.9) 开始，支持远程访问 API 的方式来升级 DataKit，此环境变量用于设置可以远程访问的客户端 IP 白名单（多个 IP 用逗号 `,` 分隔），不在白名单内的访问将被拒绝（默认是不做 IP 限制）。
- `DK_UPGRADE_LISTEN`: 指定升级服务绑定的 HTTP 地址（默认 `0.0.0.0:9542`）[:octicons-tag-24: Version-1.38.1](changelog.md#cl-1.38.1)
- `DK_HTTP_PUBLIC_APIS`: 设置 DataKit 允许远程访问的 HTTP API ，RUM 功能通常需要进行此配置，从 DataKit [1.9.2](changelog.md#cl-1.9.2) 开始支持。
- `DK_HTTP_SOCKET`: 设置 HTTP 监听的本地 Socket 路径（Windows 不支持）。[:octicons-tag-24: Version-1.80.0](changelog-2025.md#cl-1.80.0)

### DCA 相关 {#env-dca}

- `DK_DCA_ENABLE`：支持安装阶段开启 DCA 服务（默认未开启）
- `DK_DCA_WEBSOCKET_SERVER`：支持安装阶段自定义配置 DCA 的 websocket 地址

### 外部采集器相关 {#env-external-inputs}

- `DK_INSTALL_EXTERNALS`: 可用于安装未与 DataKit 一起打包的外部采集器

### Confd 配置相关 {#env-connfd}

| 环境变量名              | 类型   | 适用场景                        | 说明       | 样例值                                         |
| ----------------------- | ------ | ------------------------------- | ---------- | ---------------------------------------------- |
| DK_CONFD_BACKEND        | string | 全部                            | 后端源类型 | `etcdv3` 或 `zookeeper` 或 `redis` 或 `consul` |
| DK_CONFD_BASIC_AUTH     | string | `etcdv3` 或 `consul`            | 可选       |                                                |
| DK_CONFD_CLIENT_CA_KEYS | string | `etcdv3` 或 `consul`            | 可选       |                                                |
| DK_CONFD_CLIENT_CERT    | string | `etcdv3` 或 `consul`            | 可选       |                                                |
| DK_CONFD_CLIENT_KEY     | string | `etcdv3` 或 `consul` 或 `redis` | 可选       |                                                |
| DK_CONFD_BACKEND_NODES  | string | 全部                            | 后端源地址 | `[IP:2379, IP2:2379]`                          |
| DK_CONFD_PASSWORD       | string | `etcdv3` 或 `consul`            | 可选       |                                                |
| DK_CONFD_SCHEME         | string | `etcdv3` 或 `consul`            | 可选       |                                                |
| DK_CONFD_SEPARATOR      | string | `redis`                         | 可选默认 0 |                                                |
| DK_CONFD_USERNAME       | string | `etcdv3` 或 `consul`            | 可选       |                                                |

### Git 配置相关 {#env-gitrepo}

- `DK_GIT_URL`: 管理配置文件的远程 git repo 地址。（如 `http://username:password@github.com/username/repository.git`）
- `DK_GIT_KEY_PATH`: 本地 PrivateKey 的全路径。（如 `/Users/username/.ssh/id_rsa`）
- `DK_GIT_KEY_PW`: 本地 PrivateKey 的使用密码。（如 `passwd`）
- `DK_GIT_BRANCH`: 指定拉取的分支。**为空则是默认**，默认是远程指定的主分支，一般是 `master`。
- `DK_GIT_INTERVAL`: 定时拉取的间隔。（如 `1m`）

### WAL 磁盘缓存 {#env-wal}

- `DK_WAL_WORKERS`: 设置 WAL 消费 worker 数，默认 CPU limit 核心数 * 4
- `DK_WAL_CAPACITY`: 这是单个 WAL 最大占用磁盘大小，默认 2GB

### Sinker 相关配置 {#env-sink}

通过 `DK_SINKER_GLOBAL_CUSTOMER_KEYS` 用于设置 sinker 过滤的 tag/field key 名称，其形式如下：

<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithEnvs "DK_SINKER_GLOBAL_CUSTOMER_KEYS" "key1,key2" ) (.WithEnvs "DK_DATAWAY_ENABLE_SINKER" "on" ) }}
    ```

=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithEnvs "DK_SINKER_GLOBAL_CUSTOMER_KEYS" "key1,key2" ) (.WithEnvs "DK_DATAWAY_ENABLE_SINKER" "on" ) }}
    ```
<!-- markdownlint-enable -->

### 资源限制配置相关 {#env-cgroup}

目前仅支持 Linux 和 Windows ([:octicons-tag-24: Version-1.15.0](changelog.md#cl-1.15.0)) 操作系统。

- `DK_LIMIT_DISABLED`：关闭资源限制功能（默认开启）
- `DK_LIMIT_CPUMAX`：限制 CPU 的最大百分比使用率，默认 30.0，最大值 100（已弃用，建议使用 `DK_LIMIT_CPUCORES`）
- `DK_LIMIT_CPUCORES`：限制使用的 CPU 核数，默认 2.0（即 2 核心）
- `DK_LIMIT_MEMMAX`：限制内存（含 swap）最大用量，默认 4096（4GB）

### APM Instrumentation {#apm-instrumentation}

[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) · [:octicons-beaker-24: Experimental](index.md#experimental)

在安装命令中，指定 `DK_APM_INSTRUMENTATION_ENABLED` 可针对 Java/Python 等应用自动注入 APM：

- 开启主机注入：

```shell
DK_APM_INSTRUMENTATION_ENABLED=host \
  DK_DATAWAY=https://openway.<<<custom_key.brand_main_domain>>>?token=<TOKEN> \
  bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
```

- 开启 docker 注入：

```shell
DK_APM_INSTRUMENTATION_ENABLED=docker \
  DK_DATAWAY=https://openway.<<<custom_key.brand_main_domain>>>?token=<TOKEN> \
  bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
```

对于主机部署，在 DataKit 安装完成后，重新开启一个终端，并重启对应的 Java/Python 应用即可。

开启和关闭该功能，修改 `datakit.conf` 文件中 `[apm_inject]` 下的 `instrumentation_enabled` 配置的值：

- 值 `"host"`、`"docker"` 或 `"host,docker"`，开启
- 值 `""` 或者 `"disable"`，关闭

针对特定的主机上的进程或者容器内的进程，可以通过注入环境变量 `ENV_DATAKIT_DISABLE_APM_INS`，并把值设置为 `true` 来关闭自动注入功能。

注意事项：

1. 删除 DataKit 安装目录下的文件前，需要先卸载该功能，请执行 **`datakit tool --remove-apm-auto-inject`** 清理系统设置和 Docker 的设置。

2. 对于 Docker 注入，安装并配置 Docker 注入和删除 DataKit 安装目录下的注入相关的文件，需要执行额外的步骤

   - 安装且配置开启 Docker 注入后，若需要对已经创建的容器生效：

   ```shell
   # 停止 docker 服务
   systemctl stop docker docker.socket

   # 将已经创建的容器的 runtime 从 runc 换成 datakit 提供的 dk-runc
   datakit tool --change-docker-containers-runtime dk-runc

   # 启动 docker 服务
   systemctl start docker

   # 重新启动因 dockerd 重启，导致的容器的退出
   docker start <container_id1> <container_id2> ...
   ```

   - 在卸载该功能后（开启过 Docker 注入），若需要删除 DataKit 安装目录下的所有文件：

   ```shell
   # 停止 docker 服务
   systemctl stop docker docker.socket

   # 将已经创建的容器的 runtime 从 dk-runc 换回 runc
   datakit tool --change-docker-containers-runtime runc

   # 启动 docker 服务
   systemctl start docker

   # 重新启动因 dockerd 重启，导致的容器的退出
   docker start <container_id1> <container_id2> ...
   ```

运行环境要求：

- Linux 系统
    - CPU 架构：x86_64 或 arm64
    - C 标准库：glibc 2.4 及以上版本，或 musl
    - Java 8 及以上版本
    - Python 3.7 及以上版本

在 Kubernetes 中，可以通过 [DataKit Operator 来注入 APM](datakit-operator.md#datakit-operator-inject-lib)。

### 其它安装选项 {#env-others}

| 环境变量名                       | 取值示例                    | 说明                                                                                                                             |
| -------------------------------- | --------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `DK_INSTALL_ONLY`                | `on`                        | 仅安装，不运行                                                                                                                   |
| `DK_HOSTNAME`                    | `some-host-name`            | 支持安装阶段自定义配置主机名                                                                                                     |
| `DK_UPGRADE`                     | `1`                         | 升级到最新版本                                                   |
| `DK_UPGRADE_MANAGER`             | `on`                        | 升级 DataKit 同时是否升级 **远程升级服务**，需要和 `DK_UPGRADE` 配合使用， 从 [1.5.9](changelog.md#cl-1.5.9) 版本开始支持        |
| `DK_INSTALLER_BASE_URL`          | `https://your-url`          | 可选择不同环境的安装脚本，默认为 `https://static.<<<custom_key.brand_main_domain>>>/datakit`                                                             |
| `DK_PROXY_TYPE`                  | -                           | 代理类型。选项有：`datakit` 或 `nginx`，均为小写                                                                                 |
| `DK_NGINX_IP`                    | -                           | 代理服务器 IP 地址（只需要填 IP 不需要填端口）。这个与上面的 "HTTP_PROXY" 和 "HTTPS_PROXY" 互斥，而且优先级最高，会覆盖以上两者  |
| `DK_INSTALL_LOG`                 | -                           | 设置安装程序日志路径，默认为当前目录下的 *install.log*，如果设置为 `stdout` 则输出到命令行终端                                   |
| `HTTPS_PROXY`                    | `IP:Port`                   | 通过 DataKit 代理安装                                                                                                            |
| `DK_INSTALL_RUM_SYMBOL_TOOLS`    | `on`                        | 是否安装 RUM source map 工具集，从 DataKit [1.9.2](changelog.md#cl-1.9.2) 开始支持                                               |
| `DK_VERBOSE`                     | `on`                        | 打开安装过程中的 verbose 选项（仅 Linux/Mac 支持），将输出更多调试信息[:octicons-tag-24: Version-1.19.0](changelog.md#cl-1.19.0) |
| `DK_CRYPTO_AES_KEY`              | `0123456789abcdfg`          | 使用加密后的密码解密秘钥，用于采集器中明文密码的保护 [:octicons-tag-24: Version-1.31.0](changelog.md#cl-1.31.0)                  |
| `DK_CRYPTO_AES_KEY_FILE`         | `/usr/local/datakit/enc4dk` | 秘钥的另一种配置方式，优先于上一种。将秘钥放到该文件中，并将配置文件路径通过环境变量方式配置即可。                               |

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### 如何应付不友好的主机名 {#bad-hostname}
<!-- markdownlint-enable -->

由于 DataKit 使用主机名（Hostname）作为数据串联的依据，某些情况下，一些主机名取得不是很友好，比如 `iZbp141ahn....`，但由于某些原因，又不能修改这些主机名，这给使用带来一定的困扰。在 DataKit 中，可在主配置中覆盖这个不友好的主机名。

在 `datakit.conf` 中，修改如下配置，DataKit 将读取 `ENV_HOSTNAME` 来覆盖当前的真实主机名：

```toml
[environments]
    ENV_HOSTNAME = "your-fake-hostname-for-datakit"
```

<!-- markdownlint-disable MD046 -->
???+ note

    如果之前某个主机已经采集了一段时间的数据，更改主机名后，这些历史数据将不再跟新的主机名关联。更改主机名，相当于新增了一台全新的主机。
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### Mac 安装问题 {#mac-failed}
<!-- markdownlint-enable -->

Mac 上安装时，如果安装/升级过程中出现

```shell
"launchctl" failed with stderr: /Library/LaunchDaemons/com.datakit.plist: Service is disabled
```

执行

```shell
sudo launchctl enable system/datakit
```

然后再执行如下命令即可

```shell
sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist
```

<!-- markdownlint-disable MD013 -->
### DataKit 是否有文件以及数据的高危操作？ {#danger-ops}
<!-- markdownlint-enable -->

DataKit 在运行过程中，根据采集配置不同，会读取很多系统信息，比如进程列表、软硬件信息（比如操作系统信息、CPU、内存、磁盘、网卡等）。但它不会主动执行删除、修改其自身之外的其它数据。关于文件读写，分成两个部分，一个是和数据采集有关的读文件/端口操作，一个是 DataKit 自身运行过程中一些必要的文件读写操作。

采集需要读取的主机文件：

- 在进程信息采集、软硬件信息采集的过程中，Linux 下会读取 */proc* 目录下的相关信息；Windows 下主要通过 WMI 以及 Golang Windows SDK 来获取这些信息

- 如果配置了相关的日志采集，根据采集的配置，会扫描并且读取符合配置的日志（比如 syslog，用户应用日志等）

- 端口占用：DataKit 为了对接一些其它系统，会单独开启一些端口服务来接收外部数据。[这些端口](datakit-port.md)根据采集器不同，按需开启

- eBPF 采集：eBPF 由于其特殊性，需要更多 Linux 内核以及进程的二进制信息，会有如下一些动作：

    - 分析所有（或指定的）在运行的程序（动态库、容器内进程）的二进制文件内包含的符号地址
    - 读写内核 DebugFS 挂在点下的文件或 PMU（Performance Monitoring Unit）以放置 kprobe/uprobe/tracepoint eBPF 探针
    - uprobe 探针会修改用户进程的 CPU 指令，以读取相关数据

除了采集之外，DataKit 自身会有如下文件读写操作：

- 自身日志文件

Linux 安装时位于 */var/log/datakit/* 目录下；Windows 位于 *C:\Program Files\datakit* 目录下。

日记文件到达指定大小（默认 32MB）后会自动 Rotate，并且有最大 Rotate 个数上限（默认最大 5 + 1 个分片）。

- 磁盘缓存

部分数据采集需要用到磁盘缓存功能（需手动开启），这部分缓存会在生成和消费过程中有文件增删。磁盘缓存也有最大 capacity 设置，数据满了之后，会自动执行 FIFO 删除操作，避免写满磁盘。

<!-- markdownlint-disable MD013 -->
### DataKit 如何控制自身资源消耗？ {#resource-limit}
<!-- markdownlint-enable -->

可以通过 cgroup 等机制来限制 DataKit 自身资源使用，参见[这里](datakit-conf.md#resource-limit)。如果 DataKit 部署在 Kubernetes 中，参见[这里](datakit-daemonset-deploy.md#requests-limits)。

<!-- markdownlint-disable MD013 -->
### DataKit 自身可观测性？ {#self-obs}
<!-- markdownlint-enable -->

DataKit 在运行过程中，暴露了很多[自身的指标](datakit-metrics.md)。默认情况下，DataKit 通过[内置采集器](../integrations/dk.md)会采集这些指标并上报到用户的工作空间。

除此之外，DataKit 自身还带有一个 [monitor 命令行](datakit-monitor.md)工具，通过该工具，能查看当前的运行状态以及采集、上报情况。

## 扩展阅读 {#more-reading}

- [DataKit 使用入门](datakit-service-how-to.md)
