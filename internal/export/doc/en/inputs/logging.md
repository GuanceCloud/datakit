
# File Log
---

{{.AvailableArchs}}

---

This document focuses on local disk log collection and Socket log collection:

- Disk log collection: Collect data at the end of the file (similar to command line `tail -f`）
- Socket port collection: Send log to DataKit via TCP/UDP

## Configuration {#config}

=== "Host deployment"

    Go to the `conf.d/log` directory under the DataKit installation directory, copy `logging.conf.sample` and name it `logging.conf`. Examples are as follows:
    
    ``` toml
    [[inputs.logging]]
      # Log file list, you can specify absolute path, support batch specification using glob rules
      # Absolute path recommended
      logfiles = [
        "/var/log/*",                          # All files under the file path
        "/var/log/sys*",                       # All files prefixed with sys under the file path
        "/var/log/syslog",                     # Unix format file path
        "C:/path/space 空格中文路径/some.txt", # Windows format file path
        "/var/log/*",                          # All files under the file path
        "/var/log/sys*",                       # All files prefixed with sys under the file path
      ]
    
      ## socket currently supports two protocols: tcp/udp. It is recommended to open the intranet port to prevent potential safety hazards
      ## socket and log can only be selected at present, and can not be collected by both file and socket
      socket = [
       "tcp://0.0.0.0:9540"
       "udp://0.0.0.0:9541"
      ]
    
      # File path filtering, using glob rules, meet any one of the filtering conditions will not be collected for the file
      ignore = [""]
      
      # Data source, if empty, 'default' is used by default
      source = ""
      
      # Add tag. If it is empty, $source is used by default
      service = ""
      
      # if pipeline script path is empty, then $source.p will be used; and if $source.p does not exist, then pipeline will be used
      pipeline = ""
      
      # Filter corresponding status:
      #   `emerg`,`alert`,`critical`,`error`,`warning`,`info`,`debug`,`OK`
      ignore_status = []
      
      # Select the code, if there is a misunderstanding in the code, the data cannot be viewed. Default to null:
      #    `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
      character_encoding = ""
      
      ## Set regular expressions, such as ^\d{4}-\d{2}-\d{2} line headers to match the YYYY-MM-DD time format
      ## Data that matches this regular match will be considered valid, otherwise it will be cumulatively appended to the end of the last valid data
      ## Use three single quotation marks '''this-regexp''' to avoid escaping
      ## Regular expression link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
      # multiline_match = '''^\S'''
    
      ## Whether to turn on automatic multiline mode, it will match the applicable multiline rule in the pattern list
      auto_multiline_detection = true
      ## Configure the automatic multiline patterns list, which is an array of multiline rules, i.e. multiple multiline_matches. If it is empty, use the default rule. See the document for details
      auto_multiline_extra_patterns = []

      ## Removes ANSI escape codes from text strings.
      remove_ansi_escape_codes = false

      ## If the data sent failure, will retry forevery.
      blocking_mode = true

      ## Ignore inactive files. For example, files that were last modified 20 minutes ago and more than 10m ago will be ignored
      ## Time unit supports "ms", "s", "m", "h"
      ignore_dead_log = "1h"

      ## Read file from beginning.
      from_beginning = false
    
      ## Custom tags
      [inputs.logging.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
      # ...
    ```

=== "Kubernetes/Docker/Containerd"

    In Kubernetes, once the container collector (container.md) is started, the stdout/stderr logs of each container (including the container under Pod) will be crawled by default. The container logs are mainly configured in the following ways:
    
    - [Adjust container log collection via Annotation/Label](container.md#logging-with-annotation-or-label)
    - [Configure log collection based on container image](container.md#logging-with-image-config)
    - [Collect Pod internal logs in Sidecar form](logfwd.md)

???+ Note "关于 `ignore_dead_log` 的说明"

    If the file is already being collected, but no new log is written within 1 hour, DataKit will close the collection of the file. During this period (1h), the file **cannot** be physically deleted (for example, after `rm`, the file is only marked for deletion, and the file will not be actually deleted until DataKit closes it).

### socket Collection Log {#socket}

Comment out `logfiles` in conf and configure `sockets`. Take log4j2 as an example:

``` xml
 <!-- socket 配置日志传输到本机 9540 端口，protocol 默认 tcp -->
 <Socket name="name1" host="localHost" port="9540" charset="utf8">
     <!-- 输出格式  序列布局-->
     <PatternLayout pattern="%d{yyyy.MM.dd 'at' HH:mm:ss z} %-5level %class{36} %L %M - %msg%xEx%n"/>

     <!--注意：不要开启序列化传输到 socket 采集器上，目前 DataKit 无法反序列化，请使用纯文本形式传输-->
     <!-- <SerializedLayout/>-->
 </Socket>
```

More: For configuration and code examples of Java Go Python mainstream logging components, see: [socket client configuration](logging_socket.md)

### Multiline Log Collection {#multiline}

It can be judged whether a line of logs is a new log by identifying the characteristics of the first line of multi-line logs. If this characteristic is not met, we consider that the current row log is only an append to the previous multi-row log.

For example, logs are written in the top grid in general, but some log texts are not written in the top grid, such as the call stack log when the program crashes. Then, it is a multi-line log for this log text.

In DataKit, we identify multi-line log characteristics through regular expressions. The log line on regular matching is the beginning of a new log, and all subsequent unmatched log lines are considered as appends to this new log until another new log matching regular is encountered.

In `logging.conf`, modify the following configuration:

```toml
multiline_match = '''Fill in the specific regular expression here''' # Note that it is recommended to add three "English single quotation marks" to the regular sides here
```

Regular expression style used in log collector [reference](https://golang.org/pkg/regexp/syntax/#hdr-Syntax){:target="_blank"}

Assume that the original data is:

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

`multiline_match` is configured to `^\\d{4}-\\d{2}-\\d{2}.*` (meaning to match a line header of the form `2020-10-23`)

The cut out three line protocol points are as follows (line numbers are 1/2/8 respectively). You can see that the `Traceback ...` paragraph (lines 3-6) does not form a single log, but is appended to the `message` field of the previous log (line 2).

```
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 1.0" 1611746438938808642
testing,filename=/tmp/094318188 message="2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File \"/usr/local/lib/python3.6/dist-packages/flask/app.py\", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
" 1611746441941718584
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 5.0" 1611746443938917265
```

#### Automatic Multiline Mode {#auto-multiline}

When this function is turned on, each row of log data will be matched in the multi-row list. If the match is successful, the weight of the current multi-line rule is added by one, so that it can be matched more quickly, and then the matching cycle is exited; If there is no match at the end of the whole list, the match is considered to have failed.

Matching success and failure, subsequent operation and normal multi-line log collection are the same: if matching is successful, the existing multi-line data will be sent out and this data will be filled in; If the match fails, it will be appended to the end of the existing data.

Because there are multiple multi-row configurations for the log, their priorities are as follows:

1. `multiline_match` is not empty, only the current rule is used
2. Use source to `multiline_match` mapping configuration (`logging_source_multiline_map` exists only in the container log), using only this rule if the corresponding multiline rule can be found using source
3. Turn on `auto_multiline_detection`, which matches in these multiline rules if `auto_multiline_extra_patterns` is not empty
3. Turn on `auto_multiline_detection` and, if `auto_multiline_extra_patterns` is empty, use the default automatic multiline match rule list, namely:

```
// time.RFC3339, "2006-01-02T15:04:05Z07:00"
`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,

// time.ANSIC, "Mon Jan _2 15:04:05 2006"
`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+ \d+`,

// time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006"
`^[A-Za-z_]+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [\-\+]\d+ \d+`,

// time.UnixDate, "Mon Jan _2 15:04:05 MST 2006"
`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+( [A-Za-z_]+ \d+)?`,

// time.RFC822, "02 Jan 06 15:04 MST"
`^\d+ [A-Za-z_]+ \d+ \d+:\d+ [A-Za-z_]+`,

// time.RFC822Z, "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
`^\d+ [A-Za-z_]+ \d+ \d+:\d+ -\d+`,

// time.RFC850, "Monday, 02-Jan-06 15:04:05 MST"
`^[A-Za-z_]+, \d+-[A-Za-z_]+-\d+ \d+:\d+:\d+ [A-Za-z_]+`,

// time.RFC1123, "Mon, 02 Jan 2006 15:04:05 MST"
`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [A-Za-z_]+`,

// time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ -\d+`,

// time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z07:00"
`^\d+-\d+-\d+[A-Za-z_]+\d+:\d+:\d+\.\d+[A-Za-z_]+\d+:\d+`,

// 2021-07-08 05:08:19,214
`^\d+-\d+-\d+ \d+:\d+:\d+(,\d+)?`,

// Default java logging SimpleFormatter date format
`^[A-Za-z_]+ \d+, \d+ \d+:\d+:\d+ (AM|PM)`,

// 2021-01-31 - with stricter matching around the months/days
`^\d{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])`,
```

#### Restrictions on Processing Very Long Multi-line Logs {#too-long-logs}

At present, a single multi-line log of no more than 32MiB can be processed at most. If the actual multi-line log exceeds 32MiB, DataKit will recognize it as multiple. For example, let's assume that there are several lines of logs as follows, and we want to identify them as a single log:

```log
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
      ...                                 <---- 此处省略 32MiB - 800 字节，加上上面的 4 行，刚好超过 32MiB
        File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
          response = self.full_dispatch_request()
             ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0  <---- 全新的一条多行日志
Traceback (most recent call last):
 ...
```

Here, because of the super-long multi-line log, the first log exceeds 32MiB, DataKit ends this multi-line early, and finally gets three logs:

Number 1: 32MiB of the head

```log
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
      ...                                 <---- 此处省略 32MiB - 800 字节，加上上面的 4 行，刚好超过 32MiB
```

Number 2: Remove the 32MiB in the header, and the rest will become a log independently

```log
        File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
          response = self.full_dispatch_request()
             ZeroDivisionError: division by zero
```

Number 3: The following is a brand-new log:

```log
2020-10-23 06:41:56,688 INFO demo.py 5.0  <---- 全新的一条多行日志
Traceback (most recent call last):
 ...
```

#### Maximum Log Single Line Length {#max-log}

The maximum length of a single line (including after `multiline_match`) is 32MB, whether read from a file or from a socket, and the excess is truncated and discarded.

### Pipeline Configuring and Using {#pipeline}

[Pipeline](../developers/pipeline/index.md) is used primarily to cut unstructured text data, or to extract parts of information from structured text, such as JSON.

For log data, there are two main fields to extract:

- `time`: When the log is generated, if the `time` field is not extracted or parsing this field fails, the current system time is used by default
- `status`: The level of the log, with `stauts` set to `unknown` by default if the `status` field is not extracted

#### Available Log Levels {#status}

Valid `status` field values are as follows (case-insensitive):

| Log Availability Level          | Abbreviation    | Studio Display value |
| ------------          | :----   | ----          |
| `alert`               | `a`     | `alert`       |
| `critical`            | `c`     | `critical`    |
| `error`               | `e`     | `error`       |
| `warning`             | `w`     | `warning`     |
| `notice`              | `n`     | `notice`      |
| `info`                | `i`     | `info`        |
| `debug/trace/verbose` | `d`     | `debug`       |
| `OK`                  | `o`/`s` | `OK`          |

> Note: DataKit sets its status field to `unknown` if the log level (status) does not belong to any of the above (including abbreviations).

Example: Assume the text data is as follows:

```
12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6
```
pipeline script:

```python
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")
grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")
group_in(serverity, ["#"], "warning", status)
cast(pid, "int")
default_time(time)
```

Final result:

```python
{
    "message": "12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6",
    "msg": "Server started, Redis version 3.0.6",
    "pid": 12115,
    "role": "M",
    "serverity": "#",
    "status": "warning",
    "time": 1610127941572000000
}
```

A few considerations for Pipeline:

- Default to `<source-name>.p` if `pipeline` is empty in the logging.conf configuration file (default to `nginx` assuming `source` is `nginx.p`)
- If `<source-name.p>` does not exist, the pipeline feature will not be enabled
- All pipeline script files are stored in the pipeline directory under the DataKit installation path
- If the log file is configured with a wildcard directory, the logging collector will automatically discover new log files to ensure that new log files that meet the rules can be collected as soon as possible

### Introduction of Glob Rules {#grok-rules}

Use glob rules to specify log files more conveniently, as well as automatic discovery and file filtering.

| Wildcard character | Description                                                         | Regular Example | Matching Sample           | Mismatch                    |
| :--                | ---                                                                 | ---             | ---                       | ----                        |
| `*`                | Match any number of any characters, including none                  | `Law*`          | Law, Laws, Lawyer         | GrokLaw, La, aw             |
| `?`                | Match any single character                                          | `?at`           | Cat, cat, Bat, bat        | at                          |
| `[abc]`            | Match a character given in parentheses                              | `[CB]at`        | Cat, Bat                  | cat, bat                    |
| `[a-z]`            | Match a character in the range given in parentheses                 | `Letter[0-9]`   | Letter0, Letter1, Letter9 | Letters, Letter, Letter10   |
| `[!abc]`           | Match a character not given in parentheses                          | `[!C]at`        | Bat, bat, cat             | Cat                         |
| `[!a-z]`           | Match a character that is not within the given range in parentheses | `Letter[!3-5]`  | Letter1…                  | Letter3 … Letter5, Letterxx |

Also, in addition to the glob standard rules described above, the collector also supports `**` recursive file traversal, as shown in the sample configuration. For more information on Grok, see [here](https://rgb-24bit.github.io/blog/2018/glob.html){:target="_blank"}。

### Special Bytecode Filtering for Logs {#ansi-decode}

The log may contain some unreadable bytecodes (such as the color of terminal output, etc.), which can be deleted and filtered by setting `remove_ansi_escape_codes` to true.

<!-- markdownlint-disable MD046 -->
???+ attention

    For such color characters, it is usually recommended that they be turned off in the log output frame rather than filtered by Datakit. Filtering and filtering of special characters is handled by regular expressions, which may not provide comprehensive coverage and have some performance overhead.
<!-- markdownlint-enable -->

The benchmark results are for reference only:

```text
goos: linux
goarch: arm64
pkg: ansi
BenchmarkStrip
BenchmarkStrip-2          653751              1775 ns/op             272 B/op          3 allocs/op
BenchmarkStrip-4          673238              1801 ns/op             272 B/op          3 allocs/op
PASS
```

The processing time of each text increases by 1700 ns. If this function is not turned on, there will be no extra loss.

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.logging.tags]`:

``` toml
 [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### Why can't you see log data on the page? {#why-no-data}

After DataKit is started, the log file configured in `logfiles` ==will be collected only when new logs are generated, and the old log data will not be collected==.

In addition, once a log file is collected, a log will be automatically triggered, which reads as follows:

```
First Message. filename: /some/path/to/new/log ...
```

If you see such information, prove that the specified file ==has started to be collected, but no new log data has been generated at present==. In addition, there is a certain delay in uploading, processing and warehousing log data, and even if new data is generated, it needs to wait for a certain time (< 1min).

### Mutex of Disk Log Collection and Socket Log Collection {#exclusion}

The two collection methods are mutually exclusive at present. When collecting logs in Socket mode, the `logfiles` field in the configuration should be left blank: `logfiles=[]`

### Remote File Collection Scheme {#remote-ntfs}

On linux, you can mount the file path of the host where the log is located to the DataKit host by [NFS mode](https://linuxize.com/post/how-to-mount-an-nfs-share-in-linux/){:target="_blank"}, and the logging collector can configure the corresponding log path.

### MacOS Log Collector Error `operation not permitted` {#mac-no-permission}

In MacOS, because of system security policy, the DataKit log collector may fail to open files, error `operation not permitted`, refer to [apple developer doc](https://developer.apple.com/documentation/security/disabling_and_enabling_system_integrity_protection){:target="_blank"}.

### How to Estimate the Total Amount of Logs {#log-size}

The charge of the log is according to the number of charges, but most of the logs are written to the disk by the program in general, and only the size of the disk occupied (such as 100GB logs per day) can be seen.

A feasible way can be judged by the following simple shell:

```shell
# Count the number of rows in 1GB log
head -c 1g path/to/your/log.txt | wc -l
```

Sometimes, it is necessary to estimate the possible traffic consumption caused by log collection:

```shell
# Count the compressed size of 1GB log (bytes)
head -c 1g path/to/your/log.txt | gzip | wc -c
```

What we get here is the compressed bytes. According to the calculation method of network bits (x8), the calculation method is as follows, so that we can get the approximate bandwidth consumption:

```
bytes * 2 * 8 /1024/1024 = xxx MBit
```

But in fact, the compression ratio of DataKit will not be so high, because DataKit will not send 1GB of data at one time, and it will be sent several times, and this compression ratio is about 85% (that is, 100MB is compressed to 15MB), so a general calculation method is:

```
1GB * 2 * 8 * 0.15/1024/1024 = xxx MBit
```

??? info

    Here `*2` takes into account the actual data inflation caused by [Pipeline cutting](../developers/pipeline/index.md) and the original data should be brought after cutting in general, so according to the worst case, the calculation here is doubled.

## Extended reading {#more-reading}

- [DataKit Log Collection Overview](datakit-logging.md)
- [Pipeline: Text Data Processing](../developers/pipeline/index.md)
- [Pipeline debugging](../developers/pipeline/pipeline-quick-start.md#debug)
- [Pipeline Performance Test and Comparison](logging-pipeline-bench.md)
- [Collect container internal logs via Sidecar (logfwd)](logfwd.md)
- [Configure correctly with regular expressions](datakit-input-conf#debug-regex) 
