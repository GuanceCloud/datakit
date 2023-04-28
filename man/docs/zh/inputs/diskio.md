
# DiskIO
---

{{.AvailableArchs}}

---

磁盘 IO 采集器用于磁盘流量和时间的指标的采集。

## 前置条件 {#requests}

对于部分旧版本 Windows 操作系统，如若遇到 Datakit 报错： **"The system cannot find the file specified."**

请以管理员身份运行 PowerShell，并执行：

```powershell
diskperf -Y
```

在执行成功后需要重启 Datakit 服务。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数：
    
    | 环境变量名                            | 对应的配置参数项     | 参数示例                                                     |
    | :---                                  | ---                  | ---                                                          |
    | `ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER` | `skip_serial_number` | `true`/`false`                                               |
    | `ENV_INPUT_DISKIO_TAGS`               | `tags`               | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_DISKIO_INTERVAL`           | `interval`           | `10s`                                                        |
    | `ENV_INPUT_DISKIO_DEVICES`            | `devices`            | `'''^sdb\d*'''`                                              |
    | `ENV_INPUT_DISKIO_DEVICE_TAGS`        | `device_tags`        | `"ID_FS_TYPE", "ID_FS_USAGE"` 以英文逗号隔开                 |
    | `ENV_INPUT_DISKIO_NAME_TEMPLATES`     | `name_templates`     | `"$ID_FS_LABEL", "$DM_VG_NAME/$DM_LV_NAME"` 以英文逗号隔开   |
<!-- markdownlint-enable -->

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 扩展指标 {#extend}

[:octicons-tag-24: Version-1.5.7](changelog.md#cl-1.5.7)

### Linux 平台下采集磁盘 `await` {#linux-await}

默认情况下，DataKit 无法采集磁盘 `await` 指标，如果需要获取该指标，可以通过[自定义 Python 采集器](../developers/pythond.md)的方式来采集。

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

- 安装 `sar` 命令, 具体参考 [https://github.com/sysstat/sysstat#installation](https://github.com/sysstat/sysstat#installation)

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

新建文件 `<DataKit目录>/python.d/diskio/diskio.py`，脚本内容如下：

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

文件保存后，重启 DataKit，稍后即可在观测云平台看到相应的指标。

### 指标列表 {#ext-metrics}

`sar` 命令可以获取很多有用的[磁盘指标](https://man7.org/linux/man-pages/man1/sar.1.html)，上述脚本只采集了 `await` 和 `svctm`，如果需要采集额外的指标，可以对脚本进行相应修改。

| Metric  | Description                                                                                                                                                                      | Type  | Unit |
| ----    | ----                                                                                                                                                                             | ----  | ---- |
| `await` | The average time (in milliseconds) for I/O requests issued to the device to be served.  This includes the time spent by the requests in queue and the time spent servicing them. | float | ms   |
| `svctm` | awaitThe average service time (in milliseconds) for I/O requests that were issued to the device.                                                                                 | float | ms   |
