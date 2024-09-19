---
title     : 'Disk IO'
summary   : 'Collect metrics of disk io'
__int_icon      : 'icon/diskio'
dashboard :
  - desc  : 'Disk IO'
    path  : 'dashboard/en/diskio'
monitor   :
  - desc  : 'Host detection library'
    path  : 'monitor/en/host'
---

<!-- markdownlint-disable MD025 -->
# DiskIO
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Diskio collector is used to collect the index of disk flow and time.

## Configuration {#config}

After successfully installing and starting DataKit, the DiskIO collector will be enabled by default without the need for manual activation.

### Precondition {#requirement}

For some older versions of Windows operating systems, if you encounter an error with Datakit: **"The system cannot find the file specified."**

Run PowerShell as an administrator and execute:

```powershell
diskperf -Y
```

The Datakit service needs to be restarted after successful execution.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or it can be named by `[[inputs.diskio.tags]]` alternative host in the configuration.

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

### Extended Metric {#extend}

[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7)

### Collecting disk `await` for Linux {#linux-await}

By default, DataKit cannot collect the disk `await` metric. If you need to obtain this metric, you can collect it by [Custom Collector with Python](../../developers/pythond/).

Preconditions

- [Enable Pythond collector](../developers/pythond.md)

Enter the DataKit installation directory, copy the `pythond.conf.sample` file and rename it to `pythond.conf`. Modify the corresponding configuration as follows:

```toml

[[inputs.pythond]]

    # Python collector name 
    name = 'diskio'  # required

    # Environment variables 
    #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

    # Python collector executable path (preferably use absolute path) 
    cmd = "python3" # required. python3 is recommended.

    # Relative path of the user script
    dirs = ["diskio"]

```

- Install `sar` command. You can refer to [https://github.com/sysstat/sysstat#installation](https://github.com/sysstat/sysstat#installation){:target="_blank"}

Install from Ubuntu

```shell
sudo apt-get install sysstat

sudo vi /etc/default/sysstat
# change ENABLED="false" to ENABLED="true"

sudo service sysstat restart
```

After installation, you can ran the following command to check if it was successful.

```shell
sar -d -p 3 1

Linux 2.6.32-696.el6.x86_64 (lgh)   10/06/2019      _x86_64_        (32 CPU)

10:08:16 PM       DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
10:08:17 PM    dev8-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:17 PM  dev253-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:17 PM  dev253-1      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00

10:08:17 PM       DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
10:08:18 PM    dev8-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:18 PM  dev253-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:18 PM  dev253-1      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00

10:08:18 PM       DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
10:08:19 PM    dev8-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:19 PM  dev253-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
10:08:19 PM  dev253-1      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00

Average:          DEV       tps  rd_sec/s  wr_sec/s  avgrq-sz  avgqu-sz     await     svctm     %util
Average:       dev8-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
Average:     dev253-0      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00
Average:     dev253-1      0.00      0.00      0.00      0.00      0.00      0.00      0.00      0.00

```

### Collect Scripts {#py-script}

Create file *<DataKit Dir\>/python.d/diskio/diskio.py* and add the following content:

```python
import subprocess
import re
from datakit_framework import DataKitFramework


class DiskIO(DataKitFramework):
    name = "diskio"
    interval = 10

    def run(self):
        stats = self.getStats()

        data = []

        for s in stats:
            tags = {
                "name": s.get("DEV", "")
            }
            awaitVal = 0.0
            svctmVal = 0.0

            try:
                awaitVal = float(s.get("await"))
            except:
                awaitVal = 0.0
            try:
                svctmVal = float(s.get("svctm"))
            except:
                svctmVal = 0.0

            fields = {
                "await": awaitVal,
                "svctm": svctmVal
            }
            data.append({
                "measurement": "diskio",
                "tags": tags,
                "fields": fields
            })

        in_data = {
            "M": data,
            "input": "datakitpy"
        }

        return self.report(in_data)

    def getStats(self):
        result = subprocess.run(
            ["sar", "-d", "-p", "3", "1"], stdout=subprocess.PIPE)
        output = result.stdout.decode("utf-8")

        str_list = output.splitlines()

        columns = []
        stats = []
        pattern = r'\s+'
        isAverage = False
        for l in enumerate(str_list):
            index, content = l
            if index < 2:
                continue

            stat = re.split(pattern, content)

            if len(stat) == 0 or stat[0] == "":
                isAverage = True
                continue

            if not isAverage:
                continue
            if "await" in stat and "DEV" in stat:
                columns = stat
            else:
                stat_info = {}
                if len(stat) != len(columns):
                    continue

                for s in enumerate(columns):
                    index, name = s
                    if index == 0:
                        continue
                    stat_info[name] = stat[index]
                stats.append(stat_info) 
        return stats

```

After saving the file, restart DataKit and you will be able to see the corresponding metrics on the Guance platform shortly.

### Metric List {#ext-metrics}

The `sar` command can obtain many useful [disk metrics](https://man7.org/linux/man-pages/man1/sar.1.html){:target="_blank"}. The above script only collect `await` and `svctm`. If you need to collect additional metrics, you can modify the script accordingly.

| Metric | Description | Type | Unit |
| ---- | ---- | ---- | ---- |
| `await` | The average time (in milliseconds) for I/O requests issued to the device to be served.  This includes the time spent by the requests in queue and the time spent servicing them. | float | ms |
| `svctm` | awaitThe average service time (in milliseconds) for I/O requests that were issued to the device. | float | ms |


## FAQ {#faq}

### What is the data source on Linux hosts {#linux-diskio}

On Linux hosts, the metrics are parsed and calculated from the */proc/diskstats* file; an explanation of each column can be found in [*procfs-diskstats*](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats){:target="_blank"};

The corresponding relationship between some data source columns and indicators is as follows:

| col04: reads completed successfully        | `reads`                                                   |
| col05: reads merged                        | `merged_reads`                                            |
| col06: sectors read                        | `read_bytes = col06 * sector_size`; `read_bytes/sec = (read_bytes - last(read_bytes))/(time - last(time))`      |
| col07: time spent reading (ms)             | `read_time`                                               |
| col08: writes completed                    | `writes`                                                  |
| col09: writes merged                       | `merged_writes`                                           |
| col10: sectors written                     | `write_bytes = col10 * sector_size`; `write_bytes/sec = (write_bytes - last(write_bytes))/(time - last(time))` |
| col11: time spent writing (ms)             | `write_time`                                              |
| col12: I/Os currently in progress          | `iops_in_progress`                                        |
| col13: time spent doing I/Os (ms)          | `io_time`                                                 |
| col14: weighted time spent doing I/Os (ms) | `weighted_io_time`                                        |

attention:

1. Sector size is 512 bytes;
2. Increment all but read_bytes/sec and write_bytes/sec.
