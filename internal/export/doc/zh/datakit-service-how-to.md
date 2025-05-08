
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
├── [4.4K]  conf.d
├── [ 160]  data
├── [ 64M]  datakit
├── [ 192]  externals
├── [1.2K]  pipeline
├── [ 192]  gin.log   # Windows 平台
└── [1.2K]  log       # Windows 平台
```

其中：

- `conf.d`：存放所有采集器的配置示例。Datakit 主配置文件 *datakit.conf* 位于目录下
- `data`：存放 Datakit 运行所需的数据文件，如 IP 地址数据库等
- `datakit`：Datakit 主程序，Windows 下为 *datakit.exe*
- `externals`：部分采集器没有集成在 Datakit 主程序中，就都在这里了
- `pipeline` 存放用于文本处理的脚本代码
- `gin.log`：Datakit 可以接收外部的 HTTP 数据输入，这个日志文件相当于 HTTP 的 access-log
- `log`：Datakit 运行日志（Linux/Mac 平台下，Datakit 运行日志在 */var/log/datakit* 目录下）

<!-- markdownlint-disable MD046 -->
???+ tip "查看内核版本"

    - Linux/Mac：`uname -r`
    - Windows：执行 `cmd` 命令（按住 Win 键 + `r`，输入 `cmd` 回车），输入 `winver` 即可获取系统版本信息
<!-- markdownlint-enable -->

## Datakit 服务管理 {#manage-service}

可直接使用如下命令直接管理 Datakit：

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

有时候可能因为 Datakit 部分组件的 bug，导致服务操作失败（如 `datakit service -T` 之后，服务并未停止），可按照如下方式来强制处理。

Linux 下，如果上述命令失效，可使用以下命令来替代：

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

Mac 下，可以用如下命令代替：

```shell
# 启动 Datakit
sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist

# 停止 Datakit
sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist
```

### 服务卸载以及重装 {#uninstall-reinstall}

可直接使用如下命令直接卸载或恢复 Datakit 服务：

> 注意：此处卸载 Datakit 并不会删除 Datakit 相关文件。

```shell
# Linux/Mac shell
datakit service -I # re-install
datakit service -U # uninstall
```

## FAQ {#faq}

### :material-chat-question: Windows 下启动失败 {#windows-start-fail}

Datakit 在 Windows 下以服务的形式启动，启动后会写入很多 Event 日志，随着日志的积累，可能会出现如下报错：

``` not-set
Start service failed: The event log file is full.
```

该报错会导致 Datakit 无法启动。通过[设置一下 Windows Event](https://stackoverflow.com/a/13868216/342348){:target="_blank"} 即可（以 Windows Server 2016 为例）：

![ 修改 Windows Event 设置](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/set-windows-event-log.gif)
