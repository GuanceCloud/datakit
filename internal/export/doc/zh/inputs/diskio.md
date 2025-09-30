---
title     : '磁盘 IO'
summary   : '采集磁盘 IO 指标数据'
tags:
  - '主机'
__int_icon      : 'icon/diskio'
dashboard :
  - desc  : '磁盘 IO'
    path  : 'dashboard/zh/diskio'
monitor   :
  - desc  : '主机检测库'
    path  : 'monitor/zh/host'
---


{{.AvailableArchs}}

---

磁盘 IO 采集器用于磁盘流量和时间的指标的采集。

## 配置 {#config}

成功安装 DataKit 并启动后，会默认开启 DiskIO 采集器，无需手动开启。

### 前置条件 {#requirement}

对于部分旧版本 Windows 操作系统，如若遇到 DataKit 报错： **"The system cannot find the file specified."**

请以管理员身份运行 PowerShell，并执行：

```powershell
$ diskperf -Y
...
```

在执行成功后需要重启 DataKit 服务。

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

### 扩展指标 {#extend}

[:octicons-tag-24: Version-1.5.7](../datakit/changelog.md#cl-1.5.7)

### Linux 平台下采集磁盘 `await` {#linux-await}

默认情况下，DataKit 无法采集磁盘 `await` 指标，如果需要获取该指标，可以通过[自定义 Python 采集器](../integrations/pythond.md)的方式来采集。

进入 DataKit 安装目录，复制 `pythond.conf.sample` 文件并将其命名为 `pythond.conf`。修改相应配置如下：

```toml
[[inputs.pythond]]

    # Python 采集器名称
    name = 'diskio'  # required

    # 运行 Python 采集器所需的环境变量
    #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

    # Python 采集器可执行程序路径(尽可能写绝对路径)
    cmd = "python3" # required. python3 is recommended.

    # 用户脚本的相对路径(填写文件夹，填好后该文件夹下一级目录的模块和 py 文件都将得到应用)
    dirs = ["diskio"]

```

- 安装 `sar` 命令，具体参考 [https://github.com/sysstat/sysstat#installation](https://github.com/sysstat/sysstat#installation){:target="\_blank"}

ubuntu 安装参考如下

```shell
sudo apt-get install sysstat

sudo vi /etc/default/sysstat
# change ENABLED="false" to ENABLED="true"

sudo service sysstat restart
```

安装完成后，可以执行下述命令，看是否成功。

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

### 采集脚本 {#py-script}

新建文件 *<DataKit 目录\>/python.d/diskio/diskio.py*，脚本内容如下：

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

文件保存后，重启 DataKit，稍后即可在<<<custom_key.brand_name>>>平台看到相应的指标。

### 指标列表 {#ext-metrics}

`sar` 命令可以获取很多有用的[磁盘指标](https://man7.org/linux/man-pages/man1/sar.1.html){:target="_blank"}，上述脚本只采集了 `await` 和 `svctm`，如果需要采集额外的指标，可以对脚本进行相应修改。

| Metric  | Description         | Type  | Unit |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----- | ---- |
| `await` | The average time (in milliseconds) for I/O requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them. | float | ms   |
| `svctm` | awaitThe average service time (in milliseconds) for I/O requests that were issued to the device.                                                                                | float | ms   |


## FAQ {#faq}

### `diskio` 指标在 Linux 主机上的数据来源是什么 {#linux-diskio}

在 Linux 主机上，指标从 */proc/diskstats* 文件获取并通过解析和计算得出；其中每一列的解释可参考[文档](https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats){:target="_blank"}；

部分数据来源列和指标的对应关系为：

| `diskstats` 字段                           | `diskio` 指标                                                                                                  |
| ---                                        | ---                                                                                                            |
| col04: reads completed successfully        | `reads`                                                                                                        |
| col05: reads merged                        | `merged_reads`                                                                                                 |
| col06: sectors read                        | `read_bytes = col06 * sector_size`; `read_bytes/sec = (read_bytes - last(read_bytes))/(time - last(time))`     |
| col07: time spent reading (ms)             | `read_time`                                                                                                    |
| col08: writes completed                    | `writes`                                                                                                       |
| col09: writes merged                       | `merged_writes`                                                                                                |
| col10: sectors written                     | `write_bytes = col10 * sector_size`; `write_bytes/sec = (write_bytes - last(write_bytes))/(time - last(time))` |
| col11: time spent writing (ms)             | `write_time`                                                                                                   |
| col12: I/Os currently in progress          | `iops_in_progress`                                                                                             |
| col13: time spent doing I/Os (ms)          | `io_time`                                                                                                      |
| col14: weighted time spent doing I/Os (ms) | `weighted_io_time`                                                                                             |

注意：

1. 扇区大小（`sector_size`）为 512 字节；
1. 除 `read_bytes/sec` 和 `write_bytes/sec` 外均为递增值。
