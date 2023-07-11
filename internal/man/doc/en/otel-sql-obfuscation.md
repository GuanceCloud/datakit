# OTEL SQL Obfuscation
---

Before understanding SQL Obfuscation, please read the official preprocessing scheme：

[DB statement sanitization](https://opentelemetry.io/docs/instrumentation/java/automatic/agent-config/#db-statement-sanitization)


## DB statement sanitization {#db-statement-sanitization}

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

## Guance Branch {#guacne-branch}

If you want to get the data before cleaning and the value added by the `set` function later, you need to make a new buried point and add an environment variable:

```shell
-Dotel.jdbc.sql.obfuscation=true
# or k8s 
OTEL_JDBC_SQL_OBFUSCATION=true
```

In the end, the link details on Observation Cloud look like this:

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/otel-sql.png" style="height: 500px" alt="trace">
  <figcaption> trace </figcaption>
</figure>


## Question {#question}

1. Open `otel.jdbc.sql.obfuscation=true` and `otel.instrumentation.common.db-statement-sanitizer.enabled=true`

    ```text
        There may be a mismatch between the number of placeholders and `origin_sql_x`. The reason is that some parameters have been replaced by placeholders in the DB statement cleaning.
    ```


## More {#more}
If you have other questions, please ask in [GitHub-Guance](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues){:target="_blank"}
