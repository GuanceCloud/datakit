
# Java

---

## 简介 {#intro}

这里主要介绍一下 DDTrace-Java 的一些扩展功能。主要功能列表：

- JDBC SQL 脱敏功能
- xxl-jobs 支持
- Dubbo 2/3 支持
- Thrift 框架支持
- RocketMQ 支持
- Log Pattern 自定义
- HSF 支持
- 阿里云 RocketMQ 5.0 支持
- Redis 链路增加参数
- 获取特定函数的入参信息
- 支持 MongoDB 脱敏
- 支持达梦国产数据库


## 支持 MongoDB 数据库脱敏 {#mongo-obfuscation}
使用启动参数 `-Ddd.mongo.obfuscation=true` 或者环境变量 `DD_MONGO_OBFUSCATION` 开启脱敏。这样从观测云上就可以看见一条具体的命令。

目前可以实现脱敏的类型有：Int32/Int64/Boolean/Double/String 。 剩余的并没有参考意义，所以目前暂不支持。

支持的版本：

- [x] all

## 支持达梦国产数据库 {#dameng-db}
支持版本：

- [x] v8

## 获取特定函数的入参信息 {#dd-trace-methods}

特定函数主要是指业务指定的函数，来获取对应的入参情况。特定函数需要通过特定的参数进行定义声明，目前 DDTrace 提供了两种方式对特定的函数进行 trace 声明：

1. 通过启动参数标记 `-Ddd.trace.methods`，参考 [类或方法注入 Trace](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-param/#5-trace){:target="_blank"}
1. 通过引入 SDK 的方式，使用 `@Trace` 进行标记，参考 [函数级别埋点](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-api/#2){:target="_blank"}

通过上述方式进行声明后，会将对应的方法标记为 trace，同时生成对应的 Span 信息并包含函数（方法）的入参信息（入参名称、类型、值）。

<!-- markdownlint-disable MD046 -->
???+ attention

    由于无法对数据类型进行转化以及 JSON 序列化需要额外的依赖和开销，所以目前只是针对参数值做了 `toString()` 处理，且对于 `toString()` 结果做了二次处理，字段值长度不能超过 1024 个字符，对于超过部分做了丢弃操作。
<!-- markdownlint-enable -->

## DDTrace agent 默认远端端口 {#agent-port}

DDTrace 二次开发将默认的远端端口 8126 修改为 9529。

## Redis 链路中查看参数 {#redis-command-args}

Redis 的链路中的 Resource 只会显示 `redis.command` 信息，并不会显示参数（args）信息。

开启此功能：启动命令添加环境变量 `-Ddd.redis.command.args`，在观测云链路的详情中会增加一个 Tag：`redis.command.args=key val`。

支持版本：

- [x] Jedis 1.4.0+

## log pattern 支持自定义 {#log-pattern}

通过修改默认的 log pattern 来实现应用日志和链路互相关联，从而降低部署成本。目前已支持 Log4j2 日志框架，对于 Logback 暂不支持。

通过 `-Ddd.logs.pattern` 来调整默认的 Pattern，比如：

``` not-set
-Ddd.logs.pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger - %X{dd.service} %X{dd.trace_id} %X{dd.span_id} - %msg%n"`
```

支持版本：

- [x] log4j2

## HSF {#hsf}

[HSF](https://help.aliyun.com/document_detail/100087.html){:target="_blank"} 是在阿里巴巴广泛使用的分布式 RPC 服务框架。

支持版本：

- [x] 2.2.8.2--2019-06-stable

## SQL 脱敏 {#jdbc-sql-obfuscation}

DDTrace 默认会将 SQL 中参数转化为 `?`，这导致用户在排查问题时无法获取更准确的信息。新的探针会将占位参数单独以 Key-Value 方式提取到 Trace 数据中，便于用户查看。

在 Java 启动命令中，增加如下命令行参数来开启该功能：

```shell
# ddtrace 启动时增加参数，默认是 false
-Ddd.jdbc.sql.obfuscation=true
```

### 效果示例 {#show}

以 setString() 为例。新增探针的位置在 `java.sql.PreparedStatement/setString(key, value)`。

这里有两个参数，启动第一个是占位参数下标（从 1 开始），第二个为 string 类型，在调用 `setString(index, value)` 方法之后，会将对应的字符串值存放到 span 中。

在 SQL 被执行之后，这个 map 会填充到 Span 中。 最终的数据结构格式如下所示：

```json hl_lines="17 26 27 28 29 30 31 32"
"meta": {
  "component":
  "java-jdbc-prepared_statement",

  "db.instance":"tmalldemodb",
  "db.operation":"INSERT",

  "db.sql.origin":"INSERT product
      (product_id,
       product_name,
       product_title,
       product_price,
       product_sale_price,
       product_create_date,
       product_isEnabled,
       product_category_id)
      VALUES(null, ?, ?, ?, ?, ?, ?, ?)",

  "db.type":"mysql",
  "db.user":"root",
  "env":"test",
  "peer.hostname":"49.232.153.84",
  "span.kind":"client",
  "thread.name": "http-nio-8080-exec-6",

  "sql.params.index_1":"图书",
  "sql.params.index_2":"十万个为什么",
  "sql.params.index_3":"100.0",
  "sql.params.index_4":"99.0",
  "sql.params.index_5":"2022-11-10 14:08:21",
  "sql.params.index_6":"0",
  "sql.params.index_7":"16"
}
```

<!-- markdownlint-disable MD046 -->
???+ question "为什么没有填充到 `db.sql.origin` 中？"

    这里的 meta 信息实际是给相关开发人员排查 SQL 语句具体内容的，在拿到具体的占位参数详情后，通过替换 `db.sql.origin` 中的 `?` 实际上是可以将占位参数的值填充进去，但通过字符串替换（而不是 SQL 精确解析）并不能准确的找到正确的替换（可能导致错误的替换），故此处**尽量保留原始 SQL**，占位参数详情则单独列出来，此处 `index_1` 即表示第一个占位参数，以此类推。
<!-- markdownlint-enable -->

## xxl-jobs 支持 {#xxl-jobs}

[xxl-jobs](https://github.com/xuxueli/xxl-job){:target="_blank"} 是一个 Java 开发的分布式任务调度框架。

支持版本：

- [x] 2.3 及以上版本

## Dubbo 支持 {#dubbo}

Dubbo 是阿里云的一个开源框架，目前已经支持 Dubbo2 以及 Dubbo3。

支持版本：

- [x] Dubbo2：2.7.0+
- [x] Dubbo3：全支持

## RocketMQ {#rocketmq}

RocketMQ 是阿里云贡献 Apache 基金会的开源消息队列框架。注意：阿里云 RocketMQ 5.0 与 Apache 基金会的是两个不同的库。

引用库时有区别，`apache rocketmq artifactId: rocketmq-client`, 而阿里云 RocketMQ 5.0 的 `artifactId：rocketmq-client-java`

版本支持：目前支持 4.8.0 及以上版本。 阿里云 RocketMQ 服务支持 5.0 以上。

## Thrift 支持 {#thrift}

Thrift 属于 apache 的项目。有的客户在项目中使用 thrift RPC 进行通讯，我们就做了支持。

支持版本：

- [x] 0.9.3 及以上版本

## 批量注入 DDTrace-Java Agent {#java-attach}

原生的 DDTrace-Java 批量注入方式有一定的缺陷，不支持动态参数注入（比如 `-Ddd.agent=xxx, -Ddd.agent.port=yyy` 等）。

扩展的 DDTrace-Java 增加了动态参数注入功能。具体用法，参见[这里](ddtrace-attach.md)
