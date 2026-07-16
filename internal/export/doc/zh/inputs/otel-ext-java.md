---
title      : 'OpenTelemetry 扩展'
summary    : '<<<custom_key.brand_name>>> 对 OpenTelemetry 做了额外扩展'
__int_icon : 'icon/opentelemetry'
tags       :
  - 'OTEL'
  - '链路追踪'
  - 'APM'
---

> *作者：宋龙奇*

## SQL 脱敏 {#sql-obfuscation}

OpenTelemetry Java Agent 默认会对 SQL 进行脱敏：`db.statement` 中的参数值会被替换为 `?`，以降低敏感信息泄露风险。
官方说明见：
[DB statement sanitization](https://opentelemetry.io/docs/instrumentation/java/automatic/agent-config/#db-statement-sanitization){:target="_blank"}

默认脱敏行为包括：

- 替换值（字符串、数字）为占位符
- 压缩空白字符（多空格、换行）以便统一展示

### 示例 {#example}

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name=? and password=?");
ps.setString(1, username);   // set 了参数占位符 1
ps.setString(2, password);   // set 了参数占位符 2
```

链路中看到的会是：

`SELECT name,password,id FROM student where name=? and password=?`

如果使用内联 SQL（不建议在敏感数据场景下使用）：

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name='abc' and password='123456'");
```

链路会保留原始 SQL 文本。

### 开启脱敏参数采集（扩展） {#obfuscation}

如果需要拿到经过 `setXXX` 参数填充的 SQL 内容，请启用以下配置之一：

```shell
-Dotel.jdbc.sql.obfuscation=true
# or k8s
export OTEL_JDBC_SQL_OBFUSCATION=true
```

在 V2 扩展中也可使用官方参数：

```shell
-Dotel.instrumentation.jdbc.experimental.capture-query-parameters=true
# or k8s
export OTEL_INSTRUMENTATION_JDBC_EXPERIMENTAL_CAPTURE_QUERY_PARAMETERS=true
```

最终在 <<<custom_key.brand_name>>> 上看到的链路详情类似：

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/otel-sql.png" style="height: 500px" alt="trace">
  <figcaption> 链路详情 </figcaption>
</figure>
<!-- markdownlint-enable -->

### 常见问题 {#question}

1. 开启 `-Dotel.jdbc.sql.obfuscation=true` 后仍有参数被替换。

   部分参数可能在 `db.statement` 处理阶段已被替换，占位符与 `origin_sql_x` 数量不一致属于正常现象。

2. 开启原始 SQL 后内容很长、换行较多。

   这会产生更大的链路体积，建议结合 trace 保留策略与字段长度策略评估存储影响。

更多说明可参考：

- [OpenTelemetry Java Instrumentation 文档](https://opentelemetry.io/docs/languages/java/instrumentation/){:target="_blank"}
