
{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}

# 查看 DataKit 的 Monitor 信息

DataKit 提供了相对完善的基本可观测信息输出，通过查看 DataKit 的 monitor 输出，我们能清晰的知道当前 DataKit 的运行情况。

## 查看 Monitor

执行如下命令即可获取本机 DataKit 的运行情况。

```
datakit monitor
```

> 可通过 `datakit help monitor` 查看更多 monitor 选项。

DataKit 基本 Monitor 页面信息如下图所示：

![](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/images/datakit/monitor.png)

该图中的元素可以通过鼠标或键盘操作。被鼠标选中的块会以双边框突出显示（如上图的 `Basic Info` 块所示），另外，还能通过鼠标滚轮或者键盘上下方向键（或者 vim 的 H/J 来浏览）

上图中的每个 UI 块的信息分别是：

- `Basic Info`

用来展示 DataKit 的基本信息，如版本号、主机名、运行时长等信息。从这里我们可以对 DataKit 当前的情况有个基本了解。

- `Runtime Info` 用来展示 DataKit 的基本运行消耗（主要是内存以及 Goroutine 有关），其中：

	- `Goroutines`：当前正在运行的 Goroutine 个数
	- `Memory`：DataKit 进程当前实际消耗的内存字节数（*不含外部运行的采集器*）
	- `Stack`：当前栈中消耗的内存字节数
	- `GC Paused`：自 DataKit 启动以来，GC（垃圾回收）所消耗的时间
	- `GC Count`：自 DataKit 启动以来，GC 次数

> 关于这里的 Runtime Info，参见 [Golang 官方文档](https://pkg.go.dev/runtime#ReadMemStats)

- `Inputs Info`：用来展示每个采集器的采集情况，这里信息较多，下面一一分解
	- `Input`: 指采集器名称。某些情况下，这个名称是采集器自定义的（比如日志采集器/Prom 采集器）
	- `Category`：指该采集器所采集的数据类型（M(指标)/L(日志)/O(对象)...）
	- `Freq`：指该采集器每分钟的采集频率
	- `Avg Pts`：指该采集器每次采集所获取的行协议点数（*如果采集器频率 Freq 高，但 Avg Pts 又少，则该采集器的设定可能有问题*）
	- `Total Feed`：总的采集次数
	- `Total Pts`：采集的总的行协议点数
	- `1st Feed`：第一次采集的时间（相对当前时间）
	- `Last Feed`：最后一次采集的时间（相对当前时间）
	- `Avg Cost`：平均每次采集消耗
	- `Max Cost`：最大采集消耗
	- `Error(date)`：是否有采集错误（并附带最后一次错误相对当前的时间）

- 底部的提示文本，用于告知如何退出当前的 Monitor 程序，并且显示当前的 Monitor 刷新频率。

---

如果运行 Monitor 时，指定了 verbose 选项（`-V`），则会额外输出更多信息，如下图所示：

![](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/images/datakit/monitor-verbose.png)

- `Enabled Inputs` 展示开启的采集器列表，其中

	- `Input`：指采集器名称，该名称是固定的，不容修改
	- `Instances`：指该采集器开启的个数
	- `Crashed`：指该采集器的崩溃次数

- `Goroutine Groups` 展示 DataKit 中已有的 Goroutine 分组（该分组中的 Goroutine 个数 <= 上面面板中的 `Goroutines` 个数）

## FAQ

### 如何只展示指定采集器的运行情况？

---

A：可通过指定一个采集器名字列表（多个采集器之间以英文逗号分割）：

```shell
datakit monitor -I cpu,mem
# 或者
datakit monitor --input cpu,mem
```

### 如何展示太长的文本？

当某些采集器产生报错时，其报错信息会很长，在表格展示不全。

---

A：可通过设定展示的列宽来显示完整的信息：

```shell
datakit monitor -W 1024
# 或者
datakit monitor --max-table-width 1024
```

### 如何更改 Monitor 刷新频率？

---

A：可通过设定刷新频率来更改：

```shell
datakit monitor -R 1s
# 或者
datakit monitor --refresh 1s
```

> 注意，这里的单位需注意，必须是如下几种：s（秒）/m（分钟）/h（小时），如果时间范围小于 1s，则按照 1s 来刷新。

### 如何 Monitor 其它 DataKit？

有时候，安装的 DataKit 并不是使用默认的 9529 端口，这时候就会出现类似如下的错误：

```shell
request stats failed: Get "http://localhost:9528/stats": dial tcp ...
```

---

A: 可通过指定 datakit 地址来查看其 monitor 数据：

```shell
datakit monitor --to localhost:19528

# 也能查看另一个远程 DataKit 的 monitor
datakit monitor --to <remote-ip>:9528
```
