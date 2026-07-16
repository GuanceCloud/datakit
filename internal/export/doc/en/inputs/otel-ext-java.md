---
title      : 'OpenTelemetry Extensions'
summary    : '<<<custom_key.brand_name>>> adds extra plugins for OpenTelemetry'
__int_icon : 'icon/opentelemetry'
tags       :
  - 'OTEL'
  - 'APM'
  - 'TRACING'
---

## SQL obfuscation {#sql-obfuscation}

OpenTelemetry Java Agent sanitizes SQL by default (`db.statement` parameters are replaced with `?`) to protect sensitive data.
This is defined in [DB statement sanitization](https://opentelemetry.io/docs/instrumentation/java/automatic/agent-config/#db-statement-sanitization){:target="_blank"}.

By default:

- SQL values such as usernames, phone numbers, and card numbers are replaced.
- Multiple spaces and line breaks are normalized.

### Example {#example}

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name=? and password=?");
ps.setString(1, username);
ps.setString(2, password);
```

The span receives `db.statement` with placeholders:

`SELECT name,password,id FROM student where name=? and password=?`

If you write SQL with inline literals (not recommended for sensitive data), OTEL will keep the raw text:

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name='abc' and password='123456'");
```

### Enable raw SQL capture in extension {#extension}

To capture values passed by `setXXX` and keep sensitive information for troubleshooting, enable one of:

```shell
-Dotel.jdbc.sql.obfuscation=true
# or k8s env
export OTEL_JDBC_SQL_OBFUSCATION=true
```

Or use the V2 extension switch:

```shell
-Dotel.instrumentation.jdbc.experimental.capture-query-parameters=true
# or k8s env
export OTEL_INSTRUMENTATION_JDBC_EXPERIMENTAL_CAPTURE_QUERY_PARAMETERS=true
```

Resulting trace detail:

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/otel-sql.png" style="height: 500px" alt="trace">
  <figcaption> trace </figcaption>
</figure>
<!-- markdownlint-enable -->

### FAQ {#question}

1. I enabled `-Dotel.jdbc.sql.obfuscation=true` but obfuscation still seems active.

   This can happen when some parameters have already been replaced during DB statement sanitization before extension capture.

2. After enabling raw SQL capture, SQL appears noisy or too long.

   Unformatted SQL (many line breaks, long values) can increase storage and transfer. This is expected and should be handled at log/trace retention and query policy levels.

If you need more help:

- [OpenTelemetry Java instrumentation docs](https://opentelemetry.io/docs/languages/java/instrumentation/){:target="_blank"}
