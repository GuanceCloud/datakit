# Java

---

> *作者： 刘锐、宋龙奇*

## 简介 {#intro}

这里主要介绍一下 DDTrace-java 的一些扩展功能。主要功能列表：

- JDBC SQL 脱敏功能
- xxl-jobs 支持
- dubbo 2/3 支持
- Thrift 框架支持
- RocketMQ 支持
- log pattern 自定义
- hsf 支持
- 阿里云 RocketMQ 5.0 支持

## log pattern 支持自定义 {#log-pattern}
通过修改默认的 log pattern 来实现应用日志和链路互相关联，从而降低部署成本。目前已支持`log4j2`日志框架，对于`logback` 暂不支持。

通过`-Ddd.logs.pattern`来调整默认的pattern。比如 `-Ddd.logs.pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger - %X{dd.service} %X{dd.trace_id} %X{dd.span_id} - %msg%n"`

支持版本： log4j2

## hsf {#hsf}
[hsf](https://help.aliyun.com/document_detail/100087.html) 是在阿里巴巴广泛使用的分布式 RPC 服务框架。

支持版本：2.2.8.2--2019-06-stable

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

???+ question "为什么没有填充到 `db.sql.origin` 中？"

    这里的 meta 信息实际是给相关开发人员排查 SQL 语句具体内容的，在拿到具体的占位参数详情后，通过替换 `db.sql.origin` 中的 `?` 实际上是可以将占位参数的值填充进去，但通过字符串替换（而不是 SQL 精确解析）并不能准确的找到正确的替换（可能导致错误的替换），故此处**尽量保留原始 SQL**，占位参数详情则单独列出来，此处 `index_1` 即表示第一个占位参数，以此类推。

## xxl-jobs 支持 {#xxl-jobs}

前言： xxl-jobs 是一个基于Java开发的分布式任务调度框架, [github 地址](https://github.com/xuxueli/xxl-job) 

版本支持： 目前支持 2.3 及以上版本。

## Dubbo 支持 {#dubbo}

dubbo 是阿里云的一个开源框架，目前已经支持 dubbo2 以及 dubbo3.

版本支持： dubbo2 ：2.7.0及以上， dubbo3 无版本限制。

## RocketMQ {#rocketmq}

RocketMQ 是阿里云贡献 Apache 基金会的开源消息队列框架。注意：阿里云 RocketMQ 5.0 与 Apache基金会的是两个不同的库。 

版本支持： 目前支持 4.8.0 及以上版本。

## Thrift 支持 {#thrift}

Thrift 属于 apache 的项目。有的客户在项目中使用 thrift RPC 进行通讯，我们就做了支持。

版本支持 ： 0.9.3 及以上版本。

## 批量注入 DDTrace-Java Agent {#java-attach}

原生的 DDTrace-java 批量注入方式有一定的缺陷，不支持动态参数注入（比如 `-Ddd.agent=xxx, -Ddd.agent.port=yyy` 等）。

扩展的 DDTrace-java 增加了动态参数注入功能。具体用法，参见[这里](ddtrace-attach.md)
