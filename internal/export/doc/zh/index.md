---
icon: zy/datakit
---

# DataKit
---

## 概述 {#intro}

DataKit 是运行在您服务器上的数据采集客户端，它将采集的数据发送给<<<custom_key.brand_name>>>，在<<<custom_key.brand_name>>> Studio 上，您可以查看并分析这些数据。

DataKit 是一款开源软件，我们从 [GitHub](https://github.com/GuanceCloud/datakit){:target="_blank"} 可以获取到它的源码。

## 快速开始 {#quick-start}

在主流平台上，我们通过单个命令即可安装 DataKit。安装完成后，默认开启了[一部分采集器](datakit-input-conf.md#default-enabled-inputs)，通过这些采集器，我们能采集到主机的一些基本数据。

<div class="grid cards" markdown>
- :fontawesome-solid-computer: [主机安装](datakit-install.md#get-install)
- :fontawesome-brands-docker: [Docker 安装](datakit-docker-deploy.md)
- :material-kubernetes: [Kubernetes 安装](datakit-daemonset-deploy.md)
- :simple-amazoneks: [AWS EKS 安装](datakit-eks-deploy.md)
- :simple-awsfargate: [AWS Fargate 安装](ecs-fargate.md)
- :octicons-cloud-offline-24: [离线安装](datakit-offline-install.md)
- [IoT 精简版](datakit-install.md#lite-install)
</div>

在本部分中，会包含如下主要内容：

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 基本使用</u>: 如何管理 DataKit 服务 </font>](datakit-service-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit 工具命令</u>: DataKit 提供了很多便捷工具来辅助您的日常使用</font>](datakit-tools-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Monitor</u>: DataKit 运行状态查看</font>](datakit-monitor.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Kubernetes Operator</u>: 通过 Operator 自动化 Kubernetes 中的采集配置</font>](datakit-operator.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Proxy</u>: 如果带宽首先可以通过网络代理来上传 DataKit 流程</font>](datakit-proxy.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>日志采集</u>: 通过 DataKit 采集您的应用日志</font>](../integrations/logging.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Security</u>: 配置 DataKit 过程中涉及的一些安全问题说明</font>](datakit-conf.md#public-apis)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>资源限制</u>: 限制 DataKit 的资源开销 </font>](datakit-conf.md#resource-limit)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Troubleshooting</u>: 调试 DataKit 采集过程中的问题</font>](why-no-data.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>API</u>: DataKit HTTP API](apis.md)
</div>
</font>

## 说明 {#spec}

### 实验性功能 {#experimental}

DataKit 发布的时候，会带上一些实验性功能，这些功能往往是初次发布的新功能，部分实现可能会有一些欠缺考虑或不严谨的地方：

- 功能不太稳定
- 一些功能配置，在后续的迭代过程中不保证其兼容性
- 功能可能会被移除，但会有对应的其它措施来满足对应的需求

### 图例说明 {#legends}

| 图例                                                                                                                       | 说明                                                            |
| ---                                                                                                                        | ---                                                             |
| :fontawesome-solid-flag-checkered:                                                                                         | 表示该采集器支持选举                                            |
| :fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker: | 例分别用来表示 Linux、Windows、macOS、 Kubernetes 以及 Docker   |
| :octicons-beaker-24:                                                                                                       | 表示实验性功能（参见[实验性功能的描述](index.md#experimental)） |
