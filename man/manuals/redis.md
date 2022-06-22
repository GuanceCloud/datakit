{{.CSS}}
# Redis
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

Redis 指标采集器，采集以下数据：

- 开启 AOF 数据持久化，会收集相关指标
- RDB 数据持久化指标
- Slowlog 监控指标
- bigkey scan 监控
- 主从replication

## 场景视图
Redis 观测场景主要展示了 Redis的错误信息，性能信息，持久化信息等。

![image.png](../imgs/redis-1.png)


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

### 指标采集

1. 开启 Datakit Redis 插件，复制 sample 文件

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

> 如果是阿里云 Redis，且设置了对应的用户名密码，下面的 `<PASSWORD>` 应该设置成 `your-user:your-password`，如 `datakit:Pa55W0rd`

```toml
{{.InputSample}}
```

2. 修改 `redis.conf` 配置文件
```bash
vi redis.conf
```
参数说明

- host：要采集的redis 的地址
- port：要采集的redis 的端口
- db：要采集的redis 的数据库
- password：要采集的redis 的密码
- connect_timeout：链接超时时间
- service：配置服务名称
- interval：采集指标频率
- keys：要采集的 key 可以多选
- slow_log：开启慢日志
- slowlog-max-len：配置慢日志大小
- command_stats：获取 info 命令的结果转换成指标

```yaml
[[inputs.redis]]
    host = "localhost"
    port = 6379
    # unix_socket_path = "/var/run/redis/redis.sock"
    db = 0
    # password = "<PASSWORD>"

    ## @param connect_timeout - number - optional - default: 10s
    # connect_timeout = "10s"

    ## @param service - string - optional
    # service = "<SERVICE>"

    ## @param interval - number - optional - default: 15
    interval = "15s"

    ## @param keys - list of strings - optional
    ## The length is 1 for strings.
    ## The length is zero for keys that have a type other than list, set, hash, or sorted set.
    #
    # keys = ["KEY_1", "KEY_PATTERN"]

    ## @param warn_on_missing_keys - boolean - optional - default: true
    ## If you provide a list of 'keys', set this to true to have the Agent log a warning
    ## when keys are missing.
    #
    # warn_on_missing_keys = true

    ## @param slow_log - boolean - optional - default: false
    slow_log = true

    ## @param slowlog-max-len - integer - optional - default: 128
    slowlog-max-len = 128

    ## @param command_stats - boolean - optional - default: false
    ## Collect INFO COMMANDSTATS output as metrics.
    # command_stats = false

```

3. 重启 Datakit (如果需要开启日志，请配置日志采集再重启)
```bash
systemctl restart datakit
```

4. Redis 指标采集验证 `/usr/local/datakit/datakit -M |egrep "最近采集|redis"`

![image.png](../imgs/redis-2.png)

5. 指标预览

![image.png](../imgs/redis-3.png)


#### 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

#### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

### 日志采集
需要采集redis日志，需要开启Redis `redis.config`中日志文件输出配置

```toml
[inputs.redis.log]
    # 日志路径需要填入绝对路径
    files = ["/var/log/redis/*.log"] # 在使用日志采集时，需要将datakit安装在redis服务同一台主机中，或使用其它方式将日志挂载到外部系统中
```

参数说明

- files：日志文件路径 (通常填写访问日志和错误日志)
- ignore：要过滤的文件名
- pipeline：日志切割文件(内置)，实际文件路径 /usr/local/datakit/pipeline/redis.p
- character_encoding：日志编码格式
- match：开启多行日志收集

#### 日志 pipeline 功能切割字段说明

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
场景 - 新建场景 - Redis 监控场景
## 异常检测
异常检测库 - 新建检测库 - Redis 检测库

| 序号 | 规则名称 | 触发条件 | 级别 | 检测频率 |
| --- | --- | --- | --- | --- |
| 1 | Redis 等待阻塞命令的客户端连接数异常增加 | 客户端连接数 > 0 | 紧急 | 1m |

## 最佳实践
- [Redis可观测最佳实践](/best-practices/integrations/redis)
## 故障排查
- [无数据上报排查](why-no-data.md)

## 进一步阅读
[DataFlux pipeline 文本数据处理](/datakit/pipeline)