{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit 入门简介

本文档主要介绍 DataKit 安装完后，如何使用 DataKit 中的基本功能，包括如下几个方面：

- 安装介绍
- 采集器使用，包括如何开启、管理等
- 如何使用 DataKit 中的各种小工具

## DataKit 目录介绍

DataKit 目前支持 Linux/Windows/macOS 三种主流平台：


| 操作系统                            | 架构                | 安装路径                                                                                     |
| ---------                           | ---                 | ------                                                                                       |
| Linux 内核 2.6.23 或更高版本        | amd64/386/arm/arm64 | `/usr/local/datakit`                                                      |
| macOS 10.11 或更高版本              | amd64               | `/usr/local/datakit`                                                      |
| Windows 7, Server 2008R2 或更高版本 | amd64/386           | 64位：`C:\Program Files\datakit`<br />32位：`C:\Program Files(32)\datakit` |

安装完成年后，DataKit 目录列表大概如下：

```
├── [4.4K]  conf.d
├── [ 160]  data
├── [186M]  datakit
├── [ 192]  externals
├── [2.4K]  gin.log
├── [2.4M]  log
└── [1.2K]  pipeline
```

其中：

- `conf.d`：存放所有采集器的配置示例。DataKit 主配置文件 datakit.conf 位于目录下
- `data`：存放 DataKit 运行所需的数据文件，如 IP 地址数据库等
- `datakit`：DataKit 主程序，Windows 下为 `datakit.exe`
- `externals`：部分采集器没有集成在 DataKit 主程序中，就都在这里了
- `gin.log`：DataKit 可以接收外部的 HTTP 数据输入，这个日志文件相当于 access-log
- `log`：DataKit 运行日志
- `pipeline` 存放用于文本处理的脚本代码

## 采集器使用

DataKit 安装完成后，默认会开启一批采集器，这些采集器一般跟主机相关，列表如下：

| 采集器名称     | 说明                                           |
| ---------      | ---                                            |
| `cpu`          | 采集主机的 CPU 使用情况                        |
| `disk`         | 采集磁盘占用情况                               |
| `diskio`       | 采集主机的磁盘 IO 情况                         |
| `mem`          | 采集主机的内存使用情况                         |
| `swap`         | 采集 Swap 内存使用情况                         |
| `system`       | 采集主机操作系统负载                           |
| `net`          | 采集主机网络流量情况                           |
| `host_process` | 采集主机上常驻（存活 10min 以上）进程列表      |
| `hostobject`   | 采集主机基础信息（如操作系统信息、硬件信息等） |

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
- 对于额外追加标签，有一些特殊情况：因为 DataKit 会默认给采集的所有数据追加标签 `host=<DataKit所在主机名>`（除非采集的数据中，已经自带了这个 `host` 标签），但如果 MySQL 不在 DataKit 所在机器，肯定希望这个 `host` 标签是被采集的 MySQL 的真实主机名（或云数据库的其它标识字段），而非 DataKit 所在的主机名，此时可以在 `[inputs.mysql.tags]` 中手动增加 `host = "your-mysql-real-name"`，以此来屏蔽 DataKit 默认追加的 `host` 标签
- 我们可以在所有采集器的配置中，使用环境变量。假定该 MySQL 运行在容器中，但其主机名实际上并不可提前预知，此时可以追加额外标签 `host = $HOSTNAME`。需注意的是，指定的环境变量必须真实有效，如果 DataKit 运行时获取不到该环境变量，那么会直接使用字符串 `no-value` 作为该字段的值
- 如果要配置多个不同 MySQL 采集，可单独再复制一份出来，如下 `mysql.conf` 所示：

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

## DataKit 各种工具使用

DataKit 内置很多不同的小工具，便于大家日常使用。

### 调试 grok 和 pipeline

指定 pipeline 脚本名称（`--pl`，pipeline 脚本必须放在 `<DataKit 安装目录>/pipeline` 目录下），输入一段文本（`--txt`）即可判断提取是否成功

```shell
$ ./datakit --pl your_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.go:458  post cost 6.87021ms'
Extracted data(cost: 421.705µs): # 表示切割成功
{
	"code"   : "io/io.go: 458",       # 对应代码位置
	"level"  : "DEBUG",               # 对应日志等级
	"module" : "io",                  # 对应代码模块
	"msg"    : "post cost 6.87021ms", # 纯日志内容
	"time"   : 1610358231887000000    # 日志时间(Unix 纳秒时间戳)
}

# 提取失败示例
$ ./datakit --pl other_pipeline.p --txt '2021-01-11T17:43:51.887+0800  DEBUG io  io/io.g o:458  post cost 6.87021ms'
No data extracted from pipeline
```

由于 grok pattern 数量繁多，人工匹配较为麻烦。DataKit 提供了交互式的命令行工具 `grokq`（grok query）：

```Shell
$ ./datakit --grokq
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

### 查看帮助文档

为便于大家在服务端查看 DataKit 帮助文档，DataKit 提供如下交互式文档查看入口：

```shell
$ ./datakit --man
man > nginx
(显示 Nginx 采集文档)
man > mysql
(显示 MySQL 采集文档)
man > Q               # 输入 Q 或 exit 退出
```

### DataKit 服务管理

可直接使用如下命令直接管理 DataKit（仅 Mac/Linux 支持）

```shell
$ sudo datakit --stop
$ sudo datakit --start
$ sudo datakit --restart
$ sudo datakit --reload
```

### DataKit 更新 IP 数据库文件

可直接使用如下命令更新数据库文件（仅 Mac/Linux 支持）

```shell
$ sudo datakit --update-ip-db
```

若 DataKit 在运行中，更新成功后会自动执行 Reload 操作

### DataKit 开启外部访问配置

编辑 `conf.d/datakit.conf` 中的 `http_listen` 字段，将其改成 `0.0.0.0:9529` 或其它网卡、端口。这样就能从其它主机上请求 DataKit 接口了。

### DataKit 安装第三方软件

#### Telegraf 安装

安装
```shell
$ sudo datakit --install telegraf
```

启动
```shell
$ cd /etc/telegraf
$ sudo cp telegraf.conf.sample tg.conf
$ sudo telegraf --config tg.conf
```

若需要修改 Telegraf 配置，在 `tg.conf` 文件中修改后重启 Telegraf

#### sec-check 安装

安装
```shell
$ sudo datakit --install sec-check
```

安装成功后会自动运行，sec-check 具体使用，参见[这里](https://www.yuque.com/dataflux/sec_checker/install) 

### 其它命令

- 查看云属性数据

如果安装 DataKit 所在的机器是一台云服务器（目前支持 `aliyun/tencent/aws` 这几种），可通过如下命令查看部分云属性数据，如（标记为 `-` 表示该字段无效）：

```shell
$ datakit --show-cloud-info aws

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

