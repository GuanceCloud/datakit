
# View Monitor for Datakit
---

Datakit provides relatively complete output of basic observable information. By looking at the monitor output of Datakit, we can clearly know the current operation of Datakit.

## View Monitor {#view}

Execute the following command to get the running status of the native Datakit.

```shell
datakit monitor
```
<!-- markdownlint-disable MD046 -->
???+ tip

    You can see more monitor options through the `datakit help monitor`.
<!-- markdownlint-enable -->
The Datakit Basic Monitor page information is shown in the following figure:

![`onitor-basic-v1`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/monitor-basic-v1.png)

The elements in this diagram can be manipulated by mouse or keyboard. Blocks selected by the mouse are highlighted in bilateral boxes (as shown in the `Basic Info` block in the upper left corner of the above figure), and can also be browsed through the mouse wheel or the up and down arrow keys of the keyboard (or J/K of vim).

The information of each UI block in the above figure is:

- `Basic Info` is used to display basic information about Datakit, such as the version number, hostname, and runtime duration. From here, we can get a basic understanding of the current status of Datakit. Here are a few fields highlighted for individual explanation:
    - `Uptime`: The startup time of Datakit.
    - `Version`: The current version number of Datakit.
    - `Build`: The release date of Datakit.
    - `Branch`: The current code branch of Datakit, which is usually `master`.
    - `Build Tag`: The compilation options for Datakit; for the [Lite version](datakit-install.md#lite-install), this is `lite`.
    - `OS/Arch`: The current hardware and software platform of Datakit.
    - `Hostname`: The current hostname.
    - `Resource Limit`: Displays the current resource limit configurations for Datakit, where `mem` refers to the maximum memory limit, and `cpu` refers to the usage limit range (if displayed as `-`, it means the current cgroup is not set).
    - `Elected`: Shows the election status, see [here](election.md#status) for details.
    - `From`: The current Datakit address being monitored, such as `http://localhost:9529/metrics`.
    - `Proxy`: The current proxy server being used.

- `Runtime Info` is used to display the basic runtime consumption of Datakit (mainly memory, CPU and Golang runtime), including:

    - `Goroutines`: The number of Goroutine currently running.
    - `Total/Heap`: The memory occupied by the Golang virtual memory(`sys-alloc`) and the memory currently in use(`heap-alloc`) [^go-mem].
    - `RSS/VMS`: The RSS memory usage and VMS.
    - `GC Paused`: The time and number of times the GC (garbage collection) has consumed since Datakit started.
    - `OpenFiles`: The number of files currently open (on some platforms, it may display as `-1`, indicating that the feature is not supported).

[^go-mem]: Note that the memory usage displayed here is specific to the Golang virtual machine and does not include the memory used by external collectors that may be running.

- `Enabled Inputs` displays a list of open collectors:

    - `Input`: Refer to the collector(input) name, which is fixed and cannot be modified
    - `Count`: Refer to the number of the collector turned on
    - `Crashed`: Refer to the number of crashes of the collector

- `Inputs Info`: It is used to show the running status of each collector. There is more information here:

    - `Input`: Refer to the collector name. In some cases, this name is collector-specific (such as Log Collector/Prom Collector)
    - `Cat`: Refer to the type of data collected by the collector (M (metrics)/L (logs)/O (objects...)
    - `Feeds`: Total updates(collects) since Datakit started
    - `P90Lat`: Feed latency(blocked on queue) time(p90). The longer the duration, the slower the upload workers [:octicons-tag-24: Version-1.36.0](../datakit/changelog.md#cl-1.36.0)
    - `P90Pts`: Points(P90) collected of the collector [:octicons-tag-24: Version-1.36.0](../datakit/changelog.md#cl-1.36.0)
    - `Last Feed`: Time of last update(collect), relative to current time
    - `Avg Cost`: Average cost of each collect
    - `Errors`: Collect error count(if no error, empty here)

- The prompt text at the bottom tells you how to exit the current Monitor program and displays the current Monitor refresh rate.

---

If the verbose option (`-V`) is specified when Monitor is run, additional information is output, as shown in the following figure:

![`monitor-verbose-v1`](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/monitor-verbose-v1.png)

- `Goroutine Groups` shows the existing Goroutine Groups in the Datakit (the number of Goroutines in the group < = the number of `Goroutines` in the panel above).
- `HTTP APIs`: HTTP API request info
- `Filter`: Pull of blacklist filtering rules
- `Filter Rules`: Filtering of each type of blacklist
- `Pipeline Info`: Pipeline running info
- `WAL Info` WAL Queue Usage [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)

    The WAL queue consists of two parts: a small in-memory queue and a default 2GB disk queue. Here, `mem` refers to the number of points processed by the in-memory queue, `disk` refers to the number of points processed by the disk queue, and `drop` refers to the number of points discarded by the disk queue (for example, when the disk queue is full). Total refers to the total number of points.

- `Point Upload Info` Displays the operation of the data upload channel [^point-upload-info-on-160].
- `DataWay APIs` Displays the invocation situation of Dataway APIs.

[^point-upload-info-on-160]: [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) There have been updates here, and previous versions may show slightly different information.

## FAQ {#faq}
<!-- markdownlint-disable MD013 -->
### :material-chat-question:How to show only the operation of the specified module? {#specify-module}
<!-- markdownlint-enable -->
You can specify a list of module names (multiple modules are separated by English commas): [:octicons-tag-24: Version-1.5.7](changelog.md#cl-1.5.7)

```shell
datakit monitor -M inputs,filter
# or
datakit monitor --module inputs,filter

# use thd module abbreviation
datakit monitor -M in,f
```
<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to show only the operation of the specified collector? {#specify-inputs}
<!-- markdownlint-enable -->
You can specify a list of collector names (multiple collectors are separated by English commas):

```shell
datakit monitor -I cpu,mem
# or
datakit monitor --input cpu,mem
```
<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to display too long text? {#too-long}
<!-- markdownlint-enable -->
When some collectors report errors, their error information will be very long and incomplete in the table.

Complete information can be displayed by setting the column width of the display:

```shell
datakit monitor -W 1024
# or
datakit monitor --max-table-width 1024
```
<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to change the Monitor refresh rate? {#freq}
<!-- markdownlint-enable -->
It can be changed by setting the refresh frequency:

```shell
datakit monitor -R 1s
# or
datakit monitor --refresh 1s
```
<!-- markdownlint-disable MD046 -->
???+ attention

    Note that the units here must be the following: s (seconds)/m (minutes)/h (hours). If the time range is less than 1s, refresh according to 1s. 
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to Monitor other DataKits? {#remote-monitor}
<!-- markdownlint-enable -->

We can specify other Datakit's IP to show it's monitor:

```shell
datakit monitor --to <remote-ip>:9529
```

<!-- markdownlint-disable MD046 -->
???+ info

    By default, metrics data used by monitor are not accessible for non-localhost, we can [add API `/metrics` to API white list](datakit-conf.md#public-apis).
<!-- markdownlint-enable -->
