{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# {{.InputName}}


日志采集器支持两种模式:
- 从磁盘读取 ：采集文件尾部数据（类似命令行 `tail -f`），上报到观测云。
- socket端口获取：可通过tcp/udp 将文件发送到datakit

> 注意：两种采集方式目前互斥，当需要从socket传输日志时 请修改配置文件 *logfiles=[]*

## 配置

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logging.conf.sample` 并命名为 `logging.conf`。示例如下：

``` toml
[[inputs.logging]]
  # 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定
  # 推荐使用绝对路径
  logfiles = [
    "/var/log/*",                          # 文件路径下所有文件
    "/var/log/sys*",                       # 文件路径下所有以 sys 前缀的文件
    "/var/log/syslog",                     # Unix 格式文件路径
    "C:/path/space 空格中文路径/some.txt", # Windows 风格文件路径
    "/var/log/*",                          # 文件路径下所有文件
    "/var/log/sys*",                       # 文件路径下所有以 sys 前缀的文件
  ]
   ## socket目前支持两种协议：tcp,udp。建议开启内网端口防止安全隐患
   ## socket和log目前是互斥行为，要开启socket采集日志 需要配置logfiles=[]
   socket = [
    	"tcp://0.0.0.0:9540"
    	"udp://0.0.0.0:9541"
  	# only two protocols are supported:TCP and UDP
    ]
  # 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
  ignore = [""]
  
  # 数据来源，如果为空，则默认使用 'default'
  source = ""
  
  # 新增标记tag，如果为空，则默认使用 $source
  service = ""
  
  # pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 pipeline
  pipeline = ""
  
  # 过滤对应 status:
  #   `emerg`,`alert`,`critical`,`error`,`warning`,`info`,`debug`,`OK`
  ignore_status = []
  
  # 选择编码，如果编码有误会导致数据无法查看。默认为空即可:
  #    `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030` or ""
  character_encoding = ""
  
  ## 设置正则表达式，例如 ^\d{4}-\d{2}-\d{2} 行首匹配 YYYY-MM-DD 时间格式
  ## 符合此正则匹配的数据，将被认定为有效数据，否则会累积追加到上一条有效数据的末尾
  ## 使用3个单引号 '''this-regexp''' 避免转义
  ## 正则表达式链接：https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  ## 是否删除 ANSI 转义码，例如标准输出的文本颜色等
  remove_ansi_escape_codes = false
  
  # 自定义 tags
  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

>  注意：DataKit 启动后，`logfiles` 中配置的日志文件有新的日志产生才会采集上来，**老的日志数据是不会采集的**。

### socket采集日志

将logfiles设置为`[]` 并配置socket。以log4j2为例:
``` xml
 <!--socket配置日志传输到本机9540端口，protocol默认tcp-->
 <Socket name="name1" host="localHost" port="9540" charset="utf8">
            <!-- 输出格式  序列布局-->
           <PatternLayout pattern="%d{yyyy.MM.dd 'at' HH:mm:ss z} %-5level %class{36} %L %M - %msg%xEx%n"/>
            <!--注意：不要开启序列化传输到socket采集器上，dk无法反序列化，请使用纯文本形式传输-->
            <!-- <SerializedLayout/>-->
        </Socket>
```


### 多行日志采集

通过识别多行日志的第一行特征，即可判定某行日志是不是一条新的日志。如果不符合这个特征，我们即认为当前行日志只是前一条多行日志的追加。

举例说明一下，一般情况下，日志都是顶格写的，但有些日志文本不是顶格写的，比如程序崩溃时的调用栈日志，那么，对于这种日志文本，就是多行日志。

在 DataKit 中，我们通过正则表达式来识别多行日志特征，正则匹配上的日志行，就是一条新的日志的开始，后续所有不匹配的日志行，都认为是这条新日志的追加，直到遇到另一行匹配正则的新日志为止。 

在 `logging.conf` 中，修改如下配置：

```toml
match = '''这里填写具体的正则表达式''' # 注意，这里的正则俩边，建议分别加上三个「英文单引号」
```

日志采集器中使用的正则表达式风格[参考](https://golang.org/pkg/regexp/syntax/#hdr-Syntax)

假定原数据为：

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

Match 配置为 `^\d{4}-\d{2}-\d{2}.*`（意即匹配形如 `2020-10-23` 这样的行首）

切割出的行协议如下。可以看到 `Traceback ...` 这一行没有单独形成一条（行协议）日志，而是追加在上一条日志中。

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

### Pipeline 配置和使用

[Pipeline](pipeline) 主要用于切割非结构化的文本数据，或者用于从结构化的文本中（如 JSON）提取部分信息。

对日志数据而言，主要提取两个字段：

- `time`：即日志的产生时间，如果没有提取 `time` 字段或解析此字段失败，默认使用系统当前时间
- `status`：日志的等级，如果没有提取出 `status` 字段，则默认将 `stauts` 置为 `info`

有效的 `status` 字段值（不区分大小写）：

| status 有效字段值                | 对应值     |
| :---                             | ---        |
| `a`, `alert`                     | `alert`    |
| `c`, `critical`                  | `critical` |
| `e`, `error`                     | `error`    |
| `w`, `warning`                   | `warning`  |
| `n`, `notice`                    | `notice`   |
| `i`, `info`                      | `info`     |
| `d`, `debug`, `trace`, `verbose` | `debug`    |
| `o`, `s`, `OK`                   | `OK`       |

示例：假定文本数据如下：

```
12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6
```
pipeline 脚本：

```python
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")
grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")
group_in(serverity, ["#"], "warning", status)
cast(pid, "int")
default_time(time)
```

最终结果：

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

Pipeline 的几个注意事项：

- 如果 logging.conf 配置文件中 `pipeline` 为空，默认使用 `<source-name>.p`（假定 `source` 为 `nginx`，则默认使用 `nginx.p`）
- 如果 `<source-name.p>` 不存在，将不启用 pipeline 功能
- 所有 pipeline 脚本文件，统一存放在 DataKit 安装路径下的 pipeline 目录下
- 如果日志文件配置的是通配目录，logging 采集器会自动发现新的日志文件，以确保符合规则的新日志文件能够尽快采集到

### glob 规则简述（图表数据[来源](https://rgb-24bit.github.io/blog/2018/glob.html)）

使用 glob 规则更方便地指定日志文件，以及自动发现和文件过滤。

| 通配符   | 描述                               | 正则示例       | 匹配示例                  | 不匹配                      |
| :--      | ---                                | ---            | ---                       | ----                        |
| `*`      | 匹配任意数量的任何字符，包括无     | `Law*`         | Law, Laws, Lawyer         | GrokLaw, La, aw             |
| `?`      | 匹配任何单个字符                   | `?at`          | Cat, cat, Bat, bat        | at                          |
| `[abc]`  | 匹配括号中给出的一个字符           | `[CB]at`       | Cat, Bat                  | cat, bat                    |
| `[a-z]`  | 匹配括号中给出的范围中的一个字符   | `Letter[0-9]`  | Letter0, Letter1, Letter9 | Letters, Letter, Letter10   |
| `[!abc]` | 匹配括号中未给出的一个字符         | `[!C]at`       | Bat, bat, cat             | Cat                         |
| `[!a-z]` | 匹配不在括号内给定范围内的一个字符 | `Letter[!3-5]` | Letter1…                  | Letter3 … Letter5, Letterxx |

另需说明，除上述 glob 标准规则外，采集器也支持 `**` 进行递归地文件遍历，如示例配置所示。

## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 

## 常见问题

### 远程文件采集方案

在 linux 上，可通过 [NFS 方式](https://linuxize.com/post/how-to-mount-an-nfs-share-in-linux/)，将日志所在主机的文件路径，挂载到 DataKit 主机下，logging 采集器配置对应日志路径即可。

### 日志的特殊字节码过滤

日志可能会包含一些不可读的字节码（比如终端输出的颜色等），可以将 `remove_ansi_escape_codes` 设置为 `true` 对其删除过滤。

此配置可能会影响日志的处理性能，基准测试结果如下：

```
goos: linux
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/test
cpu: Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz
BenchmarkRemoveAnsiCodes
BenchmarkRemoveAnsiCodes-8        636033              1616 ns/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/test       1.056s
```

每一条文本的处理耗时增加 `1616 ns` 不等。如果不开启此功能将无额外损耗。

### MacOS 日志采集器报错 `operation not permitted`

在 MacOS 中，因为系统安全策略的原因，DataKit 日志采集器可能会无法打开文件，报错 `operation not permitted`，解决方法参考 [apple developer doc](https://developer.apple.com/documentation/security/disabling_and_enabling_system_integrity_protection)。

### 更多参考

- pipeline 性能测试和对比[文档](logging-pipeline-bench)
