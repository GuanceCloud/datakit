
# View Monitor for DataKit
---

DataKit provides relatively complete output of basic observable information. By looking at the monitor output of DataKit, we can clearly know the current operation of DataKit.

## View Monitor {#view}

Execute the following command to get the running status of the native DataKit.

```
datakit monitor
```

> You can see more monitor options through the `datakit help monitor`.

The DataKit Basic Monitor page information is shown in the following figure:

![基础Monitor信息展示](https://static.guance.com/images/datakit/monitor-basic-v1.gif) 

The elements in this diagram can be manipulated by mouse or keyboard. Blocks selected by the mouse are highlighted in bilateral boxes (as shown in the `Basic Info` block in the upper left corner of the above figure), and can also be browsed through the mouse wheel or the up and down arrow keys of the keyboard (or J/K of vim).

The information of each UI block in the above figure is:

- `Basic Info` is used to show the basic information of DataKit, such as version number, host name, runtime and so on. From here, we can have a basic understanding of the current situation of DataKit. Now select a few fields to explain separately:
  - `Version`: Current version number of DataKit
	- `Branch`: The current code branch of DataKit, which is generally master
	- `Uptime`: Startup time of DataKit
	- `CGroup`: Show the cgroup configuration of the current DataKit, where mem refers to the maximum memory limit and cpu refers to the usage limit
	- `OS/Arch`: Current software and hardware platforms of DataKit
	- `IO`: Show the current congestion of the DataKit IO channel
	- `Pipeline`: Show the current Pipeline processing of DataKit
	- `Elected`: Show the election situation
	  - If the election is not open, display `<namespace-name>::disabled|<none>`
	  - If the election is open, display `<namespace-name>::<disabled-or-success>|<elected-datakit-host-name>`, such as `my-namespace::success|my-host123`
	- `From`: The DataKit address of the current Monitor, such as `http://localhost:9529/stats`

- `Runtime Info` Runtime Info is used to show the basic running consumption of DataKit (mainly related to memory and Goroutine):

	- `Goroutines`: The number of Goroutines currently running
	- `Mem`: The actual number of bytes of memory currently consumed by the DataKit process (*excluding externally running collectors*)
	- `System`: Virtual memory currently consumed by the DataKit process (*excluding externally running collectors*)
	- `Stack`: Number of bytes of memory consumed in the current Stack
	- `GC Paused`: Time elapsed by GC (garbage collection) since DataKit started
	- `GC Count`: Number of GCs since DataKit started

> For Runtime Info here, see [Golang doc](https://pkg.go.dev/runtime#ReadMemStats){:target="_blank"}

- `Enabled Inputs` displays a list of open collectors:

	- `Input`: Refer to the collector name, which is fixed and cannot be modified
	- `Instances`: Refer to the number of the collector turned on
	- `Crashed`: Refer to the number of crashes of the collector

- `Inputs Info`: It is used to show the collection situation of each collector. There is more information here, which is decomposed one by one below
	- `Input`: Refer to the collector name. In some cases, this name is collector-specific (such as Log Collector/Prom Collector)
	- `Category`: Refer to the type of data collected by the collector (M (metrics)/L (logs)/O (objects...)
	- `Freq`: Refer to the acquisition frequency per minute of the collector
	- `Avg Pts`: Refer to the number of line protocol points collected by the collector per collection (*if the collector frequency Freq is high but Avg Pts is low, the collector setting may be problematic*)
	- `Total Feed`: Total collection times
	- `Total Pts`: Total line protocol points collected
	- `1st Feed`: Time of first collection (relative to current time)
	- `Last Feed`: Time of last collection (relative to current time)
	- `Avg Cost`: Average consumption per collection
	- `Max Cost`: Maximum collection consumption
	- `Error(date)`: Whether there is a collection Error (with the last Error relative to the current time)

- The prompt text at the bottom tells you how to exit the current Monitor program and displays the current Monitor refresh rate.

---

If the verbose option (`-V`) is specified when Monitor is run, additional information is output, as shown in the following figure:

![完整Monitor信息展示](imgs/monitor-verbose-v1.gif) 

- `Goroutine Groups` shows the existing Goroutine Groups in the DataKit (the number of Goroutines in the group < = the number of `Goroutines` in the panel above).
- `HTTP APIs` show API calls in DataKit.
- `Filter` shows the pull of blacklist filtering rules in DataKit.
- `Filter Rules` shows the filtering of each type of blacklist.

- `Sender Info` shows the operation of each Sink managed by Sender.
	- `Sink`: Sink name
	- `Uptime`: Runtime
	- `Count`: Number of Write
	- `Failed`: Number of Write failures
	- `Pts`: Write Points
	- `Raw Bytes`: Number of Write bytes (before compression)
	- `Bytes`: Number of Write Bytes (compressed)
	- `2XX`: HTTP status code 2XX times
	- `4XX`: HTTP status code 4XX times
	- `5XX`: HTTP status code 5XX times
	- `Timeout`: The number of HTTP timeouts

## FAQ {#faq}

### How to show only the operation of the specified collector? {#specify-inputs}

---

A: You can specify a list of collector names (multiple collectors are separated by English commas):

```shell
datakit monitor -I cpu,mem
# or
datakit monitor --input cpu,mem
```

### How to display too long text? {#too-long}

When some collectors report errors, their error information will be very long and incomplete in the table.

---

A: Complete information can be displayed by setting the column width of the display:

```shell
datakit monitor -W 1024
# or
datakit monitor --max-table-width 1024
```

### How to change the Monitor refresh rate? {#freq}

---

A: It can be changed by setting the refresh frequency:

```shell
datakit monitor -R 1s
# or
datakit monitor --refresh 1s
```

> Note that the units here must be the following: s (seconds)/m (minutes)/h (hours). If the time range is less than 1s, refresh according to 1s. 

### How to Monitor other DataKits? {#remote-monitor}

Sometimes, the DataKit installed does not use the default 9529 port, and this time, an error like the following will occur:

```shell
request stats failed: Get "http://localhost:9528/stats": dial tcp ...
```

---

A: You can view its monitor data by specifying the datakit address:

```shell
datakit monitor --to localhost:19528

# You can also view the monitor of another remote DataKit
datakit monitor --to <remote-ip>:9528
```

### How to view error messages for a specific collector? {#view-errors}

---

A: Click on the error message directly to display a detailed error message at the bottom. After clicking the error message, the display of the error message can be closed by ESC or Enter.
