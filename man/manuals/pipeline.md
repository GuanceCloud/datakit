{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Pipeline 使用文档

以下是文本处理器定义。随着不同语法的逐步支持，该文档会做不同程度的调整和增删。

## 基本规则：

- 函数名大小写不敏感
- 以 `#` 为行注释字符。不支持行内注释
- 标识符：只能出现 `[_a-zA-Z0-9]` 这些字符，且首字符不能是数字。如 `_abc, _abc123, _123ab`
- 字符串值可用双引号和单引号： `"this is a string"` 和 `'this is a string'` 是等价的
- 数据类型：支持浮点（`123.4`, `5.67E3`）、整形（`123`, `-1`）、字符串（`'张三'`, `"hello world"`）、Boolean（`true`, `false`）四种类型
- 多个函数之间，可以用空白字符（空格、换行、Tab 等）分割
- 切割出来的字段中如果带有特殊字符（如 `@`、`$`、中文字符、表情包等），在代码中引用时，需额外修饰，如 `` `@some-variable` ``，`` `这是一个表情包变量👍` ``
	- 字段名中出现的字符只能是 `[_a-zA-Z0-9]`，即只能是下划线（`_`）、大小写英文字母以及数字。另外，**首字符不能是数字**

## 快速开始

### 如何在 DataKit 中配置 pipeline：

- 第一步：编写如下 pipeline 文件，假定名为 *nginx.p*。将其存放在 `<datakit安装目录>/pipeline` 目录下。

```python
# 假定输入是一个 Nginx 日志（以下字段都是 yy 的...）
# 注意，脚本是可以加注释的

grok(_, "some-grok-patterns")  # 对输入的文本，进行 grok 提取
rename('client_ip', ip)        # 将 ip 字段改名成 client_ip
rename("网络协议", protocol)   # 将 protocol 字段改名成 `网络协议`

# 将时间戳(如 1610967131)换成 RFC3339 日期格式：2006-01-02T15:04:05Z07:00
datetime(access_time, "s", "RFC3339")

url_decode(request_url)      # 将 HTTP 请求路由翻译成明文

# 当 status_code 介于 200 ~ 300 之间，新建一个 http_status = "HTTP_OK" 的字段
group_between(status_code, [200, 300], "HTTP_OK", "http_status")

# 丢弃原内容
drop_origin_data()
```

> 注意，切割过程中，需避免[可能出现的跟 tag key 重名的问题](datakit-pl-how-to#5cf855c0)

- 第二步：配置对应的采集器来使用上面的 pipeline

以 logging 采集器为例，配置字段 `pipeline_path` 即可，注意，这里配置的是 pipeline 的脚本名称，而不是路径。所有这里引用的 pipeline 脚本，必须存放在 `<DataKit 安装目录/pipeline>` 目录下：

```python
[[inputs.logging]]
    logfiles = ["/path/to/nginx/log"]

    # required
    source = "nginx"

    # 所有脚本必须放在 /path/to/datakit/pipeline 目录下
    # 如果开启了 gitrepos 功能，则优先以 gitrepos 中的同名文件为准
    # 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名
    # 的脚本（如 nginx -> nginx.p），作为其默认 pipeline 配置
    pipeline = "nginx.p"

    ... # 其它配置
```

重启采集器，即可切割对应的日志。关于 Pipeline 编写、调试以及注意事项，参见[这里](datakit-pl-how-to)。

## Grok 模式分类

DataKit 中 grok 模式可以分为两类：全局模式与局部模式，`pattern` 目录下的模式文件都是全局模式，所有 pipeline 脚本都可使用，而在 pipeline 脚本中通过 `add_pattern()` 函数新增的模式属于局部模式，只针对当前 pipeline 脚本有效。

当 DataKit 内置模式不能满足所有用户需求，用户可以自行在 pipeline 目录中增加模式文件来扩充。若自定义模式是全局级别，则需在 `pattern` 目录中新建一个文件并把模式添加进去，不要在已有内置模式文件中添加或修改，因为datakit启动过程会把内置模式文件覆盖掉。

### 添加局部模式

grok 本质是预定义一些正则表达式来进行文本匹配提取，并且给预定义的正则表达式进行命名，方便使用与嵌套引用扩展出无数个新模式。比如 DataKit 有 3 个如下内置模式：

```python
_second (?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)    #匹配秒数，_second为模式名
_minute (?:[0-5][0-9])                            #匹配分钟数，_minute为模式名
_hour (?:2[0123]|[01]?[0-9])                      #匹配年份，_hour为模式名
```

基于上面三个内置模式，可以扩展出自己内置模式且命名为 `time`:

```python
# 把 time 加到 pattern 目录下文件中，此模式为全局模式，任何地方都能引用 time
time ([^0-9]?)%{hour:hour}:%{minute:minute}(?::%{second:second})([^0-9]?)

# 也可以通过 add_pattern() 添加到 pipeline 文件中，则此模式变为局部模式，只有当前 pipeline 脚本能使用 time
add_pattern(time, "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")

# 通过 grok 提取原始输入中的时间字段。假定输入为 12:30:59，则提取到 {"hour": 12, "minute": 30, "second": 59}
grok(_, %{time})
```

注意：

- 相同模式名以脚本级优先（即局部模式覆盖全局模式）
- pipeline 脚本中，`add_pattern()` 需在 `grok()` 函数前面调用，否则会导致第一条数据提取失败。

### 内置的 Pattern 列表

DataKit 内置了一些常用的 Pattern，我们在使用 Grok 切割的时候，可以直接使用：

```
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

## 脚本执行流

pipeline 支持 `if/elif/else` 语法，`if` 后面的语句仅支持条件表达式，即 `<`、`<=`、`==`、`>`、`>=` 和 `!=`， 且支持小括号优先级和多个条件表达式的 `AND` 和 `OR` 连接。
表达式两边可以是已存在的 key 或固定值（数值、布尔值、字符串和 nil ），例如：

```python
# 数值比较
add_key(score, 95)

if score == 100  {
	add_key(level, "S")
} elif score >= 90 && score < 100 {
	add_key(level, "A")
} elif score >= 60 {
	add_key(level, "C")
} else {
	add_key(level, "D")
}

# 字符串比较
add_key(name, "张三")

if name == "法外狂徒" {
	# 这是不可能的，不要污蔑我
}
```

和大多数编程/脚本语言相同，根据 `if/elif` 的条件是否成立，来决定其执行顺序。

注意：如果是进行数值比较，需要先用 `cast()` 进行类型转换，比如：

```
# status_code 是 grok 切出来的 string 类型
cast(status_code, "int")

if status == 200 {
	add_key(level, "OK")
} elif status >= 400 && status < 500 {
	add_key(level, "ERROR")
} elif stauts > 500 {
	add_key(level, "FATAL")
}

```

## Pipeline 脚本存放目录

### 内置的 pipeline 目录
### Git 管理的 pipeline 目录
### Remote Pipeline 目录

## 脚本函数

函数参数说明：

- 函数参数中，匿名参数（`_`）指原始的输入文本数据
- json 路径，直接表示成 `x.y.z` 这种形式，无需其它修饰。例如 `{"a":{"first":2.3, "second":2, "third":"abc", "forth":true}, "age":47}`，json 路径为 `a.thrid` 表示待操作数据为 `abc`
- 所有函数参数的相对顺序，都是固定的，引擎会对其做具体检查
- 以下提到的所有 `key` 参数，都指已经过初次提取（通过 `grok()` 或 `json()`）之后，生成的 `key`
- 待处理json的路径，支持标识符的写法，不能使用字符串，如果是生成新key，需要使用字符串

{{.PipelineFuncs}}
