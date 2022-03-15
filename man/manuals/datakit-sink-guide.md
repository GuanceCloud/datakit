# DataKit Sink 使用文档

Sink 是一个强大的存储写入模块，只需要几步简单配置，就能够支持用户写入不同的后端存储。

# 目前支持的 Sink 实例

- influxdb

# 通用参数说明

无论哪种 Sink 实例，都支持以下参数:

- `id`: 实例编号。如 `influxdb_1`。
- `target`: sink 实例目标，即要写入的存储是什么。如 `influxdb`。具体支持哪些见本文档中上面的 `目前支持的 Sink 实例` 节。
- `categories`: 汇报数据的类型。如 `["M", "N", "K", "O", "CO", "L", "T", "R", "S"]`。

<b>以上参数都是是必须参数</b>。

`categories` 中各字符串对应的上报指标集如下:

| `categories` 字符串 | 指标集 |
|  ----  | ----  |
| `M`  |  `Metric` |
| `N`  |  `Network` |
| `K`  |  `KeyEvent` |
| `O`  |  `Object` |
| `CO`  | `CustomObject` |
| `L`  |  `Logging` |
| `T`  |  `Tracing` |
| `R`  |  `RUM` |
| `S`  |  `Security` |

# 如何使用

只需要以下简单三步:

- 第一步: 搭建后端存储。

搭建好你想要的后端存储。我们这里称之为 `sink 实例`；

- 第二步: 增加配置。

在 `datakit.conf` 配置中增加上述 sink 实例的参数；

- 第三步: 重启 datakit。

`$ sudo datakit --restart`

>后端存储支持配置多个相同实例（比方说，2 个 influxdb，1 个生产，1 个备份），只需要将实例编号即 `id` 配置成不同的就行。<b>实例编号不可重复</b>。

# 各实例使用教程

## influxdb sink 使用教程

### 第一步: 搭建后端存储

自己搭建一个 influxdb 的 server 环境。

### 第二步: 增加配置

在 `datakit.conf` 中增加以下片段:

```conf
...
[sinks]

  [[sinks.sink]]
    id = "influxdb_1"
    target = "influxdb"
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S"]
    addr = "http://172.16.239.130:8086"
    database = "db0"
    timeout = "10s"
...
```

influxdb 的 sink 实例目前支持以下参数:

- `id`(必须): 实例编号，<b>唯一</b>。
- `addr`(必须): HTTP addr should be of the form `http://host:port` or `http://[ipv6-host%zone]:port`. UDP addr should be of the form `udp://host:port` or `udp://[ipv6-host%zone]:port`.
- `database`(必须): Database is the database to write points to.
- `precision`: Precision is the write precision of the points, defaults to "ns".
- `username`: Username is the influxdb username, optional.
- `password`: Password is the influxdb password, optional.
- `timeout`: Timeout for influxdb writes, defaults to no timeout.
- `user_agent`: UserAgent is the http User Agent, defaults to "InfluxDBClient".
- `retention_policy`: RetentionPolicy is the retention policy of the points.
- `write_consistency`: Write consistency is the number of servers required to confirm write.
- `write_encoding`: WriteEncoding specifies the encoding of write request
- `payload_size`(UDP 协议专用): PayloadSize is the maximum size of a UDP client message, optional. Tune this based on your network. Defaults to 512.

### 第三步: 重启 datakit

`$ sudo datakit --restart`
