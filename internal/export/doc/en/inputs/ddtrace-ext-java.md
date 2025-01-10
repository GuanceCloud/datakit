---
title     : 'DDTrace Extensions'
summary   : 'GuanceCloud Extensions on DDTrace'
__int_icon: 'icon/ddtrace'
tags      :
  - 'DDTRACE'
  - 'APM'
  - 'TRACING'
---

## Introduction {#intro}

Here we mainly introduce some extended functions of DDTrace-Java. List of main features:

- JDBC SQL obfuscation
- xxl-jobs
- Dubbo 2/3
- Thrift
- RocketMQ
- log pattern
- hsf
- Support Alibaba Cloud RocketMQ 5.0
- redis trace parameters
- Get the input parameter information of a specific function
- MongoDB obfuscation
- Supported DM8 Database
- Supported `trace-128-bit-id`
- Supported Apache Pulsar MQ
- Support placing `trace_id` in the response header
- Support putting the requested header information into the span tags
- Support add HTTP `Response Body` information in the trace data
- Support add HTTP `Request Body` information in the trace data
- Use `-Ddd.http.error.enabled=true` to change the HTTP 4xx request link status to error
- Support `Mybatis-plus:batch`
- Support Redis tag:peer_ip


## HTTP Response,Request Body in the trace {#response_body}

The command line opening parameter is `-Ddd.trace.response.body.enabled=true`, the corresponding environment variable is `DD_TRACE_RESPONSE_BODY_ENABLED=true`, and the default value is `false`.

The command line opening parameter is `-Ddd.trace.request.body.enabled=true`, the corresponding environment variable is `DD_TRACE_REQUEST_BODY_ENABLED=true`, and the default value is `false`.

Since getting `response body` causes damage to `response`, the encoding adjustment of `response body` defaults to `utf-8`. If you need to adjust it, use `-Ddd.trace.response.body.encoding=gbk`.


## Tracing Header {#trace_header}

The link information will put the header information of the request and response into the tag.The default state is off. If it needs to be turned on, add the parameter `-Ddd.trace.headers.enabled=true`  during startup.

DDTrace supported version: [v1.25.2](ddtrace-ext-changelog.md#cl-1.25.2-guance)

## supported trace-128-id {#trace_128_bit_id}

[:octicons-tag-24: Datakit-1.8.0](../datakit/changelog.md#cl-1.8.0)
[:octicons-tag-24: DDTrace-1.4.0-guance](ddtrace-ext-changelog.md#cl-1.14.0-guance)

The default trace-id of the DDTrace agent is 64 bit, and the Datakit also supports 64 bit trace-id in the received link data.
Starting from `v1.11.0`, it supports the `W3C protocol` and supports receiving 128 bit trace-id. However, the trace id sent to the link is still 64 bit.

To this end, secondary development was carried out on the Guance Cloud, which incorporated `trace_128_bit_id` is placed in the link data and sent to the Datakit, the DDTrace and OTEL links can be concatenated.

how to config:

```shell
# open trace.128.bit, and use W3C propagation.
-Ddd.trace.128.bit.traceid.generation.enabled=true -Ddd.trace.propagation.style=tracecontext
```

This is  [GitHub issue](https://github.com/GuanceCloud/dd-trace-java/issues/37){:target="_blank"}

At present, only DDTrace and OTEL are connected in series, and there is currently no testing with other APM manufacturers.


## supported MongoDB obfuscation {#mongo-obfuscation}

Use startup parameter `-DDd.mongo.obfuscation=true` or environment variable `DD_MONGO_OBFUSION` Turn on desensitization. This way, a specific command can be seen from the Guance Cloud.

Currently, the types that can achieve desensitization include Int32, Int64, Boolean, Double, and String. The remaining ones have no reference significance, so they are currently not supported.

supported version：

- [x] all

DDTrace supported version: [v1.12.1](ddtrace-ext-changelog.md#cl-1.12.1-guance)

## supported DM8 Database {#dameng-db}
Add DM8 Database trace information.
supported version：

- [x] v8

<!-- markdownlint-disable MD013 -->
## Get the input parameter information of a specific function {#dd_trace_methods}
<!-- markdownlint-enable -->
**Specific function** mainly refers to the function specified by the business to obtain the corresponding input parameters.

**Specific functions** need to be defined and declared through specific parameters. Currently, ddtrace provides two ways to trace specific functions:

1. Marked by startup parameters: -Ddd.trace.methods ，reference documents： [Class or method injection Trace](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-param/#5-trace){:target="_blank"}

2. By introducing the SDK, use @Trace to mark, refer to the document [function level burying point](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-api/#2){:target="_blank"}

After the declaration is made in the above way, the corresponding method will be marked as trace, and the corresponding Span information will be generated at the same time, including the input parameter information of the function (input parameter name, type, value).

<!-- markdownlint-disable MD046 -->
???+ info

    Since the data type cannot be converted and JSON serialization requires additional dependencies and overhead, so far only `toString()` processing is done for the parameter value, and secondary processing is done for the result of `toString()`, the length of the field value It cannot exceed <font color="red">1024 characters</font>, and the excess part is discarded.

<!-- markdownlint-enable -->

DDTrace supported version： [v1.12.1](ddtrace-ext-changelog.md#cl-1.12.1-guance)

## ddtrace agent default port {#agent_port}

ddtrace changes the default remote port 8126 to 9529.

## redis command args {#redis-command-args}
The Resource in the redis link will only display redis.command information, and will not display parameter information.

Enable this function: start the command to add the environment variable `-Ddd.redis.command.args`, and a tag will be added in the details of the Guance Cloud link: `redis.command.args=key val`.

Supported version:

- [x] `Jedis1.4.0` and above
- [x] Lettuce
- [x] Redisson

## log pattern {#log-pattern}
By modifying the default log pattern, application logs and links are correlated, thereby reducing deployment costs. The logging framework `log4j2` is currently supported, but `logback` is not currently supported.

`-Ddd.logs.pattern` like `-Ddd.logs.pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger - %X{dd.service} %X{dd.trace_id} %X{dd.span_id} - %msg%n"`

supported version： log4j2

## SQL obfuscation {#jdbc-sql-obfuscation}

By default, DDTrace converts parameters in SQL to `?`, which prevents users from obtaining more accurate information when troubleshooting. The new probe will extract the parameters into the Trace data separately in Key-Value mode, which is convenient for users to view.

In the Java startup command, add the following command line parameters to enable this function:

```shell
-Ddd.jdbc.sql.obfuscation=true
```

### Display of results {#show}

Take setString() as an example. The location of the new probe is at `java.sql.PreparedStatement/setString(key, value)`。

There are two parameters here, the first one is the subscript of the placeholder parameter (starting from 1), the second one is the string type, after calling the `setString(index, value)` method, the corresponding string value will be stored into the span.

After the SQL is executed, this map will be filled into the Span. The final data structure format is as follows:

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
???+ question "Why is it not filled into the span？"

    The meta information here is actually for the relevant developers to check the specific content of the SQL statement. 
    After getting the specific details of the placeholder parameters, by replacing the `?` in `db.sql.origin`, the placeholder parameters can actually be The value is filled in, 
    but the correct replacement cannot be accurately found through string replacement (rather than SQL precise parsing) (may lead to wrong replacement), so **try to keep the original SQL** here, and the details of placeholder parameters are listed separately Listed, here `index_1` means the first placeholder parameter, and so on.
<!-- markdownlint-enable -->

supported version： Version 2.3 and above are currently supported.

DDTrace supported version：[v0.113.0](ddtrace-ext-changelog.md#ccl-0.113.0-new)

## Dubbo supported {#dubbo}

Dubbo is an open source framework of Alibaba Cloud, which currently supports Dubbo2 and Dubbo3.

supported version: Dubbo2: 2.7.0 and above, Dubbo3 has no version restrictions.

## RocketMQ {#rocketmq}

RocketMQ is an open source message queuing framework contributed by Alibaba Cloud to the Apache Foundation. Note: Alibaba Cloud RocketMQ 5.0 and the Apache Foundation are two different libraries.

There is a difference when referencing the library, the apache RocketMQ artifactId: `rocketmq-client`, and the artifactId of Alibaba Cloud RocketMQ 5.0: `rocketmq-client-java`

supported version: Currently supports version 4.8.0 and above. Alibaba Cloud RocketMQ service supports version 5.0 and above.

## Thrift supported {#thrift}

Thrift is an apache project. Some customers use thrift RPC for communication in the project, and we support it.

supported version: 0.9.3 and above.

## batch injection DDTrace-Java Agent {#java-attach}

The native DDTrace-Java batch injection method has certain flaws, and does not support dynamic parameter injection (such as `-Ddd.agent=xxx, -Ddd.agent.port=yyy`, etc.).

The extended DDTrace-Java adds dynamic parameter injection. For specific usage, see [here](ddtrace-attach.md)
