---
title     : 'Profiling'
summary   : '采集应用程序的运行时性能数据'
__int_icon: 'icon/profiling'
tags:
  - 'PROFILE'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

Profile 支持采集使用 Java, Python 和 Go 等不同语言环境下应用程序运行过程中的动态性能数据，帮助用户查看 CPU、内存、IO 的性能问题。

## 配置 {#config}

目前 DataKit 采集 profiling 数据有两种方式：

- 推送方式：需要开启 DataKit Profile 服务，由客户端向 DataKit 主动推送数据
- 拉取方式：目前仅 [Go](profile-go.md) 支持，需要手动配置相关信息

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/profile` 目录，复制 `profile.conf.sample` 并命名为  `profile.conf` 。配置文件说明如下：
    
    ```shell
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) ，开启 Profile 服务。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 客户端应用配置 {#app-config}

客户的应用根据编程语言需要分别进行配置，目前支持的语言如下：

- [Java](profile-java.md)
- [Go](profile-go.md)
- [Python](profile-python.md)
- [C/C++](profile-cpp.md)
- [NodeJS](profile-nodejs.md)
- [.NET](profile-dotnet.md)
