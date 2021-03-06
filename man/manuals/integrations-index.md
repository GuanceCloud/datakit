---
icon: zy/integrations
---

# 集成
---

## 概述

观测云支持上百种数据的采集，通过配置采集源，可实时采集如主机、进程、容器、日志、应用性能、用户访问等多种数据。

## 前提条件

在配置采集源之前，需要先 [安装DataKit](../datakit/datakit-install.md) 。

## 配置

DataKit 安装完成后，可在DataKit的安装目录下配置采集源。DataKit 目前支持 `Linux/MacOS/Windows` 三种平台，其安装目录如下：

| 操作系统                            | 架构                | 安装路径                                                     |
| ----------------------------------- | ------------------- | ------------------------------------------------------------ |
| Linux 内核 2.6.23 或更高版本        | amd64/386/arm/arm64 | `/usr/local/datakit`                                         |
| macOS 10.11 或更高版本              | amd64               | `/usr/local/datakit`                                         |
| Windows 7, Server 2008R2 或更高版本 | amd64/386           | 64位：`C:\\Program Files\\datakit`<br>32位：`C:\\Program Files(32)\\datakit` |
|                                     |                     |                                                              |

进入DataKit 安装目录，打开采集源配置文件夹 `conf.d`，即可进行数据采集配置。采集源配置完成后，即可在观测云工作空间查看相应的数据。
