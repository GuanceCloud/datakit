# OTEL SQL 脱敏
---

> *作者： 宋龙奇*

SQL 脱敏 也是狭义的 DB 语句清理。

按照 OTEL 官方说法是：

```text
agent 在设置 `db.statement` 语义属性之前清理所有数据库查询/语句。查询字符串中的所有值（字符串、数字）都替换为问号 ( ?)，并对整条 sql 进行格式化（用空格替换换行符，空格只留一个）等操作。


例子：

- SQL 查询 SELECT a from b where password="secret" 那么在 span 中将出现 SELECT a from b where password=?


默认情况下，所有数据库检测都启用此行为。使用以下属性禁用它：

系统属性： otel.instrumentation.common.db-statement-sanitizer.enabled
环境变量： OTEL_INSTRUMENTATION_COMMON_DB_STATEMENT_SANITIZER_ENABLED

默认值：true
说明：启用 DB 语句清理。
```

## DB 语句清理和结果 {#why}

大部分语句中都包括一些敏感数据包括：用户名，手机号，密码，卡号等等。通过敏感处理能够过滤掉这些数据，另一个原因就是方便进行分组筛选操作。

SQL 语句的写法有两种：

比如 ：

```java
ps = conn.prepareStatement("SELECT name,password,id FROM student where name=? and password=?");
ps.setString(1,username);   // set   替换第一个?
ps.setString(2,pw);        //  替换第二个?
```

这是 JDBC 的写法 如库无关（ oracle 和 mysql 都是这么写的）。

结果就是，这种写法链路中拿到的就是 两个 '?' 的 `db.statement`

较少的另一种写法：

```java
    ps = conn.prepareStatement("SELECT name,password,id FROM student where name='guance' and password='123456'");
   // ps.setString(1,username);  不再使用 set
   // ps.setString(2,pw);
```

这时候 agent 拿到的就是 不带占位符的 SQL 语句。

上面说的 `OTEL_INSTRUMENTATION_COMMON_DB_STATEMENT_SANITIZER_ENABLED` 是作用是这里的。

究其原因，是因为 agent 的探针 是在函数 `prepareStatement` 或者 `Statement` 上。


从根本上解决脱敏问题。需要加探针加在 `set` 上。先将参数缓存之后才是 `exectue()` , 最终将参数放到 Attributes 中。

## 观测云二次开发 {#guacne-branch}

想要获取清洗前的数据以及后续通过 `set` 函数添加的值，就需要进行新的埋点， 并添加环境变量：

```shell
-Dotel.jdbc.sql.obfuscation=true
# or k8s 
OTEL_JDBC_SQL_OBFUSCATION=true
```

最终，在观测云上的链路详情里是这样的：

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/songlongqi/otel-sql.png" style="height: 500px" alt="链路详情">
  <figcaption> 链路详情 </figcaption>
</figure>


## 常见问题 {#question}

1. 开启 `-Dotel.jdbc.sql.obfuscation=true` 但是没有关闭 DB 语句清洗

    ```text
        可能会出现占位符和 `origin_sql_x` 数量对不上。原因是因为有的参数已经在 DB 语句清洗中被占位符替换掉了。
    ```

2. 开启 `-Dotel.jdbc.sql.obfuscation=true` 关闭 DB 语句清洗

 ```text
     如果语句过长或者换行符很多，没有进行格式化的情况下语句会很混乱。同时也会造成没必要的流量浪费。
 ```

## 更多 {#more}

如果有其他问题请前往：[GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues){:target="_blank"}
