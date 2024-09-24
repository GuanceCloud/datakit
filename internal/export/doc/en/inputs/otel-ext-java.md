---
title      : 'OpenTelemetry Extensions'
summary    : 'Guance Cloud added more OpenTelemetry plugins'
__int_icon : 'icon/opentelemetry'
tags       :
  - 'OTEL'
  - 'APM'
  - 'TRACING'
---

## SQL obfuscation {#sql-obfuscation}

Before understanding SQL Obfuscation, please read the official preprocessing scheme：

[DB statement sanitization](https://opentelemetry.io/docs/instrumentation/java/automatic/agent-config/#db-statement-sanitization)

### DB statement sanitization {#db-statement-sanitization}

Most of the sentences contain some sensitive data including: user name, mobile phone number, password, card number and so on. Another reason why these data can be filtered out through sensitive processing is to facilitate group filtering operations.

There are two ways to write SQL statements:

example:

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name=? and password=?");
ps.setString(1,username);   // set first?
ps.setString(2,pw);        //  second ?
```

This is the way of writing JDBC, such as the three-party library has nothing to do (both Oracle and Mysql are written in this way).

The result is that what you get in the link of this writing method is `db.statement` with two '?'

Another way of writing less:

```java
    ps = conn.prepareStatement("SELECT name,password,id FROM student where name='guance' and password='123456'");
   // ps.setString(1,username); 
   // ps.setString(2,pw);
```

At this time, what the agent gets is the SQL statement without placeholders.

The `OTEL_INSTRUMENTATION_COMMON_DB_STATEMENT_SANITIZER_ENABLED` mentioned above is here.

The reason is that the agent's probe is on the function `prepareStatement()` or `Statement()`.

Solve the desensitization problem fundamentally. Need to add probes to `set`. The parameters are cached before `executue()`, and finally the parameters are put into Attributes.

### Guance Cloud extension {#guacne-branch}

If you want to get the data before cleaning and the value added by the `set` function later, you need to make a new buried point and add an environment variable:

```shell
-Dotel.jdbc.sql.obfuscation=true
# or k8s 
OTEL_JDBC_SQL_OBFUSCATION=true
```

In the end, the link details on GuanceCloud look like this:

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/otel-sql.png" style="height: 500px" alt="trace">
  <figcaption> trace </figcaption>
</figure>

### Question {#question}

1. Enabling `-Dotel.jdbc.sql.obfuscation=true`, but SQL obfuscation not disabled.

   You may encounter a mismatch between placeholders and the number of `origin_sql_x` fields. This discrepancy occurs because some parameters have already been replaced by placeholders during the DB statement obfuscation process.

1. Enabling `-Dotel.jdbc.sql.obfuscation=true` and disabling DB statement obfuscation

   If the statements are excessively long or contain numerous line breaks, the formatting may become chaotic without proper formatting. Additionally, this can lead to unnecessary traffic consumption.

## More {#more}

If you have other questions, please ask in [GitHub-Guance](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues){:target="_blank"}
