
# DataKit Service Management
---

After [DataKit installation](datakit-install.md), it is necessary to do some basic introduction to the installed DataKit.

## DataKit Introduction {#install-dir}

DataKit currently supports three major platforms, Linux/Windows/Mac:

| Operating System                            | Structure                | Installation Path                                                                   |
| :--------                           | :---                | :-----                                                                     |
| Linux kernel version 2.6. 23 or later        | amd64/386/arm/arm64 | `/usr/local/datakit`                                                       |
| macOS 10.13 or later [^1]          | amd64               | `/usr/local/datakit`                                                       |
| Windows 7, Server 2008R2 Or above | amd64/386           | 64 bit: `C:\Program Files\datakit`<br />32 bit: `C:\Program Files(32)\datakit` |

[^1]: Golang 1.18 requires macOS-amd64 version 10.13.

After installation, the list of DataKit directories is roughly as follows:

```txt
├── [4.4K]  conf.d
├── [ 160]  data
├── [ 64M]  datakit
├── [ 192]  externals
├── [1.2K]  pipeline
├── [ 192]  gin.log   # Windows platform
└── [1.2K]  log       # Windows platform
```

Among them:

- `conf.d`: Store configuration examples for all collectors. The DataKit main configuration file `datakit.conf` is located in the directory.
- `data`: Store data files needed for DataKit to run, such as IP address database, etc.
- `datakit`: DataKit main program, `datakit.exe` in Windows
- `externals`: Part of the collector is not integrated in the DataKit main program, it's all here.
- `pipeline` holds script code for text processing.
- `gin.log`: DataKit can receive external HTTP data input, and this log file is equivalent to HTTP access-log.
- `log`: Datakit run log (under Linux/Mac platform, DataKit run log is in */var/log/datakit* directory).
<!-- markdownlint-disable MD046 -->
???+ tip "View kernel version"

    - Linux/Mac：`uname -r`
    - Windows: Execute the `cmd` command (hold down the Win key + `r`, enter `cmd` carriage return) and enter `winver` to get system version information
<!-- markdownlint-enable -->
## DataKit Service Management {#manage-service}

DataKit can be directly managed using the following command:

```shell
# Linux/Mac may need to add sudo
# stop
datakit service -T
# start
datakit service -S
# restart
datakit service -R
```
<!-- markdownlint-disable MD046 -->
???+ tip

    You can view more help information through `datakit help service`.
<!-- markdownlint-enable -->
### Service Management Failure Handling {#when-service-failed}

Sometimes a service operation may fail due to a bug in some DataKit components (for example, the service does not stop after `datakit service -T`), which can be enforced as follows.

Under Linux, if the above command fails, the following command can be used instead:

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

Under Mac, you can use the following command instead:

```shell
# Start DataKit
sudo launchctl load -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
# or
sudo launchctl load -w /Library/LaunchDaemons/com.guance.datakit.plist

# Stop DataKit
sudo launchctl unload -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
# or
sudo launchctl unload -w /Library/LaunchDaemons/com.guance.datakit.plist
```

### Service Uninstall and Reinstall {#uninstall-reinstall}

You can uninstall or restore the DataKit service directly using the following command:

> Note: Uninstalling the DataKit here does not delete the DataKit-related files.

```shell
# Linux/Mac shell
datakit service -I # re-install
datakit service -U # uninstall
```
