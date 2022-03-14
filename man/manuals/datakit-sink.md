# DataKit Sink 简介

Sink 是一个强大的存储写入模块，只需要几步简单配置，就能够支持用户写入不同的后端存储。

# 目前支持的 Sink 实例

- influxdb

## 如何自定义 Sink 实例（仅供会写 Go 语言代码的高级用户参考，一般用户跳过这节）

<b>此节仅供会写 Go 语言代码的 高级用户/coder/码农 参考，一般用户跳过这节。</b>

目前官方只实现了上述实例，如果想要其它的，可以自己写代码实现（用 Go 语言），非常简单，大致分为以下几步（为了让大家更能形象理解，我以 `influxdb` 举例）:

- 第一步: 克隆 [datakit 代码](https://github.com/DataFlux-cn/datakit)，在 `io/sink` 下面新建一个包，名字叫 `sinkinfluxdb`（建议都以 `sink` 开头），小写。

- 第二步: 在上面的包下新建一个源文件 `sink_influxdb.go`，新建一个常量 `creatorID`，不能与其它包里面的 `creatorID` 重名；实现 `ISink` 的 `interface`，具体是实现以下几个函数:

|  函数名   | 作用  |
|  ----  | ----  |
| `GetID() string`  | 返回实例编号 |
| `LoadConfig(mConf map[string]interface{}) error`  | 加载外部配置到内部 |
| `Write(pts []ISinkPoint) error`  | 写入数据 |

大致代码如下:

```golang
const creatorID = "influxdb"

type SinkInfluxDB struct {
  // 这里写连接、写入等操作内部需要用到的一些参数，比如保存连接用到的参数等。
  ...
}

func (s *SinkInfluxDB) GetID() string {
  // 返回实例编号
  ...
}

func (s *SinkInfluxDB) LoadConfig(mConf map[string]interface{}) error {
  // 加载外部配置到内部
  ...
}

func (s *SinkInfluxDB) Write(pts []sinkcommon.ISinkPoint) error {
  // 写入数据
  ...
}
```

> 大体上可以参照 `influxdb` 的代码实现，还是非常简单的。一切以简单为首要设计原则，写的复杂了你自己也不愿维护。欢迎大家向 github 社区提交代码，大家一起来维护。

一些开发常见问题见 [这里](https://www.yuque.com/dataflux/datakit/development)。

- 第三步: 在 `datakit.conf` 里面增加配置，`target` 写上自定义的实例名，即 `creatorID`，唯一。比如:

```conf
...
[sinks]

  [[sinks.sink]]
    id = "influxdb_1" # 实例编号
    target = "influxdb"
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S"]
    addr = "http://172.16.239.130:8086"
    database = "db0"
    timeout = "10s"
...
```

# 通用参数说明

无论哪种 Sink 实例，都支持以下参数:

- `id`: 实例编号。如 `influxdb_1`。
- `target`: sink 实例目标，即要写入的存储是什么。如 `influxdb`。具体支持哪些见本文档中上面的 `目前支持的 Sink 实例` 节。
- `categories`: 汇报数据的类型。如 `["M", "N", "K", "O", "CO", "L", "T", "R", "S"]`。

<b>以上参数都是是必须参数</b>。

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
