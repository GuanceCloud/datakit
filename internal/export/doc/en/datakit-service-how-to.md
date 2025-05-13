# DataKit Service Management
---
After [installing DataKit](datakit-install.md), it is necessary to provide some basic information about the installed DataKit.

## Introduction to DataKit Directories {#install-dir}

DataKit currently supports three mainstream platforms: Linux, Windows, and Mac:

| Operating System                      | Architecture        | Installation Path                                                              |
| ---------                             | ---                 | ------                                                                         |
| Linux kernel version 2.6.23 or higher | amd64/386/arm/arm64 | `/usr/local/datakit`                                                           |
| macOS version 10.13 or higher[^1]     | amd64               | `/usr/local/datakit`                                                           |
| Windows 7, Server 2008R2 or higher    | amd64/386           | 64-bit: `C:\Program Files\datakit`<br />32-bit: `C:\Program Files(32)\datakit` |

[^1]: Golang 1.18 requires macOS-amd64 version 10.13.

After the installation is complete, the DataKit directory list is roughly as follows:

``` not-set
├── [   12]  apm_inject/
├── [    0]  gitrepos/
├── [    0]  python.d/
├── [  430]  pipeline/
├── [   26]  pipeline_remote/
├── [   42]  cache/
├── [   36]  externals/
├── [  316]  data/
├── [ 138M]  datakit
├── [  958]  conf.d/
└── [    7]  .pid
```

| Directory Name    | Description                                                                                                                            |
| ---               | ---                                                                                                                                    |
| `apm_inject`      | After enabling the APM auto-injection function, this directory is used to store some dependent files.                                  |
| `cache`           | Store some data caches used during the collection process.                                                                             |
| `conf.d`          | Store configuration examples of all collectors. The DataKit main configuration file *datakit.conf* is located in this directory.       |
| `data`            | Store data files required for DataKit operation, such as the IP address database.                                                      |
| `datakit`         | The main DataKit program. On Windows, it is *datakit.exe*. Most of the collection functions of DataKit are integrated in this program. |
| `externals`       | Some collectors are not integrated in the DataKit main program and are compiled separately.                                            |
| `gitrepos`        | If Git is used to manage collector configurations, store these configurations here.                                                    |
| `pipeline`        | Store Pipeline scripts.                                                                                                                |
| `pipeline_remote` | Store Pipeline scripts written in Studio.                                                                                              |
| `python.d`        | Store Python scripts.                                                                                                                  |
| `.pid`            | Store the process ID of the currently running DataKit.                                                                                 |

There are two DataKit log files:

| Directory Name            | Description                                                                                 |
| ---                       | ---                                                                                     |
| `gin.log`                 | DataKit can receive external HTTP data input. This log file is equivalent to the HTTP access log. |
| `log`                     | DataKit operation log (On Linux/Mac platforms, the DataKit operation log is located in the */var/log/datakit* directory. On Windows, it is located in the *C:\Program Files\datakit\* directory). |

<!-- markdownlint-disable MD046 -->
???+ tip "Check the Kernel Version"

    - Linux/Mac: `uname -r`
    - Windows: Execute the `cmd` command (Press Win key + `r`, enter `cmd` and press Enter), and input `winver` to get the system version information.
<!-- markdownlint-enable -->

## DataKit Service Management {#manage-service}

You can directly use the following commands to manage DataKit:

```shell
# Linux/Mac may require sudo
datakit service -T # stop
datakit service -S # start
datakit service -R # restart
```

<!-- markdownlint-disable MD046 -->
???+ tip

    You can use `datakit help service` to view more help information.
<!-- markdownlint-enable -->

### Handling of Service Management Failures {#when-service-failed}

Sometimes, due to bugs in some components of DataKit, the service operation may fail (for example, after `datakit service -T`, the service does not stop). You can force the processing in the following way.

On Linux, if the above command fails, you can use the following commands instead:

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

On Mac, you can use the following commands instead:

```shell
# Start DataKit
sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist

# Stop DataKit
sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist
```

### Service Uninstall and Reinstall {#uninstall-reinstall}

You can directly use the following commands to uninstall or restore the DataKit service:

> Note: Uninstalling DataKit here will not delete DataKit-related files.

```shell
# Linux/Mac shell
datakit service -I # re-install
datakit service -U # uninstall
```

## Impact of DataKit on the Host Environment {#datakit-overhead}

During the use of DataKit, the existing system may be affected in the following ways:

1. Log collection will lead to high-speed disk reading. The larger the log volume, the higher the iops of reading.
1. If the RUM SDK is added to a Web/App application, continuous RUM-related data upload will occur. If there are restrictions on the upload bandwidth, it may cause the Web/App page to freeze.
1. After [eBPF collection](../integrations/ebpf.md) is enabled, due to the large amount of collected data, a certain amount of memory and CPU will be occupied. After bpf-netlog is enabled, a large number of logs will be generated based on all TCP packets of the host and container network cards.
1. When DataKit is busy (a large number of logs/Traces are accessed, and external data is imported, etc.), it will occupy a considerable amount of CPU and memory resources. It is recommended to set reasonable [resource limit configurations](datakit-conf.md#resource-limit) for control.
1. When DataKit is [deployed in Kubernetes](datakit-daemonset-deploy.md), there will be a certain request pressure on the API server.
1. When the [default collector](datakit-input-conf.md#default-enabled-inputs) is enabled, the memory (RSS) consumption is approximately 100MB, and the CPU consumption is controlled within 10%. In addition to its own logs, the disk consumption also includes additional [disk cache](datakit-conf.md#dataway-wal). The network traffic depends on the specific amount of collected data. The traffic uploaded by DataKit is compressed and uploaded using GZip by default.

## FAQ {#faq}

### Failure to Start on Windows {#windows-start-fail}

DataKit is started as a service on Windows. After startup, a lot of Event logs will be written. As the logs accumulate, the following error may occur:

``` not-set
Start service failed: The event log file is full.
```

This error will prevent DataKit from starting. You can [set the Windows Event](https://stackoverflow.com/a/13868216/342348){:target="_blank"} to solve this problem.

## Further References {#further-reading}

Other documents related to the basic use of DataKit:

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit Update</u>: Update the DataKit version </font>](datakit-update.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Monitor</u>: View the running status of DataKit</font>](datakit-monitor.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit Tool Commands</u>: DataKit provides many convenient tools to assist your daily use</font>](datakit-tools-how-to.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>DataKit Port Occupancy</u>: The list of ports used by DataKit by default</font>](datakit-port.md)
</div>
</font>
