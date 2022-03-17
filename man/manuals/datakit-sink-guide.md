# DataKit Sink 使用文档

## 导言

本文将讲述什么是 DataKit 的 Sink 模块(以下简称 Sink 模块、Sink)、以及如何使用 Sink 模块。适合于想了解 Sink 功能和有意愿使用 Sink 的同学。

## 如何阅读本文

本文开篇介绍了 Sink 的定义、使用步骤, 紧接着是各个 Sink 实例的使用例子。

这里有两个新概念, 一个是 Sink, 一个是 Sink 实例, 会在后面的片段中分别讲到, 读者务必理解好这两个概念才能读懂本文。

本文尽量做到极致简洁、以实际应用为主。

以下是正文。难度: 2 星(5 星最难)。

## 什么是 Sink

Sink 是一个强大的存储写入模块。只需要几步简单配置, 就能够支持用户将 DataKit 采集到的数据写入到不同的后端存储。

### 什么情况下可以使用 Sink

在以前, DataKit 采集到的数据是往 [观测云](https://console.guance.com/) 汇报的。近来为了响应部分用户把数据存储在本地的诉求, 特地开发了 Sink 功能。

## 什么是 Sink 实例

Sink 实例即 Sink 模块实例化的一个对象。举两个例子:
- 我们将 DataKit 采集到的数据写入到自建的 influxdb 中, 那么 influxdb 就是一个 "Sink 实例";
- 我们将 DataKit 采集到的数据写入到自建的 elasticsearch 集群中, 那么 elasticsearch 就是一个 "Sink 实例"。

### 目前支持的 Sink 实例

- influxdb

由于时间有限, 目前仅支持以上的实例。读者如果有一定技术基础的话可以自己开发其它的 Sink 实例, 开发方法可以阅读 [这篇文章](datakit-sink-dev.md)。

## 如何使用

只需要以下简单三步:

- 第一步: 搭建后端存储。

搭建好你想要的后端存储。

- 第二步: 增加配置。

在 `datakit.conf` 配置中增加 sink 实例的相关参数;

>后端存储支持配置多个相同实例(比方说, 2 个 influxdb, 1 个生产, 1 个备份), 只需要将实例编号即 `id` 配置成不同的就行(比方说, `influxdb_1`, `influxdb_2`)。<b>实例编号不可重复</b>。

- 第三步: 重启 DataKit。

`$ sudo datakit --restart`

初看有点抽象, 没关系, 后面会以例子的形式讲述如何实践以上三步, 每个已实现的 Sink 实例都会被覆盖到。

## 关于配置的注意事项: 通用参数的说明

无论哪种 Sink 实例, 都必须支持以下参数:

- `id`: 实例编号。如 `influxdb_1`。
- `target`: sink 实例目标, 即要写入的存储是什么。如 `influxdb`。具体支持哪些见本文档中上面的 `目前支持的 Sink 实例` 节。
- `categories`: 汇报数据的类型。如 `["M", "N", "K", "O", "CO", "L", "T", "R", "S"]`。

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

## 各实例使用教程

"实践是检验真理的唯一标准"。看了这么多，还感觉到抽象? 那是时候来实践一把了。本节将讲述各个已实现 Sink 实例的使用方法，以举例子形式展开，旨在通俗易懂。

### influxdb sink 使用教程

#### 第一步: 搭建后端存储

自己搭建一个 influxdb 的 server 环境。

#### 第二步: 增加配置

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

- `id`(必须): 实例编号, <b>唯一</b>。
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

#### 第三步: 重启 DataKit

`$ sudo datakit --restart`
