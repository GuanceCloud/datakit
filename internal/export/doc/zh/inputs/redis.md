---
title     : 'Redis'
summary   : 'Redis 指标和日志采集'
tags:
  - '缓存'
  - '中间件'
__int_icon      : 'icon/redis'
dashboard :
  - desc  : 'Redis'
    path  : 'dashboard/zh/redis'
monitor:
  - desc: 'Redis'
    path: 'monitor/zh/redis'
---

{{.AvailableArchs}}

---

Redis 指标采集器，采集以下数据：

- 开启 AOF 数据持久化，会收集相关指标
- RDB 数据持久化指标
- Slow Log 监控指标
- Big Key scan 监控
- 主从 Replication
- 集群

## 配置 {#config}

已测试的版本：

- [x] 7.0.11
- [x] 6.2.12
- [x] 6.0.8
- [x] 5.0.14
- [x] 4.0.14

### 前置条件 {#reqirement}

- 在采集主从架构下数据时，配置从节点或主节点的主机信息进行数据采集，可以得到不同的主从相关的指标信息。
- 创建监控用户（**可选**）：redis 6.0+ 进入 `redis-cli` 命令行，创建用户并且授权：

```sql
ACL SETUSER username >password
ACL SETUSER username on +@dangerous +ping
```

- 授权统计 `hotkey/bigkey` 信息，进入 `redis-cli` 命令行：

```sql
CONFIG SET maxmemory-policy allkeys-lfu
ACL SETUSER username on +get +@read +@connection +@keyspace ~*
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

---

???+ note

    如果是阿里云 Redis，且设置了对应的用户名密码，conf 中的 `<PASSWORD>` 应该设置成 `your-user:your-password`，如 `datakit:Pa55W0rd`
<!-- markdownlint-enable -->

#### 集群/主从节点采集 {#cluster}

Redis 集群节点和主从节点的采集和单点节点（standalone）的配置不同，具体而言，我们需要单独配置如下两个部分：

```toml
# cluster
[inputs.redis.cluster]
  hosts = [ "localhost:6379" ]

# master/slave
[inputs.redis.master_slave]
  hosts       = [ "localhost:26380" ] # master or/and slave ip/host
  [inputs.redis.master_slave.sentinel]
    hosts       = [ "localhost:26380" ] # sentinel ip/host
    master_name = "your-master-name"
    password    = "sentinel-requirepassword"
```

对集群/主从节点的采集，默认遵循如下原则：

- 对于常规指标采集，比如 `INFO/CLIENT LIST` 等采集，每次采集会对主从节点都实施采集
- 低频的深度采集（比如 hot/big key 采集），由于其消耗较大（aka 对采集目标有一定的性能开销），可配置随机选择一个从库实施采集（集群不可以）

### 日志采集配置 {#logging-config}

需要采集 Redis 日志，需要开启 Redis `redis.config` 中日志文件输出配置：

```toml
[inputs.redis.log]
    # 日志路径需要填入绝对路径
    files = ["/var/log/redis/*.log"]
```

<!-- markdownlint-disable MD046 -->
???+ note

    在配置日志采集时，需要将 DataKit 安装在 Redis 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器。

    在 K8s 中，可以将 Redis 日志暴露到 stdout，DataKit 能自动找到其对应的日志。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 自定义对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.DescZh}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable -->

### Pipeline 日志切割 {#pipeline}

原始日志为

```log
122:M 14 May 2019 19:11:40.164 * Background saving terminated with success
```

切割后的字段列表如下：

| 字段名      | 字段值                                      | 说明                         |
| ---         | ---                                         | ---                         |
| `pid`       | `122`                                       | 进程 id                      |
| `role`      | `M`                                         | 角色                         |
| `serverity` | `*`                                         | 服务                         |
| `statu`     | `notice`                                    | 日志级别                     |
| `msg`       | `Background saving terminated with success` | 日志内容                     |
| `time`      | `1557861100164000000`                       | 纳秒时间戳（作为行协议时间） |
