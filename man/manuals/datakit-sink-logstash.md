
# Logstash Sink 使用教程

Logstash 仅支持写入 Logging 种类的数据。

## 第一步: 搭建后端存储

自己搭建一个 Logstash 的环境, 需要开启 [HTTP 模块](https://www.elastic.co/cn/blog/introducing-logstash-input-http-plugin), 开启的方法也非常简单, 直接在 Logstash 的 pipeline 文件中配置即可。

### 新建 Logstash 的 pipeline 文件

新建一个 Logstash 的 pipeline 文件 `pipeline-http.conf`, 如下所示:

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

这个文件可以任意放, 我这里放在 `/opt/elastic/logstash` 下, 即全路径是 `/opt/elastic/logstash/pipeline-http.conf`。

### 配置 Logstash 使用上面的 pipeline 文件

有两种方式: 配置文件方式和命令行方式。选其一即可。

#### 配置文件方式

在配置文件 `logstash/config/logstash.yml` 中增加一行:

```yml
path.config: /opt/elastic/logstash/pipeline-http.conf
```

#### 命令行方式

在命令行中指定 pipeline 文件:

```shell
$ logstash/bin/logstash -f /opt/elastic/logstash/pipeline-http.conf
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

## 安装阶段指定 LogStash Sink 设置

LogStash 支持安装时环境变量开启的方式。

```shell
DK_SINK_L="logstash://localhost:8080?protocol=http&request_path=/index/type/id&timeout=5s&write_type=json" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```
