{{.CSS}}
# Redis
---

- DataKit 版本：{{.Version}}
- 操作系统支持：{{.AvailableArchs}}

Redis 指标采集器，采集以下数据：

- 开启 AOF 数据持久化，会收集相关指标
- RDB 数据持久化指标
- Slowlog 监控指标
- bigkey scan 监控
- 主从replication

![](imgs/input-redis-1.png)

## 前置条件

- Redis 版本 v5.0+

在采集主从架构下数据时，请配置从节点的主机信息进行数据采集，可以得到主从相关的指标信息。

创建监控用户

redis6.0+ 进入redis-cli命令行,创建用户并且授权

```sql
ACL SETUSER username >password
ACL SETUSER username on +@dangerous
ACL SETUSER username on +ping
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

> 如果是阿里云 Redis，且设置了对应的用户名密码，下面的 `<PASSWORD>` 应该设置成 `your-user:your-password`，如 `datakit:Pa55W0rd`

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标预览

![](imgs/input-redis-2.png)


## 日志预览

![](imgs/input-redis-3.png)

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### 指标 {#metric}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### 日志 {#logging}

[:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 日志采集 {#redis-logging}

需要采集 Redis 日志，需要开启 Redis `redis.config`中日志文件输出配置：

```toml
[inputs.redis.log]
    # 日志路径需要填入绝对路径
    files = ["/var/log/redis/*.log"]
```

???+ attention

    在配置日志采集时，需要将 DataKit 安装在 Redis 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器。

    在 K8s 中，可以将 Redis 日志暴露到 stdout，DataKit 能自动找到其对应的日志。

### Pipeline 日志切割 {#pipeline}

原始日志为

```
122:M 14 May 2019 19:11:40.164 * Background saving terminated with success
```

切割后的字段列表如下：

| 字段名      | 字段值                                      | 说明                         |
| ---         | ---                                         | ---                          |
| `pid`       | `122`                                       | 进程id                       |
| `role`      | `M`                                         | 角色                         |
| `serverity` | `*`                                         | 服务                         |
| `statu`     | `notice`                                    | 日志级别                     |
| `msg`       | `Background saving terminated with success` | 日志内容                     |
| `time`      | `1557861100164000000`                       | 纳秒时间戳（作为行协议时间） |

## 场景视图

<场景 - 新建场景 - Redis 监控场景>

## 异常检测

<异常检测库 - 新建检测库 - Redis 检测库>

## 更多阅读

- [Redis 可观测最佳实践](../best-practices/integrations/redis.md)
