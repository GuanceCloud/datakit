
# How to Write Pipeline Scripts
---

Pipeline writing is relatively troublesome, so Datakit has built-in simple debugging tools to help you write Pipeline scripts.

## Debug grok and pipeline {#debug}

Specify the pipeline script name and enter a piece of text to determine whether the extraction is successful or not.

> The pipeline script must be placed in the *[Datakit 安装目录]/pipeline* directory.

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
$ datakit pipeline -P your_pipeline.p -F sample.log
```

For more Pipeline debugging commands, see `datakit help pipeline`.

### Grok Wildcard Search {#grokq}

Manual matching is troublesome due to the large number of Grok patterns. Datakit provides an interactive command-line tool, `grokq`（grok query）：

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

???+ attention

    In Windows environment, debug in Powershell.

### How to Handle with Multiple Lines {#multiline}

When dealing with some call stack related logs, the logs of the following situations cannot be handled directly with the pattern `GREEDYDATA` since the number of log lines is not fixed:

```
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
grok(_, '%{_dklog_date:log_time}\\s+%{LOGLEVEL:Level}\\s+%{NUMBER:Level_value}\\s+---\\s+\\[%{NOTSPACE:thread_name}\\]\\s+%{GREEDYDATA:Logger_name}\\s+(\\n)?(%{GREEDYLINES:stack_trace})'

# Remove the message field here for easy debugging
drop_origin_data()
```

Save the above multi-line log as *multi-line.log* and debug it:

```shell
$ datakit pipeline -P test.p -T "$(<multi-line.log)"
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

In all the fields cut out by Pipeline, they are a field rather than a tag. We should not cut out any fields with the same name as tag due to the [line protocol constraint](apis.md#lineproto-limitation). These tags include the following categories:

- [Global Tag](datakit-conf.md#set-global-tag) in Datakit
- [Custom Tag](logging.md#measurements) in Log Collector

In addition, all collected logs have the following reserved fields. We should not override these fields, otherwise the data may not appear properly on the observer page.

| Field Name | Type          | Description                                                                  |
| ---        | ----          | ----                                                                         |
| `source`   | string(tag)   | Log source                                                                   |
| `service`  | string(tag)   | The service corresponding to the log is the same as the `service` by default |
| `status`   | string(tag)   | The [level](logging.md#status)  corresponding to the log                     |
| `message`  | string(field) | Original log                                                                 |
| `time`     | int           | Timestamp corresponding to log                                               |


???+ tip

    Of course, we can override the values of these tags by [specific Pipeline function](pipeline.md#fn-set-tag).

Once the Pipeline cut-out field has the same name as the existing Tag (case sensitive), it will cause the following data error. Therefore, it is recommended to bypass these field naming in Pipeline cutting.

```shell
# This error is visible in the Datakit monitor
same key xxx in tag and field
```

### Complete Pipeline Sample {#example}

Take Datakit's own log cutting as an example. Datakit's own log form is as follows:

```
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms
```

Write the corresponding pipeline：

```python
# pipeline for datakit log
# Mon Jan 11 10:42:41 CST 2021
# auth: tanb

grok(_, '%{_dklog_date:log_time}%{SPACE}%{_dklog_level:level}%{SPACE}%{_dklog_mod:module}%{SPACE}%{_dklog_source_file:code}%{SPACE}%{_dklog_msg:msg}')
rename("time", log_time) # rename log_time to time
default_time(time)       # use the time field as the timestamp of the output data
drop_origin_data()       # discard the original log text (not recommended)
```

Several user-defined patterns are referenced, such as `_dklog_date`、`_dklog_level`. We put these rules under `<datakit安装目录>/pipeline/pattern` .

> Note that the user-defined pattern must be placed in the *[Datakit 安装目录]/pipeline/pattern/* directory) if it needs to be globally effective (that is, applied in other pipeline scripts):

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

Now that you have the pipeline and its referenced pattern, you can cut this line of logs through Datakit's built-in pipeline debugging tool:

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

### :material-chat-question: Why can't variables be referenced when Pipeline is debugging? {#ref-variables}

Pipeline:

```python
json(_, message, "message")
json(_, thread_name, "thread")
json(_, level, "status")
json(_, @timestamp, "time")
```

The error reported is as follows:

```
[E] new piepline failed: 4:8 parse error: unexpected character: '@'
```

---

A: For variables with special characters, you need to decorate them with two `` ` ``:

```python
json(_, `@timestamp`, "time")
```

See [Basic syntax rules of Pipeline](pipeline.md#basic-syntax)

### :material-chat-question: When debugging Pipeline, why can't you find the corresponding Pipeline script? {#pl404}

The order is as follows:

```shell
$ datakit pipeline -P test.p -T "..."
[E] get pipeline failed: stat /usr/local/datakit/pipeline/test.p: no such file or directory
```

---

A: Pipeline scripts for debugging. Place them in *<Datakit 安装目录>/pipeline* Directory.

### :material-chat-question: How to cut logs in many different formats in one Pipeline? {#if-else}

In daily logs, because of different services, logs will take on various forms. At this time, multiple Grok cuts need to be written. In order to improve the running efficiency of Grok, you can give priority to matching the Grok with higher frequency according to the frequency of logs, so that high probability logs can be matched in the previous Groks, avoiding invalid matching.

???+ attention

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

### :material-chat-question: How to discard field cut? {#drop-keys}

In some cases, all we need is a few fields in the middle of log, but it is difficult to skip the previous parts, such as: 

```
200 356 1 0 44 30032 other messages
```

Where we only need the value of `44` , which may be code response delay, we can cut it like this (that is, the `:some_field` part is not included in Grok):

```python
grok(_, "%{INT} %{INT} %{INT} %{INT:response_time} %{GREEDYDATA}")
```

### :material-chat-question: `add_pattern()` Escape Problem {#escape}

When you use `add_pattern()` to add local patterns, you are prone to escape problems, such as the following pattern (used to match file paths and file names):

```
(/?[\w_%!$@:.,-]?/?)(\S+)?
```

If we put it in the global pattern directory (that is, *pipeline/pattern* directory), we can write this:

```
# my-test
source_file (/?[\w_%!$@:.,-]?/?)(\S+)?
```

If you use `add_pattern()`, you need to write this:

```python
# my-test.p
add_pattern('source_file', '(/?[\\w_%!$@:.,-]?/?)(\\S+)?')
```

That is, the backslash needs to be escaped.
