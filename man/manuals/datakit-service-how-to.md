{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 服务管理

[DataKit 安装](datakit-install)完后，有必要对安装好的 DataKit 做一些基本的介绍。

## DataKit 目录介绍

DataKit 目前支持 Linux/Windows/Mac 三种主流平台：

| 操作系统                                                                  | 架构                | 安装路径                                                                   |
| ---------                                                                 | ---                 | ------                                                                     |
| Linux 内核 2.6.23 或更高版本                                              | amd64/386/arm/arm64 | `/usr/local/datakit`                                                       |
| macOS 10.12 或更高版本([原因](https://github.com/golang/go/issues/25633)) | amd64               | `/usr/local/datakit`                                                       |
| Windows 7, Server 2008R2 或更高版本                                       | amd64/386           | 64位：`C:\Program Files\datakit`<br />32位：`C:\Program Files(32)\datakit` |

> Tips：查看内核版本

- Linux/Mac：`uname -r`
- Windows：执行 `cmd` 命令（按住 Win键 + `r`，输入 `cmd` 回车），输入 `winver` 即可获取系统版本信息

安装完成以后，DataKit 目录列表大概如下：

```
├── [4.4K]  conf.d
├── [ 160]  data
├── [ 64M]  datakit
├── [ 192]  externals
├── [1.2K]  pipeline
├── [ 192]  gin.log   # Windows 平台
└── [1.2K]  log       # Windows 平台
```

其中：

- `conf.d`：存放所有采集器的配置示例。DataKit 主配置文件 datakit.conf 位于目录下
- `data`：存放 DataKit 运行所需的数据文件，如 IP 地址数据库等
- `datakit`：DataKit 主程序，Windows 下为 `datakit.exe`
- `externals`：部分采集器没有集成在 DataKit 主程序中，就都在这里了
- `pipeline` 存放用于文本处理的脚本代码
- `gin.log`：DataKit 可以接收外部的 HTTP 数据输入，这个日志文件相当于 HTTP 的 access-log
- `log`：DataKit 运行日志

> 注：Linux/Mac 平台下，DataKit 运行日志在 `/var/log/datakit` 目录下。

## DataKit 服务管理

可直接使用如下命令直接管理 DataKit：

```shell
# Linux/Mac 可能需加上 sudo
datakit --stop
datakit --start
datakit --restart
```

#### 服务管理失败处理

有时候可能因为 DataKit 部分组件的 bug，导致服务操作失败（如 `--stop` 之后，服务并未停止），可按照如下方式来强制处理。

Linux 下，如果上述命令失效，可使用以下命令来替代：

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

Mac 下，可以用如下命令代替：

```shell
# 启动 DataKit
sudo launchctl load -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
# 或者
sudo launchctl load -w /Library/LaunchDaemons/com.guance.datakit.plist

# 停止 DataKit
sudo launchctl unload -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
# 或者
sudo launchctl unload -w /Library/LaunchDaemons/com.guance.datakit.plist
```

#### 服务卸载以及重装

可直接使用如下命令直接卸载或恢复 DataKit 服务：

> 注意：此处卸载 DataKit 并不会删除 DataKit 相关文件。

```shell
# Linux/Mac shell
sudo datakit --uninstall
sudo datakit --reinstall

# Windows Powershell
datakit --uninstall
datakit --reinstall
```
