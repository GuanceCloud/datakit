# DataKit Sink 开发文档

## 如何自定义 Sink 实例

此节仅供有 Go 语言基础的 高级用户/coder/程序员 参考，一些常见开发问题见 [这里](https://www.yuque.com/dataflux/datakit/development)。

目前官方只实现了部分实例，如果想要其它的，可以自己写代码实现（用 Go 语言），非常简单，大致分为以下几步（为了让大家更能形象理解，我以 `influxdb` 举例）:

- 第一步: 克隆 [datakit 代码](https://github.com/DataFlux-cn/datakit)，在 `io/sink` 下面新建一个包，名字叫 `sinkinfluxdb`（建议都以 `sink` 开头），小写。

- 第二步: 在上面的包下新建一个源文件 `sink_influxdb.go`，新建一个常量 `creatorID`，不能与其它包里面的 `creatorID` 重名；实现 `ISink` 的 `interface`，具体是实现以下几个函数:

- `GetID() string`: 返回实例编号
- `LoadConfig(mConf map[string]interface{}) error`: 加载外部配置到内部
- `Write(pts []ISinkPoint) error`: 写入数据

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
  // 这里你可能要熟悉下 ISinkPoint 这个 interface，里面有两个方法 ToPoint 和 ToJSON 供使用。
  //   ToPoint 返回的是 influxdb 的 point;
  //   ToJSON 返回的是结构体，如果不想使用 influxdb 的东西可以使用这个。
  ...
}
```

> 大体上可以参照 `influxdb` 的代码实现，还是非常简单的。一切以简单为首要设计原则，写的复杂了你自己也不愿维护。欢迎大家向 github 社区提交代码，大家一起来维护。


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
