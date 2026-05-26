# Grok Pattern
---

## Grok Pattern Introduction {#grok-pattern}

DataKit Pipeline provides [grok()](pipeline-built-in-function.md#fn-grok) function to implement support for executing Grok patterns (in terms of implementation, the grok() function will translate Grok patterns into regular expressions), And provide [add_pattern()](pipeline-built-in-function.md#fn-add-pattern) function to add custom naming patterns.

The Grok pattern is based on regular expressions. After the pattern is named, it can be used in other patterns in the following three ways. Be careful not to use circular references:

- `%{pattern_name}`
- `%{pattern_name:key_name}`
- `%{pattern_name:key_name:type}`

The value of `type` can be in the range of {`float`, `int`, `str`, `bool`}; more complex Grok patterns can be obtained by combining Grok patterns.

Any regular expression can be regarded as a legal Grok pattern, and supports the mixed use of named Grok patterns and regular expressions to write Grok patterns;

For the pattern notation `%{pattern_name:key_name}`, it is equivalent to the named capture group in the regular expression:

```regexp
(?P<key_name>pattern)
```

## Grok Pattern Classification in DataKit {#grok-pattern-class}

Grok patterns in DataKit can be divided into two categories:

- Global pattern: The pattern files in the *pattern* directory are all global patterns, which can be used by all Pipeline scripts
- Partial pattern: The pattern added by the [add_pattern()](pipeline-built-in-function.md#fn-add-pattern) function in the Pipeline script is a partial pattern, which is only valid for the current Pipeline script

The following takes Nginx access-log as an example to explain how to write the corresponding Grok mode. The original nginx access log is as follows:

```log
127.0.0.1 - - [26/May/2022:20:53:52 +0800] "GET /server_status HTTP/1.1" 404 134 "-" "Go-http-client/1.1"
```

Assuming that we need to obtain client_ip, time (request), http_method, http_url, http_version, status_code from the access log, the grok mode can initially be written as:

```python
grok(_,"%{NOTSPACE:client_ip} %{NOTSPACE} %{NOTSPACE} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT} \"%{NOTSPACE}\" \"%{NOTSPACE}\"")

cast(status_code, "int")
group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)
default_time(time)
```

Optimize it again and extract the corresponding features respectively:

```python
# client_ip, http_ident, http_auth at the header of the log as a pattern
add_pattern("p1", "%{NOTSPACE:client_ip} %{NOTSPACE} %{NOTSPACE}")

# The middle http_method, http_url, http_version, status_code as a pattern,
# And specify the data type int of status_code in the pattern instead of the cast function used
add_pattern("p3", '"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{INT:status_code:int}')

grok(_, "%{p1} \\[%{HTTPDATE:time}\\] %{p3} %{INT} \"%{NOTSPACE}\" \"%{NOTSPACE}\"")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warning", status)
group_between(status_code, [500,599], "error", status)

default_time(time)
```


The optimized cutting is more readable than the preliminary single-line pattern. Since the default data type of the field resolved by grok is string, specifying the data type of the field here avoids the subsequent use of the [cast()](pipeline-built-in-function.md#fn-cast) function for type conversion.

### Custom Grok Pattern {#custom-pattern}

The essence of grok is to predefine some regular expressions for text matching extraction, and name the predefined regular expressions, which is convenient to use and expand countless new patterns with nested references. For example, DataKit has three built-in modes as follows:

```python
_second (?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)    #Number of seconds to match, _ second is the schema name
_minute (?:[0-5][0-9])                            #Match minutes, _ minute is the schema name
_hour (?:2[0123]|[01]?[0-9])                      #Match year, _ hour is the schema name
```

Based on the above three built-in patterns, you can extend your own built-in pattern and name it `time`:

```python
# Add time to the file under the pattern directory. This mode is global mode, and time can be referenced anywhere.
time ([^0-9]?)%{hour:hour}:%{minute:minute}(?::%{second:second})([^0-9]?)

# It can also be added to the pipeline file via add_pattern (), then the mode becomes local and only the current pipeline script can use time.
add_pattern("time", "(?:[^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})(?:[^0-9]?)")

# Extract the time field from the original input through grok. Assuming the input is 12:30:59, the {"hour": 12, "minute": 30, "second": 59}
grok(_, "%{time}")
```
<!-- markdownlint-disable MD046 -->
???+ note

    - If a pattern with the same name occurs, the local pattern takes precedence (that is, the local pattern overrides the global pattern).
    - In the Pipeline script, [add_pattern()](#fn-add-pattern) needs to be called before the [grok()](#fn-grok) function, otherwise the first data fetch will fail.
<!-- markdownlint-enable -->

### Grok Fast Path Optimization {#grok-fast-path}

Pipeline automatically checks whether a Grok pattern can use the Fast Path when the pattern is compiled. Fast Path is designed for structured log patterns that look like "fixed text + explicit fields + stable delimiters". If a pattern is not suitable for this optimization, Pipeline automatically falls back to the standard regular expression path, so Grok semantic compatibility is preserved.

The following patterns are usually good Fast Path candidates:

```python
grok(_, "%{TIMESTAMP_ISO8601:time} %{LOGLEVEL:level} service=%{NOTSPACE:service} msg=\"%{GREEDYDATA:msg}\"")
```

```python
grok(_, "%{IPORHOST:client} - - \\[%{HTTPDATE:time}\\] \"%{WORD:method} %{URIPATHPARAM:path} HTTP/%{NUMBER:http_version}\" %{INT:status} %{INT:bytes}")
```

These patterns are composed of fixed literals, built-in Grok primitives, and clear field boundaries. Common primitives that can be optimized directly, or after expansion, include:

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

Some built-in patterns, such as `IP`, `USER`, `USERNAME`, and `PATH`, are expanded into more specific regular expressions first. They use the Fast Path only when the expanded expression is within the supported subset; otherwise they automatically fall back to the regular expression path.

Custom patterns can also use the Fast Path, as long as they expand mainly into the primitives above and fixed literals. For example:

```python
add_pattern("APPDATE", "%{YEAR}[./]%{MONTHNUM}[./]%{MONTHDAY} %{TIME}")
grok(_, "%{APPDATE:time} \\[%{LOGLEVEL:level}\\] %{GREEDYDATA:msg}")
```

To improve the chance of using the Fast Path:

- Keep stable delimiters between fields, such as spaces, `=`, `,`, `[]`, and `""`
- Prefer built-in Grok primitives instead of writing many complex raw regular expressions directly
- Put broad fields such as `%{DATA}` and `%{GREEDYDATA}` in positions with clear boundaries
- Avoid placing multiple unbounded broad fields next to each other

For example, the following pattern is unlikely to use the Fast Path:

```python
grok(_, "%{DATA:a}%{DATA:b}%{GREEDYDATA:c}")
```

In this pattern, both `DATA` and `GREEDYDATA` can match arbitrary text, and there is no delimiter between fields. For the same input, the boundaries of `a`, `b`, and `c` depend on regular expression non-greedy, greedy, and backtracking behavior, so they cannot be split reliably with deterministic scanning. Prefer a bounded form:

```python
grok(_, "a=%{DATA:a} b=%{DATA:b} msg=%{GREEDYDATA:c}")
```

Or:

```python
grok(_, "%{NOTSPACE:a} %{NOTSPACE:b} %{GREEDYDATA:c}")
```

The following benchmark results were collected on `linux/amd64` with an AMD Ryzen 7 9700X, sorted by log scenario. Each `ns/op` value in the table is the median of 3 runs. They are intended to show the performance scale only; actual gains depend on the log content, pattern shape, and runtime environment:

Among the 51 benchmark scenarios covered here, 50 use the Fast Path and 1 falls back to the regular expression path (`Solr request log`). Fallback cases preserve Grok semantic compatibility, but their performance is close to the regular expression path.

| Category | Grok pattern example | Fast Path | Regexp path | Speedup |
| --- | --- | ---: | ---: | ---: |
| Web / access | Apache combined log | 773.4 ns/op | 73,924 ns/op | ~96x |
| Web / access | Apache access log | 284.5 ns/op | 3,176 ns/op | ~11x |
| Web / access | Nginx access log | 425.4 ns/op | 64,414 ns/op | ~151x |
| Web / access | Nginx error log | 384.8 ns/op | 8,116 ns/op | ~21x |
| Web / access | Gateway access log | 270.1 ns/op | 25,421 ns/op | ~94x |
| Web / access | Tomcat access log | 315.5 ns/op | 1,720 ns/op | ~5.5x |
| Web / access | Python Gunicorn access log | 479.1 ns/op | 41,150 ns/op | ~86x |
| Application | Go logfmt service log | 355.7 ns/op | 2,752 ns/op | ~7.7x |
| Application | Go Gin access log | 364.9 ns/op | 4,651 ns/op | ~13x |
| Application | Consul log | 837.6 ns/op | 16,830 ns/op | ~20x |
| Application | Jenkins log | 134.0 ns/op | 2,892 ns/op | ~22x |
| Database | PostgreSQL duration log | 305.9 ns/op | 1,180 ns/op | ~3.9x |
| Database | PostgreSQL log | 1,140 ns/op | 9,715 ns/op | ~8.5x |
| Database | MySQL log | 142.3 ns/op | 1,682 ns/op | ~12x |
| Database | MySQL slow log | 1,313 ns/op | 4,012 ns/op | ~3.1x |
| Database | SQLServer log | 122.9 ns/op | 1,527 ns/op | ~12x |
| Middleware | Kafka server log | 432.6 ns/op | 2,447 ns/op | ~5.7x |
| Middleware | RabbitMQ default log | 113.9 ns/op | 1,670 ns/op | ~15x |
| Middleware | Redis log | 159.6 ns/op | 1,055 ns/op | ~6.6x |
| Search / analytics | Elasticsearch log | 419.8 ns/op | 10,587 ns/op | ~25x |
| Search / analytics | Elasticsearch search slow log | 180.5 ns/op | 8,194 ns/op | ~45x |
| Search / analytics | Elasticsearch index slow log | 189.1 ns/op | 10,109 ns/op | ~54x |
| Search / analytics | Solr request log | 2,851 ns/op | 2,853 ns/op | ~1.00x |
| Search / analytics | Solr log | 180.7 ns/op | 1,643 ns/op | ~9.1x |
| Search / analytics | TDengine log | 280.4 ns/op | 5,137 ns/op | ~18x |
| Custom pattern | Custom alias Nginx error log | 391.8 ns/op | 22,605 ns/op | ~58x |
| Custom pattern | Custom alias Postfix log | 199.4 ns/op | 10,288 ns/op | ~52x |

Main pattern examples used by the benchmarks above:

#### Web / access {#benchmark-pattern-web-access}

- Apache combined log: `%{COMBINEDAPACHELOG}`
- Apache access log: `%{GREEDYDATA:ip_or_host} - - \[%{HTTPDATE:time}\] "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{NUMBER:http_code}`
- Nginx access log: `%{IPORHOST:remote_addr} - %{DATA:remote_user} \[%{HTTPDATE:time_local}\] "%{WORD:method} %{URIPATHPARAM:request} HTTP/%{NUMBER:http_version}" %{INT:status} %{INT:body_bytes_sent} "%{DATA:http_referer}" "%{DATA:http_user_agent}"`
- Nginx error log: `%{date2:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}, client: %{NOTSPACE:client_ip}, server: %{NOTSPACE:server}, request: "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}", (upstream: "%{GREEDYDATA:upstream}", )?host: "%{NOTSPACE:ip_or_host}"`
- Gateway access log: `%{IPORHOST:client} %{WORD:method} %{URIPATHPARAM:path} status=%{INT:status} bytes=%{INT:bytes} duration=%{NUMBER:duration}`
- Tomcat access log: `%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \[%{HTTPDATE:time}\] "%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}" %{INT:status_code} %{INT:bytes}`
- Python Gunicorn access log: `%{IPORHOST:client} - - \[%{HTTPDATE:time}\] "%{WORD:method} %{URIPATHPARAM:path} HTTP/%{NUMBER:http_version}" %{INT:status} %{INT:bytes} "%{GREEDYDATA:referrer}" "%{GREEDYDATA:agent}"`

#### Application {#benchmark-pattern-application}

- Go logfmt service log: `time=%{TIMESTAMP_ISO8601:time} level=%{LOGLEVEL:level} logger=%{NOTSPACE:logger} msg="%{GREEDYDATA:msg}" err="%{GREEDYDATA:error}"`
- Go Gin access log: `\[%{GINTIME:time}\] %{INT:status} %{WORD:method} %{URIPATHPARAM:path} %{IPORHOST:client} %{NUMBER:latency}ms`
- Consul log: `%{SYSLOGTIMESTAMP}%{SPACE}%{SYSLOGHOST}%{SPACE}consul\[%{POSINT}\]:%{SPACE}%{_clog_date:date}%{SPACE}\[%{_clog_level:level}\]%{SPACE}%{_clog_character:character}:%{SPACE}%{_clog_message:msg}`
- Jenkins log: `%{TIMESTAMP_ISO8601:time} \[id=%{GREEDYDATA:id}\]\t%{GREEDYDATA:status}\t`

#### Database {#benchmark-pattern-database}

- PostgreSQL duration log: `%{TIMESTAMP_ISO8601:time} \[%{POSINT:pid}\] %{USER:user}@%{WORD:database} %{WORD:severity}:  duration: %{NUMBER:duration_ms} ms  statement: %{GREEDYDATA:statement}`
- PostgreSQL log: `%{log_date:time}%{SPACE}\[%{INT:process_id}\]%{SPACE}(%{WORD:db_name}?%{SPACE}%{application_name}%{SPACE}%{USER:user}?%{SPACE}%{remote_host}%{SPACE})?%{session_id:session_id}%{SPACE}(%{status:status}:)?`
- MySQL log: `%{TIMESTAMP_ISO8601:time}\s+%{INT:thread_id}\s+%{WORD:operation}\s+%{GREEDYDATA:raw_query}`
- MySQL slow log: `%{timeline}\n%{userline}\n%{kvline01}(\n)?(%{kvline02})?(\n)?(%{kvline03})?\n%{sqls:db_slow_statement}`
- SQLServer log: `%{TIMESTAMP_ISO8601:time} %{NOTSPACE:origin}\s+%{GREEDYDATA:msg}`

#### Middleware {#benchmark-pattern-middleware}

- Kafka server log: `^\[%{date1:time}\] %{WORD:status} %{DATA:msg} \(%{DATA:name}\)`
- RabbitMQ default log: `%{DATA:time} \[%{LOGLEVEL:status}\] %{GREEDYDATA:msg}`
- Redis log: `%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}`

#### Search / analytics {#benchmark-pattern-search-analytics}

- Elasticsearch log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[%{NOTSPACE:name}%{SPACE}\]%{SPACE}(\[%{HOSTNAME:nodeId}\])?.*`
- Elasticsearch search slow log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[i.s.s.(?:query|fetch)%{SPACE}\] (?:\[%{HOSTNAME:nodeId}\] )?\[%{NOTSPACE:index}\]\[%{INT}\] took\[.*\], took_millis\[%{INT:duration}\].*`
- Elasticsearch index slow log: `^\[%{TIMESTAMP_ISO8601:time}\]\[%{LOGLEVEL:status}%{SPACE}\]\[i.i.s.index%{SPACE}\] (?:\[%{HOSTNAME:nodeId}\] )?\[%{NOTSPACE:index}/%{NOTSPACE}\] took\[.*\], took_millis\[%{INT:duration}\].*`
- Solr request log: `%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\(%{NOTSPACE:thread}\)%{SPACE}\[%{SPACE}%{NOTSPACE}?\]%{SPACE}%{solrReporter:reporter}%{SPACE}\[%{NOTSPACE:core}\]%{SPACE}webapp=%{NOTSPACE:webapp}%{SPACE}path=%{solrPath:path}%{SPACE}params=\{%{solrParams:params}\}(?:%{SPACE}hits=%{NUMBER:hits})?%{SPACE}status=%{NUMBER:qstatus}%{SPACE}QTime=%{NUMBER:qtime}`
- Solr log: `%{TIMESTAMP_ISO8601:time}%{SPACE}%{LOGLEVEL:status}%{SPACE}\(%{NOTSPACE:thread}\)%{SPACE}\[%{SPACE}%{NOTSPACE}?\]%{SPACE}%{solrReporter:reporter}.*`
- TDengine log: `%{GREEDYDATA:temp}%{SPACE}TAOS_%{NOTSPACE:module}%{SPACE}%{NOTSPACE:level}%{SPACE}%{GREEDYDATA:http_url}`

#### Custom pattern {#benchmark-pattern-custom}

- Custom alias Nginx error log: `%{APPDATE:time} \[%{LOGLEVEL:level}\] %{GREEDYDATA:msg}, client: %{IPORHOST:client}, server: %{NOTSPACE:server}, request: "%{WORD:method} %{GREEDYDATA:path} HTTP/%{NUMBER:http_version}", %{UPBLOCK}host: "%{NOTSPACE:host}"`
- Custom alias Postfix log: `%{SYSLOGTIMESTAMP:timestamp} %{SYSLOGHOST:host} %{WORD:program}\[%{POSINT:pid}\]: %{QUEUEID:queue_id}: %{GREEDYDATA:msg}`

### Built-in Pattern List {#builtin-patterns}

DataKit has some commonly used Patterns built in, which we can use directly when using Grok cutting:

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
