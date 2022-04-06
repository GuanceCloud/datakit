
# Logstash Sink 使用教程

Logstash 仅支持写入 Logging 种类的数据。

## 第一步: 搭建后端存储

自己搭建一个 Logstash 的环境, 并开启 [HTTP 模块](https://www.elastic.co/cn/blog/introducing-logstash-input-http-plugin):

```conf
input {
    http {
        host => "0.0.0.0" # default: 0.0.0.0
        port => 8080 # default: 8080
    }
}

output {
    elasticsearch {
        hosts => [ "localhost:9200"] # 我这里配置的是往 elasticsearch 写入数据
    }
}
```

## 第二步: 增加配置

在 `datakit.conf` 中增加以下片段:

```conf
...
[sinks]

  [[sinks.sink]]
    categories = ["L"]
    host = "1.1.1.1:8080"
    protocol = "http"
    request_path = "/index/type/id"
    target = "logstash"
    timeout = "5s"
    write_type="json"
...
```

除了 Sink 必须配置[通用参数](datakit-sink-guide)外, Logstash 的 Sink 实例目前支持以下参数:

- `host`(必须): HTTP host should be of the form `host:port` or `[ipv6-host%zone]:port`.
- `protocol`(必须): `http` or `https`.
- `write_type`(必须): 写入的源数据的格式类型: `json` 或者 `plain`。
- `request_path`: 请求 URL 的路径.
- `timeout`: Timeout for influxdb writes, defaults to 10 seconds.

## 第三步: 重启 DataKit

`$ sudo datakit --restart`
