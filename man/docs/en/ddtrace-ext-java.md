# Java

---

> *作者： 刘锐、宋龙奇*

## Introduction {#intro}

Here we mainly introduce some extended functions of DDTrace-java. List of main features:

- JDBC SQL obfuscation
- xxl-jobs
- dubbo 2/3
- Thrift
- RocketMQ
- log pattern
- hsf
- Support Alibaba Cloud RocketMQ 5.0
- redis trace parameters
- Get the input parameter information of a specific function

## Get the input parameter information of a specific function {#dd_trace_methods}
**Specific function** mainly refers to the function specified by the business to obtain the corresponding input parameters.

**Specific functions** need to be defined and declared through specific parameters. Currently, ddtrace provides two ways to trace specific functions:

1. Marked by startup parameters: -Ddd.trace.methods ，reference documents： [Class or method injection Trace](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-param/#5-trace)

2. By introducing the SDK, use @Trace to mark, refer to the document [function level burying point](https://docs.guance.com/integrations/apm/ddtrace/ddtrace-skill-api/#2)

After the declaration is made in the above way, the corresponding method will be marked as trace, and the corresponding Span information will be generated at the same time, including the input parameter information of the function (input parameter name, type, value).

???+ info

    Since the data type cannot be converted and json serialization requires additional dependencies and overhead, so far only `toString()` processing is done for the parameter value, and secondary processing is done for the result of `toString()`, the length of the field value It cannot exceed <font color="red">1024 characters</font>, and the excess part is discarded.

## ddtrace agent default port {#agent_port}
ddtrace changes the default remote port 8126 to 9529.

## redis command args {#redis-command-args}
The Resource in the redis link will only display redis.command information, and will not display parameter information.

Enable this function: start the command to add the environment variable `-Ddd.redis.command.args`, and a tag will be added in the details of the observation cloud link: `redis.command.args=key val`.

Supported version: jedis1.4.0 and above.

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

???+ question "Why is it not filled into the span？"

    The meta information here is actually for the relevant developers to check the specific content of the SQL statement. 
    After getting the specific details of the placeholder parameters, by replacing the `?` in `db.sql.origin`, the placeholder parameters can actually be The value is filled in, 
    but the correct replacement cannot be accurately found through string replacement (rather than SQL precise parsing) (may lead to wrong replacement), so **try to keep the original SQL** here, and the details of placeholder parameters are listed separately Listed, here `index_1` means the first placeholder parameter, and so on.


supported version： Version 2.3 and above are currently supported.

## Dubbo supported {#dubbo}

dubbo is an open source framework of Alibaba Cloud, which currently supports dubbo2 and dubbo3.

supported version: dubbo2: 2.7.0 and above, dubbo3 has no version restrictions.

## RocketMQ {#rocketmq}

RocketMQ is an open source message queuing framework contributed by Alibaba Cloud to the Apache Foundation. Note: Alibaba Cloud RocketMQ 5.0 and the Apache Foundation are two different libraries.

There is a difference when referencing the library, the apache rocketmq artifactId: `rocketmq-client`, and the artifactId of Alibaba Cloud rocketmq 5.0: `rocketmq-client-java`

supported version: Currently supports version 4.8.0 and above. Alibaba Cloud Rocketmq service supports version 5.0 and above.

## Thrift supported {#thrift}

Thrift is an apache project. Some customers use thrift RPC for communication in the project, and we support it.

supported version: 0.9.3 and above.

## batch injection DDTrace-Java Agent {#java-attach}

The native DDTrace-java batch injection method has certain flaws, and does not support dynamic parameter injection (such as `-Ddd.agent=xxx, -Ddd.agent.port=yyy`, etc.).

The extended DDTrace-java adds dynamic parameter injection. For specific usage, see [here](ddtrace-attach.md)
