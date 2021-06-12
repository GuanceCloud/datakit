{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 选举

当集群中只有一个被采集对象（如 Kubernetes），但是在批量部署情况下，多个 DataKit 的配置完全相同，都开启了对该中心对象的采集，为了避免重复采集，我们可以开启 DataKit 的选举功能。

编辑 `conf.d/datakit.conf`，将如下开关开启：

```
enable_election = true
```

## 选举原理

以 Kubernetes 为例，在同一个集群中，假定有 10 DataKit 如果都开启了选举，且都开启了 Kubernetes 采集器：

- 一旦某个 DataKit 被选举上，那么所有 Kubernetes（其它选举类的采集也一样）的数据采集，都将由该 DataKit 来采集，不管被采集对象是一个还是多个，赢者通吃
- DataFlux 中心会判断当前选上的 DataKit 是否正常，如果异常，则强行踢掉该 DataKit，其它待命状态的 DataKit 将替代它
- 未开启选举的 DataKit，如果也配置了 Kubernetes 采集，不受选举约束，故不应出现这种情况，否则会出现 Kubernetes 被多次采集，一来造成数据浪费，二来也给 Kubernetes 造成无意义的 IO 开销
- 选举的范围是工作空间级别的，单个工作空间中，一次最多只能有一个 DataKit 被选上
	- 关于工作空间，在 datakit.conf 中，通过 DataWay 地址串中的 `token` URL 参数来表示，每个工作空间，都有其对应 token

## 支持选举的采集列表

目前支持选举的采集器列表有以下：

- [kubernetes](kubernetes)
- [gitlab](gitlab)
