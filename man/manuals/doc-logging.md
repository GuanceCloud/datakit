# Datakit 日志采集系统的设计和实现

## 前言

日志采集（logging）是观测云 Datakit 重要的一项，它将或主动采集、或被动接收的日志数据加以处理，最终上传到观测云中心。日志采集按照数据来源可以分为 “网络流数据” 和 “本地磁盘文件” 两种。

### 网络流数据

基本都是以订阅网络接口的方式，被动接收日志产生端的发送过来的数据。

最常见的例子就是查看 Docker 日志，当执行 `docker logs -f CONTAIENR_NAME` 命令时，Docker 会启动一个单独的进程并连接到主进程，接收主进程发送过来的数据输出到终端。虽然 Docker 日志进程和主进程在同一台主机，但是它们的交互是通过本地环回网络。

更加复杂的网络日志场景比如 kubenetes 集群，它们的日志分布在不同 Node 上面，需要以 api-server 进行中转，比 Docker 的单一访问链路复杂一倍。

但是大部分通过网络获取日志都存在一个问题 —— 无法指定日志位置。日志接收端只能选择从首部开始接收日志，有多少收多少，可能一次收到几十万条；或从尾部开始，类似 `tail -f` 只接收当前产生的最新的数据，如果日志接收端的进程重启，那么这期间的日志就丢失了。

Datakit 的容器日志采集最初是使用网络接收的方式，被上述问题困扰许久，后通过逐步调整改为下文提到的 “本地磁盘文件” 采集方式。

### 本地磁盘文件

采集本地日志文件是最常见和最高效的方式，省去了中间复杂的传输步骤，直接对磁盘文件进行访问，可操控性更高，但是实现更复杂，会遇到一系列的细节问题，比如：

- 怎样在磁盘上读取数据更高效？
- 文件被删除或者执行翻转（rotate）该怎么办？
- 重新打开文件时该怎样定位上次的位置再进行读取？

这些问题等同是将 Docker 日志服务给铺展开，各种细节和执行都交由自己来处理，只省去最后的网络传输部分，实现的复杂度比单纯用网络接收要麻烦很多。

本文将主要针对 “本地磁盘文件”，自底向上，分为 “发现文件”、“采集数据并处理”、“发送和同步” 三个方面，依次介绍 Datakit 日志采集系统的设计和实现细节。

补充，Datakit 日志采集执行流如下，涵盖和细分了上述的 “三个方面” ：

```
    glob 发现文件       Docker API 发现文件      Containerd（CRI）发现文件
         |                       |                            |
         ------------------------------------------------------
                                 |
                   添加到日志调度器，分配到指定 lines
                                 |
         ---------------------------------------------------
         |                |                |               |
       line1            line2            line3          line4
                          |
                          |          |- 采集数据，分行    <--|
                          |          |                       |
                          |          |- 数据转码             |
               |----->    |          |                       |
               |          |          |- 特殊字符处理         |
               |          |-  文件 A |                       |
               |          |          |- 多行处理             |
               |          |          |                       | 一个采集周期
               |          |          |- Pipeline 处理        |
               |          |          |                       |
               |          |          |- 发送                 |
               |          |          |                       |
               |          |          |- 同步文件采集位置     |
               |          |          |                       |
               | 流水线   |          |- 文件状态检测      ---|
               | 循环     |
               |          |
               |          |-  文件 B |-
               |          |
               |          |
               |          |-  文件 C |-
               |          |
               |----------|
```

## 发现和定位日志文件

既然要读取和采集日志文件，那么首先要在磁盘上定位文件位置。在 Datakit 中主要有三种文件日志，其中两种容器日志，一种普通日志，它们的采集方式大同小异，本文也主要介绍这种三种，它们分别是：

- 普通日志文件
- Docker Stdout/Stderr，由 Docker 服务本身进行日志管理和落盘（Datakit 目前只支持解析 `json-file` 驱动）
- Containerd Stdout/Stderr，Containerd 没有输出日志的策略，现阶段的 Containerd Stdout/Stderr 都是由 Kubenetes 的 kubelet 组件进行管理，后续会统称为 `Containerd（CRI）`

### 发现普通日志文件

普通日志文件是最常见的一种，它们是进程直接将可读的记录数据写到磁盘文件，像著名的 “log4j” 框架或者执行 `echo "this is log" >> /tmp/log` 命令都会产生日志文件。

这种日志的文件路径大部分情况都是固定的，像 MySQL 在 Linux 平台的日志路径是 `/var/log/mysql/mysql.log`，如果运行 Datakit MySQL 采集器，默认会去找个路径找寻日志文件。但是日志存储路径是可配的，Datakit 无法兼顾所有情况，所以必须支持手动指定文件路径。

在 Datakit 中使用 [glob](https://en.wikipedia.org/wiki/Glob_(programming) 模式配置文件路径，它使用通配符来定位文件名（当然也可以不使用通配符）。

举个例子，现在有以下的文件：

```
$ tree /tmp
/tmp
├── datakit
│   ├── datakit-01.log
│   ├── datakit-02.log
│   └── datakit-03.log
└── mysql.d
    └── mysql
        └── mysql.log

3 directories, 4 files
```

在 Datakit logging 采集器中可以通过配置 `logfiles` 参数项，指定要采集的日志文件，比如：

- 采集 `datakit` 目录下所有文件，glob 为`/tmp/datakit/*`
- 采集所有带有 `datakit` 名字的文件，对应的 glob 为`/tmp/datakit/datakit-*log`
- 采集 `mysql.log`，但是中间有 `mysql.d` 和 `mysql` 两层目录，有好几种方法定位到 `mysql.log` 文件：
   - 直接指定：`/tmp/mysql.d/mysql/mysql.log`
   - 单星号指定：`/tmp/*/*/mysql.log`，这种方法基本用不到
   - 双星号（`double star`）：`/tmp/**/mysql.log`，使用双星号 `**` 代替中间的多层目录结构，是较为简洁、常用的一种方式

在配置文件中使用 glob 指定文件路径后，Datakit 会定期在磁盘中搜寻符合规则的文件，如果发现没有在采集列表中，便将其添加并进行采集。

### 定位容器 Stdout/Stderr 日志文件

在容器中输出日志有两种方式：

- 一是直接写到挂载的磁盘目录，这种方式在主机看来和上述的 “普通日志文件” 相同，都是在磁盘固定位置的文件
- 另一种方式是输出到 Stdout/Stderr，由容器的 runtime 来收集并管理落盘，这也是较为常见的方式。这个落盘路径通过访问 runtime API 可以获取到

Datakit 通过连接 Docker 或 Containerd 的 sock 文件，访问它们的 API 获取指定容器的 `LogPath`，类似在命令行执行 `docker inspect --format='{{.LogPath}}' $INSTANCE_ID`：

```
$ docker inspect --format='{{.LogPath}}' cf681e
/var/lib/docker/containers/cf681eXXXX/cf681eXXXX-json.log
```

获取到容器 `LogPath` 后，使用这个路径和相应配置创建日志采集。

## 采集日志数据并加以处理

### 日志采集调度器

在拿到一个日志文件路径后，因为 Datakit 使用 Golang 编写的，通常会选择开启一个或多个 goroutine 独立对文件进行采集，模型简单、实现方便，Datakit 在之前确实是这么做的。

但是如果文件数量太多，要开启的 goroutine 数量也会随之增加，这对 goroutine 的管理很不利。所以 Datakit 实现了一个日志采集的调度器。

和大多数调度器模型一样，Datakit 会在调度器下层实现多条流水线（line）。当一个新的日志采集注册到调度器，Datakit 会根据各条流水线的权重进行分配。

每条流水线都是循环执行，即 A 文件采集一次（或是连续采集 N 秒，视情况而定），再采集 B 文件，再采集 C 文件，这样可以对 goroutine 数量有效控制，而且避免了大量 goroutine 的底层调度和资源争夺。

如果发现这个日志采集出现错误，可能会将其从流水线撤下，下次循环就不会再采集。

### 读取数据并分割成行

提到读取日志数据，大部分情况都会先想到类似 `Readline()` 这种方法函数，每次调用都会返回完整的一行日志。但是在 Datakit 没有这样实现。

为了确保更细致的操控和更高的性能，Datakit 只使用了最基础的 `Read()` 方法，每次读取 4KiB 数据（buff 大小是 4KiB，实际读取到可能更少），手动将这 4KiB 数据通过换行符 `\n` 分割成 N 份。这样会出现两种情况：

- 这 4KiB 数据最后一个字符刚好是换行符，可以分割成 N 份，没有剩余
- 这 4KiB 数据最后一个字符不是换行符，对比上文，本次的分割只有 N-1 份，有剩余部分，这段剩余的部分将补充到下一个 4KiB 数据的首部，依次类推

在 Datakit 的代码中，此处对同一个 buff 不断进行 `update CursorPosition`、`copy` 和 `truncate`，以实现最大化的内存复用，此处不过多提及。

经过处理，读取到的数据已经变成一行一行，可以走向执行流的下一层也就是转码和特殊字符处理。

### 转码和特殊字符处理

转码和特殊字符的处理要在数据成型之后再进行，否则会出现从字符中间截断、对一段截断的数据做处理的情况。比如一个 UTF-8 的中文字符占 3 字节，在采集到第 1 个字节时就做转码处理，这属于 Undefined 行为。

数据转码是很常见的行为，需要指定编码类型和大端小端（如果有），本文重点讲述一下 “特殊字符处理”。

“特殊字符” 在此处代指数据中的颜色字符，比如以下命令会在命令行终端输出一个红色 `rea` 单词：

```
$ RED='\033[0;31m' && NC='\033[0m' && print "${RED}red${NC}"
```

如果不进行处理，不删除颜色字符，那么最终日志数据也会带有 `\033[0;31m`，不仅缺乏美观、占用存储，而且可能对后续的数据处理产生负影响。所以要在此处筛除特殊颜色字符。

开源社区有许多案例，但是大部分都使用正则表达式进行实现，性能一般。

但是对于一个成熟的日志输出框架，一定有关闭颜色字符的方法，Datakit 更推荐这种做法。

### 解析行数据

“解析行数据” 主要是针对容器 Stdout/Stderr 日志。容器 runtime 管理和落盘日志时会添加一些额外的信息字段，比如产生时间，来源是 `stdout` 还是 `stderr`，本条日志是否被截断等等。Datakit 需要对这种数据做解析，提取对应字段。

- Docker 日志单条格式如下，是简单的 JSON 格式，正文在 `log` 字段中。如果 `log` 内容的结尾是 `\n` 表示这一行数据是完整的，没有被截断；如果不是 `\n`，则表明数据太长超过 16KB 被截断了，其剩余部分在下一个 JSON 中。
    ```
    {"log":"2022/09/14 15:11:11 Bash For Loop Examples. Hello, world! Testing output.\n","stream":"stdout","time":"2022-09-14T15:11:11.125641305Z"}
    ```
- Containerd（CRI）单条日志格式如下，各项字段以空格分割。和 Docker 相同的是，Containerd（CRI）也有日志截断的标记，即第三个字段 `P`，此外还有 `F`。`P` 表示 `Partia`，即不完整的、被截断的；`F` 表示 `Full`。 
    ```
    2016-10-06T00:17:09.669794202Z stdout P log content 1
    2016-10-06T00:17:09.669794202Z stdout F log content 2
    ```
    拼接之后的日志数据是 `log content 1 log content 2`。

通过解析行数据，可以获得日志正文、stdout/sterr 等信息，根据标记确定是否是不完整的截断日志，要进行日志拼接。在普通日志文件中不存在截断，文件中的单行数据理论上可以无限长。

此外，日志单行被截断，拼接之后也属于一行日志，而不是下文要提到的多行日志，这是两个不同的概念。

### 多行数据

多行处理是日志采集非常重要的一项，它将一些不符合特征的数据，在不丢失数据的前提下变得符合特征。比如日志文件中有以下数据，这是一段常见的 Python 栈打印：

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]
Traceback (most recent call last):
  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

如果没有多行处理，那么最终数据就是以上 7 行，和原文一模一样。这不利于后续的 Pipeline 切割，因为像第 3 行的 `Traceback (most recent call last):` 或 第 4 行的 `File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app` 都不是固定格式。

如果经过有效的多行处理，这 7 行数据会变成 3 行，结果如下：

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
2020-10-23 06:54:20,164 ERROR /usr/local/lib/python3.6/dist-packages/flask/app.py Exception on /0 [GET]\nTraceback (most recent call last):\n  File "/usr/local/lib/python3.6/dist-packages/flask/app.py", line 2447, in wsgi_app\n    response = self.full_dispatch_request()
2020-10-23 06:41:56,688 INFO demo.py 5.0
```

可以看到，现在每行日志数据都以 `2020-10-23` 这样的特征字符串开头，原文中不符合特征的第 3、4、5 行被追加到第 2 行的末尾。这样看起来要美观很多，而且有利于后续的 Pipeline 字段切割。

这一功能并不复杂，只需要指定特征字符串的正则表达式即可。

在 Datakit logging 采集器配置有 `multiline_match` 项，以上文的例子，该项的配置应该是 `^\d{4}-\d{2}-\d{2}`，即匹配形如 `2020-10-23` 这样的行首。

具体实现上类似一个数据结构中的栈（stack）结构，符合特征就将前一条出栈并再把自己入栈进去，不符合特征就只将自己入栈追加到前一条末尾，这样从外面收到的出栈数据都是符合特征的。

此外，Datakit 还支持自动多行，在 logging 采集器的配置项中是 `auto_multiline_detection` 和 `auto_multiline_extra_patterns`，它的逻辑非常简单，就是提供一组的 `multiline_match`，根据原文遍历匹配所有规则，匹配成功就提高它的权重以便下次优先选择它。

自动多行是简化配置的一种方式，除了用户配置外，还提供 “默认自动多行规则列表”，详情链接见文章末尾。

### Pipeline 切割和日志 status

Pipeline 是一种简单的脚本语言，提供各种函数和语法，用以编写对一段文本数据的执行规则，主要用于切割非结构化的文本数据，例如把一行字符串文本切割出多个有意义的字段，或者用于从结构化的文本中（如 JSON）提取部分信息。

Pipeline 的实现比较复杂，它由抽象语法树（AST）和一系列内部状态机、功能纯函数组成，此处不过多描述。

只看使用场景，举个简单的例子，原文如下：

```
2020-10-23 06:41:56,688 INFO demo.py 1.0
```

pipeline 脚本：

```python
grok(_, "%{date:time} %{NOTSPACE:status} %{GREEDYDATA:msg}")
default_time(time)
```

最终结果：

```python
{
    "message": "2020-10-23 06:41:56,688 INFO demo.py 1.0",
    "msg": "demo.py 1.0",
    "status": "info",
    "time": 1603435316688000000
}
```

*注意：Pipeline 切割后的 `status` 字段是 `INFO`，但是 Datakit 有做映射处理，所以严谨起见显示为小写的 `info`*

Pipeline 是日志数据处理最后一步，Datakit 会使用 Pipeline 的结果构建行协议，序列化对象并准备打包发送给 Dataway。

## 发送数据和同步

数据发送是很常见的行为，在 Datakit 中基本就三步 —— “打包”、“转码” 和 “发送”。

但是发送之后的操作必不可少，而且至关重要，分别是 “同步当前文件的读取位置” 和 “检测文件状态”。

### 同步

在文章第一节介绍 “网络流数据” 时提到，为了能够进行日志文件的定点续读，而不是只支持 “从文件首部开始读取” 或者 `tail -f` 模式，Datakit 引入一个重要的操作 —— 记录当前文件的读取位置（position）。

日志采集每次从磁盘文件中读取数据，都会记录这段数据在文件的位置，只有当一系列处理和发送完成，才会将这个位置信息连通其他数据，同步到单独一个磁盘文件中。

这样做的好处是 Datakit 每次开启日志采集，如果这个日志文件之前被采集过，这次就能定位到上次的位置继续采集，不会造成数据采集重复，也不会丢失中间某段数据。

功能实现并不复杂：

Datakit 开启日志采集时，使用 `文件绝对路径 + 文件 inode + 文件首部的 N 个字节` 拼凑成一个专属的 key 值，使用这个 key 去指定路径的文件中找寻 position

- 如果能找到 position，表示这个文件上次已经被采集过，会从当前 position 再进行读取
- 如果没有找到 position，说明这是一个新的文件，会根据情况选择从文件首部读取，还是文件尾部读取

### 检测文件的状态

磁盘文件的状态不是一成不变的，这个文件可能会被删除、被重命名，或者长时间没有改动，Datakit 要对这些情况做处理。

- 文件长时间没有修改：
    - Datakit 会定期获取该文件的修改时间（`file Modification Date`），如果发现距离当前时间超过某个限定值，就会认为这文件已经 “不活跃”（inactive)，从而将其关闭不再采集
    - 这个逻辑在使用 glob 规则搜寻日志文件时也存在，如果找到一个符合 glob 规则的文件，但是它长久没有修改（也可以说 “更新”），不会对其进行日志采集

- 文件发生反转（rotate）：
    - 文件 rotate 是一个很复杂的逻辑，通俗一点来说就是文件名不变，但是文件名指向的具体文件发生改变，比如它的 inode。典型的例子就是 Docker 日志落盘。
    
Datakit 会定期检查当前正在采集的文件是否发生 rotate，检查的逻辑是：使用此文件名打开一个新的文件句柄，调用类似 `SameFile()` 的函数，判断两个句柄是否指向一致，如果不一致表示当前这个文件名已经发生 rotate。

一旦检测到文件发生 rotate，Datakit 会将当前文件的剩余数据（直到 EOF）采集完，再重新打开文件，此时已经是一个新的文件，然后操作流程一切照旧。

## 总结

日志采集是一个很复杂的系统，涉及到非常多的细节处理和优化逻辑。本文旨在介绍 Datakit 日志采集系统的设计和实现，没有 Benchmark 报告以及跟同类项目的性能对比，后续可以视情况补全。

补充链接：

- [glob 模式介绍](https://en.wikipedia.org/wiki/Glob_(programming)
- [Datakit 自动多行配置](https://docs.guance.com/integrations/logging/#auto-multiline)
- [Datakit Pipeline 处理](https://docs.guance.com/datakit/pipeline/)
- [Docker 截断超过 16KiB 日志的讨论](https://github.com/moby/moby/issues/34855)
- [Docker 截断超过 16KiB 的源码](https://github.com/nalind/docker/blob/master/daemon/logger/copier.go#L13)
- [Docker logging driver 描述 rotate 条件（max-file/max-size）](https://docs.docker.com/config/containers/logging/local/)
