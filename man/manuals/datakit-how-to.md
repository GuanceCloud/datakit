{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 入门简介

本文档主要介绍 [DataKit 安装](datakit-install)完后，如何使用 DataKit 中的基本功能，包括如下几个方面：

- 安装介绍
- 采集器使用，包括如何开启、管理等
- DataKit 中各种小工具使用

## DataKit 目录介绍

DataKit 目前支持 Linux/Windows/Mac 三种主流平台：


| 操作系统                                                                  | 架构                | 安装路径                                                                   |
| ---------                                                                 | ---                 | ------                                                                     |
| Linux 内核 2.6.23 或更高版本                                              | amd64/386/arm/arm64 | `/usr/local/datakit`                                                       |
| macOS 10.12 或更高版本([原因](https://github.com/golang/go/issues/25633)) | amd64               | `/usr/local/datakit`                                                       |
| Windows 7, Server 2008R2 或更高版本                                       | amd64/386           | 64位：`C:\Program Files\datakit`<br />32位：`C:\Program Files(32)\datakit` |

> Tips：查看内核版本

- Linux/Mac：`uname -r`
- Windows：执行 `cmd` 命令（按住 Win键 + `r`，输入 `cmd` 回车），输入 `winver` 即可获取系统版本信息

安装完成以后，DataKit 目录列表大概如下：

```
├── [4.4K]  conf.d
├── [ 160]  data
├── [186M]  datakit
├── [ 192]  externals
├── [1.2K]  pipeline
├── [ 192]  gin.log   # Windows 平台
└── [1.2K]  log       # Windows 平台
```

其中：

- `conf.d`：存放所有采集器的配置示例。DataKit 主配置文件 datakit.conf 位于目录下
- `data`：存放 DataKit 运行所需的数据文件，如 IP 地址数据库等
- `datakit`：DataKit 主程序，Windows 下为 `datakit.exe`
- `externals`：部分采集器没有集成在 DataKit 主程序中，就都在这里了
- `pipeline` 存放用于文本处理的脚本代码
- `gin.log`：DataKit 可以接收外部的 HTTP 数据输入，这个日志文件相当于 HTTP 的 access-log（DataKit 日志需[开启 `debug` 选项](datakit-how-to#d294fa14)才能看到 gin.log，否则 gin.log 内容为空）
- `log`：DataKit 运行日志

> 注：Linux/Mac 平台下，DataKit 运行日志在 `/var/log/datakit` 目录下。

## 采集器使用

DataKit 安装完成后，默认会开启一批采集器，这些采集器一般跟主机相关，列表如下：

| 采集器名称       | 说明                                           |
| ---------        | ---                                            |
| `cpu`            | 采集主机的 CPU 使用情况                        |
| `disk`           | 采集磁盘占用情况                               |
| `diskio`         | 采集主机的磁盘 IO 情况                         |
| `mem`            | 采集主机的内存使用情况                         |
| `swap`           | 采集 Swap 内存使用情况                         |
| `system`         | 采集主机操作系统负载                           |
| `net`            | 采集主机网络流量情况                           |
| `host_processes` | 采集主机上常驻（存活 10min 以上）进程列表      |
| `hostobject`     | 采集主机基础信息（如操作系统信息、硬件信息等） |
| `container`      | 采集主机上可能的容器对象以及容器日志           |

### 配置文件格式

DataKit 的配置均使用 [Toml 文件](https://toml.io/cn)，一个典型的配置文件格式，大概如下所示

```toml
[[inputs.some_name]] # 这一行是必须的，它表明这个 toml 文件是哪一个采集器的配置
	key = value
	...

[[inputs.some_name.other_options]] # 这一行则可选，有些采集器配置有这一行，有些则没有
	key = value
	...
```

### DataKit 主配置修改

以下涉及 DataKit 主配置的修改，均需重启 DataKit：

```shell
sudo datakit --restart
```

#### HTTP 绑定端口

出于安全考虑，DataKit 的 HTTP 服务默认绑定在 `localhost:9529` 上，如果希望从外部访问 DataKit API，需编辑 `conf.d/datakit.conf` 中的 `listen` 字段，这样就能从其它主机上请求 DataKit 接口了：

```toml
[http_api]
	listen = "0.0.0.0:9529" # 将其改成 `0.0.0.0:9529` 或其它网卡、端口
```

当你需要做如下操作时，一般都需要修改 `listen` 配置：

- [远程查看 DataKit 运行情况](http://localhost:9529/monitor)
- [远程查看 DataKit 文档](http://localhost:9529/man)
- [RUM 采集](rum)
- 其它诸如 [APM](ddtrace)/[安全巡检](sec-checker) 等，看具体的部署情况，可能也需要修改 `listen` 配置

#### 全局标签（tag）的开启

DataKit 允许在 `datakit.conf` 中配置全局标签，这些标签会默认添加到该 DataKit 采集的每一条数据上（前提是采集的原始数据上不带有这里配置的标签）。这里是一个全局标签配置示例：

```toml
[global_tags]
	ip         = "__datakit_ip"
	datakit_id = "$datakit_id"
	project    = "some_of_your_online_biz"
	other_tags = "..."                    # 可追加其它更多标签
```

注意，如下几个变量可用于这里的全局标签设置（双下划线（`__`）前缀和 `$` 都是可以的）：

- `__datakit_ip/$datakit_ip`：标签值会设置成 DataKit 获取到的第一个主网卡 IP
- `__datakit_id/$datakit_id`：标签值会设置成 DataKit 的 ID

另外，即使这里不设置全局 Tag，DataKit 也会将每一条数据追加上名为 `host` 的标签，其值为 DataKit 所在的主机名。这么做的原因是便于建立异类数据之间的关联（如关联容器数据和容器所在主机的数据）。如果要禁用这一行为，[参见这里](datakit-how-to#1aab7c29)

另外注意的是，这里的标签值必须用双引号包围，否则会导致主配置解析失败。

#### 日志配置修改

DataKit 默认日志等级为 `info`。编辑 `conf.d/datakit.conf`，可修改日志等级：

```toml
[logging]
	level = "debug" # 将 info 改成 debug
```

置为 `debug` 后，即可看到更多日志（目前只支持 `debug/info` 两个级别）。

DataKit 默认会对日志进行分片，默认分片大小（`rotate`）为 32MB，总共 6 个分片（1 个当前写入分片加上 5 个切割分片，分片个数尚不支持配置）。如果嫌弃 DataKit 日志占用太多磁盘空间（最多 32 x 6 = 192MB），可减少 `rotate` 大小（比如改成 4，单位为 MB）。

gin.log 也会按照同样的方式自动切割。

### 采集器配置文件

各个采集器的配置文件均存放在 `conf.d` 目录下，且分门别类，存放在各个子分类中，如 `conf.d/host` 目录下存放着各种主机相关的采集器配置示例，以 Linux 为例：

```
├── cpu.conf.sample
├── disk.conf.sample
├── diskio.conf.sample
├── host_processes.conf.sample
├── hostobject.conf.sample
├── kernel.conf.sample
├── mem.conf.sample
├── net.conf.sample
├── swap.conf.sample
└── system.conf.sample
```

同样数据库相关的配置示例，在 `conf.d/db` 目录下：

```
├── elasticsearch.conf.sample
├── mysql.conf.sample
├── oracle.conf.sample
├── postgresql.conf.sample
└── sqlserver.conf.sample
```

还有其它更多的分类，某些具体的采集器，无法分类，则单独成行，如 rabbitmq 就是一个单独的分类。如果不清楚具体某个采集器的示例文档位置，可参考该采集器的使用文档，其中必有类似如下描述：

> 进入 DataKit 安装目录下的 `conf.d/xxx` 目录，复制 `yyy.conf.sample` 并命名为 `yyy.conf`...

此处需注意的是，由于 DataKit 只会搜索 `conf.d/` 目录下的 `.conf` 文件，复制出来的 `yyy.conf`，必须放在 `conf.d` 目录下（不一定要在特定的 `conf.d/xxx` 目录中），且必须以 `.conf` 作为文件后缀，不然 DataKit 会忽略该配置文件的处理。

> Tips：如果要暂时移除掉某个采集配置，只需将其后缀改一下即可，如 `yyy.conf` 改成 `yyy.conf.bak`。

### 具体采集的开启

以 MySQL 采集器为例：

```toml
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项...
```

其中：

| 配置                  | 描述                                                               |
| ---------             | ---                                                                |
| `[[inputs.mysql]]`    | 这一行是必须的，它表明「这是一个 mysql 采集器」，便于 DataKit 识别 |
| `host/user/...`       | 这些属于基础配置项，连接 MySQL 必须要这些配置                      |
| `[inputs.mysql.log]`  | 采集 MySQL 日志配置入口                                            |
| `[inputs.mysql.tags]` | 对采集的 MySQL 数据追加额外的标签                                  |

几个注意点：

- 我们将 MySQL 的日志采集和指标采集放在一起，主要是便于大家使用，无需单独用额外的采集器配置来收集日志
- 在采集器的配置中，我们可以使用形如 `$XXXXX` 这样的环境变量（注意，DataKit 主配置中不支持这种）。例如，假定该 MySQL 运行在容器中，但其主机名实际上并不可提前预知，此时可以追加额外标签 `host = $HOSTNAME`。需注意的是，指定的环境变量必须真实有效，如果 DataKit 运行时获取不到该环境变量，那么会直接使用字符串 `no-value` 作为该字段的值。

#### 单个采集器如何开启多份采集

如果要配置多个不同 MySQL 采集，可单独再复制一份出来，如下 `mysql.conf` 所示：

```toml
# 第一个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项...

# 再来一个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项...

# 下面继续再加一个
[[inputs.mysql]]
	...
```

#### 全局 host 标签问题

因为 DataKit 会默认给采集到的所有数据追加标签 `host=<DataKit所在主机名>`，但某些情况这个默认追加的 `host` 会带来困扰。

以 MySQL 为例，如果 MySQL 不在 DataKit 所在机器，肯定希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名。此时可在 `[inputs.mysql.tags]` 中手动增加 `host = "<your-mysql-real-hostname>"`，以此来屏蔽 DataKit 默认追加的 `host` 标签。在 DataKit 看来，如果采集到的数据中就带有 `host` 标签，那么就不再追加 DataKit 所在主机的 host 信息了。

## DataKit 各种工具使用

DataKit 内置很多不同的小工具，便于大家日常使用。可通过如下命令来查看 DataKit 的命令行帮助：

```shell
datakit -h
```

>注意：因不同平台的差异，具体帮助内容会有差别。

### 查询 DQL

DataKit 支持以交互式方式执行 DQL 查询，在交互模式下，DataKit 自带语句补全功能：

```shell
datakit --dql      # 或者 datakit -Q
dql > cpu limit 1
-----------------[ 1.cpu ]-----------------
             cpu 'cpu-total'
            host 'tan-air.local'
            time 2021-06-23 10:06:03 +0800 CST
       usage_irq 0
      usage_idle 56.928839
      usage_nice 0
      usage_user 19.825218
     usage_guest 0
     usage_steal 0
     usage_total 43.071161
    usage_iowait 0
    usage_system 23.245943
   usage_softirq 0
usage_guest_nice 0
---------
1 rows, cost 13.55119ms
```

Tips：

- 输入 `echo_explain` 即可看到后端查询语句
- 为避免显示太多 `nil` 查询结果，可通过 `disable_nil/enable_nil` 来开关
- 支持查询语句模糊搜，如 `echo_explain` 只需要输入 `echo` 或 `exp` 即可弹出提示，**通过 `Tab` 即可选择下拉提示**
- DataKit 会自动保存前面多次运行的 DQL 查询历史（最大 5000 条），可通过上下方向键来选择

> 注：Windows 下，请在 Powershell 中执行 `datakit --dql` 或 `datakit -Q`

关于 DQL 执行，DataKit 支持如下一些额外选项：

- `--run-dql`：单次执行一条查询语句：`datakit --run-dql 'cpu limit 1'`
- `--json`：以 JSON 形式输出结果，但 JSON 模式下，不会输出一些统计信息，如返回行数、时间消耗等

```shell
datakit --run-dql 'O::HOST:(os, message)' --json
datakit -Q --json
```

- `--token`：通过指定不同的 Token 来查询其它工作空间的数据

```shell
datakit --run-dql 'O::HOST:(os, message)' --token tkn_<another-token>
datakit -Q --token tkn_<another-token>
```

- `--auto-json`：DQL 查询的结果，如果字段值是 JSON 字符串，则自动做 JSON 美化

```shell
datakit --run-dql 'O::HOST:(os, message)' --auto-json
-----------------[ r1.HOST.s1 ]-----------------
message ----- json -----  # JSON 开始处有明显标志，此处 message 为字段名
{
  "host": {
    "meta": {
      "host_name": "www",
  ....                    # 此处省略长文本
  "config": {
    "ip": "10.100.64.120",
    "enable_dca": false,
    "http_listen": "localhost:9529",
    "api_token": "tkn_f2b9920f05d84d6bb5b14d9d39db1dd3"
  }
}
----- end of json -----   # JSON 结束处有明显标志
     os 'darwin'
   time 2021-09-13 16:56:22 +0800 CST
---------
8 rows, 1 series, cost 4ms
```

> 注意：JSON 模式下，`--auto-json` 选项无效。

### 查看 DataKit 运行情况

在终端即可查看 DataKit 运行情况，其效果跟浏览器端 monitor 页面相似：

```shell
datakit --monitor     # 或者 datakit -M

# 同时可查看采集器开启情况：
datakit -M --vvv
```

> 注：Windows 下暂不支持在终端查看 monitor 数据，只能在浏览器端查看。

### 检查采集器配置是否正确

编辑完采集器的配置文件后，可能某些配置有误（如配置文件格式错误），通过如下命令可检查是否正确：

```shell
sudo datakit --check-config
------------------------
checked 13 conf, all passing, cost 22.27455ms
```

### 调试 grok 和 pipeline

指定 pipeline 脚本名称（`--pl`，pipeline 脚本必须放在 `<DataKit 安装目录>/pipeline` 目录下），输入一段文本（`--txt`）即可判断提取是否成功

```shell
datakit --pl your_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs): # 表示切割成功
{
	"code"   : "io/io.go: 458",       # 对应代码位置
	"level"  : "DEBUG",               # 对应日志等级
	"module" : "io",                  # 对应代码模块
	"msg"    : "post cost 6.87021ms", # 纯日志内容
	"time"   : 1610358231887000000    # 日志时间(Unix 纳秒时间戳)
}

# 提取失败示例
datakit --pl other_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
No data extracted from pipeline
```

由于 grok pattern 数量繁多，人工匹配较为麻烦。DataKit 提供了交互式的命令行工具 `grokq`（grok query）：

```Shell
datakit --grokq
grokq > Mon Jan 25 19:41:17 CST 2021   # 此处输入你希望匹配的文本
        2 %{DATESTAMP_OTHER: ?}        # 工具会给出对应对的建议，越靠前匹配月精确（权重也越大）。前面的数字表明权重。
        0 %{GREEDYDATA: ?}

grokq > 2021-01-25T18:37:22.016+0800
        4 %{TIMESTAMP_ISO8601: ?}      # 此处的 ? 表示你需要用一个字段来命名匹配到的文本
        0 %{NOTSPACE: ?}
        0 %{PROG: ?}
        0 %{SYSLOGPROG: ?}
        0 %{GREEDYDATA: ?}             # 像 GREEDYDATA 这种范围很广的 pattern，权重都较低
                                       # 权重越高，匹配的精确度越大

grokq > Q                              # Q 或 exit 退出
Bye!
```

> 注：Windows 下，请在 Powershell 中执行调试。

### 查看帮助文档

为便于大家在服务端查看 DataKit 帮助文档，DataKit 提供如下交互式文档查看入口（Windows 不支持）：

```shell
datakit --man
man > nginx
(显示 Nginx 采集文档)
man > mysql
(显示 MySQL 采集文档)
man > Q               # 输入 Q 或 exit 退出
```

### DataKit 服务管理

可直接使用如下命令直接管理 DataKit：

```shell
# Linux/Mac 可能需加上 sudo
datakit --stop
datakit --start
datakit --restart
```

#### 服务管理失败处理

有时候可能因为 DataKit 部分组件的 bug，导致服务操作失败（如 `--stop` 之后，服务并未停止），可按照如下方式来强制处理。

Linux 下，如果上述命令失效，可使用以下命令来替代：

```shell
sudo service datakit stop/start/restart
sudo systemctl stop/start/restart datakit
```

Mac 下，可以用如下命令代替：

```shell
# 启动 DataKit
sudo launchctl load -w /Library/LaunchDaemons/cn.dataflux.datakit.plist

# 停止 DataKit
sudo launchctl unload -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
```

#### 服务卸载以及重装

可直接使用如下命令直接卸载或恢复 DataKit 服务：

> 注意：此处卸载 DataKit 并不会删除 DataKit 相关文件。

```shell
# Linux/Mac shell
sudo datakit --uninstall
sudo datakit --reinstall

# Windows Powershell
datakit --uninstall
datakit --reinstall
```

### DataKit 更新 IP 数据库文件

可直接使用如下命令更新数据库文件（仅 Mac/Linux 支持）

```shell
sudo datakit --update-ip-db
```

若 DataKit 在运行中，更新成功后会自动更新 IP-DB 文件。

### DataKit 安装第三方软件

#### Telegraf 集成

> 注意：建议在使用 Telegraf 之前，先确 DataKit 是否能满足期望的数据采集。如果 DataKit 已经支持，不建议用 Telegraf 来采集，这可能会导致数据冲突，从而造成使用上的困扰。

安装 Telegraf 集成

```shell
sudo datakit --install telegraf
```

启动 Telegraf

```shell
cd /etc/telegraf
sudo cp telegraf.conf.sample telegraf.conf
sudo telegraf --config telegraf.conf
```

关于 Telegraf 的使用事项，参见[这里](telegraf)。

#### Security Checker 集成

安装 Security Checker

```shell
sudo datakit --install scheck
sudo datakit --install sec-checker  # 该命名即将废弃
```

安装成功后会自动运行，Security Checker 具体使用，参见[这里](https://www.yuque.com/dataflux/sec_checker/install) 

### DataKit 限制运行资源

通过 cgourp 限制 DataKit 运行资源（例如 CPU 使用率等），仅支持 Linux 操作系统。

进入 DataKit 安装目录下的 `conf.d` 目录，修改 `datakit.conf` 配置文件，将 `enable` 设置为 `true`，示例如下：

```
[cgroup]
  # 是否开启资源限制，默认关闭
  enable = true

  # 允许 CPU 最大使用率（百分制）
  cpu_max = 40.0

  # 允许 CPU 最使用率（百分制）
  cpu_min = 5.0
```

配置好后，重启 DataKit 即可。

#### CPU使用率说明

DataKit 会持续以当前 CPU 使用率为基准，动态调整自身能使用的 CPU 资源。假设现在 CPU 使用率较高，DataKit 可能会将自身限制在 `cpu_min` 值以下，反之 CPU 较为空闲时，可能会将限制调整到 `cpu_max`。

`cpu_max` 和 `cpu_min` 是正浮点数，且最大值不能超过 `100`。此值为主机 CPU 使用率，而非某个 CPU 核心使用率。

例如 `cpu_max` 为 `40.0`，8 核心 CPU 满负载使用率为 `800%`，则 DataKit 能使用的最大 CPU 资源是 `800% * 40% = 320%` 左右，是占全局 CPU 资源的 40%，而非单核心 CPU 的 40%。


### 其它命令

- 查看云属性数据

如果安装 DataKit 所在的机器是一台云服务器（目前支持 `aliyun/tencent/aws` 这几种），可通过如下命令查看部分云属性数据，如（标记为 `-` 表示该字段无效）：

```shell
datakit --show-cloud-info aws

           cloud_provider: aws
              description: -
     instance_charge_type: -
              instance_id: i-09b37dc1xxxxxxxxx
            instance_name: -
    instance_network_type: -
          instance_status: -
            instance_type: t2.nano
               private_ip: 172.31.22.123
                   region: cn-northwest-1
        security_group_id: launch-wizard-1
                  zone_id: cnnw1-az2
```
