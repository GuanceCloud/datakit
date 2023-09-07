
# View Monitor for DataKit
---

DataKit provides relatively complete output of basic observable information. By looking at the monitor output of DataKit, we can clearly know the current operation of DataKit.

## View Monitor {#view}

Execute the following command to get the running status of the native DataKit.

```
datakit monitor
```

???+ tip

    You can see more monitor options through the `datakit help monitor`.

The DataKit Basic Monitor page information is shown in the following figure:

![](https://static.guance.com/images/datakit/monitor-basic-v1.png)

The elements in this diagram can be manipulated by mouse or keyboard. Blocks selected by the mouse are highlighted in bilateral boxes (as shown in the `Basic Info` block in the upper left corner of the above figure), and can also be browsed through the mouse wheel or the up and down arrow keys of the keyboard (or J/K of vim).

The information of each UI block in the above figure is:

- `Basic Info` is used to show the basic information of DataKit, such as version number, host name, runtime and so on. From here, we can have a basic understanding of the current situation of DataKit. Now select a few fields to explain separately:
    - `Version`: Current version number of DataKit
    - `Uptime`: Startup time of DataKit
    - `Branch`: The current code branch of DataKit, which is generally master
    - `Build`：DataKit release date
    - `Resource Limit`: Show the resource limit configuration of the current DataKit, where mem refers to the maximum memory limit and cpu refers to the usage limit  (If cgroup not set, the value is `-`)
    - `Hostname`: Current hostname
    - `OS/Arch`: Current software and hardware platforms of DataKit
    - `Elected`: Election info(See [Election](election.md#status))
    - `From`: The DataKit address of the current Monitor, such as `http://localhost:9529/metrics`

- `Runtime Info` Runtime Info is used to show the basic running consumption of Datakit (mainly related to memory and Goroutine):

    - `Goroutines`: The number of Goroutines currently running
    - `Mem`: The actual number of bytes of memory currently consumed by the DataKit process (*excluding externally running collectors*)
    - `System`: Virtual memory currently consumed by the DataKit process (*excluding externally running collectors*)
    - `GC Paused`: Time elapsed and count of Golang GC (garbage collection) since DataKit started

???+ info

    For Runtime Info here, see [Golang doc](https://pkg.go.dev/runtime#ReadMemStats){:target="_blank"}

- `Enabled Inputs` displays a list of open collectors:

    - `Input`: Refer to the collector(input) name, which is fixed and cannot be modified
    - `Count`: Refer to the number of the collector turned on
    - `Crashed`: Refer to the number of crashes of the collector
    
- `Inputs Info`: It is used to show the running status of each collector. There is more information here:

    - `Input`: Refer to the collector name. In some cases, this name is collector-specific (such as Log Collector/Prom Collector)
    - `Cat`: Refer to the type of data collected by the collector (M (metrics)/L (logs)/O (objects...)
    - `Feeds`: Total updates(collects) since Datakit started
    - `TotalPts`: Total points collected of the collector
    - `Last Feed`: Time of last update(collect), relative to current time
    - `Avg Cost`: Average cost of each collect
    - `Errors`: Collect error count(if no error, empty here)

- The prompt text at the bottom tells you how to exit the current Monitor program and displays the current Monitor refresh rate.

---

If the verbose option (`-V`) is specified when Monitor is run, additional information is output, as shown in the following figure:

![](https://static.guance.com/images/datakit/monitor-verbose-v1.png)

- `Goroutine Groups` shows the existing Goroutine Groups in the DataKit (the number of Goroutines in the group < = the number of `Goroutines` in the panel above).
- `HTTP APIs`: HTTP API request info
- `Filter`: Pull of blacklist filtering rules
- `Filter Rules`: Filtering of each type of blacklist
- `Pipeline Info`: Pipeline running info
- `IO Info`: Data upload info
- `DataWay APIs`: Dataway API request info

## FAQ {#faq}

### :material-chat-question:How to show only the operation of the specified module? {#specify-module}

You can specify a list of module names (multiple modules are separated by English commas): [:octicons-tag-24: Version-1.5.7](changelog.md#cl-1.5.7)

```shell
datakit monitor -M inputs,filter
# or
datakit monitor --module inputs,filter

# use thd module abbreviation
datakit monitor -M in,f
```

### :material-chat-question: How to show only the operation of the specified collector? {#specify-inputs}

You can specify a list of collector names (multiple collectors are separated by English commas):

```shell
datakit monitor -I cpu,mem
# or
datakit monitor --input cpu,mem
```

### :material-chat-question: How to display too long text? {#too-long}

When some collectors report errors, their error information will be very long and incomplete in the table.

Complete information can be displayed by setting the column width of the display:

```shell
datakit monitor -W 1024
# or
datakit monitor --max-table-width 1024
```

### :material-chat-question: How to change the Monitor refresh rate? {#freq}

It can be changed by setting the refresh frequency:

```shell
datakit monitor -R 1s
# or
datakit monitor --refresh 1s
```

???+ attention

    Note that the units here must be the following: s (seconds)/m (minutes)/h (hours). If the time range is less than 1s, refresh according to 1s. 

### :material-chat-question: How to Monitor other DataKits? {#remote-monitor}

Sometimes, the DataKit installed does not use the default 9529 port, and this time, an error like the following will occur:

```shell
request stats failed: Get "http://localhost:9528/stats": dial tcp ...
```

We can view its monitor data by specifying the datakit address:

```shell
datakit monitor --to localhost:19528

# We can also view the monitor of another remote DataKit
datakit monitor --to <remote-ip>:9528
```
