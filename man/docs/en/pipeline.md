
# Text Data Processing（Pipeline）
---

The following is the text processor definition. With the gradual support of different syntax, the document will be adjusted, added and deleted to varying degrees.

## Ground Rule {#basic-syntax}

- Function name is case-insensitive
- Take `#` as the line comment character. Inline comments are not supported.
- Identifier: Only `[_a-zA-Z0-9]` can appear, and the first character cannot be a number. Such as `_abc, _abc123, _123ab`
- Double or single quotation marks can be used for string values, and triple double or single quotation marks can be used for multi-line strings:
  - `"this is a string"` 
  - `'this is a string'`
  - ```
    """%{NUMBER:a:int}
    %{WORD:B:bool} %{NUMBER:b:float}"""
    ```
  - ```
    '''%{NUMBER:a:int}
    %{WORD:B:bool} %{NUMBER:b:float}'''
    ```

- Data types: support floating point (`123.4`, `5.67E3`), shaping (`123`, `-1`), string (`'Zhang San'`, `"hello world"`), Boolean (`true`, `false`)
- Multiple functions can be separated by white space characters (space, line break, Tab, etc.)
- If the cut-out field contains special characters (such as `@`, `$`, Chinese characters, emoticons, etc.), it needs additional modification when referencing in the code, such as `` `@some-variable` ``，`` `这是一个表情包变量👍` ``
- Only `[_a-zA-Z0-9]` can appear in the field name, i.e. only underscores (`_`), upper and lower case letters, and numbers. In addition, the first character of **cannot be a number**.

### Special Character {#special}

To maintain forward compatibility of Pipeline semantics, for logs, `_` is an alias for `message`,  which only takes effect in log class data.

## Quick Start {#quick-start}

- Configure the pipeline in the DataKit by writing the following pipeline file, assuming the name is *nginx.p*. Store it in the `<datakit安装目录>/pipeline` directory.

```python
# Assume the input is an Nginx log (the following fields are all yy's...)
# Note that scripts can be annotated

grok(_, "some-grok-patterns")  # grok the input text
rename('client_ip', ip)        # renames the ip field to client_ip
rename("网络协议", protocol)   # renames the protocol field to ` network protocol '

# Change timestamp (e.g. 1610967131) to RFC3339 date format: 2006-01-02T15:04:05Z07:00
datetime(access_time, "s", "RFC3339")

url_decode(request_url)      # translates HTTP request routing into clear text

# When status_code is between 200 and 300, create a new field with http_status = "HTTP_OK"
group_between(status_code, [200, 300], "HTTP_OK", "http_status")

# Drop the original content
drop_origin_data()
```

???+ attention

    During cutting, avoid [possible duplicate name with tag key](datakit-pl-how-to.md#naming)

- Configure the Corresponding Collector to Use the Pipeline Above

Taking the logging collector as an example, configure the field `pipeline_path`, noting that the script name of the pipeline is configured here, not the path. All pipeline scripts referenced here must be stored in the `<DataKit 安装目录/pipeline>` directory:

```python
[[inputs.logging]]
    logfiles = ["/path/to/nginx/log"]

    # required
    source = "nginx"

    # All scripts must be placed in the/path/to/datakit/pipeline directory.
    # If gitrepos functionality is turned on, the file with the same name in gitrepos takes precedence.
    # If the pipeline is not configured, look for the same name as source in the pipeline directory.
    # As its default pipeline configuration, the script for (such as nginx -> nginx.p).
    pipeline = "nginx.p"

    ... # Other configurations
```

Restart the collector to cut the corresponding log.

???+ info

    For Pipeline writing, debugging, and considerations, see [here](datakit-pl-how-to.md).

## Grok Pattern Classification {#grok}

Grok patterns in DataKit can be divided into two categories:

- Global mode: The schema files in the *pattern* directory are global mode and can be used by all pipeline scripts.
- Local pattern: The new pattern in the pipeline script through the [add_pattern()](#fn-add-pattern) function is a local pattern and is only valid for the current pipeline script.

Here's an example of how to write the corresponding grok, using the Nginx access-log, the original of which is as follows:

```log
127.0.0.1 - - [26/May/2022:20:53:52 +0800] "GET /server_status HTTP/1.1" 404 134 "-" "Go-http-client/1.1"
```

Assuming that we need to get client_ip、time (request), http_method, http_url, http_version, status_code from the access log, the grok pattern can initially be written as:

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

The optimized cutting is more readable than the preliminary single-line pattern. Since the default data type of the field resolved by grok is string, specifying the data type of the field here avoids the subsequent use of the [cast()](#fn-cast) function for type conversion.

### grok Combination {#grok-compose}

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
add_pattern(time, "([^0-9]?)%{HOUR:hour}:%{MINUTE:minute}(?::%{SECOND:second})([^0-9]?)")

# Extract the time field from the original input through grok. Assuming the input is 12:30:59, the {"hour": 12, "minute": 30, "second": 59}
grok(_, %{time})
```

???+ attention

    - If a pattern with the same name occurs, the local pattern takes precedence (that is, the local pattern overrides the global pattern).
    - In the pipeline script, [add_pattern()](#fn-add-pattern) needs to be called before the [grok()](#fn-grok) function, otherwise the first data fetch will fail.

### Built-in Pattern List {#builtin-patterns}

DataKit has some commonly used Patterns built in, which we can use directly when using Grok cutting:

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

## if/else Branch {#if-else}

pipeline supports the `if/elif/else` syntax, and statements following `if` only support conditional expressions, namely `<`、`<=`、`==`、`>`、`>=` and `!=`, and it supports parenthesis precedence `AND` and `OR` joins of multiple conditional expressions.
An expression can be flanked by an existing key or fixed value (numeric, Boolean, string, and nil), such as:

```python
# Numerical comparison
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

# String comparison
add_key(name, "张三")

if name == "法外狂徒" {
  # It's impossible, don't slander me
}
```

Like most programming/scripting languages, the order of execution depends on whether the `if/elif` condition holds.

Note: For numeric comparison, you need to use `cast()` for type conversion, such as:

```
# status_code is the string type cut out by grok
cast(status_code, "int")

if status == 200 {
  add_key(level, "OK")
} elif status >= 400 && status < 500 {
  add_key(level, "ERROR")
} elif stauts > 500 {
  add_key(level, "FATAL")
}
```

## Pipeline Script Storage Directory {#pl-dirs}

Pipeline's directory search priority is:

1. Remote Pipeline directory
2. Git-managed pipeline directory
3. Built-in pipeline directory

Find from 1 to 3 directions, and return directly after matching.

Absolute path writing is not allowed.

### Remote Pipeline Directory {#remote-pl}

Under the `pipeline_remote` directory under the Datakit installation directory, the directory structure is as follows:

```
.
├── conf.d
├── datakit
├── pipeline
│   ├── root_apache.p
│   └── root_consul.p
├── pipeline_remote
│   ├── remote_elasticsearch.p
│   └── remote_jenkins.p
├── gitrepos
│   └── mygitproject
│       ├── conf.d
│       ├── pipeline
│       │   └── git_kafka.p
│       │   └── git_mongod.p
│       └── python.d
└── ...
```

### Git Managed Pipeline Directory {#git-pl}

Under the `project name/pipeline` directory under the `gitrepos` directory, the directory structure is shown above.

### Built-in Pipeline Directory {#internal-pl}

Under the `pipeline` directory under the Datakit installation directory, the directory structure is shown above.

## Script Functions {#functions}

Function parameter description:

- In function arguments, the anonymous argument (`_`) refers to the original input text data
- json path, expressed directly as `x.y.z`, without any other modifications. For example, `{"a":{"first":2.3, "second":2, "third":"abc", "forth":true}, "age":47}`, where the json path is `a.thrid` to indicate that the data to be manipulated is `abc`
- The relative order of all function arguments is fixed, and the engine will check it concretely
- All of the `key` parameters mentioned below refer to the `key` generated after the initial extraction (via `grok()` or `json()`)
- The path of the json to be processed, supports the writing of identifiers, and cannot use strings. If you are generating new keys, you need to use strings

### `add_key()` {#fn-add-key}

Function prototype: `add_key(key-name=required, key-value=required)`

Function description: Add a field

Function parameter

- `key-name`: New key name
- `key-value`: Key value (only type string/number/bool/nil)

Example:

```python
# Data to be processed: {"age": 17, "name": "zhangsan", "height": 180}

# Processing scripts
add_key(city, "shanghai")

# Processing results
{
    "age": 17,
    "height": 180,
    "name": "zhangsan",
    "city": "shanghai"
}
```
### `add_pattern()` {#fn-add-pattern}

Function prototype: `add_pattern(name=required, pattern=required)`

Function description: Create a custom grok pattern. The grok pattern has scope restrictions, such as a new scope generated within the if else statement, and the pattern is valid only within this scope. This function cannot override grok patterns that already exist in the same scope or in the previous scope.

Parameters:

- `name`: Schema naming
- `pattern`: Custom pattern content

Example:

```python
# Data to be processed: "11,abc,end1", "22,abc,end1", "33,abc,end3"

# pipline script
add_pattern("aa", "\\d{2}")
grok(_, "%{aa:aa}")
if false {

} else {
    add_pattern("bb", "[a-z]{3}")
    if aa == "11" {
        add_pattern("cc", "end1")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    } elif aa == "22" {
        # Using pattern cc here will cause compilation to fail: no pattern found for% {cc}
        grok(_, "%{aa:aa},%{bb:bb},%{INT:cc}")
    } elif aa == "33" {
        add_pattern("bb", "[\\d]{5}")	# Failed to overwrite bb here
        add_pattern("cc", "end3")
        grok(_, "%{aa:aa},%{bb:bb},%{cc:cc}")
    }
}

# Processing results
{
    "aa":      "11"
    "bb":      "abc"
    "cc":      "end1"
    "message": "11,abc,end1"
}
{
    "aa":      "22"
	 "message": "22,abc,end1"
}
{
    "aa":      "33"
    "bb":      "abc"
    "cc":      "end3"
    "message": "33,abc,end3"
}
```
### `adjust_timezone()` {#fn-adjust-timezone}

Function prototype: `adjust_timezone(key=required, minute=optional)`

Function parameter

- `key`: Nanosecond timestamp, such as the timestamp after processing by the `default_time(time)`
- `minute`: The number of minutes (integer) that the return value is allowed to exceed the current time, the value range is [0, 15], and the default value is 2 minutes

Function description: Make the difference between the incoming timestamp minus the timestamp at the time when the function is executed within (-60 + minute, minute] minutes; does not apply to data with time difference beyond this range, otherwise it will lead to wrong data. Calculation flow:

1. Add a few hours to the value of key so that it is within the current hour
2. At this time, calculate the minute difference between the two. The minute value range of the two is [0, 60], and the difference range is (-60, 0] and [0, 60)
3. If the difference is less than or equal to-60 + minute, plus 1 hour, and if it is greater than minute, minus 1 hour
4. The default value of minute is 2, then the range of difference is allowed to be (-58, 2). If this time is 11:10, the log time is 3:12:00. 001, and the final result is 10:12:00. 001; If this time is 11:59: 1.000, the log time is 3:01: 1.000, and the final result is 12:01: 1.000

For example:

```json
# input 1 
{
    "time":"11 Jul 2022 12:49:20.937", 
    "second":2,
    "third":"abc",
    "forth":true
}
```

Script:

```python
json(_, time)      # Extract the time field (if the time zone in the container is UTC+0000)
default_time(time) # Convert the extracted time field into a timestamp
                   # (Use local time zone UTC+0800/UTC+0900... parsing for data without time zone)
adjust_timezone(time)
                   # Automatically (re-) select the time zone and calibrate the time deviation

```

execute `datakit pipeline -P <name>.p -F <input_file_name>  --date`:

```json
# output 1
{
  "message": "{\n    \"time\":\"11 Jul 2022 12:49:20.937\",\n    \"second\":2,\n    \"third\":\"abc\",\n    \"forth\":true\n}",
  "status": "unknown",
  "time": "2022-07-11T20:49:20.937+08:00"
}
```

Local time: `2022-07-11T20:55:10.521+08:00`

Using only default_time, the time parsed by default native time zone (UTC+8) is as follows:
  - Input 1 Result: `2022-07-11T12:49:20.937+08:00`

After using adjust_timezone, you will get:
  - Input 1 Result: `2022-07-11T20:49:20.937+08:00`

### `cast()` {#fn-cast}

Function prototype: `cast(key=required, type=required)`

Function description: splits key value conversion into specified type

Function parameter

- `key`: A field that has been extracted
- `type`: Target type of conversion, supported `\"str\", \"float\", \"int\", \"bool\"` , and the target type should be enclosed in English state double quotation marks

For example:

```python
# Data to be processed: {"first": 1,"second":2,"third":"aBC","forth":true}

# Processing script
json(_, first) cast(first, "str")

# Processing result
{
  "first": "1"
}
```

### `cover()` {#fn-cover}

Function prototype: `cover(key=required, range=require)`

Function description: The string data obtained on the specified field is desensitized according to the range

Function parameter

- `key`: Field to be extracted
- `range`: The index range of the desensitized string `[start,end]`）both start and end support negative subscripts to express the semantics of tracing back from the tail. If the interval is reasonable, the end will default to the maximum length if it is greater than the maximum length of the string.

For example:

```python
# Data to be processed {"str": "13789123014"}
json(_, str) cover(str, [8, 13])

# Data to be processed {"str": "13789123014"}
json(_, str) cover(str, [2, 4])

# Data to be processed {"str": "13789123014"}
json(_, str) cover(str, [1, 1])

# Data to be processed {"str": "小阿卡"}
json(_, str) cover(str, [2, 2])
```

### `datetime()` {#fn-datetime}

Function prototype: `datetime(key=required, precision=required, fmt=required)`

Function description: Convert the timestamp to the specified date format.

Function parameter

- `key`: Extracted timestamp (required)
- `precision`: Input timestamp precision (s, ms)
- `fmt`: Date format, time format, support the following templates

```python
ANSIC       = "Mon Jan _2 15:04:05 2006"
UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
RFC822      = "02 Jan 06 15:04 MST"
RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
RFC3339     = "2006-01-02T15:04:05Z07:00"
RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
Kitchen     = "3:04PM"
```

For example:

```python
# Data to be processed:
#    {
#        "a":{
#            "timestamp": "1610960605000",
#            "second":2
#        },
#        "age":47
#    }

# Processing script
json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')
```

### `decode()` {#fn-decode}

Function prototype: `decode(text, text-encode)`

Function Description: text into UTF8 encoding to deal with the original log for non-UTF8 encoding problem. The currently supported encoding is utf-16le/utf-16be/gbk/gb18030 (these encoding names can only be lowercase).

```python
decode("wwwwww", "gbk")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "wwwwww",
# }
```

### `default_time()` {#fn-defalt-time}

Function prototype: `default_time(key=required, timezone=optional)`

Function Description: Take one of the extracted fields as the timestamp of the final data

Function parameter

- `key`: The specified key, whose data type needs to be a string type
- `timezone`: Specify the time zone used by the time text to be formatted, default to the current local time zone, example time zone `+8/-8/+8:30`

The data to be processed supports the following formatting times

| Date Format                                           | Date Format                                                | Date Format                                       | Date Format                          |
| -----                                              | ----                                                    | ----                                           | ----                              |
| `2014-04-26 17:24:37.3186369`                      | `May 8, 2009 5:57:51 PM`                                | `2012-08-03 18:31:59.257000000`                | `oct 7, 1970`                     |
| `2014-04-26 17:24:37.123`                          | `oct 7, '70`                                            | `2013-04-01 22:43`                             | `oct. 7, 1970`                    |
| `2013-04-01 22:43:22`                              | `oct. 7, 70`                                            | `2014-12-16 06:20:00 UTC`                      | `Mon Jan  2 15:04:05 2006`        |
| `2014-12-16 06:20:00 GMT`                          | `Mon Jan  2 15:04:05 MST 2006`                          | `2014-04-26 05:24:37 PM`                       | `Mon Jan 02 15:04:05 -0700 2006`  |
| `2014-04-26 13:13:43 +0800`                        | `Monday, 02-Jan-06 15:04:05 MST`                        | `2014-04-26 13:13:43 +0800 +08`                | `Mon, 02 Jan 2006 15:04:05 MST`   |
| `2014-04-26 13:13:44 +09:00`                       | `Tue, 11 Jul 2017 16:28:13 +0200 (CEST)`                | `2012-08-03 18:31:59.257000000 +0000 UTC`      | `Mon, 02 Jan 2006 15:04:05 -0700` |
| `2015-09-30 18:48:56.35272715 +0000 UTC`           | `Thu, 4 Jan 2018 17:53:36 +0000`                        | `2015-02-18 00:12:00 +0000 GMT`                | `Mon 30 Sep 2018 09:09:09 PM UTC` |
| `2015-02-18 00:12:00 +0000 UTC`                    | `Mon Aug 10 15:44:11 UTC+0100 2015`                     | `2015-02-08 03:02:00 +0300 MSK m=+0.000000001` | `Thu, 4 Jan 2018 17:53:36 +0000`  |
| `2015-02-08 03:02:00.001 +0300 MSK m=+0.000000001` | `Fri Jul 03 2015 18:04:07 GMT+0100 (GMT Daylight Time)` | `2017-07-19 03:21:51+00:00`                    | `September 17, 2012 10:09am`      |
| `2014-04-26`                                       | `September 17, 2012 at 10:09am PST-08`                  | `2014-04`                                      | `September 17, 2012, 10:10:09`    |
| `2014`                                             | `2014:3:31`                                             | `2014-05-11 08:20:13,787`                      | `2014:03:31`                      |
| `3.31.2014`                                        | `2014:4:8 22:05`                                        | `03.31.2014`                                   | `2014:04:08 22:05`                |
| `08.21.71`                                         | `2014:04:2 03:00:51`                                    | `2014.03`                                      | `2014:4:02 03:00:51`              |
| `2014.03.30`                                       | `2012:03:19 10:11:59`                                   | `20140601`                                     | `2012:03:19 10:11:59.3186369`     |
| `20140722105203`                                   | `2014年04月08日`                                        | `1332151919`                                   | `2006-01-02T15:04:05+0000`        |
| `1384216367189`                                    | `2009-08-12T22:15:09-07:00`                             | `1384216367111222`                             | `2009-08-12T22:15:09`             |
| `1384216367111222333`                              | `2009-08-12T22:15:09Z`                                  |

JSON Extraction Example:

```python
# Raw json
{
    "time":"06/Jan/2017:16:16:37 +0000",
    "second":2,
    "third":"abc",
    "forth":true
}

# pipeline script
json(_, time)      # extract time field
default_time(time) # convert the extracted time field into a timestamp

# Processing result
{
  "time": 1483719397000000000,
}
```

Example of text extraction:

```python
# Original log text
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms

# pipeline script
grok(_, '%{TIMESTAMP_ISO8601:log_time}')   # extract the log time and names the field log_time
default_time(log_time)                     # convert the extracted log_time field into a timestamp

# Processing results
{
  "log_time": 1610358231887000000,
}

# For data collected by logging, it is best to name the time field time, otherwise the logging collector will populate with the current time
rename("time", log_time)

# Processing results
{
  "time": 1610358231887000000,
}
```


### `drop()` {#fn-drop}

Function prototype:`drop()`

Function description: Discard the whole log and do not upload it

```python
# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
if str_a == "2"{
  drop()
  exit()
}
json(_, str_b)

# Extracted data(drop: true, cost: 30.02µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a": "2"
# }
```


### `drop_key()` {#fn-drop-key}

Function prototype: `drop_key(key=required)`

Function Description: Delete Extracted Fields

Function parameter

- `key`: Field name to be deleted

Fo r:

```python
data = `{\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}`

# Processing script
json(_, age,)
json(_, name)
json(_, height)
drop_key(height)

# Processing result
{
    "age": 17,
    "name": "zhangsan"
}
```


### `drop_origin_data()` {#fn-drop-origin-data}

Function prototype: `drop_origin_data()`

Handling script function description: Discard the initialization text, otherwise the initial text will be placed in the message field.

For example:

```python
# Data to be processed: {"age": 17, "name": "zhangsan", "height": 180}

# Delete message content in the result set
drop_origin_data()
```


### `duration_precision()` {#fn-duration-precision}

Function prototype: `duration_precision(key=required, old_precision=require, new_precision=require)`

Function Description: duration precision conversion, through the parameters to specify the current precision and target precision. Support conversion between s, ms, us and ns.

```python
# in << {"ts":12345}
json(_, ts)
cast(ts, "int")
duration_precision(ts, "ms", "ns")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "{\"ts\":12345}",
#   "ts": 12345000000
# }
```


### `exit()` {#fn-exit}

Function prototype:`exit()`

Function description: End the parsing of the current log. If the function drop () is not called, the parsed part will still be output.

```python
# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
if str_a == "2"{
  exit()
}
json(_, str_b)

# Extracted data(drop: false, cost: 48.233µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a": "2"
# }
```


### `geoip()` {#fn-geoip}

Function prototype:`geoip(key=required)`

Function description: Add more IP information on IP. `geoip()` produces additional fields, such as:

- `isp`: Operator
- `city`: City
- `province`: Provice
- `country`: Country

Parameter:

- `key`:  IP fields that have been extracted, supporting IPv4/6

For example:

```python
# Data to be processed: {"ip":"1.2.3.4"}

# Processing script
json(_, ip)
geoip(ip)

# Processing script
{
  "city"     : "Brisbane",
  "country"  : "AU",
  "ip"       : "1.2.3.4",
  "province" : "Queensland",
  "isp"      : "unknown"
  "message"  : "{\"ip\": \"1.2.3.4\"}",
}
```
### `grok()` {#fn-grok}

Function prototype:`grok(input=required, pattern=required, trim_space=optional)`

Function description: Extract the contents of the text string `input` by `pattern`.

Parameter:

- `input`: The text to be extracted, either the original text (`_`) or some `key` after the initial extraction
- `pattern`: Grok expression, in which data types for specifying key are supported: bool, float, int, string, default to string
- `trim_space`: Delete the white space beginning and ending characters in the extracted characters, default is true

```python
grok(_, pattern)    # uses the input text directly as raw data
grok(key, pattern)  # grok again for a key that has been extracted before
```

For example:

```python
# Data to be processed: "12/01/2021 21:13:14.123"

# pipline script
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour:string}:%{_minute:minute:int}(?::%{_second:second:float})([^0-9]?)
grok(_, "%{DATE_US:date} %{time}")

# Processing script
{
  "date": "12/01/2021",
  "hour": "21",
  "message": "12/01/2021 21:13:14.123",
  "minute": 13,
  "second": 14.123
}
```

### `group_between()` {#fn-group-between}

Function prototype:`group_between(key=required, between=required, new-value=required, new-key=optional)`

Function description: If the `key` value is within the specified range `between` (note: it can only be a single interval, such as `[0,100]`), you can create a new field and assign a new value. If no new field is provided, the original field value is overwritten

For example:

```python
# Data to be processed: {"http_status": 200, "code": "success"}

json(_, http_status)

# If the field http_status value is within the specified range, change its value to "OK"
group_between(http_status, [200, 300], "OK")
`

# Processing script
{
    "http_status": "OK"
}
```

For example:

```python
# Data to be processed: {"http_status": 200, "code": "success"}

json(_, http_status)

# If the field http_status value is within the specified range, change its value to "OK".
group_between(http_status, [200, 300], "OK", status)

# Processing script
{
    "http_status": 200,
    "status": "OK"
}
```


### `group_in()` {#fn-group-in}

Function prototype:`group_in(key=required, in=required, new-value=required, new-key=optional)`

Function description: If the `key` value is in the list `in`, you can create a new field and assign a new value. If no new field is provided, the original field value is overwritten

For example:

```python
# Change the value of the field log_level to "OK" if it is in the list
group_in(log_level, ["info", "debug"], "OK")

# If the field http_status value is in the specified list, creates a new status field with a value of "not-ok"
group_in(log_level, ["error", "panic"], "not-ok", status)
```


### `json()` {#fn-json}

Function prototype:`json(input=required, jsonPath=required, newkey=required, trim_space=optional)`

Function description: Extract the specified field in json and name it as a new field.

Parameter:

- `input`: The json to be extracted can be the original text (`_`) or some `key` after the initial extraction
- `jsonPath`: json path information
- `newkey`: The extracted data is written to the new key
- `trim_space`: Delete the white space beginning and ending characters in the extracted characters, default is true

```python
# Extract the x.y field from the original input json directly and name it as a new field abc
json(_, x.y, abc)

# One of the `key` that has been extracted, extract it again `x.y', with the extracted field named `x.y'
json(key, x.y) 
```

For example:

```python
# Data to be processed: {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# Processing script
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "年龄")

# Processing script
{
    "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}
    "zhangsan": {
        "age": 17,
        "height": 180,
        "name": "zhangsan"
    }
}
```

For example:

```python
# Data to be processed
#    data = {
#        "name": {"first": "Tom", "last": "Anderson"},
#        "age":37,
#        "children": ["Sara","Alex","Jack"],
#        "fav.movie": "Deer Hunter",
#        "friends": [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#        ]
#    }

# Processing script
json(_, name) json(name, first)
```

For example:

```python
# Data to be processed
#    [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#    ]
    
# Processing script, json数组处理
json(_, [0].nets[-1])
```

### `lowercase()` {#fn-lowercase}

Function prototype:`lowercase(key=required)`

Function description: Convert the contents in the extracted key to lowercase

Function parameter

- `key`: Specify the extracted field name to be converted

For example:

```python
# Data to be processed: {"first": "HeLLo","second":2,"third":"aBC","forth":true}

# Processing script
json(_, first) lowercase(first)

# Processing result
{
		"first": "hello"
}
```

### `mquery_refer_table()` {#fn-mquery-refer-table}

Function prototype:`mquery_refer_table(table_name=requierd, keys=required, values=required)`

Function description: Query the external reference table by specifying multiple keys, and append all columns in the first row of the query result to the field.

Parameter:

- `table_name`: The name of the table to find
- `keys`: A list of multiple column names
- `values`: The value for each column

For example:

```python
json(_, table)
json(_, key)
json(_, value)

# Query and append the data of the current column, which is added to the data as field by default
mquery_refer_table(table, values=[value, false], keys=[key, "col4"])
```

Example result:

```json
{
  "col": "ab",
  "col2": 1234,
  "col3": 1235,
  "col4": false,
  "key": "col2",
  "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
  "status": "unknown",
  "table": "table_abc",
  "time": "2022-08-16T16:23:31.940600281+08:00",
  "value": 1234
}

```

### `nullif()` {#fn-nullif}

Function prototype:`nullif(key=required, value=required)`

Function description: If the extracted field specified by `key` is equal to the value of `value`, delete this field

Function parameter

- `key`: Specify the field
- `value`: Target value

For example:

```python
# Data to be processed: {"first": 1,"second":2,"third":"aBC","forth":true}

# Processing script
json(_, first) json(_, second) nullif(first, "1")

# Processing script
{
    "second":2
}
```

> Note: This function can be achieved through the `if/else` semantics:

```python
if first == "1" {
	drop_key(first)
}
```


### `parse_date()` {#fn-parse-date}

Function prototype:`parse_date(new-key=required, yy=require, MM=require, dd=require, hh=require, mm=require, ss=require, ms=require, zone=require)`

Function description: Converts the value of each part of the passed-in date field into a timestamp

Function parameter

- `key`: Newly inserted field
- `yy` : Year numeric string, supports four-digit or two-digit string, if it is empty, the current year will be taken when processing
- `MM`: Month string, supported numbers, English, English abbreviations
- `dd`: Day string
- `hh`: Hour string
- `mm`: Minute string
- `ss`: Second string
- `ms`: Millisecond string
- `zone`: Time zone string in the form of "+8" or\ "Asia/Shanghai\"

For example:

```python
parse_date(aa, "2021", "May", "12", "10", "10", "34", "", "Asia/Shanghai") # result aa=1620785434000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "", "Asia/Shanghai") # result aa=1639275034000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "100", "Asia/Shanghai") # result aa=1639275034000000100

parse_date(aa, "20", "February", "12", "10", "10", "34", "", "+8") result aa=1581473434000000000
```


### `parse_duration()` {#fn-parse-duration}

Function prototype:`parse_duration(key=required)`

Function description: If the value of `key` is a duration string of golang (such as `123ms`), `key` is automatically parsed to an integer in nanoseconds.

At present, the duration units in golang are as follows:

- `ns` nanoseconds
- `us/µs` microseconds
- `ms` milliseconds
- `s` seconds
- `m` minutes
- `h` hours

Function parameter

- `key`: Fields to be resolved

For example:

```python
# Assuming abc = "3.5s"
parse_duration(abc) # Result abc = 3500000000

# Supports negative numbers: abc = "-3.5s"
parse_duration(abc) # Result abc = -3500000000

# Support floating point: abc = "-2.3s"
parse_duration(abc) # Result abc = -2300000000

```

### `query_refer_table()` {#fn-query-refer-table}

Function prototype:`query_refer_table(table_name=requierd, key=required, value=required)`

Function description: Query the external reference table with the specified key, and appends all columns in the first row of the query result to the field.

Parameter:

- `table_name`: The name of the table to find
- `key`: Column name
- `value`: The value of the column

For example:

```python
# Extract table name, column name, column value from input
json(_, table)
json(_, key)
json(_, value)

# Queries and appends the data for the current column, adding it to the data as a field by default
query_refer_table(table, key, value)

```

Example results:

```json
{
  "col": "ab",
  "col2": 1234,
  "col3": 123,
  "col4": true,
  "key": "col2",
  "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
  "status": "unknown",
  "table": "table_abc",
  "time": "2022-08-16T15:02:14.158452592+08:00",
  "value": 1234
}
```

### `rename()` {#fn-rename}

Function prototype:`rename(new-key=required, old-key=required)`

Function description: Rename the extracted field

Parameter:

- `new-key`: New field name
- `old-key`: Extracted field name

```python
# Rename the extracted abc field to abc1
rename('abc1', abc)

# or 

rename(abc1, abc)
```

For example:

```python
# Data to be processed: {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# Processing script
json(_, info.name, "姓名")

# Processing script
{
  "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
  "zhangsan": {
    "age": 17,
    "height": 180,
    "姓名": "zhangsan"
  }
}
```


### `replace()` {#fn-replace}

Function prototype:`replace(key=required, regex=required, replaceStr=required)`

Function description: Replace the string data obtained on the specified field according to regularity.

Function parameter

- `key`: Field to be extracted
- `regex`: Regular expression
- `replaceStr`: Replaced string

For example:

```python
# Telephone number: {"str": "13789123014"}
json(_, str)
replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")

# English name {"str": "zhang san"}
json(_, str)
replace(str, "([a-z]*) \\w*", "$1 ***")

# ID number {"str": "362201200005302565"}
json(_, str)
replace(str, "([1-9]{4})[0-9]{10}([0-9]{4})", "$1**********$2")

# Chinese name {"str": "小阿卡"}
json(_, str)
replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")
```


### `set_measurement()` {#fn-set-measurement}

Function prototype:`set_measurement(key=required, disable_delete_key=optional)`

Function description: Change the mesaurement name of the row protocol.

Function parameter

- `key`: Take the key value as the mesaurement name
- `value`: Default is false delete key, can be true or false


### `set_tag()` {#fn-set-tag}

Function prototype:`set_tag(key=required, value=optional)`

Function Description: The specified field is marked as tag output, and after being set to tag, other functions can still operate on this variable. If the key set to tag is a cut-out field, it will not appear in the field, which prevents the cut-out field key from having the same name as the tag key on the existing data.

Function parameter

- `key`: Fields to be marked as tag
- `value`: be a literal string or a variable

```python
# in << {"str": "13789123014"}
set_tag(str)
json(_, str)          # str == "13789123014"
replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")
# Extracted data(drop: false, cost: 49.248µs):
# {
#   "message": "{\"str\": \"13789123014\", \"str_b\": \"3\"}",
#   "str#": "137****3014"
# }
# * 字符 `#` 仅为 datakit --pl <path> --txt <str> 输出展示时字段为 tag 的标记

# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
set_tag(str_a, "3")   # str_a == 3
# Extracted data(drop: false, cost: 30.069µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a#": "3"
# }


# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
json(_, str_b)
set_tag(str_a, str_b) # str_a == str_b == "3"
# Extracted data(drop: false, cost: 32.903µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a#": "3",
#   "str_b": "3"
# }
				"```

### `sql_cover()` {#fn-sql-cover}

Function prototype:`sql_cover(sql_test)`

Function description: desensitized sql statement

```python
# in << {"select abc from def where x > 3 and y < 5"}
sql_cover(_)

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "select abc from def where x > ? and y < ?"
# }
```

### `strfmt()` {#fn-strfmt}

Function prototype:`strfmt(key=required, fmt=required, key1=optional, key2, ...)`

Function description: Format the field contents specified by the extracted `key1,key2...` according to `fmt` , and write the formatted contents into the `key` field

Function parameter

- `key`: Specify the formatted data write field name
- `fmt`: Formatting string template
- `key1，key2`: Field name to be formatted extracted

For example:

```python
# Data to be processed: {"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}

# Processing script
json(_, a.second)
json(_, a.thrid)
cast(a.second, "int")
json(_, a.forth)
strfmt(bb, "%v %s %v", a.second, a.thrid, a.forth)
```

### `trim()` {#fn-trim}

Function prototype:`trim(key=required, cutset=optional)`

Function description: Delete the characters specified in the beginning and end of key, and delete all white space characters by default when cutset is an empty string.

Function arguments:

- `key`: A field extracted, string type
- `cutset`: Delete the first and last characters in the key that appear in the cutset string

For example:

```python
# Data to be processed: "trim(key, cutset)"

# Processing script
add_key(test_data, "ACCAA_test_DataA_ACBA")
trim(test_data, "ABC_")

# Processing script
{
  "test_data": "test_Data"
}
```

### `uppercase()` {#fn-uppercase}

Function prototype:`uppercase(key=required)`

Function description: Convert the contents in the extracted key to uppercase.

Function parameter

- `key`: Specify the extracted field name to be converted, converting `key` contents to uppercase.

For example:

```python
# Data to be processed: {"first": "hello","second":2,"third":"aBC","forth":true}

# Processing script
json(_, first) uppercase(first)

# Processing script
{
   "first": "HELLO"
}
```


### `url_decode()` {#fn-url-decode}

Function prototype:`url_decode(key=required)`

Function description: Parse URL in extracted `key` into clear text

Parameter:

- `key`: One of the `key` that has been extracted

For example:

```python
# Data to be processed: {"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}

# Processing script
json(_, url) url_decode(url)

# Processing script
{
  "message": "{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}",
  "url": "http://www.baidu.com/s?wd=测试"
}
```

### `use()` {#fn-use}

Function prototype:`use(name=required)`

Parameter:

- `name`: script name, such as abp.p

Function description: Call other scripts, you can access all the current data in the called script
For example:

```python
# Data to be processed: {"ip":"1.2.3.4"}

# Processing script a.p
use(\"b.p\")

# Processing script b.p
json(_, ip)
geoip(ip)

# Execute the processing result of script a.p
{
  "city"     : "Brisbane",
  "country"  : "AU",
  "ip"       : "1.2.3.4",
  "province" : "Queensland",
  "isp"      : "unknown"
  "message"  : "{\"ip\": \"1.2.3.4\"}",
}
```

### `user_agent()` {#fn-user-agent}

Function prototype:`user_agent(key=required)`

Function description: Obtain client information on the specified field.

Function parameter

- `key`: Field to be extracted

`user_agent()` produces multiple fields, such as:

- `os`: Operating system
- `browser`: Browser

For example:

```python
# Data to be processed
#    {
#        "userAgent" : "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
#        "second"    : 2,
#        "third"     : "abc",
#        "forth"     : true
#    }

json(_, userAgent) user_agent(userAgent)
```
### `xml()` {#fn-xml}

Function prototype:`xml(input=required, xpath_expr=required, key_name=required)`

Function description: Extract fields from XML through xpath expression.

Parameters:

- input: XML to be extracted
- xpath_expr: xpath expression
- key_name: The extracted data is written to the new key

Example:

```python
# Data to be processed
       <entry>
        <fieldx>valuex</fieldx>
        <fieldy>...</fieldy>
        <fieldz>...</fieldz>
        <fieldarray>
            <fielda>element_a_1</fielda>
            <fielda>element_a_2</fielda>
        </fieldarray>
    </entry>

# Processing script
xml(_, '/entry/fieldarray//fielda[1]/text()', field_a_1)

# Processing script
{
  "field_a_1": "element_a_1",  # 提取了 element_a_1
  "message": "\t\t\u003centry\u003e\n        \u003cfieldx\u003evaluex\u003c/fieldx\u003e\n        \u003cfieldy\u003e...\u003c/fieldy\u003e\n        \u003cfieldz\u003e...\u003c/fieldz\u003e\n        \u003cfieldarray\u003e\n            \u003cfielda\u003eelement_a_1\u003c/fielda\u003e\n            \u003cfielda\u003eelement_a_2\u003c/fielda\u003e\n        \u003c/fieldarray\u003e\n    \u003c/entry\u003e",
  "status": "unknown",
  "time": 1655522989104916000
}
```

Example:

```python
# Data to be processed
<OrderEvent actionCode = "5">
 <OrderNumber>ORD12345</OrderNumber>
 <VendorNumber>V11111</VendorNumber>
</OrderEvent>

# Processing script
xml(_, '/OrderEvent/@actionCode', action_code)
xml(_, '/OrderEvent/OrderNumber/text()', OrderNumber)

# Processing script
{
  "OrderNumber": "ORD12345",
  "action_code": "5",
  "message": "\u003cOrderEvent actionCode = \"5\"\u003e\n \u003cOrderNumber\u003eORD12345\u003c/OrderNumber\u003e\n \u003cVendorNumber\u003eV11111\u003c/VendorNumber\u003e\n\u003c/OrderEvent\u003e",
  "status": "unknown",
  "time": 1655523193632471000
}
```

{{.PipelineFuncsEN}}
