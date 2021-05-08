{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 选举文档

在批量安装 DataKit 时，可能会因为某些采集器使用相同的配置文件，导致采集到的数据重复。

例如集群中有 10 台 DataKit，所有 DataKit 都开启 Kubernetes 采集器且配置相同，如此将会 10 台机器全部采集同一个 Kubernetes 服务，出现数据重复的问题。

为此，DataKit 提供选举功能，特定采集器（详见下方列表）会参与选举。

## 选举使用方式

在需要开启选举的 DataKit 配置文件中，将 `enable_election` 参数设为 `true`，即表明此 DataKit 会参与选举。

DataKit 选举以工作空间 ID 划分，同一个工作空间如果只有一台 DataKit 开启选举，那它自己就是执行者。

以上述例子延续，如果集群中的 10 台 DataKit 全部开启选举功能，将只有一台 DataKit 采集 Kubernetes 数据，其他全部处于待命状态。

注意：如果这 10 台 DataKit 的 Kubernetes 配置不相同，但是全部开启了选举功能，会导致处于待命状态的 Kubernetes 采集器没执行从而未采集数据。

建议选举功能只在支持选举的采集器配置相同的情况下开启。

## 支持选举的采集列表

目前支持选举的采集器列表有以下：

| 采集器名称 |
| :---       |
| Kubernetes |
