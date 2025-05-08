
# 查看 Datakit 的 Monitor
---

Datakit 提供了相对完善的基本可观测信息输出，通过查看 Datakit 的 monitor 输出，我们能清晰的知道当前 Datakit 的运行情况。

## 查看 Monitor {#view}

执行如下命令即可获取本机 Datakit 的运行情况。

``` shell
datakit monitor
```

<!-- markdownlint-disable MD046 -->
???+ tip

    可通过 `datakit help monitor` 查看更多 monitor 选项。
<!-- markdownlint-enable -->

Datakit 基本 Monitor 页面信息如下图所示：

![not-set](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/monitor-basic-v1.png)

该图中的元素可以通过鼠标或键盘操作。被鼠标选中的块会以双边框突出显示（如上图左上角的 `Basic Info` 块所示），另外，还能通过鼠标滚轮或者键盘上下方向键（或者 vim 的 J/K）来浏览。

上图中的每个 UI 块的信息分别是：

- `Basic Info` 用来展示 Datakit 的基本信息，如版本号、主机名、运行时长等信息。从这里我们可以对 Datakit 当前的情况有个基本了解。现挑选几个字段出来单独说明：
    - `Uptime`：Datakit 的启动时间
    - `Version`：Datakit 当前的版本号
    - `Build`：Datakit 的发布时间
    - `Branch`：Datakit 当前的代码分支，一般情况下都是 master
    - `Build Tag`：Datakit 的编译选项，[精简版](datakit-install.md#lite-install)此处是 `lite`
    - `OS/Arch`：当前 Datakit 的软硬件平台
    - `Hostname`：当前主机名
    - `Resource Limit`：展示当前 Datakit 的资源限制配置，其中 `mem` 指最大内存限制，`cpu` 指使用率限制范围 （如果展示为 `-` 表示当前 cgroup 未设置）
    - `Elected`：展示选举情况，详见[这里](election.md#status)
    - `From`：当前被 Monitor 的 Datakit 地址，如 `http://localhost:9529/metrics`
    - `Proxy`：当前使用的代理服务器

- `Runtime Info` 用来展示 Datakit 的基本运行消耗（主要是内存、CPU 以及 Golang runtime），其中：

    - `Goroutines`：当前正在运行的 Goroutine 个数
    - `Total/Heap`：Golang vm 占用内存以及正在使用中的内存（*不含外部运行的采集器*）[^go-mem]
    - `RSS/VMS`：RSS 内存占用以及 VMS（*不含外部运行的采集器*）
    - `GC Paused`：自 Datakit 启动以来，GC（垃圾回收）所消耗的时间以及次数
    - `OpenFiles`：当前打开的文件个数（部分平台可能显示为 `-1`，表示不支持该功能）

[^go-mem]: 关于这里的 Runtime Info，参见 [Golang 官方文档](https://pkg.go.dev/runtime#ReadMemStats){:target="_blank"}

- `Enabled Inputs` 展示开启的采集器列表，其中

    - `Input`：指采集器名称，该名称是固定的，不容修改
    - `Count`：指该采集器开启的个数
    - `Crashed`：指该采集器的崩溃次数

- `Inputs Info`：用来展示每个采集器的采集情况，这里信息较多，下面一一分解
    - `Input`: 指采集器名称。某些情况下，这个名称是采集器自定义的（比如日志采集器/Prom 采集器）
    - `Cat`：指该采集器所采集的数据类型（M(指标)/L(日志)/O(对象)...）
    - `Feeds`：指该采集器自启动以来更新数据（采集）的次数
    - `P90Lat`：指该采集器在上报数据点时的阻塞时长（P90），如果时间越长，表示当前数据发送越慢 [:octicons-tag-24: Version-1.36.0](../datakit/changelog.md#cl-1.36.0)
    - `P90Pts`：采集器采集的点数（P90）[:octicons-tag-24: Version-1.36.0](../datakit/changelog.md#cl-1.36.0)
    - `Filtered`：被黑名单筛选掉的点数
    - `Last Feed`：最后一次更新数据（采集）的时间（相对当前时间）
    - `Avg Cost`：平均每次采集消耗
    - `Errors`：采集错误次数（如果没有则不显示）

- 底部的提示文本，用于告知如何退出当前的 Monitor 程序，并且显示当前的 Monitor 刷新频率。

---

如果运行 Monitor 时，指定了 verbose 选项（`-V`），则会额外输出更多信息，如下图所示：

![not-set](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/monitor-verbose-v1.png)

- `Goroutine Groups` 展示 Datakit 中已有的 Goroutine 分组（该分组中的 Goroutine 个数 <= 上面面板中的 `Goroutines` 个数）
- `HTTP APIs` 展示 Datakit 中 API 调用情况
- `Filter` 展示 Datakit 中黑名单过滤规则拉取情况
- `Filter Rules` 展示每类黑名单的过滤情况
- `Pipeline Info` 展示 Pipeline 运行情况
- `WAL Info` WAL 队列的使用情况 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)

    WAL 队列由两部分组成，少量的内存队列以及默认 2GB 的磁盘队列。此处，`mem` 指内存队列处理的点数，`disk` 指磁盘队列处理的点数，`drop` 指磁盘队列丢弃的点数（比如磁盘队列满了）。Total 指总点数。

- `Point Upload Info` 展示数据上传通道的运行情况 [^point-upload-info-on-160]
- `DataWay APIs` 展示 Dataway API 的调用情况

[^point-upload-info-on-160]: [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 对这里有更新，之前的版本在这里的显示稍有差异。

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 如何展示 Datakit 指定模块的运行情况？ {#specify-module}
<!-- markdownlint-enable -->

可指定一个模块名字列表（多个模块之间以英文逗号分割）：[:octicons-tag-24: Version-1.5.7](changelog.md#cl-1.5.7)

```shell
datakit monitor -M inputs,filter
# 或者
datakit monitor --module inputs,filter

# 也可是模块名的简称
datakit monitor -M in,f
```

### :material-chat-question: 如何只展示指定采集器的运行情况？ {#specify-inputs}

可通过指定一个采集器名字列表（多个采集器之间以英文逗号分割）：

```shell
datakit monitor -I cpu,mem
# 或者
datakit monitor --input cpu,mem
```

### :material-chat-question: 如何展示太长的文本？ {#too-long}

当某些采集器产生报错时，其报错信息会很长，在表格展示不全。可通过设定展示的列宽来显示完整的信息：

```shell
datakit monitor -W 1024
# 或者
datakit monitor --max-table-width 1024
```

### :material-chat-question: 如何更改 Monitor 刷新频率？ {#freq}

可通过设定刷新频率来更改：

```shell
datakit monitor -R 1s
# 或者
datakit monitor --refresh 1s
```

<!-- markdownlint-disable MD046 -->
???+ info

    这里的单位需注意，必须是如下几种：s（秒）/m（分钟）/h（小时），如果时间范围小于 1s，则按照 1s 来刷新。
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 如何 Monitor 其它 Datakit？ {#remote-monitor}
<!-- markdownlint-enable -->

可通过指定 Datakit 地址来查看其 monitor 数据：

```shell
datakit monitor --to <remote-ip>:9529
```

<!-- markdownlint-disable MD046 -->
???+ info

    默认情况下，monitor 数据是不能通过非 localhost 访问的，可以[手动将其添加到 API 白名单中](datakit-conf.md#public-apis)。
<!-- markdownlint-enable -->
