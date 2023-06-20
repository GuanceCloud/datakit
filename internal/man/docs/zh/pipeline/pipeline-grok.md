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

对于模式写法 `%{pattern_name:key_name}`，其等价与正则表达式中的命名捕获组：  

```regexp
(?P<pattern_name>pattern)
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
add_pattern(time, "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")

# 通过 grok 提取原始输入中的时间字段。假定输入为 12:30:59，则提取到 {"hour": 12, "minute": 30, "second": 59}
grok(_, "%{time}")
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 如果出现同名模式，则以局部模式优先（即局部模式覆盖全局模式）
    - Pipeline 脚本中，[add_pattern()](pipeline-built-in-function.md#fn-add-pattern) 需在 [grok()](pipeline-built-in-function.md#fn-grok) 函数前面调用，否则会导致第一条数据提取失败
<!-- markdownlint-enable -->

### 内置的 Pattern 列表 {#built-in-patterns}

DataKit 内置了一些常用的 Pattern，我们在构造 Grok 模式的时候，可以直接使用：

<!-- markdownlint-disable MD046 -->
???- "内置 Patterns"

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
<!-- markdownlint-enable -->
