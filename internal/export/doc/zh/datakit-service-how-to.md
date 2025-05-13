
# DataKit 服务管理
---

[DataKit 安装](datakit-install.md)完后，有必要对安装好的 DataKit 做一些基本的介绍。

## DataKit 目录介绍 {#install-dir}

DataKit 目前支持 Linux/Windows/Mac 三种主流平台：

| 操作系统                            | 架构                | 安装路径                                                                       |
| ---------                           | ---:                | ------                                                                         |
| Linux 内核 2.6.23 或更高版本        | amd64/386/arm/arm64 | `/usr/local/datakit`                                                           |
| macOS 10.13 或更高版本[^1]          | amd64               | `/usr/local/datakit`                                                           |
| Windows 7, Server 2008R2 或更高版本 | amd64/386           | 64-bit：`C:\Program Files\datakit`<br />32-bit：`C:\Program Files(32)\datakit` |

[^1]: Golang 1.18 要求 macOS-amd64 版本为 10.13。

安装完成以后，DataKit 目录列表大概如下：


``` not-set
├── [   12]  apm_inject/
├── [    0]  gitrepos/
├── [    0]  python.d/
├── [  430]  pipeline/
├── [   26]  pipeline_remote/
├── [   42]  cache/
├── [   36]  externals/
├── [  316]  data/
├── [ 138M]  datakit
├── [  958]  conf.d/
└── [    7]  .pid
```

| 目录名            | 说明                                                                                 |
| ---               | ---                                                                                  |
| `apm_inject`      | 启用了 APM 自动注入功能后，这里用来存放一些依赖文件                                  |
| `cache`           | 存放采集过程中用到的一些数据缓存                                                     |
| `conf.d`          | 存放所有采集器的配置示例。DataKit 主配置文件 *datakit.conf* 位于目录下               |
| `data`            | 存放 DataKit 运行所需的数据文件，如 IP 地址数据库等                                  |
| `datakit`         | DataKit 主程序，Windows 下为 *datakit.exe*，DataKit 绝大部分采集功能都集成在该程序中 |
| `externals`       | 部分采集器没有集成在 DataKit 主程序中，它们是分离编译的                              |
| `gitrepos`        | 如果是用 Git 来管理采集器配置，这里存放这些采集配置                                  |
| `pipeline`        | 存放 Pipeline 脚本                                                                   |
| `pipeline_remote` | 存放 Studio 中编写的 Pipeline 脚本                                                   |
| `python.d`        | 存放 Python 脚本                                                                     |
| `.pid`            | 存放 DataKit 当前运行的进程号                                                        |

DataKit 日志文件有两个：

| 目录名            | 说明                                                                                 |
| ---               | ---                                                                                  |
| `gin.log`         | DataKit 可以接收外部的 HTTP 数据输入，这个日志文件相当于 HTTP 的 access log          |
| `log`             | DataKit 运行日志（Linux/Mac 平台下，DataKit 运行日志在 */var/log/datakit* 目录下，Windows 位于 *C:\Program Files\datakit\* 目录下）   |

<!-- markdownlint-disable MD046 -->
???+ tip "查看内核版本"

    - Linux/Mac：`uname -r`
    - Windows：执行 `cmd` 命令（按住 Win 键 + `r`，输入 `cmd` 回车），输入 `winver` 即可获取系统版本信息
<!-- markdownlint-enable -->

## DataKit 服务管理 {#manage-service}

可直接使用如下命令直接管理 DataKit：

```shell
# Linux/Mac 可能需加上 sudo
datakit service -T # stop
datakit service -S # start
datakit service -R # restart
```

<!-- markdownlint-disable MD046 -->
???+ tip

    可通过 `datakit help service` 查看更多帮助信息。
<!-- markdownlint-enable -->

### 服务管理失败处理 {#when-service-failed}

有时候可能因为 DataKit 部分组件的 bug，导致服务操作失败（如 `datakit service -T` 之后，服务并未停止），可按照如下方式来强制处理。

Linux 下，如果上述命令失效，可使用以下命令来替代：

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

Mac 下，可以用如下命令代替：

```shell
# 启动 DataKit
sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist

# 停止 DataKit
sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist
```

### 服务卸载以及重装 {#uninstall-reinstall}

可直接使用如下命令直接卸载或恢复 DataKit 服务：

> 注意：此处卸载 DataKit 并不会删除 DataKit 相关文件。

```shell
# Linux/Mac shell
datakit service -I # re-install
datakit service -U # uninstall
```

## DataKit 对宿主环境的影响 {#datakit-overhead}

在使用 DataKit 过程中，对已有的系统可能会有如下一些影响：

1. 日志采集会导致的磁盘高速读取，日志量越大，读取的 iops 越高
1. 如果在 Web/App 应用中加入了 RUM SDK，那么会有持续的 RUM 相关的数据上传，如果上传的带宽有相关限制，可能会导致 Web/App 的页面卡顿
1. [eBPF 采集](../integrations/ebpf.md) 开启后，由于采集的数据量比较大，会占用一定量的内存和 CPU。其中 bpf-netlog 开启后，会根据主机和容器网卡的所有 TCP 数据包，产生大量的日志
1. 在 DataKit 繁忙的时候（接入了大量的日志/Trace 以及外部数据导入等），其会占用相当量的 CPU 和内存资源，建议设置合理的[资源限制配置](datakit-conf.md#resource-limit)来加以控制
1. 当 DataKit [部署在 Kubernetes](datakit-daemonset-deploy.md) 中时，对 API server 会有一定的请求压力
1. 开启[默认采集器](datakit-input-conf.md#default-enabled-inputs)的情况下，内存（RSS）消耗大概在 100MB 左右，CPU 消耗控制在 10% 以内；磁盘消耗除了自身日志外，还有额外的[磁盘缓存](datakit-conf.md#dataway-wal)。网络流量则视具体采集的数据量而定，DataKit 上传的流量默认以 GZip 压缩上传。

## FAQ {#faq}

### Windows 下启动失败 {#windows-start-fail}

DataKit 在 Windows 下以服务的形式启动，启动后会写入很多 Event 日志，随着日志的积累，可能会出现如下报错：

``` not-set
Start service failed: The event log file is full.
```

该报错会导致 DataKit 无法启动。通过[设置一下 Windows Event](https://stackoverflow.com/a/13868216/342348){:target="_blank"} 即可。

## 更多参考 {#further-reading}

其它跟 DataKit 基本使用相关的文档：

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 更新</u>: 更新 DataKit 版本 </font>](datakit-update.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Monitor</u>: DataKit 运行状态查看</font>](datakit-monitor.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 工具命令</u>: DataKit 提供了很多便捷工具来辅助您的日常使用</font>](datakit-tools-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 端口占用</u>: DataKit 默认使用的端口列表</font>](datakit-port.md)
</div>
</font>
