{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Pipeline 调试

Pipeline 编写较为麻烦，为此，DataKit 中内置了简单的调试工具，用以辅助大家来编写 Pipeline 脚本。

## 调试 grok 和 pipeline

指定 pipeline 脚本名称（`--pl`，pipeline 脚本必须放在 `<DataKit 安装目录>/pipeline` 目录下），输入一段文本（`--txt`）即可判断提取是否成功

```shell
datakit --pl your_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs): # 表示切割成功
{
	"code"   : "io/io.go: 458",       # 对应代码位置
	"level"  : "DEBUG",               # 对应日志等级
	"module" : "io",                  # 对应代码模块
	"msg"    : "post cost 6.87021ms", # 纯日志内容
	"time"   : 1610358231887000000    # 日志时间(Unix 纳秒时间戳)
}

# 提取失败示例
datakit --pl other_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
No data extracted from pipeline
```

> 如果调试文本比较复杂，可以将它们写入一个文件（sample.log），用如下方式调试：

```shell
datakit --pl your_pipeline.p --txt "$(< sample.log)"
```

由于 grok pattern 数量繁多，人工匹配较为麻烦。DataKit 提供了交互式的命令行工具 `grokq`（grok query）：

```Shell
datakit --grokq
grokq > Mon Jan 25 19:41:17 CST 2021   # 此处输入你希望匹配的文本
        2 %{DATESTAMP_OTHER: ?}        # 工具会给出对应对的建议，越靠前匹配月精确（权重也越大）。前面的数字表明权重。
        0 %{GREEDYDATA: ?}

grokq > 2021-01-25T18:37:22.016+0800
        4 %{TIMESTAMP_ISO8601: ?}      # 此处的 ? 表示你需要用一个字段来命名匹配到的文本
        0 %{NOTSPACE: ?}
        0 %{PROG: ?}
        0 %{SYSLOGPROG: ?}
        0 %{GREEDYDATA: ?}             # 像 GREEDYDATA 这种范围很广的 pattern，权重都较低
                                       # 权重越高，匹配的精确度越大

grokq > Q                              # Q 或 exit 退出
Bye!
```

> 注：Windows 下，请在 Powershell 中执行调试。

### 多行如何处理

在处理一些调用栈相关的日志时，由于其日志行数不固定，直接用 `GREEDYDATA` 这个 pattern 无法处理如下情况的日志：

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

此处可以使用 `GREEDYLINES` 规则来通配，如（*/usr/local/datakit/pipeline/test.p*）：

```python
add_pattern('_dklog_date', '%{YEAR}-%{MONTHNUM}-%{MONTHDAY} %{HOUR}:%{MINUTE}:%{SECOND}%{INT}')
grok(_, '%{_dklog_date:log_time}\\s+%{LOGLEVEL:Level}\\s+%{NUMBER:Level_value}\\s+---\\s+\\[%{NOTSPACE:thread_name}\\]\\s+%{GREEDYDATA:Logger_name}\\s+(\\n)?(%{GREEDYLINES:stack_trace})'

# 此处移除 message 字段便于调试
drop_origin_data()
```

将上述多行日志存为 *multi-line.log*，调试一下：

```shell
datakit --pl test.p --txt "$(<multi-line.log)"
```

得到如下切割结果：

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

### 完整 Pipeline 示例

这里以 DataKit 自身的日志切割为例。DataKit 自身的日志形式如下：

```
2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms
```

编写对应 pipeline：

```python
# pipeline for datakit log
# Mon Jan 11 10:42:41 CST 2021
# auth: tanb

grok(_, '%{_dklog_date:log_time}%{SPACE}%{_dklog_level:level}%{SPACE}%{_dklog_mod:module}%{SPACE}%{_dklog_source_file:code}%{SPACE}%{_dklog_msg:msg}')
rename("time", log_time) # 将 log_time 重名命名为 time
default_time(time)       # 将 time 字段作为输出数据的时间戳
drop_origin_data()       # 丢弃原始日志文本(不建议这么做)
```

这里引用了几个用户自定义的 pattern，如 `_dklog_date`、`_dklog_level`。我们将这些规则存放 `<datakit安装目录>/pipeline/pattern` 下（**注意，用户自定义 pattern 如果需要全局生效，必须放置在 `<DataKit安装目录/pipeline/pattern/>` 目录下**）:

```Shell
$ cat pipeline/pattern/datakit
# 注意：自定义的这些 pattern，命名最好加上特定的前缀，以免跟内置的命名冲突（内置 pattern 名称不允许覆盖）
# 自定义 pattern 格式为：
#    <pattern-name><空格><具体 pattern 组合>
_dklog_date %{YEAR}-%{MONTHNUM}-%{MONTHDAY}T%{HOUR}:%{MINUTE}:%{SECOND}%{INT}
_dklog_level (DEBUG|INFO|WARN|ERROR|FATAL)
_dklog_mod %{WORD}
_dklog_source_file (/?[\w_%!$@:.,-]?/?)(\S+)?
_dklog_msg %{GREEDYDATA}
```

现在 pipeline 以及其引用的 pattern 都有了，就能通过 DataKit 内置的 pipeline 调试工具，对这一行日志进行切割：

```Shell
# 提取成功示例
$ ./datakit --pl dklog_pl.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs):
{
    "code": "io/io.go:458",
    "level": "DEBUG",
    "module": "io",
    "msg": "post cost 6.87021ms",
    "time": 1610358231887000000
}
```

### Pipeline 字段命名注意事项

由于[行协议约束](apis#f54b954f)，在切割出来的字段中（在行协议中，它们都是 Field），不宜有任何 tag 字段，这些 Tag 包含如下几类：

- 各个具体采集器中，用户自行配置增加的 Tag，如 `[inputs.nginx.tags]` 下可增加各种 Tag
- DataKit 全局 Tag，如 `host`。当然，这个全局 Tag 用户也能自行配置
- 日志采集器默认会增加 `source/service` 这两个 Tag，在 Pipeline 中也不宜出现这两个字段切割

一旦 Pipeline 切割出来的字段中带有上述任何一个 Tag key（大小写敏感），都会导致如下数据报错，故建议在 Pipeline 切割中，绕开这些字段命名。

```shell
# 该错误在 DataKit monitor 中能看到
same key xxx in tag and field
```
