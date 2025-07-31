---
title     : '日志采集'
summary   : '采集主机上的日志数据'
tags:
  - '日志'
__int_icon      : 'icon/logging'
dashboard :
  - desc  : '日志'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

本文档主要介绍本地磁盘日志采集和 Socket 日志采集：

- 磁盘日志采集 ：采集文件尾部数据（类似命令行 `tail -f`）
- Socket 端口获取：通过 TCP/UDP 方式将日志发送给 DataKit

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logging.conf.sample` 并命名为 `logging.conf`。示例如下：
    
    ``` toml
    {{ CodeBlock .InputSample 4 }}
    ```

    ???+ info "关于 `ignore_dead_log` 的说明"
    
        如果文件已经在采集，但 1h 内没有新日志写入的话，DataKit 会关闭该文件的采集。在这期间（1h），该文件**不能**被物理删除（如 `rm` 之后，该文件只是标记删除，DataKit 关闭该文件后，该文件才会真正被删除）。

=== "Kubernetes/Docker/Containerd"

    在 Kubernetes 中，一旦[容器采集器](container.md)启动，则会默认去抓取各个容器（含 Pod 下的容器）的 stdout/stderr 日志，容器日志主要有以下几个配置方式：

    - [通过 Annotation/Label 调整容器日志采集](container.md#logging-with-annotation-or-label)
    - [根据容器 image 配置日志采集](container.md#logging-with-image-config)
    - [通过 Sidecar 形式采集 Pod 内部日志](logfwd.md)

=== "Windows Event"

    参见 [Windows Event 采集](windows_event.md)。

=== "TCP/UDP"

    在 *logging.conf* 中开启 `sockets` 配置：

    ```toml
      sockets = [
       "tcp://0.0.0.0:9540",
       "udp://0.0.0.0:9541", # or via UDP
      ]
    ```

    然后配置应用日志的输出，以 log4j2 为例：
    
    ``` xml
     <!-- socket 配置日志传输到本机 9540 端口，protocol 默认 TCP -->
     <Socket name="name1" host="localHost" port="9540" charset="utf8">
         <!-- 输出格式  序列布局-->
         <PatternLayout pattern="%d{yyyy.MM.dd 'at' HH:mm:ss z} %-5level %class{36} %L %M - %msg%xEx%n"/>
    
         <!--注意：不要开启序列化传输到 socket 采集器上，目前 DataKit 无法反序列化，请使用纯文本形式传输-->
         <!-- <SerializedLayout/>-->
     </Socket>
    ```

    更多 Java/Go/Python 主流日志组件的配置及代码示例，请参阅 [Socket 日志采集](logging_socket.md)。

<!-- markdownlint-enable -->

## 高级主题 {#deepin-topics}

以下涉及日志采集更深入的一些介绍，如果您感兴趣，可以了解一下。

### 多行日志采集 {#multiline}

通过识别多行日志的第一行特征，即可判定某行日志是不是一条新的日志。如果不符合这个特征，我们即认为当前行日志只是前一条多行日志的追加。

举例说明一下，一般情况下，日志都是顶格写的，但有些日志文本不是顶格写的，比如程序崩溃时的调用栈日志，那么，对于这种日志文本，就是多行日志。

在 DataKit 中，我们通过正则表达式来识别多行日志特征，正则匹配上的日志行，就是一条新的日志的开始，后续所有不匹配的日志行，都认为是这条新日志的追加，直到遇到另一行匹配正则的新日志为止。

在 `logging.conf` 中，修改如下配置：

```toml
multiline_match = ''' 这里填写具体的正则表达式 ''' # 注意，这里的正则俩边，建议分别加上三个「英文单引号」
```

日志采集器中使用的正则表达式风格[参考这里](https://golang.org/pkg/regexp/syntax/#hdr-Syntax){:target="_blank"}。

假定原数据为：

```not-set
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

`multiline_match` 配置为 `^\\d{4}-\\d{2}-\\d{2}.*` 时，（意即匹配形如 `2020-10-23` 这样的行首）

切割出的三个行协议点如下（行号分别是 1/2/8）。可以看到 `Traceback ...` 这一段（第 3 ~ 6 行）没有单独形成一条日志，而是追加在上一条日志（第 2 行）的 `message` 字段中。

```not-set
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 1.0" 1611746438938808642
testing,filename=/tmp/094318188 message="2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File \"/usr/local/lib/python3.6/dist-packages/flask/app.py\", line 2447, in wsgi_app
    response = self.full_dispatch_request()
ZeroDivisionError: division by zero
" 1611746441941718584
testing,filename=/tmp/094318188 message="2020-10-23 06:41:56,688 INFO demo.py 5.0" 1611746443938917265
```

???+ tip "正则表达式性能优化建议"
    - 行首显式添加 `^` —— 精确限定匹配范围，避免不必要的全文回溯
    - 尾部避免使用 `.*` —— 匹配成功后立即终止扫描，减少无效字符遍历
    - 保持表达式简洁 —— 短正则的编译速度和执行效率显著优于复杂长表达式


#### 自动多行模式 {#auto-multiline}

开启此功能后，每一行日志数据都会在多行列表中匹配。如果匹配成功，就将当前的多行规则权重加一，以便后面能更快速的匹配到，然后退出匹配循环；如果整个列表结束依然没有匹配到，则认为匹配失败。

匹配成功与失败，后续操作和正常的多行日志采集是一样的：匹配成功，会将现存的多行数据发送出去，并将本条数据填入；匹配失败，会追加到现存数据的尾端。

因为日志存在多个多行配置，它们的优先级如下：

1. `multiline_match` 不为空，只使用当前规则
1. 使用 source 到 `multiline_match` 的映射配置（只在容器日志中存在 `logging_source_multiline_map`），如果使用 source 能找到对应的多行规则，只使用此规则
1. 开启 `auto_multiline_detection`，如果 `auto_multiline_extra_patterns` 不为空，会在这些多行规则中匹配
1. 开启 `auto_multiline_detection`，如果 `auto_multiline_extra_patterns` 为空，使用默认的自动多行匹配规则列表，即：

```not-set
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

#### 超长多行日志处理的限制 {#too-long-logs}

单条多行日志不超过 DataKit 配置项 `MaxRawBodySize * 0.8`（默认 819KiB) 大小，如果超过这个值，DataKit 会将剩余的日志也拼接起来，即使它们不是有效的多行数据。举例如下，假定有如下多行日志：

```log
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
      ...                                 <---- 此处省略 819KiB
        File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
          response = self.full_dispatch_request()
             ZeroDivisionError: division by zero
2020-10-23 06:41:56,688 INFO demo.py 5.0  <---- 全新的一条多行日志
Traceback (most recent call last):
 ...
```

此处，由于有超长的多行日志，第一条日志超过 819KiB，DataKit 提前结束这条多行，最终得到三条日志：

第一条：即头部的 819KiB

```log
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
      ...                                 <---- 此处省略 819KiB
```

第二条：除去头部的 819KiB，剩余的部分拼接成为一条日志

```log
        File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
          response = self.full_dispatch_request()
             ZeroDivisionError: division by zero
```

第三条：下面一条全新的日志：

```log
2020-10-23 06:41:56,688 INFO demo.py 5.0  <---- 全新的一条多行日志
Traceback (most recent call last):
 ...
```

#### 日志单行最大长度 {#max-log}

无论从文件还是从 TCP/UDP 中读取的日志，单行（包括经过 `multiline_match` 处理后）最大长度默认约 800KiB 左右，超出部分会被分割成多条上报。

### Pipeline 配置和使用 {#pipeline}

[Pipeline](../pipeline/use-pipeline/index.md) 主要用于切割非结构化的文本数据，或者用于从结构化的文本中（如 JSON）提取部分信息。

对日志数据而言，主要提取两个字段：

- `time`：即日志的产生时间，如果没有提取 `time` 字段或解析此字段失败，默认使用系统当前时间
- `status`：日志的等级，如果没有提取出 `status` 字段，则默认将 `stauts` 置为 `info`

#### 可用日志等级 {#status}

有效的 `status` 字段值如下（不区分大小写）：

| 日志可用等级          | 简写    | Studio 显示值 |
| ------------          | :----   | ----          |
| `alert`               | `a`     | `alert`       |
| `critical`            | `c`     | `critical`    |
| `error`               | `e`     | `error`       |
| `warning`             | `w`     | `warning`     |
| `notice`              | `n`     | `notice`      |
| `info`                | `i`     | `info`        |
| `debug/trace/verbose` | `d`     | `debug`       |
| `OK`                  | `o`/`s` | `OK`          |

示例：假定文本数据如下：

```not-set
12115:M 08 Jan 17:45:41.572 # Server started, Redis version 3.0.6
```

Pipeline 脚本：

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
- 如果 `<source-name.p>` 不存在，将不启用 Pipeline 功能
- 所有 Pipeline 脚本文件，统一存放在 DataKit 安装路径下的 Pipeline 目录下
- 如果日志文件配置的是通配目录，logging 采集器会自动发现新的日志文件，以确保符合规则的新日志文件能够尽快采集到

### Glob 规则简述 {#glob-rules}

使用 Glob 规则更方便地指定日志文件，以及自动发现和文件过滤。

| 通配符   | 描述                               | 正则示例       | 匹配示例                    | 不匹配                        |
| :--      | ---                                | ---            | ---                         | ----                          |
| `*`      | 匹配任意数量的任何字符，包括无     | `Law*`         | `Law, Laws, Lawyer`         | `GrokLaw, La, aw`             |
| `?`      | 匹配任何单个字符                   | `?at`          | `Cat, cat, Bat, bat`        | `at`                          |
| `[abc]`  | 匹配括号中给出的一个字符           | `[CB]at`       | `Cat, Bat`                  | `cat, bat`                    |
| `[a-z]`  | 匹配括号中给出的范围中的一个字符   | `Letter[0-9]`  | `Letter0, Letter1, Letter9` | `Letters, Letter, Letter10`   |
| `[!abc]` | 匹配括号中未给出的一个字符         | `[!C]at`       | `Bat, bat, cat`             | `Cat`                         |
| `[!a-z]` | 匹配不在括号内给定范围内的一个字符 | `Letter[!3-5]` | `Letter1…`                  | `Letter3 … Letter5, Letterxx` |

另需说明，除上述 glob 标准规则外，采集器也支持 `**` 进行递归地文件遍历，如示例配置所示。更多 Grok 介绍，参见[这里](https://rgb-24bit.github.io/blog/2018/glob.html){:target="_blank"}。

### 文件读取的偏移位置 {#read-position}

*支持 DataKit [:octicons-tag-24: Version-1.5.5](../datakit/changelog.md#cl-1.5.5) 及以上版本。*

文件读取的偏移是指打开文件后，从哪个位置开始读取。一般是 “首部（head）” 或 “尾部（tail）”。

在 DataKit 中主要是 3 种情况，按照优先级划分如下：

- 优先使用该文件的 position cache，如果能够得到 position 值，且该值小于等于文件大小（说明这是一个没有被 truncated 的文件），使用这个 position 作为读取的偏移位置
- 其次是配置 `from_beginning` 为 `true`，会从文件首部读取
- 最后是默认的 `tail` 模式，即从尾部读取

<!-- markdownlint-disable MD046 -->
???+ info "关于 `position cache` 的说明"

    `position cache` 是日志采集的一项内置功能，它是多个 K/V 键值对，存放在 `cahce/logtail.history` 文件中：

    - key 是根据日志文件路径、inode 等信息生成的唯一值
    - value 是此文件的读取偏移位置（position），并且实施更新

    日志采集在启动时，会根据 key 取得 position 作为读取偏移量，避免漏采和重复采集。
<!-- markdownlint-enable -->

### 日志的特殊字节码处理 {#ansi-decode}

日志可能会包含一些不可读的字节码（比如终端输出的颜色等），可以将 `remove_ansi_escape_codes` 设置为 true 对其删除过滤。

<!-- markdownlint-disable MD046 -->
??? warning "颜色字符处理会带来额外的采集开销"

    对于此类颜色字符，通常建议在日志输出框架中关闭，而不是由 DataKit 进行过滤。特殊字符的筛选和过滤是由正则表达式处理，可能覆盖不够全面，且有一定的性能开销。

    处理性能基准测试结果如下，仅供参考：
    
    ```text
    goos: linux
    goarch: arm64
    pkg: ansi
    BenchmarkStrip
    BenchmarkStrip-2  653751  1775 ns/op  272 B/op  3 allocs/op
    BenchmarkStrip-4  673238  1801 ns/op  272 B/op  3 allocs/op
    PASS
    ok      ansi      2.422s
    ```
    
    每一条文本的处理耗时增加 1700 ns 不等。如果不开启此功能将无额外损耗。
<!-- markdownlint-enable -->

### 根据白名单保留指定字段 {#field-whitelist}

容器日志采集有以下基础字段：

| 字段名           |
| -----------      |
| `service`        |
| `status`         |
| `filepath`       |
| `log_read_lines` |

在特殊场景下，很多基础字段不是必要的。现在提供一个白名单（whitelist）功能，只保留指定的字段。

字段白名单配置例如 `'["service", "filepath"]'`，具体细节如下：

- 如果 whitelist 为空，则添加所有基础字段
- 如果 whitelist 不为空，且值有效，例如 `["service", "filepath"]`，则只保留这两个字段
- 如果 whitelist 不为空，且全部是无效字段，例如 `["no-exist"]` 或 `["no-exist-key1", "no-exist-key2"]`，则这条数据被丢弃

对于其他来源的 tags 字段，有以下几种情况：

- whitelist 对 DataKit 的全局标签（`global tags`）不生效
- 通过 `ENV_ENABLE_DEBUG_FIELDS = "true"` 开启的 debug 字段不受影响，包括日志采集的 `log_read_offset` 和 `log_file_inode` 两个字段，以及 `pipeline` 的 debug 字段

## 数据字段 {#logging}

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

{{$m.MarkdownTable}}

{{ end }}

<!-- markdownlint-disable MD053 -->
[^1]: 早期 Pipeline 实现的时候，只能切割出 field，而 status 大部分都是通过 Pipeline 切割出来的，故将其归类到 field 中。但语义上，它应该属于 tag 的范畴。
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
## FAQ {#faq}

### 为什么在页面上看不到日志数据？ {#why-no-data}

DataKit 启动后，`logfiles` 中配置的日志文件**有新的日志产生才会采集上来，老的日志数据是不会采集的**。

另外，一旦开始采集某个日志文件，将会自动触发一条日志，内容大概如下：

``` not-set
First Message. filename: /some/path/to/new/log ...
```

如果看到这样的信息，说明指定的文件「已经开始采集，只是当前尚无新的日志数据产生」。另外，日志数据的上传、处理、入库也有一定的时延，即使有新的数据产生，也需要等待一定时间（< 1min）。

### 磁盘日志采集和 Socket 日志采集的互斥性 {#exclusion}

两种采集方式目前互斥，当以 Socket 方式采集日志时，需将配置中的 `logfiles` 字段置空：`logfiles=[]`

### 远程文件采集方案 {#remote-ntfs}

在 Linux 上，可通过 [NFS 方式](https://linuxize.com/post/how-to-mount-an-nfs-share-in-linux/){:target="_blank"}，将日志所在主机的文件路径，挂载到 DataKit 主机下，logging 采集器配置对应日志路径即可。

### MacOS 日志采集器报错 `operation not permitted` {#mac-no-permission}

在 MacOS 中，因为系统安全策略的原因，DataKit 日志采集器可能会无法打开文件，报错 `operation not permitted`，解决方法参考 [apple developer doc](https://developer.apple.com/documentation/security/disabling_and_enabling_system_integrity_protection){:target="_blank"}。

### 如何估算日志的总量 {#log-size}

由于日志的收费是按照条数来计费的，但一般情况下，大部分的日志都是程序写到磁盘的，只能看到磁盘占用大小（比如每天 100GB 日志）。

一种可行的方式，可以用以下简单的 shell 来判断：

```shell
# 统计 1GB 日志的行数
head -c 1g path/to/your/log.txt | wc -l
```

有时候，要估计一下日志采集可能带来的流量消耗：

```shell
# 统计 1GB 日志压缩后大小（字节）
head -c 1g path/to/your/log.txt | gzip | wc -c
```

这里拿到的是压缩后的字节数，按照网络 bit 的计算方法（x8），其计算方式如下，以此可拿到大概的带宽消耗：

``` not-set
bytes * 2 * 8 /1024/1024 = xxx MBit
```

但实际上 DataKit 的压缩率不会这么高，因为 DataKit 不会一次性发送 1GB 的数据，而且分多次发送的，这个压缩率在 85% 左右（即 100MB 压缩到 15MB），故一个大概的计算方式是：

``` not-set
1GB * 2 * 8 * 0.15/1024/1024 = xxx MBit
```

<!-- markdownlint-disable MD046 -->
??? info

    此处 `*2` 考虑到了 [Pipeline 切割](../pipeline/use-pipeline/index.md)导致的实际数据膨胀，而一般情况下，切割完都是要带上原始数据的，故按照最坏情况考虑，此处以加倍方式来计算。
<!-- markdownlint-enable -->

## 延伸阅读 {#more-reading}

- [DataKit 日志采集综述](datakit-logging.md)
- [Pipeline: 文本数据处理](../pipeline/use-pipeline/index.md)
- [Pipeline 调试](../developers/datakit-pl-how-to.md)
- [Pipeline 性能测试和对比](logging-pipeline-bench.md)
- [通过 Sidecar(logfwd) 采集容器内部日志](logfwd.md)
- [正确使用正则表达式来配置](../datakit/datakit-input-conf.md#debug-regex)
