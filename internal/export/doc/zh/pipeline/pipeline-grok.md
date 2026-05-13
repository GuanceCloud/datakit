# Grok 模式
---

## Grok 模式简介 {#grok-pattern}

DataKit Pipeline 提供 [grok()](pipeline-built-in-function.md#fn-grok) 函数实现对执行 Grok 模式的支持（实现上，grok() 函数会将 Grok 模式翻译为正则表达式），并提供 [add_pattern()](pipeline-built-in-function.md#fn-add-pattern) 函数来添加自定义命名模式。

Grok 模式基于正则表达式，模式在命名后可以通过以下三种写法使用在其他模式中，注意不要循环引用：

- `%{pattern_name}`
- `%{pattern_name:key_name}`
- `%{pattern_name:key_name:type}`

其中 `type` 的值可以是范围是 {`float`, `int`, `str`, `bool`}；可以通过组合 Grok 模式获得更复杂的 Grok 模式。

任意正则表达式都可视为一个合法的 Grok 模式，并支持混合使用命名 Grok 模式和正则表达式编写 Grok 模式；

对于模式写法 `%{pattern_name:key_name}`，其等价于正则表达式中的命名捕获组：  

```regexp
(?P<key_name>pattern)
```

## DataKit 中 Grok 模式分类 {#grok-pattern-class}

DataKit 中 Grok 模式可以分为两类：

- 全局模式：*pattern* 目录下的模式文件都是全局模式，所有 Pipeline 脚本都可使用
- 局部模式：在 Pipeline 脚本中通过 [add_pattern()](pipeline-built-in-function.md#fn-add-pattern) 函数新增的模式为局部模式，只针对当前 Pipeline 脚本有效

以下以 Nginx access-log 为例，说明一下如何编写对应的 Grok 模式，原始 nginx access log 如下：

```log
127.0.0.1 - - [26/May/2022:20:53:52 +0800] "GET /server_status HTTP/1.1" 404 134 "-" "Go-http-client/1.1"
```

假设我们需要从该访问日志中获取 client_ip、time (request)、http_method、http_url、http_version、status_code 这些内容，那么 Grok 模式初步可以写成：

```python
grok(_,"%{NOTSPACE:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT} \"%{NOTSPACE}\" \"%{NOTSPACE}\"")

cast(status_code, "int")
group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)
default_time(time)
```

再优化一下，分别将对应的特征提取一下：

```python
# 日志首部的 client_ip、http_ident、http_auth 作为一个 pattern
add_pattern("p1", "%{NOTSPACE:client_ip} %{NOTSPACE} %{NOTSPACE}")

# 中间的 http_method、http_url、http_version、status_code 作为一个 pattern，
# 并在 pattern 内指定 status_code 的数据类型 int 来替代使用的 cast 函数
add_pattern("p3", '"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{INT:status_code:int}')

grok(_, "%{p1} \\[%{HTTPDATE:time}\\] %{p3} %{INT} \"%{NOTSPACE}\" \"%{NOTSPACE}\"")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)

default_time(time)
```

优化之后的切割，相较于初步的单行 pattern 来说可读性更好。由于根据 Grok 模式解析出的字段默认数据类型是 string，在此处指定字段的数据类型后，可以避免后续再使用 [cast()](pipeline-built-in-function.md#fn-cast) 函数来进行类型转换。

### 自定义 Grok 模式 {#custom-pattern}

Grok 本质是预定义一些正则表达式来进行文本匹配提取，并且给预定义的正则表达式进行命名，方便使用与嵌套引用扩展出无数个新模式。比如 DataKit 有 3 个如下内置模式：

```python
# pattern_name pattern 
_second (?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)    # 匹配秒数，_second 为模式名
_minute (?:[0-5][0-9])                            # 匹配分钟数，_minute 为模式名
_hour (?:2[0123]|[01]?[0-9])                      # 匹配年份，_hour 为模式名
```

基于上面三个内置模式，可以扩展出自己内置模式且命名为 `time`:

```python
# 把 time 加到 pattern 目录下文件中，此模式为全局模式，任何地方都能引用 time
# 如：time ([^0-9]?)%{hour:hour}:%{minute:minute}(?::%{second:second})([^0-9]?)

# 也可以通过 add_pattern() 添加到 pipeline 文件中，则此模式变为局部模式，只有当前 pipeline 脚本能使用 time
add_pattern("time", "(?:[^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})(?:[^0-9]?)")

# 通过 grok 提取原始输入中的时间字段。假定输入为 12:30:59，则提取到 {"hour": 12, "minute": 30, "second": 59}
grok(_, "%{time}")
```

<!-- markdownlint-disable MD046 -->
???+ note

    - 如果出现同名模式，则以局部模式优先（即局部模式覆盖全局模式）
    - Pipeline 脚本中，[add_pattern()](pipeline-built-in-function.md#fn-add-pattern) 需在 [grok()](pipeline-built-in-function.md#fn-grok) 函数前面调用，否则会导致第一条数据提取失败
<!-- markdownlint-enable -->

### Grok Fast Path 优化 {#grok-fast-path}

Pipeline 会在编译 Grok 模式时自动判断是否可以使用 Fast Path。Fast Path 主要面向“固定文本 + 明确字段 + 明确分隔符”的结构化日志模式；如果模式不满足优化条件，会自动回退到标准正则匹配路径，不影响 Grok 语义兼容性。

通常以下写法更容易命中 Fast Path：

```python
grok(_, "%{TIMESTAMP_ISO8601:time} %{LOGLEVEL:level} service=%{NOTSPACE:service} msg=\"%{GREEDYDATA:msg}\"")
```

```python
grok(_, "%{IPORHOST:client} - - \\[%{HTTPDATE:time}\\] \"%{WORD:method} %{URIPATHPARAM:path} HTTP/%{NUMBER:http_version}\" %{INT:status} %{INT:bytes}")
```

这类模式由固定字面量、内置 Grok 基础字段和清晰边界组成。当前常见可直接优化或展开后优化的基础字段包括：

```text
WORD, NOTSPACE,
INT, POSINT, NONNEGINT, NUMBER, BASE10NUM,
MONTH, MONTHNUM, MONTHNUM2, MONTHDAY, DAY, YEAR,
HOUR, MINUTE, SECOND, TIME,
HTTPDATE, TIMESTAMP_ISO8601, LOGLEVEL,
IPORHOST, HOST, HOSTNAME,
URIPATH, URIPATHPARAM,
DATA, GREEDYDATA, GREEDYLINES,
SPACE, QS, QUOTEDSTRING
```

部分内置模式（如 `IP`、`USER`、`USERNAME`、`PATH`）会先展开为更具体的正则表达式，只有展开后的表达式满足 Fast Path 支持范围时才会被优化；否则会自动回退到正则路径。

自定义模式也可以命中 Fast Path，只要展开后仍然主要由上述基础字段和固定字面量组成。例如：

```python
add_pattern("APPDATE", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")
grok(_, "%{APPDATE:time} \\[%{LOGLEVEL:level}\\] %{GREEDYDATA:msg}")
```

为了提高命中 Fast Path 的概率，建议：

- 字段之间保留稳定的分隔符，例如空格、`=`、`,`、`[]`、`""` 等
- 优先使用内置 Grok 基础字段组合，而不是大量直接书写复杂正则表达式
- 将宽泛字段如 `%{DATA}`、`%{GREEDYDATA}` 放在有明确前后边界的位置
- 避免多个无边界的宽泛字段连续出现

例如下面的模式不容易命中 Fast Path：

```python
grok(_, "%{DATA:a}%{DATA:b}%{GREEDYDATA:c}")
```

该模式中的 `DATA` 和 `GREEDYDATA` 都可以匹配任意内容，且字段之间没有分隔符。对于同一段文本，字段 `a`、`b`、`c` 的边界只能依赖正则引擎的非贪婪、贪婪和回溯规则来决定，无法通过确定性扫描稳定切分。可以改写为带边界的形式：

```python
grok(_, "a=%{DATA:a} b=%{DATA:b} msg=%{GREEDYDATA:c}")
```

或：

```python
grok(_, "%{NOTSPACE:a} %{NOTSPACE:b} %{GREEDYDATA:c}")
```

以下是在 `linux/amd64`、AMD Ryzen 7 9700X 环境下实际运行 benchmark 得到的结果，按日志场景排序。表格中的 `ns/op` 取 3 次运行的中位数。结果仅用于展示不同模式的性能量级，实际收益会受日志内容、Pattern 写法和运行环境影响：

当前文档 benchmark 覆盖的 51 个场景中，50 个使用 Fast Path，1 个回退到正则路径（`Solr request log`）。回退场景仍保持 Grok 语义一致，只是性能接近正则路径。

| 场景 | Grok 模式样例 | Fast Path | 正则路径 | 提升 |
| --- | --- | ---: | ---: | ---: |
| Web / 访问日志 | Apache combined log | 773.4 ns/op | 73,924 ns/op | 约 96x |
| Web / 访问日志 | Apache access log | 284.5 ns/op | 3,176 ns/op | 约 11x |
| Web / 访问日志 | Nginx access log | 425.4 ns/op | 64,414 ns/op | 约 151x |
| Web / 访问日志 | Nginx error log | 384.8 ns/op | 8,116 ns/op | 约 21x |
| Web / 访问日志 | Gateway access log | 270.1 ns/op | 25,421 ns/op | 约 94x |
| Web / 访问日志 | Tomcat access log | 315.5 ns/op | 1,720 ns/op | 约 5.5x |
| Web / 访问日志 | Python Gunicorn access log | 479.1 ns/op | 41,150 ns/op | 约 86x |
| 应用日志 | Go logfmt service log | 355.7 ns/op | 2,752 ns/op | 约 7.7x |
| 应用日志 | Go Gin access log | 364.9 ns/op | 4,651 ns/op | 约 13x |
| 应用日志 | Consul log | 837.6 ns/op | 16,830 ns/op | 约 20x |
| 应用日志 | Jenkins log | 134.0 ns/op | 2,892 ns/op | 约 22x |
| 数据库 | PostgreSQL duration log | 305.9 ns/op | 1,180 ns/op | 约 3.9x |
| 数据库 | PostgreSQL log | 1,140 ns/op | 9,715 ns/op | 约 8.5x |
| 数据库 | MySQL log | 142.3 ns/op | 1,682 ns/op | 约 12x |
| 数据库 | MySQL slow log | 1,313 ns/op | 4,012 ns/op | 约 3.1x |
| 数据库 | SQLServer log | 122.9 ns/op | 1,527 ns/op | 约 12x |
| 中间件 | Kafka server log | 432.6 ns/op | 2,447 ns/op | 约 5.7x |
| 中间件 | RabbitMQ default log | 113.9 ns/op | 1,670 ns/op | 约 15x |
| 中间件 | Redis log | 159.6 ns/op | 1,055 ns/op | 约 6.6x |
| 搜索 / 分析 | Elasticsearch log | 419.8 ns/op | 10,587 ns/op | 约 25x |
| 搜索 / 分析 | Elasticsearch search slow log | 180.5 ns/op | 8,194 ns/op | 约 45x |
| 搜索 / 分析 | Elasticsearch index slow log | 189.1 ns/op | 10,109 ns/op | 约 54x |
| 搜索 / 分析 | Solr request log | 2,851 ns/op | 2,853 ns/op | 约 1.00x |
| 搜索 / 分析 | Solr log | 180.7 ns/op | 1,643 ns/op | 约 9.1x |
| 搜索 / 分析 | TDengine log | 280.4 ns/op | 5,137 ns/op | 约 18x |
| 自定义 Pattern | Custom alias Nginx error log | 391.8 ns/op | 22,605 ns/op | 约 58x |
| 自定义 Pattern | Custom alias Postfix log | 199.4 ns/op | 10,288 ns/op | 约 52x |

上述 benchmark 使用的主要 Pattern 示例：

#### Web / 访问日志 {#benchmark-pattern-web-access}

- Apache combined log: `%{COMBINEDAPACHELOG}`
- Apache access log: `%{GREEDYDATA:ip_or_host} - - \[%{HTTPDATE:time}\] "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{NUMBER:http_code}`
- Nginx access log: `%{IPORHOST:remote_addr} - %{DATA:remote_user} \[%{HTTPDATE:time_local}\] "%{WORD:method} %{URIPATHPARAM:request} HTTP/%{NUMBER:http_version}" %{INT:status} %{INT:body_bytes_sent} "%{DATA:http_referer}" "%{DATA:http_user_agent}"`
- Nginx error log: `%{date2:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}", (upstream: "%{GREEDYDATA:upstream}", )?host: "%{NOTSPACE:ip_or_host}"`
- Gateway access log: `%{IPORHOST:client} %{WORD:method} %{URIPATHPARAM:path} status=%{INT:status} bytes=%{INT:bytes} duration=%{NUMBER:duration}`
- Tomcat access log: `%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \[%{HTTPDATE:time}\] "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{INT:status_code} %{INT:bytes}`
- Python Gunicorn access log: `%{IPORHOST:client} - - \[%{HTTPDATE:time}\] "%{WORD:method} %{URIPATHPARAM:path} HTTP/%{NUMBER:http_version}" %{INT:status} %{INT:bytes} "%{GREEDYDATA:referrer}" "%{GREEDYDATA:agent}"`

#### 应用日志 {#benchmark-pattern-application}

- Go logfmt service log: `time=%{TIMESTAMP_ISO8601:time} level=%{LOGLEVEL:level} logger=%{NOTSPACE:logger} msg="%{GREEDYDATA:msg}" err="%{GREEDYDATA:error}"`
- Go Gin access log: `\[%{GINTIME:time}\] %{INT:status} %{WORD:method} %{URIPATHPARAM:path} %{IPORHOST:client} %{NUMBER:latency}ms`
- Consul log: `%{SYSLOGTIMESTAMP}%{SPACE}%{SYSLOGHOST}%{SPACE}consul\[%{POSINT}\]:%{SPACE}%{_clog_date:date}%{SPACE}\[%{_clog_level:level}\]%{SPACE}%{_clog_character:character}:%{SPACE}%{_clog_message:msg}`
- Jenkins log: `%{TIMESTAMP_ISO8601:time} \[id=%{GREEDYDATA:id}\]\t%{GREEDYDATA:status}\t`

#### 数据库 {#benchmark-pattern-database}

- PostgreSQL duration log: `%{TIMESTAMP_ISO8601:time} \[%{POSINT:pid}\] %{USER:user}@%{WORD:database} %{WORD:severity}:  duration: %{NUMBER:duration_ms} ms  statement: %{GREEDYDATA:statement}`
- PostgreSQL log: `%{log_date:time}%{SPACE}\[%{INT:process_id}\]%{SPACE}(%{WORD:db_name}?%{SPACE}%{application_name}%{SPACE}%{USER:user}?%{SPACE}%{remote_host}%{SPACE})?%{session_id:session_id}%{SPACE}(%{status:status}:)?`
- MySQL log: `%{TIMESTAMP_ISO8601:time}\s+%{INT:thread_id}\s+%{WORD:operation}\s+%{GREEDYDATA:raw_query}`
- MySQL slow log: `%{timeline}\n%{userline}\n%{kvline01}(\n)?(%{kvline02})?(\n)?(%{kvline03})?\n%{sqls:db_slow_statement}`
- SQLServer log: `%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\s+%{GREEDYDATA:msg}`

#### 中间件 {#benchmark-pattern-middleware}

- Kafka server log: `^\[%{date1:time}\] %{WORD:status} %{DATA:msg} \(%{DATA:name}\)`
- RabbitMQ default log: `%{DATA:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}`
- Redis log: `%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}`

#### 搜索 / 分析 {#benchmark-pattern-search-analytics}

- Elasticsearch log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[%{NOTSPACE:name}%{SPACE}\]%{SPACE}(\[%{HOSTNAME:nodeId}\])?.*`
- Elasticsearch search slow log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[i.s.s.(?:query|fetch)%{SPACE}\] (?:\[%{HOSTNAME:nodeId}\] )?\[%{NOTSPACE:index}\]\[%{INT}\] took\[.*\], took_millis\[%{INT:duration}\].*`
- Elasticsearch index slow log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[i.i.s.index%{SPACE}\] (?:\[%{HOSTNAME:nodeId}\] )?\[%{NOTSPACE:index}/%{NOTSPACE}\] took\[.*\], took_millis\[%{INT:duration}\].*`
- Solr request log: `%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\(%{NOTSPACE:thread}\)%{SPACE}\[%{SPACE}%{NOTSPACE}?\]%{SPACE}%{solrReporter:reporter}%{SPACE}\[%{NOTSPACE:core}\]%{SPACE}webapp=%{NOTSPACE:webapp}%{SPACE}path=%{solrPath:path}%{SPACE}params=\{%{solrParams:params}\}(?:%{SPACE}hits=%{NUMBER:hits})?%{SPACE}status=%{NUMBER:qstatus}%{SPACE}QTime=%{NUMBER:qtime}`
- Solr log: `%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\(%{NOTSPACE:thread}\)%{SPACE}\[%{SPACE}%{NOTSPACE}?\]%{SPACE}%{solrReporter:reporter}.*`
- TDengine log: `%{GREEDYDATA:temp}%{SPACE}TAOS_%{NOTSPACE:module}%{SPACE}%{NOTSPACE:level}%{SPACE}%{GREEDYDATA:http_url}`

#### 自定义 Pattern {#benchmark-pattern-custom}

- Custom alias Nginx error log: `%{APPDATE:time} \[%{LOGLEVEL:level}\] %{GREEDYDATA:msg}, client: %{IPORHOST:client}, server: %{NOTSPACE:server}, request: "%{WORD:method} %{GREEDYDATA:path} HTTP/%{NUMBER:http_version}", %{UPBLOCK}host: "%{NOTSPACE:host}"`
- Custom alias Postfix log: `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:host} %{WORD:program}\[%{POSINT:pid}\]: %{QUEUEID:queue_id}: %{GREEDYDATA:msg}`

### 内置的 Pattern 列表 {#built-in-patterns}

DataKit 内置了一些常用的 Pattern，我们在构造 Grok 模式的时候，可以直接使用：

``` not-set
USERNAME             : [a-zA-Z0-9._-]+
USER                 : %{USERNAME}
EMAILLOCALPART       : [a-zA-Z][a-zA-Z0-9_.+-=:]+
EMAILADDRESS         : %{EMAILLOCALPART}@%{HOSTNAME}
HTTPDUSER            : %{EMAILADDRESS}|%{USER}
INT                  : (?:[+-]?(?:[0-9]+))
BASE10NUM            : (?:[+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)
NUMBER               : (?:%{BASE10NUM})
BASE16NUM            : (?:0[xX]?[0-9a-fA-F]+)
POSINT               : \b(?:[1-9][0-9]*)\b
NONNEGINT            : \b(?:[0-9]+)\b
WORD                 : \b\w+\b
NOTSPACE             : \S+
SPACE                : \s*
DATA                 : .*?
GREEDYDATA           : .*
GREEDYLINES          : (?s).*
QUOTEDSTRING         : "(?:[^"\\]*(?:\\.[^"\\]*)*)"|\'(?:[^\'\\]*(?:\\.[^\'\\]*)*)\'
UUID                 : [A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12}
MAC                  : (?:%{CISCOMAC}|%{WINDOWSMAC}|%{COMMONMAC})
CISCOMAC             : (?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})
WINDOWSMAC           : (?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})
COMMONMAC            : (?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})
IPV6                 : (?:(?:(?:[0-9A-Fa-f]{1,4}:){7}(?:[0-9A-Fa-f]{1,4}|:))|(?:(?:[0-9A-Fa-f]{1,4}:){6}(?::[0-9A-Fa-f]{1,4}|(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){5}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,2})|:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){4}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,3})|(?:(?::[0-9A-Fa-f]{1,4})?:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){3}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,4})|(?:(?::[0-9A-Fa-f]{1,4}){0,2}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){2}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,5})|(?:(?::[0-9A-Fa-f]{1,4}){0,3}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){1}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,6})|(?:(?::[0-9A-Fa-f]{1,4}){0,4}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?::(?:(?:(?::[0-9A-Fa-f]{1,4}){1,7})|(?:(?::[0-9A-Fa-f]{1,4}){0,5}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(?:%.+)?
IPV4                 : (?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)
IP                   : (?:%{IPV6}|%{IPV4})
HOSTNAME             : \b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(?:\.?|\b)
HOST                 : %{HOSTNAME}
IPORHOST             : (?:%{IP}|%{HOSTNAME})
HOSTPORT             : %{IPORHOST}:%{POSINT}
PATH                 : (?:%{UNIXPATH}|%{WINPATH})
UNIXPATH             : (?:/[\w_%!$@:.,-]?/?)(?:\S+)?
TTY                  : (?:/dev/(?:pts|tty(?:[pq])?)(?:\w+)?/?(?:[0-9]+))
WINPATH              : (?:[A-Za-z]:|\\)(?:\\[^\\?*]*)+
URIPROTO             : [A-Za-z]+(?:\+[A-Za-z+]+)?
URIHOST              : %{IPORHOST}(?::%{POSINT:port})?
URIPATH              : (?:/[A-Za-z0-9$.+!*'(){},~:;=@#%_\-]*)+
URIPARAM             : \?[A-Za-z0-9$.+!*'|(){},~@#%&/=:;_?\-\[\]<>]*
URIPATHPARAM         : %{URIPATH}(?:%{URIPARAM})?
URI                  : %{URIPROTO}://(?:%{USER}(?::[^@]*)?@)?(?:%{URIHOST})?(?:%{URIPATHPARAM})?
MONTH                : \b(?:Jan(?:uary|uar)?|Feb(?:ruary|ruar)?|M(?:a|ä)?r(?:ch|z)?|Apr(?:il)?|Ma(?:y|i)?|Jun(?:e|i)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|O(?:c|k)?t(?:ober)?|Nov(?:ember)?|De(?:c|z)(?:ember)?)\b
MONTHNUM             : (?:0?[1-9]|1[0-2])
MONTHNUM2            : (?:0[1-9]|1[0-2])
MONTHDAY             : (?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])
DAY                  : (?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)
YEAR                 : (\d\d){1,2}
HOUR                 : (?:2[0123]|[01]?[0-9])
MINUTE               : (?:[0-5][0-9])
SECOND               : (?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)
TIME                 : (?:[^0-9]?)%{HOUR}:%{MINUTE}(?::%{SECOND})(?:[^0-9]?)
DATE_US              : %{MONTHNUM}[/-]%{MONTHDAY}[/-]%{YEAR}
DATE_EU              : %{MONTHDAY}[./-]%{MONTHNUM}[./-]%{YEAR}
ISO8601_TIMEZONE     : (?:Z|[+-]%{HOUR}(?::?%{MINUTE}))
ISO8601_SECOND       : (?:%{SECOND}|60)
TIMESTAMP_ISO8601    : %{YEAR}-%{MONTHNUM}-%{MONTHDAY}[T ]%{HOUR}:?%{MINUTE}(?::?%{SECOND})?%{ISO8601_TIMEZONE}?
DATE                 : %{DATE_US}|%{DATE_EU}
DATESTAMP            : %{DATE}[- ]%{TIME}
TZ                   : (?:[PMCE][SD]T|UTC)
DATESTAMP_RFC822     : %{DAY} %{MONTH} %{MONTHDAY} %{YEAR} %{TIME} %{TZ}
DATESTAMP_RFC2822    : %{DAY}, %{MONTHDAY} %{MONTH} %{YEAR} %{TIME} %{ISO8601_TIMEZONE}
DATESTAMP_OTHER      : %{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{TZ} %{YEAR}
DATESTAMP_EVENTLOG   : %{YEAR}%{MONTHNUM2}%{MONTHDAY}%{HOUR}%{MINUTE}%{SECOND}
HTTPDERROR_DATE      : %{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{YEAR}
SYSLOGTIMESTAMP      : %{MONTH} +%{MONTHDAY} %{TIME}
PROG                 : [\x21-\x5a\x5c\x5e-\x7e]+
SYSLOGPROG           : %{PROG:program}(?:\[%{POSINT:pid}\])?
SYSLOGHOST           : %{IPORHOST}
SYSLOGFACILITY       : <%{NONNEGINT:facility}.%{NONNEGINT:priority}>
HTTPDATE             : %{MONTHDAY}/%{MONTH}/%{YEAR}:%{TIME} %{INT}
QS                   : %{QUOTEDSTRING}
SYSLOGBASE           : %{SYSLOGTIMESTAMP:timestamp} (?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource} %{SYSLOGPROG}:
COMMONAPACHELOG      : %{IPORHOST:clientip} %{HTTPDUSER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)
COMBINEDAPACHELOG    : %{COMMONAPACHELOG} %{QS:referrer} %{QS:agent}
HTTPD20_ERRORLOG     : \[%{HTTPDERROR_DATE:timestamp}\] \[%{LOGLEVEL:loglevel}\] (?:\[client %{IPORHOST:clientip}\] ){0,1}%{GREEDYDATA:errormsg}
HTTPD24_ERRORLOG     : \[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel}\] \[pid %{POSINT:pid}:tid %{NUMBER:tid}\]( \(%{POSINT:proxy_errorcode}\)%{DATA:proxy_errormessage}:)?( \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}
HTTPD_ERRORLOG       : %{HTTPD20_ERRORLOG}|%{HTTPD24_ERRORLOG}
LOGLEVEL             : (?:[Aa]lert|ALERT|[Tt]race|TRACE|[Dd]ebug|DEBUG|[Nn]otice|NOTICE|[Ii]nfo|INFO|[Ww]arn?(?:ing)?|WARN?(?:ING)?|[Ee]rr?(?:or)?|ERR?(?:OR)?|[Cc]rit?(?:ical)?|CRIT?(?:ICAL)?|[Ff]atal|FATAL|[Ss]evere|SEVERE|EMERG(?:ENCY)?|[Ee]merg(?:ency)?)
COMMONENVOYACCESSLOG : \[%{TIMESTAMP_ISO8601:timestamp}\] \"%{DATA:method} (?:%{URIPATH:uri_path}(?:%{URIPARAM:uri_param})?|%{DATA:}) %{DATA:protocol}\" %{NUMBER:status_code} %{DATA:response_flags} %{NUMBER:bytes_received} %{NUMBER:bytes_sent} %{NUMBER:duration} (?:%{NUMBER:upstream_service_time}|%{DATA:tcp_service_time}) \"%{DATA:forwarded_for}\" \"%{DATA:user_agent}\" \"%{DATA:request_id}\" \"%{DATA:authority}\" \"%{DATA:upstream_service}\"
```
