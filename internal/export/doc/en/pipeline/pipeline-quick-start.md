# Quick Start
---

## First Script {#first-script}

- To configure Pipeline in DataKit, write the following Pipeline file, which is assumed to be named *nginx.p*. Store it in the *[DataKit installation directory]/pipeline* directory.

```python
# Assume input is an Nginx log
# Note that scripts can be commented

grok(_, "some-grok-patterns")  # Perform grok extraction on the input text
rename('client_ip', ip)        # Rename the ip field to client_ip
rename("network_protocol", protocol)   # Rename the protocol field to "network_protocol"

# Replace timestamp (eg 1610967131) with RFC3339 date format: 2006-01-02T15:04:05Z07:00
datetime(access_time, "s", "RFC3339")

url_decode(request_url)      # Translate HTTP request routing into clear text

# When the status_code is between 200 and 300, create a new http_status = "HTTP_OK" field
group_between(status_code, [200, 300], "HTTP_OK", "http_status")

# Drop original content
drop_origin_data()
```

- Configure the corresponding collector to use the above script

Take the logging collector as an example, just configure the field `pipeline_path`. Note that the script name of the Pipeline is configured here, not the path. All the Pipeline scripts referenced here must be stored in the `<DataKit installation directory/pipeline>` directory:

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

    ... # other configuration
```

Restart the collector to cut the corresponding log.

## Debug grok and Pipeline {#debug}

Specify the Pipeline script name and enter a piece of text to determine whether the extraction is successful or not.

> The Pipeline script must be placed in the `[DataKit installation path]/pipeline` directory.

```shell
$ datakit pipeline -P your_pipeline.p -T '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs): # Indicate successful cutting
{
    "code"   : "io/io.go: 458",       # Corresponding code position
    "level"  : "DEBUG",               # Corresponding log level
    "module" : "io",                  # Corresponding code module
    "msg"    : "post cost 6.87021ms", # Pure log attributes
    "time"   : 1610358231887000000    # Log time (Unix nanosecond timestamp)
    "message": "2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms"
}
```

Extraction failure example (only `message` is left, indicating that other fields have not been extracted):

```shell
$ datakit pipeline -P other_pipeline.p -T '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
{
    "message": "2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms"
}
```

> If the debug text is complex, you can write it to a file (sample.log) and debug it as follows:

```shell
datakit pipeline -P your_pipeline.p -F sample.log
```

For more Pipeline debugging commands, see `datakit help pipeline`.

### Grok Wildcard Search {#grokq}

Manual matching is troublesome due to the large number of Grok patterns. DataKit provides an interactive command-line tool, `grokq`（grok query）：

```Shell
datakit tool --grokq
grokq > Mon Jan 25 19:41:17 CST 2021   # Enter the text you want to match here
        2 %{DATESTAMP_OTHER: ?}        # The tool will give corresponding suggestions, and the more accurate the matching month is (the greater the weight is). The previous figures indicate the weights.
        0 %{GREEDYDATA: ?}

grokq > 2021-01-25T18:37:22.016+0800
        4 %{TIMESTAMP_ISO8601: ?}      # Here ? indicates that you need to name the matching text with a field
        0 %{NOTSPACE: ?}
        0 %{PROG: ?}
        0 %{SYSLOGPROG: ?}
        0 %{GREEDYDATA: ?}             # A wide range of patterns like GREEDYDATA have low weights
                                       # The higher the weight, the greater the matching accuracy

grokq > Q                              # Q or exit 
Bye!
```
<!-- markdownlint-disable MD046 -->
???+ note

    In Windows environment, debug in Powershell.
<!-- markdownlint-enable -->
### How to Handle with Multiple Lines {#multiline}

When dealing with some call stack related logs, the logs of the following situations cannot be handled directly with the pattern `GREEDYDATA` since the number of log lines is not fixed:

```txt
2022-02-10 16:27:36.116 ERROR 1629881 --- [scheduling-1] o.s.s.s.TaskUtils$LoggingErrorHandler    : Unexpected error occurred in scheduled task

    java.lang.NullPointerException: null
    at com.xxxxx.xxxxxxxxxxx.xxxxxxx.impl.SxxxUpSxxxxxxImpl.isSimilarPrize(xxxxxxxxxxxxxxxxx.java:442)
    at com.xxxxx.xxxxxxxxxxx.xxxxxxx.impl.SxxxUpSxxxxxxImpl.lambda$getSimilarPrizeSnapUpDo$0(xxxxxxxxxxxxxxxxx.java:595)
    at java.util.stream.ReferencePipeline$3$1.accept(xxxxxxxxxxxxxxxxx.java:193)
    at java.util.ArrayList$ArrayListSpliterator.forEachRemaining(xxxxxxxxx.java:1382)
    at java.util.stream.AbstractPipeline.copyInto(xxxxxxxxxxxxxxxx.java:481)
    at java.util.stream.AbstractPipeline.wrapAndCopyInto(xxxxxxxxxxxxxxxx.java:471)
    at java.util.stream.ReduceOps$ReduceOp.evaluateSequential(xxxxxxxxx.java:708)
    at java.util.stream.AbstractPipeline.evaluate(xxxxxxxxxxxxxxxx.java:234)
    at java.util.stream.ReferencePipeline.collect(xxxxxxxxxxxxxxxxx.java:499)
```

Here you can use the `GREEDYLINES` rule for generalization, such as (*/usr/local/datakit/pipeline/test.p*):

```python
add_pattern('_dklog_date', '%{YEAR}-%{MONTHNUM}-%{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND}%{INT}')
grok(_, '%{_dklog_date:log_time}\\s+%{LOGLEVEL:Level}\\s+%{NUMBER:Level_value}\\s+---\\s+\\[%{NOTSPACE:thread_name}\\]\\s+%{GREEDYDATA:Logger_name}\\s+(\\n)?(%{GREEDYLINES:stack_trace})')

# Remove the message field here for easy debugging
drop_origin_data()
```

Save the above multi-line log as *multi-line.log* and debug it:

```shell
datakit pipeline -P test.p -T "$(<multi-line.log)"
```

The following cutting results are obtained:

```json
{
  "Level": "ERROR",
  "Level_value": "1629881",
  "Logger_name": "o.s.s.s.TaskUtils$LoggingErrorHandler    : Unexpected error occurred in scheduled task",
  "log_time": "2022-02-10 16:27:36.116",
  "stack_trace": "java.lang.NullPointerException: null\n\tat com.xxxxx.xxxxxxxxxxx.xxxxxxx.impl.SxxxUpSxxxxxxImpl.isSimilarPrize(xxxxxxxxxxxxxxxxx.java:442)\n\tat com.xxxxx.xxxxxxxxxxx.xxxxxxx.impl.SxxxUpSxxxxxxImpl.lambda$getSimilarPrizeSnapUpDo$0(xxxxxxxxxxxxxxxxx.java:595)\n\tat java.util.stream.ReferencePipeline$3$1.accept(xxxxxxxxxxxxxxxxx.java:193)\n\tat java.util.ArrayList$ArrayListSpliterator.forEachRemaining(xxxxxxxxx.java:1382)\n\tat java.util.stream.AbstractPipeline.copyInto(xxxxxxxxxxxxxxxx.java:481)\n\tat java.util.stream.AbstractPipeline.wrapAndCopyInto(xxxxxxxxxxxxxxxx.java:471)\n\tat java.util.stream.ReduceOps$ReduceOp.evaluateSequential(xxxxxxxxx.java:708)\n\tat java.util.stream.AbstractPipeline.evaluate(xxxxxxxxxxxxxxxx.java:234)\n\tat java.util.stream.ReferencePipeline.collect(xxxxxxxxxxxxxxxxx.java:499)",
  "thread_name": "scheduling-1"
}
```

### Pipeline Field Naming Notes {#naming}

In all the fields cut out by Pipeline, they are a field rather than a tag. We should not cut out any fields with the same name as tag due to the [line protocol constraint](../../datakit/apis.md#lineproto-limitation). These tags include the following categories:

- [Global Tag](../../datakit/datakit-conf.md#set-global-tag) in DataKit
- [Custom Tag](../../datakit/logging.md#measurements) in Log Collector

In addition, all collected logs have the following reserved fields. We should not override these fields, otherwise the data may not appear properly on the observer page.

| Field Name | Type          | Description                                                                 |
| ---        | ----          | ----                                                                        |
| `source`   | string(tag)   | Log source                                                                  |
| `service`  | string(tag)   | The service corresponding to the log is the same as the `source` by default |
| `status`   | string(tag)   | The [level](../../datakit/logging.md#status)  corresponding to the log      |
| `message`  | string(field) | Original log                                                                |
| `time`     | int           | Timestamp corresponding to log                                              |

<!-- markdownlint-disable MD046 -->
???+ tip

    Of course, we can override the values of these tags by [specific Pipeline function](pipeline-built-in-function.md#fn-set-tag).
<!-- markdownlint-enable -->

Once the Pipeline cut-out field has the same name as the existing Tag (case sensitive), it will cause the following data error. Therefore, it is recommended to bypass these field naming in Pipeline cutting.

```shell
# This error is visible in the DataKit monitor
same key xxx in tag and field
```

### Complete Pipeline Sample {#example}

Take DataKit's own log cutting as an example. DataKit's own log form is as follows:

```txt
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms
```

Write the corresponding Pipeline：

```python
# pipeline for datakit log
# Mon Jan 11 10:42:41 CST 2021
# auth: tanb

grok(_, '%{_dklog_date:log_time}%{SPACE}%{_dklog_level:level}%{SPACE}%{_dklog_mod:module}%{SPACE}%{_dklog_source_file:code}%{SPACE}%{_dklog_msg:msg}')
rename("time", log_time) # rename log_time to time
default_time(time)       # use the time field as the timestamp of the output data
drop_origin_data()       # discard the original log text (not recommended)
```

Several user-defined patterns are referenced, such as `_dklog_date`、`_dklog_level`. We put these rules under `<DataKit installation path>/pipeline/pattern` .

> Note that the user-defined pattern must be placed in the *[DataKit installation path]/pipeline/pattern/* directory) if it needs to be globally effective (that is, applied in other Pipeline scripts):

```Shell
$ cat pipeline/pattern/datakit
# Note: For these custom patterns, it is best to add a specific prefix to the name so as not to conflict with the built-in naming (the built-in pattern name is not allowed to be overwritten)
# Custom pattern format is:
#    <pattern-name><空格><具体 pattern 组合>
_dklog_date %{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}%{INT}
_dklog_level (DEBUG|INFO|WARN|ERROR|FATAL)
_dklog_mod %{WORD}
_dklog_source_file (/?[\w_%!$@:.,-]?/?)(\S+)?
_dklog_msg %{GREEDYDATA}
```

Now that you have the Pipeline and its referenced pattern, you can cut this line of logs through DataKit's built-in Pipeline debugging tool:

```Shell
# Extract successful examples
$ ./datakit pipeline -P dklog_pl.p -T '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs):
{
    "code": "io/io.go:458",
    "level": "DEBUG",
    "module": "io",
    "msg": "post cost 6.87021ms",
    "time": 1610358231887000000
}
```

## FAQ {#faq}
<!-- markdownlint-disable MD013 -->
### Why can't variables be referenced when Pipeline is debugging? {#ref-variables}
<!-- markdownlint-enable -->
Pipeline:

```python
json(_, message, "message")
json(_, thread_name, "thread")
json(_, level, "status")
json(_, @timestamp, "time")
```

The error reported is as follows:

```txt
[E] new piepline failed: 4:8 parse error: unexpected character: '@'
```

---

A: For variables with special characters, you need to decorate them with two `` ` ``:

```python
json(_, `@timestamp`, "time")
```

See [Basic syntax rules of Pipeline](pipeline-platypus-grammar.md)
<!-- markdownlint-disable MD013 -->
### When debugging Pipeline, why can't you find the corresponding Pipeline script? {#pl404}
<!-- markdownlint-enable -->
The order is as follows:

```shell
$ datakit pipeline -P test.p -T "..."
[E] get pipeline failed: stat /usr/local/datakit/pipeline/test.p: no such file or directory
```

---

A: Pipeline scripts for debugging. Place them in *[DataKit installation path]/pipeline* Directory.
<!-- markdownlint-disable MD013 -->
### How to cut logs in many different formats in one Pipeline? {#if-else}
<!-- markdownlint-enable -->
In daily logs, because of different services, logs will take on various forms. At this time, multiple Grok cuts need to be written. In order to improve the running efficiency of Grok, you can give priority to matching the Grok with higher frequency according to the frequency of logs, so that high probability logs can be matched in the previous Groks, avoiding invalid matching.
<!-- markdownlint-disable MD046 -->
???+ note

    In log cutting, Grok matching is the most expensive part, so avoiding repeated Grok matching can greatly improve the cutting performance of Grok.

    ```python
    grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} ...")
    if client_ip != nil {
        # Prove that the above grok has matched at this time, then continue the subsequent processing according to the log
        ...
    } else {
        # Here shows that there is a different log, and the above grok does not match the current log
        grok(_, "%{date2:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg} ...")
    
        if status != nil {
            # Here you can check whether the grok above matches...
        } else {
            # Unrecognized logs, or a grok can be added here to process them, so as to step by step
        }
    }
    ```
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### How to discard field cut? {#drop-keys}

In some cases, all we need is a few fields in the middle of log, but it is difficult to skip the previous parts, such as:

```txt
200 356 1 0 44 30032 other messages
```

Where we only need the value of `44` , which may be code response delay, we can cut it like this (that is, the `:some_field` part is not included in Grok):

```python
grok(_, "%{INT} %{INT} %{INT} %{INT:response_time} %{GREEDYDATA}")
```

### `add_pattern()` Escape Problem {#escape}

When you use `add_pattern()` to add local patterns, you are prone to escape problems, such as the following pattern (used to match file paths and file names):

```txt
(/?[\w_%!$@:.,-]?/?)(\S+)?
```

If we put it in the global pattern directory (that is, *pipeline/pattern* directory), we can write this:

```txt
# my-test
source_file (/?[\w_%!$@:.,-]?/?)(\S+)?
```

If you use `add_pattern()`, you need to write this:

```python
# my-test.p
add_pattern('source_file', '(/?[\\w_%!$@:.,-]?/?)(\\S+)?')
```

That is, the backslash needs to be escaped.
<!-- markdownlint-enable -->
