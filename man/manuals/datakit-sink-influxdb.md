
# InfluxDB Sink 使用教程

InfluxDB 仅支持写入 Metric 种类的数据。

## 第一步: 搭建后端存储

自己搭建一个 InfluxDB 的 server 环境。

## 第二步: 增加配置

在 `datakit.conf` 中增加以下片段:

```conf
...
[sinks]
  [[sinks.sink]]
    categories = ["M"]
    target = "influxdb"
    host = "localhost:8087"
    protocol = "http"
    database = "db1"
    precision = "ns"
    timeout = "15s"
...
```

除了 Sink 必须配置[通用参数](datakit-sink-guide)外, InfluxDB 的 Sink 实例目前支持以下参数:

- `host`(必须): HTTP/UDP host should be of the form `host:port` or `[ipv6-host%zone]:port`.
- `protocol`(必须): `http` or `udp`.
- `database`(必须): Database is the database to write points to.
- `precision`: Precision is the write precision of the points, defaults to "ns".
- `username`: Username is the influxdb username, optional.
- `password`: Password is the influxdb password, optional.
- `timeout`: Timeout for influxdb writes, defaults to 10 seconds.
- `user_agent`: UserAgent is the http User Agent, defaults to "InfluxDBClient".
- `retention_policy`: RetentionPolicy is the retention policy of the points.
- `write_consistency`: Write consistency is the number of servers required to confirm write.
- `write_encoding`: WriteEncoding specifies the encoding of write request
- `payload_size`(UDP 协议专用): PayloadSize is the maximum size of a UDP client message, optional. Tune this based on your network. Defaults to 512.

## 第三步: 重启 DataKit

`$ sudo datakit --restart`

## 附: 通过环境变量开启 InfluxDB Sink 实例

InfluxDB 支持安装时环境变量开启的方式。

```shell
DK_SINK_M="influxdb://localhost:8087?protocol=http&database=db1&precision=ns&timeout=15s" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/community/install.sh)"
```
